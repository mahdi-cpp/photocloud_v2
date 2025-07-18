package storage

import (
	"context"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type UserStorage struct {
	config            Config
	mu                sync.RWMutex // Protects all indexes and maps
	user              model.User
	assets            map[int]*model.PHAsset
	albumManager      *CollectionManager[*model.Album]
	tripManager       *CollectionManager[*model.Trip]
	personManager     *CollectionManager[*model.Person]
	pinnedManager     *CollectionManager[*model.Pinned]
	metadata          *MetadataManager
	thumbnail         *ThumbnailManager
	lastID            int
	lastRebuild       time.Time
	maintenanceCtx    context.Context
	cancelMaintenance context.CancelFunc
	statsMu           sync.Mutex
	stats             Stats
}

func (us *UserStorage) UploadAsset(userID int, file multipart.File, header *multipart.FileHeader) (*model.PHAsset, error) {

	// Check file size
	if header.Size > us.config.MaxUploadSize {
		return nil, ErrFileTooLarge
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Handler asset filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", 1, ext)
	assetPath := filepath.Join(us.config.AssetsDir, filename)

	// Save asset file
	if err := os.WriteFile(assetPath, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to save asset: %w", err)
	}

	// Initialize the ImageExtractor with the path to exiftool
	extractor := NewMetadataExtractor("/usr/local/bin/exiftool")

	// Extract metadata
	width, height, camera, err := extractor.ExtractMetadata(assetPath)
	if err != nil {
		log.Printf("Metadata extraction failed: %v", err)
	}
	mediaType := GetMediaType(ext)

	// Handler asset
	asset := &model.PHAsset{
		ID:           us.lastID,
		UserID:       userID,
		Filename:     filename,
		CreationDate: time.Now(),
		MediaType:    mediaType,
		PixelWidth:   width,
		PixelHeight:  height,
		CameraModel:  camera,
	}

	// Save metadata
	if err := us.metadata.SaveMetadata(asset); err != nil {
		// Clean up asset file if metadata save fails
		os.Remove(assetPath)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	// Add to indexes
	//us.addToIndexes(asset)

	// Update stats
	us.statsMu.Lock()
	us.stats.TotalAssets++
	us.stats.Uploads24h++
	us.statsMu.Unlock()

	return asset, nil
}

func (us *UserStorage) GetAsset(assetId int) (*model.PHAsset, bool) {

	//var selectAsset model.PHAsset
	//for _, asset := range us.assets {
	//	if asset.ID == assetId {
	//		selectAsset = asset
	//		break
	//	}
	//}

	asset, exists := us.assets[assetId]
	return asset, exists
}

func (us *UserStorage) GetAssetContent(id int) ([]byte, error) {
	// Get asset to resolve filename
	asset, exists := us.GetAsset(id)
	if !exists {
		return nil, fmt.Errorf("asset not found")
	}

	assetPath := filepath.Join(us.config.AssetsDir, asset.Filename)
	return os.ReadFile(assetPath)
}

func (us *UserStorage) UpdateAsset(assetIds []int, update model.AssetUpdate) (string, error) {

	us.mu.Lock()
	defer us.mu.Unlock()

	for _, id := range assetIds {

		// Load current asset
		//asset, err := us.metadata.LoadMetadata(id)
		//if err != nil {
		//	return "", err
		//}

		asset, exists := us.GetAsset(id)
		if !exists {
			continue
		}

		// Apply updates
		if update.Filename != nil {
			asset.Filename = *update.Filename
		}
		if update.CameraMake != nil {
			asset.CameraMake = *update.CameraMake
		}
		if update.CameraModel != nil {
			asset.CameraModel = *update.CameraModel
		}

		if update.IsFavorite != nil {
			asset.IsFavorite = *update.IsFavorite
		}
		if update.IsScreenshot != nil {
			asset.IsScreenshot = *update.IsScreenshot
		}
		if update.IsHidden != nil {
			asset.IsHidden = *update.IsHidden
		}

		// Handle album operations
		switch {
		case update.Albums != nil:
			// Full replacement
			asset.Albums = *update.Albums
		case len(update.AddAlbums) > 0 || len(update.RemoveAlbums) > 0:

			// Handler a set for efficient lookups
			albumSet := make(map[int]bool)
			for _, id := range asset.Albums {
				albumSet[id] = true
			}

			// Add new items (avoid duplicates)
			for _, id := range update.AddAlbums {
				if !albumSet[id] {
					asset.Albums = append(asset.Albums, id)
					albumSet[id] = true
				}
			}

			// Remove specified items
			if len(update.RemoveAlbums) > 0 {
				removeSet := make(map[int]bool)
				for _, id := range update.RemoveAlbums {
					removeSet[id] = true
				}

				newAlbums := make([]int, 0, len(asset.Albums))
				for _, id := range asset.Albums {
					if !removeSet[id] {
						newAlbums = append(newAlbums, id)
					}
				}
				asset.Albums = newAlbums
			}
		}

		// Handle trip operations
		switch {
		case update.Trips != nil:
			// Full replacement
			asset.Trips = *update.Trips
		case len(update.AddTrips) > 0 || len(update.RemoveTrips) > 0:

			// Handler a set for efficient lookups
			tripSet := make(map[int]bool)
			for _, id := range asset.Trips {
				tripSet[id] = true
			}

			// Add new Persons (avoid duplicates)
			for _, id := range update.AddTrips {
				if !tripSet[id] {
					asset.Trips = append(asset.Trips, id)
					tripSet[id] = true
				}
			}

			// Remove specified trips
			if len(update.RemoveTrips) > 0 {
				removeSet := make(map[int]bool)
				for _, id := range update.RemoveTrips {
					removeSet[id] = true
				}

				newTrips := make([]int, 0, len(asset.Trips))
				for _, id := range asset.Trips {
					if !removeSet[id] {
						newTrips = append(newTrips, id)
					}
				}
				asset.Trips = newTrips
			}
		}

		// Handle person operations
		switch {
		case update.Persons != nil:
			// Full replacement
			asset.Persons = *update.Persons
		case len(update.AddPersons) > 0 || len(update.RemovePersons) > 0:

			// Handler a set for efficient lookups
			personSet := make(map[int]bool)
			for _, id := range asset.Persons {
				personSet[id] = true
			}

			// Add new Persons (avoid duplicates)
			for _, id := range update.AddPersons {
				if !personSet[id] {
					asset.Persons = append(asset.Persons, id)
					personSet[id] = true
				}
			}

			// Remove specified Persons
			if len(update.RemovePersons) > 0 {
				removeSet := make(map[int]bool)
				for _, id := range update.RemovePersons {
					removeSet[id] = true
				}

				newPersons := make([]int, 0, len(asset.Persons))
				for _, id := range asset.Persons {
					if !removeSet[id] {
						newPersons = append(newPersons, id)
					}
				}
				asset.Persons = newPersons
			}
		}

		asset.ModificationDate = time.Now()

		// Save updated metadata
		if err := us.metadata.SaveMetadata(asset); err != nil {
			return "", err
		}

		//for _, asset := range us.assets {
		//	if asset.ID == asset.ID {
		//		us.assets
		//		break
		//	}
		//}

		// Update indexes
		//us.updateIndexesForAsset(asset)

		// Update memory
		//us.memory.Put(id, asset)
	}

	// Merging strings with the integer ID
	merged := fmt.Sprintf(" %s, %d:", "update assets count: ", len(assetIds))

	return merged, nil
}

func (us *UserStorage) PrepareAlbums() {
	us.prepareAlbums()
}

func (us *UserStorage) GetSystemStats() Stats {
	us.statsMu.Lock()
	defer us.statsMu.Unlock()
	return us.stats
}

func (us *UserStorage) FetchAssets(with model.PHFetchOptions) ([]*model.PHAsset, int, error) {

	us.mu.RLock()
	defer us.mu.RUnlock()

	startTime := time.Now()

	// Step 1: Build criteria from with
	criteria := assetBuildCriteria(with)

	// Step 2: Find all matching assets (store pointers to original assets)
	var matches []*model.PHAsset
	totalCount := 0

	for _, asset := range us.assets {
		if criteria(*asset) {
			matches = append(matches, asset)
			totalCount++
		}
	}

	//for i := range us.assets {
	//	if criteria(us.assets[i]) {
	//		matches = append(matches, &us.assets[i])
	//		totalCount++
	//	}
	//}

	// Apply sorting
	assetSortAssets(matches, with.SortBy, with.SortOrder)

	// Step 3: Apply pagination
	start := with.FetchOffset
	if start < 0 {
		start = 0
	}
	if start > len(matches) {
		start = len(matches)
	}

	end := start + with.FetchLimit
	if end > len(matches) || with.FetchLimit <= 0 {
		end = len(matches)
	}

	paginated := matches[start:end]

	//Log performance
	duration := time.Since(startTime)
	log.Printf("Search: scanned %d assets, found %d matches, returned %d (in %v)", len(us.assets), totalCount, len(paginated), duration)

	//fmt.Println("matches[start:end]: ", start, end)
	//fmt.Println("matches: ", with.FetchOffset)
	fmt.Println("paginated: ", len(paginated))

	return paginated, totalCount, nil
}

func (us *UserStorage) prepareAlbums() {

	items, err := us.albumManager.GetAll()
	if err != nil {
	}

	for _, album := range items {

		with := model.PHFetchOptions{
			UserID:     4,
			Albums:     []int{album.ID},
			SortBy:     "modificationDate",
			SortOrder:  "gg",
			FetchLimit: 6,
		}

		assets, count, err := us.FetchAssets(with)
		if err != nil {
			continue
		}
		album.Count = count
		us.albumManager.itemAssets[album.ID] = assets
	}
}

func (us *UserStorage) DeleteAsset(id int) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Get asset
	//asset, err := us.GetAsset(id)
	//if err != nil {
	//	return err
	//}

	// Delete asset file
	//assetPath := filepath.Join(us.config.AssetsDir, asset.Filename)
	//if err := os.Remove(assetPath); err != nil {
	//	return fmt.Errorf("failed to delete asset file: %w", err)
	//}

	// Delete metadata
	if err := us.metadata.DeleteMetadata(id); err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	// Delete thumbnail (if exists)
	//us.thumbnail.DeleteThumbnails(id)

	// Remove from indexes
	//us.removeFromIndexes(id)

	// Remove from memory
	//us.memory.Remove(id)

	// Update stats
	us.statsMu.Lock()
	us.stats.TotalAssets--
	us.statsMu.Unlock()

	return nil
}

func assetBuildCriteria(with model.PHFetchOptions) searchCriteria[model.PHAsset] {

	return func(asset model.PHAsset) bool {

		// Filter by UserID (if non-zero)
		//if with.UserID != 0 && asset.UserID != with.UserID {
		//	return false
		//}

		// Filter by Query (case-insensitive service in Filename/URL)
		if with.Query != "" {
			query := strings.ToLower(with.Query)
			filename := strings.ToLower(asset.Filename)
			url := strings.ToLower(asset.Url)
			if !strings.Contains(filename, query) && !strings.Contains(url, query) {
				return false
			}
		}

		//Filter by MediaType (if specified)
		if with.MediaType != "" && asset.MediaType != with.MediaType {
			return false
		}

		// Filter by CameraModel (exact match)
		if with.CameraMake != "" && asset.CameraMake != with.CameraMake {
			return false
		}
		if with.CameraModel != "" && asset.CameraModel != with.CameraModel {
			return false
		}

		// Filter by CreationDate range
		if with.StartDate != nil && asset.CreationDate.Before(*with.StartDate) {
			return false
		}
		if with.EndDate != nil && asset.CreationDate.After(*with.EndDate) {
			return false
		}

		// Filter by boolean flags (if specified)
		if with.IsFavorite != nil && asset.IsFavorite != *with.IsFavorite {
			return false
		}
		if with.IsScreenshot != nil && asset.IsScreenshot != *with.IsScreenshot {
			return false
		}
		if with.IsHidden != nil && asset.IsHidden != *with.IsHidden {
			return false
		}

		if with.HideScreenshot != nil && *with.HideScreenshot == false && asset.IsScreenshot == true {
			return false
		}

		// Filter by  int
		if with.PixelWidth != 0 && asset.PixelWidth != with.PixelWidth {
			return false
		}
		if with.PixelHeight != 0 && asset.PixelHeight != with.PixelHeight {
			return false
		}

		// Filter by landscape orientation
		if with.IsLandscape != nil {
			isLandscape := asset.PixelWidth > asset.PixelHeight
			if isLandscape != *with.IsLandscape {
				return false
			}
		}

		// Album filtering
		if len(with.Albums) > 0 {
			found := false
			for _, albumID := range with.Albums {
				for _, assetAlbumID := range asset.Albums {
					if assetAlbumID == albumID {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				return false
			}
		}

		// Location filtering
		//if len(asset.Location) == 2 {
		//
		//	// Near point + radius search
		//	if len(with.NearPoint) == 2 && with.WithinRadius > 0 {
		//		distance := indexer.haversineDistance(
		//			with.NearPoint[0], with.NearPoint[1],
		//			asset.Location[0], asset.Location[1],
		//		)
		//		if distance > with.WithinRadius {
		//			return false
		//		}
		//	}
		//
		//	// Bounding box search
		//	if len(with.BoundingBox) == 4 {
		//		if !indexer.isInBoundingBox(asset.Location, with.BoundingBox) {
		//			return false
		//		}
		//	}
		//}

		return true // Asset matches all active with
	}
}

type IndexedItemV2[T any] struct {
	Index int
	Value T
}

type assetSearchCriteria[T any] func(T) bool

func assetSearch[T any](slice []T, criteria assetSearchCriteria[T]) []IndexedItemV2[T] {
	var results []IndexedItemV2[T]

	for i, item := range slice {
		if criteria(item) {
			results = append(results, IndexedItemV2[T]{Index: i, Value: item})
		}
	}
	return results
}

func assetSortAssets(assets []*model.PHAsset, sortBy, sortOrder string) {

	if sortBy == "" {
		return // No sorting requested
	}

	sort.Slice(assets, func(i, j int) bool {
		a := assets[i]
		b := assets[j]

		switch sortBy {
		case "id":
			if sortOrder == "asc" {
				return a.ID < b.ID
			}
			return a.ID > b.ID
		case "creationDate":
			if sortOrder == "asc" {
				return a.CreationDate.Before(b.CreationDate)
			}
			return a.CreationDate.After(b.CreationDate)

		case "modificationDate":
			if sortOrder == "asc" {
				return a.ModificationDate.Before(b.ModificationDate)
			}
			return a.ModificationDate.After(b.ModificationDate)

		case "filename":
			if sortOrder == "asc" {
				return a.Filename < b.Filename
			}
			return a.Filename > b.Filename

		default:
			return false // No sorting for unknown fields
		}
	})
}

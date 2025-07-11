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
	assets            []model.PHAsset
	albumManager      *AlbumManager
	tripManager       *TripManager
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

	// Create asset filename
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

	// Create asset
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

func (us *UserStorage) GetAsset(assetId int) (*model.PHAsset, error) {

	var selectAsset model.PHAsset
	for _, asset := range us.assets {
		if asset.ID == assetId {
			selectAsset = asset
			break
		}
		fmt.Println(asset)
	}

	return &selectAsset, nil
}

func (us *UserStorage) GetAssetContent(id int) ([]byte, error) {
	// Get asset to resolve filename
	asset, err := us.GetAsset(id)
	if err != nil {
		return nil, err
	}

	assetPath := filepath.Join(us.config.AssetsDir, asset.Filename)
	return os.ReadFile(assetPath)
}

func (us *UserStorage) UpdateAsset(assetIds []int, update model.AssetUpdate) (string, error) {

	us.mu.Lock()
	defer us.mu.Unlock()

	for _, id := range assetIds {

		// Load current asset
		asset, err := us.metadata.LoadMetadata(id)
		if err != nil {
			return "", err
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

			// Create a set for efficient lookups
			albumSet := make(map[int]bool)
			for _, id := range asset.Albums {
				albumSet[id] = true
			}

			// Add new albums (avoid duplicates)
			for _, id := range update.AddAlbums {
				if !albumSet[id] {
					asset.Albums = append(asset.Albums, id)
					albumSet[id] = true
				}
			}

			// Remove specified albums
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

			// Create a set for efficient lookups
			tripSet := make(map[int]bool)
			for _, id := range asset.Trips {
				tripSet[id] = true
			}

			// Add new Trips (avoid duplicates)
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

			// Create a set for efficient lookups
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

		// Update indexes
		//us.updateIndexesForAsset(asset)

		// Update cache
		//us.cache.Put(id, asset)
	}

	// Merging strings with the integer ID
	merged := fmt.Sprintf(" %s, %d:", "update assets count: ", len(assetIds))

	return merged, nil
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

	// Remove from cache
	//us.cache.Remove(id)

	// Update stats
	us.statsMu.Lock()
	us.stats.TotalAssets--
	us.statsMu.Unlock()

	return nil
}

func (us *UserStorage) GetSystemStats() Stats {
	us.statsMu.Lock()
	defer us.statsMu.Unlock()
	return us.stats
}

func (us *UserStorage) FilterAssets(filters model.AssetSearchFilters) ([]*model.PHAsset, int, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	startTime := time.Now()

	// Step 1: Build criteria from filters
	criteria := assetBuildCriteria(filters)

	// Step 2: Find all matching assets (store pointers to original assets)
	var matches []*model.PHAsset
	totalCount := 0

	for i := range us.assets {
		if criteria(us.assets[i]) {
			matches = append(matches, &us.assets[i])
			totalCount++
		}
	}

	// Apply sorting
	assetSortAssets(matches, filters.SortBy, filters.SortOrder)

	// Step 3: Apply pagination
	start := filters.Offset
	if start < 0 {
		start = 0
	}
	if start > len(matches) {
		start = len(matches)
	}

	end := start + filters.Limit
	if end > len(matches) || filters.Limit <= 0 {
		end = len(matches)
	}

	paginated := matches[start:end]

	// Log performance
	duration := time.Since(startTime)
	log.Printf("Search: scanned %d assets, found %d matches, returned %d (in %v)", len(us.assets), totalCount, len(paginated), duration)

	return paginated, totalCount, nil
}

// ========================
// Internal Implementation
// ========================

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

func assetBuildCriteria(filters model.AssetSearchFilters) searchCriteria[model.PHAsset] {

	return func(asset model.PHAsset) bool {

		// Filter by UserID (if non-zero)
		//if filters.UserID != 0 && asset.UserID != filters.UserID {
		//	return false
		//}

		// Filter by Query (case-insensitive service in Filename/URL)
		if filters.Query != "" {
			query := strings.ToLower(filters.Query)
			filename := strings.ToLower(asset.Filename)
			url := strings.ToLower(asset.Url)
			if !strings.Contains(filename, query) && !strings.Contains(url, query) {
				return false
			}
		}

		//Filter by MediaType (if specified)
		if filters.MediaType != "" && asset.MediaType != filters.MediaType {
			return false
		}

		// Filter by CameraModel (exact match)
		if filters.CameraMake != "" && asset.CameraMake != filters.CameraMake {
			return false
		}

		if filters.CameraModel != "" && asset.CameraModel != filters.CameraModel {
			return false
		}

		// Filter by CreationDate range
		if filters.StartDate != nil && asset.CreationDate.Before(*filters.StartDate) {
			return false
		}
		if filters.EndDate != nil && asset.CreationDate.After(*filters.EndDate) {
			return false
		}

		// Filter by boolean flags (if specified)
		if filters.IsFavorite != nil && asset.IsFavorite != *filters.IsFavorite {
			return false
		}
		if filters.IsScreenshot != nil && asset.IsScreenshot != *filters.IsScreenshot {
			return false
		}
		if filters.IsHidden != nil && asset.IsHidden != *filters.IsHidden {
			return false
		}

		// Filter by  int
		if filters.PixelWidth != 0 && asset.PixelWidth != filters.PixelWidth {
			return false
		}
		if filters.PixelHeight != 0 && asset.PixelHeight != filters.PixelHeight {
			return false
		}

		// Filter by landscape orientation
		if filters.IsLandscape != nil {
			isLandscape := asset.PixelWidth > asset.PixelHeight
			if isLandscape != *filters.IsLandscape {
				return false
			}
		}

		// Album filtering
		if len(filters.Albums) > 0 {
			found := false
			for _, albumID := range filters.Albums {
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
		if len(asset.Location) == 2 {

			// Near point + radius search
			if len(filters.NearPoint) == 2 && filters.WithinRadius > 0 {
				distance := haversineDistance(
					filters.NearPoint[0], filters.NearPoint[1],
					asset.Location[0], asset.Location[1],
				)
				if distance > filters.WithinRadius {
					return false
				}
			}

			// Bounding box search
			if len(filters.BoundingBox) == 4 {
				if !isInBoundingBox(asset.Location, filters.BoundingBox) {
					return false
				}
			}
		}

		return true // Asset matches all active filters
	}
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

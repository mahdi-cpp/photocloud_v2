package storage

import (
	"context"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/pkg/collection"
	"github.com/mahdi-cpp/photocloud_v2/pkg/common_models"
	"github.com/mahdi-cpp/photocloud_v2/pkg/exif"
	"github.com/mahdi-cpp/photocloud_v2/pkg/image_loader"
	"github.com/mahdi-cpp/photocloud_v2/pkg/metadata"
	"github.com/mahdi-cpp/photocloud_v2/pkg/thumbnail"
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
	config              Config
	mu                  sync.RWMutex // Protects all indexes and maps
	user                common_models.User
	originalImageLoader *image_loader.ImageLoader
	tinyImageLoader     *image_loader.ImageLoader
	assets              map[int]*common_models.PHAsset
	cameras             map[string]*common_models.PHCollection[model.Camera]
	AlbumManager        *collection.Manager[*model.Album]
	TripManager         *collection.Manager[*model.Trip]
	PersonManager       *collection.Manager[*model.Person]
	PinnedManager       *collection.Manager[*model.Pinned]
	SharedAlbumManager  *collection.Manager[*model.SharedAlbum]
	VillageManager      *collection.Manager[*model.Village]
	metadata            *metadata.AssetMetadataManager
	thumbnail           *thumbnail.ThumbnailManager
	lastID              int
	lastRebuild         time.Time
	maintenanceCtx      context.Context
	cancelMaintenance   context.CancelFunc
	statsMu             sync.Mutex
	stats               Stats
}

func (userStorage *UserStorage) UploadAsset(userID int, file multipart.File, header *multipart.FileHeader) (*common_models.PHAsset, error) {

	// Check file size
	if header.Size > userStorage.config.MaxUploadSize {
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
	assetPath := filepath.Join(userStorage.config.AssetsDir, filename)

	// Save asset file
	if err := os.WriteFile(assetPath, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to save asset: %w", err)
	}

	// Initialize the ImageExtractor with the path to exiftool
	extractor := asset_create.NewMetadataExtractor("/usr/local/bin/exiftool")

	// Extract metadata
	width, height, camera, err := extractor.ExtractMetadata(assetPath)
	if err != nil {
		log.Printf("Metadata extraction failed: %v", err)
	}
	mediaType := asset_create.GetMediaType(ext)

	// Handler asset
	asset := &common_models.PHAsset{
		ID:           userStorage.lastID,
		UserID:       userID,
		Filename:     filename,
		CreationDate: time.Now(),
		MediaType:    mediaType,
		PixelWidth:   width,
		PixelHeight:  height,
		CameraModel:  camera,
	}

	// Save metadata
	if err := userStorage.metadata.SaveMetadata(asset); err != nil {
		// Clean up asset file if metadata save fails
		os.Remove(assetPath)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	// Add to indexes
	//userStorage.addToIndexes(asset)

	// Update stats
	userStorage.statsMu.Lock()
	userStorage.stats.TotalAssets++
	userStorage.stats.Uploads24h++
	userStorage.statsMu.Unlock()

	return asset, nil
}

func (userStorage *UserStorage) GetAsset(assetId int) (*common_models.PHAsset, bool) {
	asset, exists := userStorage.assets[assetId]
	return asset, exists
}

func (userStorage *UserStorage) GetAllAssets() map[int]*common_models.PHAsset {
	return userStorage.assets
}

func (userStorage *UserStorage) GetAssetContent(id int) ([]byte, error) {
	// Get asset to resolve filename
	asset, exists := userStorage.GetAsset(id)
	if !exists {
		return nil, fmt.Errorf("asset not found")
	}

	assetPath := filepath.Join(userStorage.config.AssetsDir, asset.Filename)
	return os.ReadFile(assetPath)
}

func (userStorage *UserStorage) UpdateAsset(update common_models.AssetUpdate) (string, error) {

	userStorage.mu.Lock()
	defer userStorage.mu.Unlock()

	for _, assetId := range update.AssetIds {

		asset, exists := userStorage.GetAsset(assetId)
		if !exists {
			continue
		}

		// Apply updates
		if update.Filename != nil {
			asset.Filename = *update.Filename
		}
		if update.MediaType != "" {
			asset.MediaType = update.MediaType
		}
		if update.CameraMake != nil {
			asset.CameraMake = *update.CameraMake
		}
		if update.CameraModel != nil {
			asset.CameraModel = *update.CameraModel
		}
		if update.IsCamera != nil {
			asset.IsCamera = *update.IsCamera
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
		if err := userStorage.metadata.SaveMetadata(asset); err != nil {
			return "", err
		}

		//for _, asset := range userStorage.assets {
		//	if asset.ID == asset.ID {
		//		userStorage.assets
		//		break
		//	}
		//}

		// Update indexes
		//userStorage.updateIndexesForAsset(asset)

		// Update memory
		//userStorage.memory.Put(assetId, asset)
	}

	// Merging strings with the integer ID
	merged := fmt.Sprintf(" %s, %d:", "update assets count: ", len(update.AssetIds))

	return merged, nil
}

func (userStorage *UserStorage) UpdateCollections() {
	userStorage.prepareAlbums()
	userStorage.prepareCameras()
	userStorage.prepareTrips()
	userStorage.preparePersons()
	userStorage.preparePinned()
}

func (userStorage *UserStorage) GetSystemStats() Stats {
	userStorage.statsMu.Lock()
	defer userStorage.statsMu.Unlock()
	return userStorage.stats
}

func (userStorage *UserStorage) FetchAssets(with common_models.PHFetchOptions) ([]*common_models.PHAsset, int, error) {

	userStorage.mu.RLock()
	defer userStorage.mu.RUnlock()

	//startTime := time.Now()

	// Step 1: Build criteria from with
	criteria := assetBuildCriteria(with)

	// Step 2: Find all matching assets (store pointers to original assets)
	var matches []*common_models.PHAsset
	totalCount := 0

	for _, asset := range userStorage.assets {
		if criteria(*asset) {
			matches = append(matches, asset)
			totalCount++
		}
	}

	//for i := range userStorage.assets {
	//	if criteria(userStorage.assets[i]) {
	//		matches = append(matches, &userStorage.assets[i])
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
	//duration := time.Since(startTime)
	//log.Printf("Search: scanned %d assets, found %d matches, returned %d (in %v)", len(userStorage.assets), totalCount, len(paginated), duration)

	//fmt.Println("matches[start:end]: ", start, end)
	//fmt.Println("matches: ", with.FetchOffset)
	//fmt.Println("paginated: ", len(paginated))

	return paginated, totalCount, nil
}

func (userStorage *UserStorage) prepareAlbums() {

	items, err := userStorage.AlbumManager.GetAll()
	if err != nil {
	}

	for _, album := range items {

		with := common_models.PHFetchOptions{
			UserID:     4,
			Albums:     []int{album.ID},
			SortBy:     "modificationDate",
			SortOrder:  "gg",
			FetchLimit: 6,
		}

		assets, count, err := userStorage.FetchAssets(with)
		if err != nil {
			continue
		}
		album.Count = count
		userStorage.AlbumManager.ItemAssets[album.ID] = assets
	}
}

func (userStorage *UserStorage) prepareTrips() {

	items, err := userStorage.TripManager.GetAll()
	if err != nil {
	}

	for _, item := range items {

		with := common_models.PHFetchOptions{
			UserID:     1,
			Trips:      []int{item.ID},
			SortBy:     "modificationDate",
			SortOrder:  "gg",
			FetchLimit: 2,
		}

		assets, count, err := userStorage.FetchAssets(with)
		if err != nil {
			continue
		}
		item.Count = count
		userStorage.TripManager.ItemAssets[item.ID] = assets
	}
}

func (userStorage *UserStorage) preparePersons() {

	items, err := userStorage.PersonManager.GetAll()
	if err != nil {
	}

	for _, item := range items {

		with := common_models.PHFetchOptions{
			UserID:     1,
			Persons:    []int{item.ID},
			SortBy:     "modificationDate",
			SortOrder:  "gg",
			FetchLimit: 1,
		}

		assets, count, err := userStorage.FetchAssets(with)
		if err != nil {
			continue
		}
		item.Count = count
		userStorage.PersonManager.ItemAssets[item.ID] = assets
	}
}

func (userStorage *UserStorage) prepareCameras() {

	//items, err := userStorage.CameraManager.GetAll()
	//if err != nil {
	//}

	if userStorage.cameras == nil {
		userStorage.cameras = map[string]*common_models.PHCollection[model.Camera]{}
	}

	for _, asset := range userStorage.assets {
		if asset.CameraModel == "" {
			continue
		}

		camera, exists := userStorage.cameras[asset.CameraModel]
		if exists {
			camera.Item.Count = camera.Item.Count + 1
			userStorage.cameras[asset.CameraModel] = camera
		} else {
			collection := &common_models.PHCollection[model.Camera]{
				Item: model.Camera{
					ID:          1,
					CameraMake:  asset.CameraMake,
					CameraModel: asset.CameraModel,
					Count:       1},
			}
			fmt.Println(collection)
			userStorage.cameras[asset.CameraModel] = collection
		}
	}

	//fmt.Println("camera count: ", len(userStorage.cameras))

	for _, collection := range userStorage.cameras {

		with := common_models.PHFetchOptions{
			UserID:      1,
			CameraMake:  collection.Item.CameraMake,
			CameraModel: collection.Item.CameraModel,
			SortBy:      "modificationDate",
			SortOrder:   "gg",
			FetchLimit:  6,
		}

		assets, _, err := userStorage.FetchAssets(with)
		if err != nil {
			continue
		}
		collection.Assets = assets
	}
}

func (userStorage *UserStorage) preparePinned() {

	items, err := userStorage.PinnedManager.GetAll()
	if err != nil {
		fmt.Println(err)
	}

	for _, item := range items {

		var with *common_models.PHFetchOptions

		switch item.Type {
		case "camera":
			with = &common_models.PHFetchOptions{
				IsCamera:   GetBoolPointer(true),
				SortBy:     "modificationDate",
				SortOrder:  "acs",
				FetchLimit: 1,
			}
			break
		case "screenshot":
			with = &common_models.PHFetchOptions{
				IsScreenshot: GetBoolPointer(true),
				SortBy:       "modificationDate",
				SortOrder:    "acs",
				FetchLimit:   1,
			}
			break
		case "favorite":
			with = &common_models.PHFetchOptions{
				IsFavorite: GetBoolPointer(true),
				SortBy:     "modificationDate",
				SortOrder:  "acs",
				FetchLimit: 1,
			}
			break
		case "video":
			with = &common_models.PHFetchOptions{
				MediaType:  "video",
				SortBy:     "modificationDate",
				SortOrder:  "acs",
				FetchLimit: 1,
			}
			break
		case "map":
			var assets []*common_models.PHAsset
			asset := common_models.PHAsset{ID: 12, MediaType: "image", Url: "map", Filename: "map"}
			assets = append(assets, &asset)
			userStorage.PinnedManager.ItemAssets[item.ID] = assets
			break
		case "album":
			album, err := userStorage.AlbumManager.Get(item.AlbumID)
			if err != nil {
				continue
			}
			item.Title = album.Title
			with = &common_models.PHFetchOptions{
				Albums:     []int{album.ID},
				SortBy:     "modificationDate",
				SortOrder:  "acs",
				FetchLimit: 1,
			}
			break
		}

		if with == nil || item.Type == "map" {
			continue
		}

		assets, count, err := userStorage.FetchAssets(*with)
		if err != nil {
			continue
		}
		item.Count = count
		userStorage.PinnedManager.ItemAssets[item.ID] = assets
	}
}

func (userStorage *UserStorage) GetAllCameras() map[string]*common_models.PHCollection[model.Camera] {
	return userStorage.cameras
}

func GetBoolPointer(b bool) *bool {
	return &b
}

func (userStorage *UserStorage) DeleteAsset(id int) error {
	userStorage.mu.Lock()
	defer userStorage.mu.Unlock()

	// Get asset
	//asset, err := userStorage.GetAsset(id)
	//if err != nil {
	//	return err
	//}

	// Delete asset file
	//assetPath := filepath.Join(userStorage.config.AssetsDir, asset.Filename)
	//if err := os.Remove(assetPath); err != nil {
	//	return fmt.Errorf("failed to delete asset file: %w", err)
	//}

	// Delete metadata
	if err := userStorage.metadata.DeleteMetadata(id); err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	// Delete thumbnail (if exists)
	//userStorage.thumbnail.DeleteThumbnails(id)

	// Remove from indexes
	//userStorage.removeFromIndexes(id)

	// Remove from memory
	//userStorage.memory.Remove(id)

	// Update stats
	userStorage.statsMu.Lock()
	userStorage.stats.TotalAssets--
	userStorage.statsMu.Unlock()

	return nil
}

func assetBuildCriteria(with common_models.PHFetchOptions) assetSearchCriteria[common_models.PHAsset] {

	return func(asset common_models.PHAsset) bool {

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
		if with.IsCamera != nil && *with.IsCamera != asset.IsCamera {
			return false
		}
		if with.IsFavorite != nil && *with.IsFavorite != asset.IsFavorite {
			return false
		}
		if with.IsScreenshot != nil && *with.IsScreenshot != asset.IsScreenshot {
			return false
		}
		if with.IsHidden != nil && *with.IsHidden != asset.IsHidden {
			return false
		}

		if with.NotInOneAlbum != nil {
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

		// Trip filtering
		if len(with.Trips) > 0 {
			found := false
			for _, tripID := range with.Trips {
				for _, assetTripID := range asset.Trips {
					if assetTripID == tripID {
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

		// Person filtering
		if len(with.Persons) > 0 {
			found := false
			for _, personID := range with.Persons {
				for _, assetPersonID := range asset.Persons {
					if assetPersonID == personID {
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

func assetSortAssets(assets []*common_models.PHAsset, sortBy, sortOrder string) {

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

		case "capturedDate":
			if sortOrder == "asc" {
				return a.CapturedDate.Before(b.CapturedDate)
			}
			return a.CapturedDate.After(b.CapturedDate)

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

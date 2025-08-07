package storage_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/pkg/exif"
	"github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"
	"github.com/mahdi-cpp/photocloud_v2/pkg/metadata"
	"github.com/mahdi-cpp/photocloud_v2/pkg/registery"
	"github.com/mahdi-cpp/photocloud_v2/pkg/thumbnail"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	earthRadius = 6371 // Earth's radius in km
)

var mahdiAssets []happle_models.PHAsset

// PhotoStorage implements the core storage functionality
type PhotoStorage struct {
	config Config
	mu     sync.RWMutex // Protects all indexes and maps
	cache  *LRUCache

	metadata  *metadata.AssetMetadataManager
	update    *UpdateManager
	thumbnail *thumbnail.ThumbnailManager

	username string
	//userService *UserManager

	albumRegistry *registery.Registry[model.Album]

	// Indexes
	assetIndex      map[int]string   // assetID -> filename
	userIndex       map[int][]int    // userID -> []assetID
	dateIndex       map[string][]int // "YYYY-MM-DD" -> []assetID
	textIndex       map[string][]int // word -> []assetID
	hiddenIndex     map[int]bool     // assetID -> isHidden
	favoriteIndex   map[int]bool     // assetID -> isFavorite
	screenshotIndex map[int]bool     // assetID -> isScreenshot
	mediaTypeIndex  map[string][]int // mediaType -> []assetID
	cameraIndex     map[string][]int // cameraModel -> []assetID

	// Indexers
	indexers map[string]Indexer

	lastID            int
	indexDirty        bool
	lastRebuild       time.Time
	maintenanceCtx    context.Context
	cancelMaintenance context.CancelFunc

	// Stats
	statsMu sync.Mutex
	stats   Stats
}

// NewPhotoStorage creates a new storage instance
func NewPhotoStorage(cfg Config) (*PhotoStorage, error) {

	// Handler context for background workers
	ctx, cancel := context.WithCancel(context.Background())

	ps := &PhotoStorage{
		config: cfg,
		cache:  NewLRUCache(cfg.CacheSize),

		username: "Mahdi_Abdolmaleki",
		//userService: NewUserManager(cfg.AppDir),

		metadata:      metadata.NewMetadataManager(cfg.MetadataDir),
		update:        NewUpdateManager(cfg.MetadataDir),
		albumRegistry: registery.NewRegistry[model.Album](),

		thumbnail:         thumbnail.NewThumbnailManager(cfg.ThumbnailsDir),
		assetIndex:        make(map[int]string),
		userIndex:         make(map[int][]int),
		dateIndex:         make(map[string][]int),
		textIndex:         make(map[string][]int),
		favoriteIndex:     make(map[int]bool),
		hiddenIndex:       make(map[int]bool),
		screenshotIndex:   make(map[int]bool),
		mediaTypeIndex:    make(map[string][]int),
		cameraIndex:       make(map[string][]int),
		maintenanceCtx:    ctx,
		cancelMaintenance: cancel,

		indexers: map[string]Indexer{
			"text": NewTextIndexer(),
			"date": NewDateIndexer(),
			//"mediaType": NewMediaTypeIndexer(),
			//"camera":    NewCameraIndexer(),
			"favorite":   NewBoolIndexer("IsFavorite"),
			"hidden":     NewBoolIndexer("IsHidden"),
			"screenshot": NewBoolIndexer("IsScreenshot"),
			"landscape":  NewBoolIndexer("IsLandscape"),
		},
	}

	// Ensure directories exist
	dirs := []string{cfg.AssetsDir, cfg.MetadataDir, cfg.ThumbnailsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Load or rebuild index
	if err := ps.loadOrRebuildIndex(); err != nil {
		return nil, fmt.Errorf("failed to initialize index: %w", err)
	}

	//assets2, err := ps.metadata.LoadUserAllMetadata()
	//if err != nil {
	//}
	//mahdiAssets = assets2

	// Start background maintenance
	go ps.periodicMaintenance()

	return ps, nil
}

// loadOrRebuildIndex initializes the storage index
func (ps *PhotoStorage) loadOrRebuildIndex() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Try to load existing index
	if _, err := os.Stat(ps.config.IndexFile); err == nil {
		if err := ps.loadIndex(); err == nil {
			return nil
		}
		log.Printf("Index load failed: %v, rebuilding...", err)
	}

	// Rebuild index from metadata
	return ps.rebuildIndex()
}

// UploadAsset handles file uploads
func (ps *PhotoStorage) UploadAsset(userID int, file multipart.File, header *multipart.FileHeader) (*happle_models.PHAsset, error) {
	// Check file size
	if header.Size > ps.config.MaxUploadSize {
		return nil, ErrFileTooLarge
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Handler asset filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", ps.nextID(), ext)
	assetPath := filepath.Join(ps.config.AssetsDir, filename)

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
	asset := &happle_models.PHAsset{
		ID:           ps.lastID,
		UserID:       userID,
		Filename:     filename,
		CreationDate: time.Now(),
		MediaType:    mediaType,
		PixelWidth:   width,
		PixelHeight:  height,
		CameraModel:  camera,
	}

	// Save metadata
	if err := ps.metadata.SaveMetadata(asset); err != nil {
		// Clean up asset file if metadata save fails
		os.Remove(assetPath)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	// Add to indexes
	ps.addToIndexes(asset)

	// Update stats
	ps.statsMu.Lock()
	ps.stats.TotalAssets++
	ps.stats.Uploads24h++
	ps.statsMu.Unlock()

	return asset, nil
}

// GetAsset retrieves an asset by ID
func (ps *PhotoStorage) GetAsset(id int) (*happle_models.PHAsset, error) {
	// Check memory first
	if asset, found := ps.cache.Get(id); found {
		return asset, nil
	}

	// Load from metadata
	asset, err := ps.metadata.LoadMetadata(id)
	if err != nil {
		return nil, err
	}

	// Add to memory
	ps.cache.Put(id, asset)

	return asset, nil
}

// GetAssetContent returns the binary content of an asset
func (ps *PhotoStorage) GetAssetContent(id int) ([]byte, error) {
	// Get asset to resolve filename
	asset, err := ps.GetAsset(id)
	if err != nil {
		return nil, err
	}

	assetPath := filepath.Join(ps.config.AssetsDir, asset.Filename)
	return os.ReadFile(assetPath)
}

// UpdateAsset updates asset metadata
func (ps *PhotoStorage) UpdateAsset(assetIds []int, update happle_models.AssetUpdate) (string, error) {

	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, id := range assetIds {

		// Load current asset
		asset, err := ps.metadata.LoadMetadata(id)
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
		if err := ps.metadata.SaveMetadata(asset); err != nil {
			return "", err
		}

		// Update indexes
		ps.updateIndexesForAsset(asset)

		// Update memory
		ps.cache.Put(id, asset)
	}

	// Merging strings with the integer ID
	merged := fmt.Sprintf(" %s, %d:", "update assets count: ", len(assetIds))

	return merged, nil
}

// Handler updates asset metadata
//func (ps *PhotoStorage) Handler(ctx context.Context, createAlbum model.Album) (string, error) {
//
//}

// DeleteAsset removes an asset and its metadata
func (ps *PhotoStorage) DeleteAsset(id int) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Get asset
	asset, err := ps.GetAsset(id)
	if err != nil {
		return err
	}

	// Delete asset file
	assetPath := filepath.Join(ps.config.AssetsDir, asset.Filename)
	if err := os.Remove(assetPath); err != nil {
		return fmt.Errorf("failed to delete asset file: %w", err)
	}

	// Delete metadata
	if err := ps.metadata.DeleteMetadata(id); err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	// Delete thumbnail (if exists)
	ps.thumbnail.DeleteThumbnails(id)

	// Remove from indexes
	ps.removeFromIndexes(id)

	// Remove from memory
	ps.cache.Remove(id)

	// Update stats
	ps.statsMu.Lock()
	ps.stats.TotalAssets--
	ps.statsMu.Unlock()

	return nil
}

// GetSystemStats returns storage statistics
func (ps *PhotoStorage) GetSystemStats() Stats {
	ps.statsMu.Lock()
	defer ps.statsMu.Unlock()
	return ps.stats
}

// GetIndexStatus returns index health information
func (ps *PhotoStorage) GetIndexStatus() IndexStatus {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return IndexStatus{
		LastRebuild:   ps.lastRebuild,
		AssetCount:    len(ps.assetIndex),
		TextIndexSize: len(ps.textIndex),
		DateIndexSize: len(ps.dateIndex),
		Dirty:         ps.indexDirty,
	}
}

// RebuildIndex rebuilds the index from metadata
func (ps *PhotoStorage) RebuildIndex() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.rebuildIndex()
}

// SearchAssets searches assets based on criteria
func (ps *PhotoStorage) SearchAssets(filters happle_models.PHFetchOptions) ([]*happle_models.PHAsset, int, error) {

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Start with all assets for user
	//results := ps.userIndex[filters.UserID]
	results := ps.userIndex[3327]
	total := len(results)

	// Start with all assets for user
	//results := ps.indexers["user"].Search(filters.UserID)
	//total := len(results)

	// Apply filters
	if filters.Query != "" {
		results = ps.indexers["text"].Filter(results, filters.Query)
	}
	if filters.MediaType != "" {
		results = ps.indexers["mediaType"].Filter(results, string(filters.MediaType))
	}
	if filters.IsFavorite != nil {
		results = ps.indexers["IsFavorite"].Filter(results, *filters.IsFavorite)
	}
	if filters.IsLandscape != nil {
		results = ps.indexers["landscape"].Filter(results, *filters.IsLandscape)
	}

	if filters.StartDate != nil || filters.EndDate != nil {
		dateRange := []time.Time{}
		if filters.StartDate != nil {
			dateRange = append(dateRange, *filters.StartDate)
		}
		if filters.EndDate != nil {
			dateRange = append(dateRange, *filters.EndDate)
		}
		results = ps.indexers["date"].Filter(results, dateRange)
	}

	// Apply filters
	if filters.Query != "" {
		results = ps.filterByText(results, filters.Query)
	}
	if filters.MediaType != "" {
		results = ps.filterByMediaType(results, string(filters.MediaType))
	}
	if filters.IsFavorite != nil {
		results = ps.filterByFavorite(results, *filters.IsFavorite)
	}
	if filters.IsScreenshot != nil {
		results = ps.filterByScreenshot(results, *filters.IsScreenshot)
	}
	if filters.StartDate != nil || filters.EndDate != nil {
		results = ps.filterByDateRange(results, filters.StartDate, filters.EndDate)
	}

	// Convert IDs to assets
	assets := make([]*happle_models.PHAsset, 0, len(results))
	for _, id := range results {
		asset, err := ps.GetAsset(id)
		if err != nil {
			continue // Skip assets that can't be loaded
		}
		assets = append(assets, asset)
	}

	// Apply pagination
	start := filters.FetchOffset
	if start > len(assets) {
		start = len(assets)
	}
	end := start + filters.FetchLimit
	if end > len(assets) {
		end = len(assets)
	}

	return assets[start:end], total, nil
}

// FilterAssets searches assets based on criteria
func (ps *PhotoStorage) FilterAssets(filters happle_models.PHFetchOptions) ([]*happle_models.PHAsset, int, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	startTime := time.Now()

	err := ps.update.updateCameraMake(7, filters.CameraMake)
	if err != nil {
	}

	// Step 1: Build criteria from filters
	criteria := buildCriteria(filters)

	// Step 2: Find all matching assets (store pointers to original assets)
	var matches []*happle_models.PHAsset
	totalCount := 0

	for i := range mahdiAssets {
		if criteria(mahdiAssets[i]) {
			matches = append(matches, &mahdiAssets[i])
			totalCount++
		}
	}

	// Apply sorting
	sortAssets(matches, filters.SortBy, filters.SortOrder)

	// Step 3: Apply pagination
	start := filters.FetchOffset
	if start < 0 {
		start = 0
	}
	if start > len(matches) {
		start = len(matches)
	}

	end := start + filters.FetchLimit
	if end > len(matches) || filters.FetchLimit <= 0 {
		end = len(matches)
	}

	paginated := matches[start:end]

	// Log performance
	duration := time.Since(startTime)
	log.Printf("Search: scanned %d assets, found %d matches, returned %d (in %v)", len(mahdiAssets), totalCount, len(paginated), duration)

	return paginated, totalCount, nil
}

// ========================
// Internal Implementation
// ========================

// IndexedItem represents an item with its index
type IndexedItem[T any] struct {
	Index int
	Value T
}

// searchCriteria defines a function type for service conditions
type searchCriteria[T any] func(T) bool

// search performs a generic service on a slice and returns matched indices
func search[T any](slice []T, criteria searchCriteria[T]) []IndexedItem[T] {
	var results []IndexedItem[T]

	for i, item := range slice {
		if criteria(item) {
			results = append(results, IndexedItem[T]{Index: i, Value: item})
		}
	}
	return results
}

func buildCriteria(filters happle_models.PHFetchOptions) searchCriteria[happle_models.PHAsset] {

	return func(asset happle_models.PHAsset) bool {

		// Filter by UserID (if non-zero)
		if filters.UserID != 0 && asset.UserID != filters.UserID {
			return false
		}

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

		// Album filtering (if items are specified)
		//if len(filters.AlbumCollection) > 0 {
		//	found := false
		//	for _, albumID := range filters.AlbumCollection {
		//		if contains(asset.AlbumCollection, albumID) {
		//			found = true
		//			break
		//		}
		//	}
		//	if !found {
		//		return false
		//	}
		//}

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
		if !asset.Place.IsEmpty() {

			// Near point + radius search
			if len(filters.NearPoint) == 2 && filters.WithinRadius > 0 {
				distance := haversineDistance(filters.NearPoint[0], filters.NearPoint[1], asset.Place.Latitude, asset.Place.Longitude)
				if distance > filters.WithinRadius {
					return false
				}
			}

			// Bounding box search
			//if len(filters.BoundingBox) == 4 {
			//	if !isInBoundingBox(asset.Place.Latitude, filters.BoundingBox) {
			//		return false
			//	}
			//}
		}

		return true // Asset matches all active filters
	}
}

func sortAssets(assets []*happle_models.PHAsset, sortBy, sortOrder string) {
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

//------------

// nextID generates the next asset ID
func (ps *PhotoStorage) nextID() int {
	ps.lastID++
	return ps.lastID
}

// addToIndexes adds an asset to all indexes
func (ps *PhotoStorage) addToIndexes(asset *happle_models.PHAsset) {

	ps.assetIndex[asset.ID] = asset.Filename
	ps.userIndex[asset.UserID] = append(ps.userIndex[asset.UserID], asset.ID)

	dateKey := asset.CreationDate.Format("2006-01-02")
	ps.dateIndex[dateKey] = append(ps.dateIndex[dateKey], asset.ID)

	words := strings.Fields(strings.ToLower(asset.Filename))
	for _, word := range words {
		if len(word) > 2 {
			ps.textIndex[word] = append(ps.textIndex[word], asset.ID)
		}
	}

	ps.favoriteIndex[asset.ID] = asset.IsFavorite
	ps.screenshotIndex[asset.ID] = asset.IsScreenshot
	ps.hiddenIndex[asset.ID] = asset.IsHidden
	ps.mediaTypeIndex[string(asset.MediaType)] = append(ps.mediaTypeIndex[string(asset.MediaType)], asset.ID)

	if asset.CameraModel != "" {
		ps.cameraIndex[asset.CameraModel] = append(ps.cameraIndex[asset.CameraModel], asset.ID)
	}

	for _, indexer := range ps.indexers {
		indexer.Add(asset)
	}

	ps.indexDirty = true
}

// removeFromIndexes removes an asset from all indexes
func (ps *PhotoStorage) removeFromIndexes(id int) {
	delete(ps.assetIndex, id)

	for userId, ids := range ps.userIndex {
		newIds := make([]int, 0, len(ids))
		for _, assetId := range ids {
			if assetId != id {
				newIds = append(newIds, assetId)
			}
		}
		ps.userIndex[userId] = newIds
	}

	for date, ids := range ps.dateIndex {
		newIds := make([]int, 0, len(ids))
		for _, assetId := range ids {
			if assetId != id {
				newIds = append(newIds, assetId)
			}
		}
		ps.dateIndex[date] = newIds
	}

	for word, ids := range ps.textIndex {
		newIds := make([]int, 0, len(ids))
		for _, assetId := range ids {
			if assetId != id {
				newIds = append(newIds, assetId)
			}
		}
		ps.textIndex[word] = newIds
	}

	delete(ps.favoriteIndex, id)
	delete(ps.hiddenIndex, id)

	for mediaType, ids := range ps.mediaTypeIndex {
		newIds := make([]int, 0, len(ids))
		for _, assetId := range ids {
			if assetId != id {
				newIds = append(newIds, assetId)
			}
		}
		ps.mediaTypeIndex[mediaType] = newIds
	}

	for camera, ids := range ps.cameraIndex {
		newIds := make([]int, 0, len(ids))
		for _, assetId := range ids {
			if assetId != id {
				newIds = append(newIds, assetId)
			}
		}
		ps.cameraIndex[camera] = newIds
	}

	for _, indexer := range ps.indexers {
		indexer.Remove(id)
	}

	ps.indexDirty = true
}

// updateIndexesForAsset updates indexes when an asset changes
func (ps *PhotoStorage) updateIndexesForAsset(asset *happle_models.PHAsset) {
	ps.removeFromIndexes(asset.ID)
	ps.addToIndexes(asset)
}

// loadIndex loads the index from disk
func (ps *PhotoStorage) loadIndex() error {
	data, err := os.ReadFile(ps.config.IndexFile)
	if err != nil {
		return err
	}

	var indexData struct {
		LastID          int
		AssetIndex      map[int]string
		UserIndex       map[int][]int
		DateIndex       map[string][]int
		TextIndex       map[string][]int
		FavoriteIndex   map[int]bool
		ScreenshotIndex map[int]bool
		HiddenIndex     map[int]bool
		MediaTypeIndex  map[string][]int
		CameraIndex     map[string][]int
	}

	if err := json.Unmarshal(data, &indexData); err != nil {
		return err
	}

	ps.lastID = indexData.LastID
	ps.assetIndex = indexData.AssetIndex
	ps.userIndex = indexData.UserIndex
	ps.dateIndex = indexData.DateIndex
	ps.textIndex = indexData.TextIndex
	ps.favoriteIndex = indexData.FavoriteIndex
	ps.screenshotIndex = indexData.ScreenshotIndex
	ps.hiddenIndex = indexData.HiddenIndex
	ps.mediaTypeIndex = indexData.MediaTypeIndex
	ps.cameraIndex = indexData.CameraIndex

	ps.lastRebuild = time.Now()
	return nil
}

type SerializableIndexer interface {
	Indexer
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

// saveIndex saves the index to disk
func (ps *PhotoStorage) saveIndex() error {

	// Serialize indexers first
	indexersData := make(map[string][]byte)
	for name, indexer := range ps.indexers {
		if serializable, ok := indexer.(SerializableIndexer); ok {
			data, err := serializable.Serialize()
			if err != nil {
				return fmt.Errorf("failed to serialize indexer %s: %w", name, err)
			}
			indexersData[name] = data
		}
	}

	indexData := struct {
		LastID          int
		AssetIndex      map[int]string
		UserIndex       map[int][]int
		DateIndex       map[string][]int
		TextIndex       map[string][]int
		FavoriteIndex   map[int]bool
		ScreenshotIndex map[int]bool
		HiddenIndex     map[int]bool
		MediaTypeIndex  map[string][]int
		CameraIndex     map[string][]int
		Indexers        map[string]string `json:"Indexers"` // Simplified for readability
	}{
		LastID:          ps.lastID,
		AssetIndex:      ps.assetIndex,
		UserIndex:       ps.userIndex,
		DateIndex:       ps.dateIndex,
		TextIndex:       ps.textIndex,
		FavoriteIndex:   ps.favoriteIndex,
		ScreenshotIndex: ps.screenshotIndex,
		HiddenIndex:     ps.hiddenIndex,
		MediaTypeIndex:  ps.mediaTypeIndex,
		CameraIndex:     ps.cameraIndex,
		Indexers:        ps.serializeIndexersForOutput(), // Add this function
	}

	// Use MarshalIndent for pretty-printed JSON
	data, err := json.MarshalIndent(indexData, "", "\t")
	if err != nil {
		return err
	}

	//data, err := json.Marshal(indexData)
	//if err != nil {
	//	return err
	//}

	tmpFile := ps.config.IndexFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, ps.config.IndexFile)
}

// Helper to simplify indexers for JSON output
func (ps *PhotoStorage) serializeIndexersForOutput() map[string]string {
	output := make(map[string]string)
	for name := range ps.indexers {
		// Just show the type for readability
		output[name] = fmt.Sprintf("%T", ps.indexers[name])
	}
	return output
}

// rebuildIndex reconstructs the index from metadata files
func (ps *PhotoStorage) rebuildIndex() error {

	// Clear existing indexes
	ps.assetIndex = make(map[int]string)
	ps.userIndex = make(map[int][]int)
	ps.dateIndex = make(map[string][]int)
	ps.textIndex = make(map[string][]int)
	ps.favoriteIndex = make(map[int]bool)
	ps.hiddenIndex = make(map[int]bool)
	ps.mediaTypeIndex = make(map[string][]int)
	ps.cameraIndex = make(map[string][]int)

	// Scan metadata directory
	files, err := os.ReadDir(ps.config.MetadataDir)
	if err != nil {
		return fmt.Errorf("failed to read metadata directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Extract ID from filename
		filename := file.Name()
		if !strings.HasSuffix(filename, ".json") {
			continue
		}

		idStr := strings.TrimSuffix(filename, ".json")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		// Load asset
		asset, err := ps.metadata.LoadMetadata(id)
		if err != nil {
			log.Printf("Skipping invalid metadata %s: %v", filename, err)
			continue
		}

		// Verify asset file exists
		assetPath := filepath.Join(ps.config.AssetsDir, asset.Filename)
		if _, err := os.Stat(assetPath); err != nil {
			log.Printf("Asset file missing for %d: %s", id, asset.ID)
			continue
		}

		// Add to indexes
		ps.addToIndexes(asset)

		// Update lastID
		if id > ps.lastID {
			ps.lastID = id
		}
	}

	// Save new index
	if err := ps.saveIndex(); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	ps.lastRebuild = time.Now()
	ps.indexDirty = false
	return nil
}

// periodicMaintenance runs background tasks
func (ps *PhotoStorage) periodicMaintenance() {
	saveTicker := time.NewTicker(5 * time.Minute)
	rebuildTicker := time.NewTicker(24 * time.Hour)
	statsTicker := time.NewTicker(30 * time.Minute)
	cleanupTicker := time.NewTicker(1 * time.Hour)

	for {
		select {
		case <-ps.maintenanceCtx.Done():
			return

		case <-saveTicker.C:
			if ps.indexDirty {
				ps.mu.Lock()
				if err := ps.saveIndex(); err != nil {
					log.Printf("Index save failed: %v", err)
				} else {
					log.Println("Index saved successfully")
					ps.indexDirty = false
				}
				ps.mu.Unlock()
			}

		case <-rebuildTicker.C:
			ps.mu.Lock()
			log.Println("Starting index rebuild...")
			if err := ps.rebuildIndex(); err != nil {
				log.Printf("Index rebuild failed: %v", err)
			} else {
				log.Println("Index rebuild completed")
			}
			ps.mu.Unlock()

		case <-statsTicker.C:
			// Reset daily upload count
			ps.statsMu.Lock()
			ps.stats.Uploads24h = 0
			ps.statsMu.Unlock()

		case <-cleanupTicker.C:
			ps.cleanupOrphanedAssets()
		}
	}
}

// cleanupOrphanedAssets removes mahdiAssets with missing files
func (ps *PhotoStorage) cleanupOrphanedAssets() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	removed := 0
	for id, filename := range ps.assetIndex {
		assetPath := filepath.Join(ps.config.AssetsDir, filename)
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			log.Printf("Removing orphaned asset %d (%s)", id, filename)
			ps.removeFromIndexes(id)
			ps.metadata.DeleteMetadata(id)
			ps.thumbnail.DeleteThumbnails(id)
			ps.cache.Remove(id)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("Removed %d orphaned assets", removed)
		ps.indexDirty = true
		ps.stats.TotalAssets -= removed
	}
}

// filterByText filters mahdiAssets by search query
func (ps *PhotoStorage) filterByText(assetIDs []int, query string) []int {
	query = strings.ToLower(query)
	words := strings.Fields(query)

	// Find matching IDs for each word
	idSets := make([]map[int]bool, len(words))
	for i, word := range words {
		ids := ps.textIndex[word]
		idSet := make(map[int]bool)
		for _, id := range ids {
			idSet[id] = true
		}
		idSets[i] = idSet
	}

	// Intersection of all word matches
	resultIDs := make(map[int]bool)
	for id := range idSets[0] {
		inAll := true
		for i := 1; i < len(idSets); i++ {
			if !idSets[i][id] {
				inAll = false
				break
			}
		}
		if inAll {
			resultIDs[id] = true
		}
	}

	// Filter original list
	filtered := make([]int, 0, len(assetIDs))
	for _, id := range assetIDs {
		if resultIDs[id] {
			filtered = append(filtered, id)
		}
	}

	return filtered
}

// filterByMediaType filters mahdiAssets by media type
func (ps *PhotoStorage) filterByMediaType(assetIDs []int, mediaType string) []int {
	// Get all assets of this type
	typeAssets := make(map[int]bool)
	for _, id := range ps.mediaTypeIndex[mediaType] {
		typeAssets[id] = true
	}

	// Filter original list
	filtered := make([]int, 0, len(assetIDs))
	for _, id := range assetIDs {
		if typeAssets[id] {
			filtered = append(filtered, id)
		}
	}

	return filtered
}

// filterByFavorite filters mahdiAssets by favorite status
func (ps *PhotoStorage) filterByFavorite(assetIDs []int, favorite bool) []int {
	filtered := make([]int, 0, len(assetIDs))
	for _, id := range assetIDs {
		if ps.favoriteIndex[id] == favorite {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

// filterByScreenshot filters mahdiAssets by favorite status
func (ps *PhotoStorage) filterByScreenshot(assetIDs []int, screenshot bool) []int {
	filtered := make([]int, 0, len(assetIDs))
	for _, id := range assetIDs {
		if ps.screenshotIndex[id] == screenshot {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

// filterByDateRange filters mahdiAssets by date range
func (ps *PhotoStorage) filterByDateRange(assetIDs []int, start, end *time.Time) []int {
	filtered := make([]int, 0, len(assetIDs))

	for _, id := range assetIDs {
		// Find date for asset
		var found bool
		for dateKey, ids := range ps.dateIndex {
			for _, assetID := range ids {
				if assetID == id {
					assetDate, _ := time.Parse("2006-01-02", dateKey)

					if start != nil && assetDate.Before(*start) {
						continue
					}
					if end != nil && assetDate.After(*end) {
						continue
					}

					filtered = append(filtered, id)
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	return filtered
}

// IndexStatus represents index health information
type IndexStatus struct {
	LastRebuild   time.Time
	AssetCount    int
	TextIndexSize int
	DateIndexSize int
	Dirty         bool
}

// Close stops background maintenance
func (ps *PhotoStorage) Close() {
	ps.cancelMaintenance()

	// Save index if dirty
	if ps.indexDirty {
		ps.saveIndex()
	}
}

// haversineDistance calculates distance between two points in km
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// isInBoundingBox checks if point is within rectangular bounds
func isInBoundingBox(point []float64, bbox []float64) bool {
	if len(point) < 2 || len(bbox) < 4 {
		return false
	}
	return point[0] >= bbox[0] && // minLat
		point[0] <= bbox[2] && // maxLat
		point[1] >= bbox[1] && // minLon
		point[1] <= bbox[3] // maxLon
}

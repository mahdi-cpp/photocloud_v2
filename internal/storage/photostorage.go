package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrAssetNotFound     = errors.New("asset not found")
	ErrThumbnailNotFound = errors.New("thumbnail not found")
	ErrFileTooLarge      = errors.New("file size exceeds limit")
	ErrInvalidUpdate     = errors.New("invalid asset update")
	ErrMetadataCorrupted = errors.New("metadata corrupted")
	ErrIndexCorrupted    = errors.New("index corrupted")
)

// PhotoStorage implements the core storage functionality
type PhotoStorage struct {
	config StorageConfig
	mu     sync.RWMutex // Protects all indexes and maps
	cache  *LRUCache
	//indexers  map[string]Indexer
	metadata  *MetadataManager
	thumbnail *ThumbnailManager

	// Indexes
	assetIndex      map[int]string   // assetID -> filename
	userIndex       map[int][]int    // userID -> []assetID
	dateIndex       map[string][]int // "YYYY-MM-DD" -> []assetID
	textIndex       map[string][]int // word -> []assetID
	favoriteIndex   map[int]bool     // assetID -> isFavorite
	hiddenIndex     map[int]bool     // assetID -> isHidden
	screenshotIndex map[int]bool     // assetID -> isScreenshot
	mediaTypeIndex  map[string][]int // mediaType -> []assetID
	cameraIndex     map[string][]int // cameraModel -> []assetID

	lastID            int
	indexDirty        bool
	lastRebuild       time.Time
	maintenanceCtx    context.Context
	cancelMaintenance context.CancelFunc

	// Stats
	statsMu sync.Mutex
	stats   StorageStats
}

// StorageConfig defines storage system configuration
type StorageConfig struct {
	AssetsDir     string
	MetadataDir   string
	ThumbnailsDir string
	IndexFile     string
	CacheSize     int
	MaxUploadSize int64
}

// StorageStats holds storage system statistics
type StorageStats struct {
	TotalAssets   int
	CacheHits     int64
	CacheMisses   int64
	Uploads24h    int
	ThumbnailsGen int
}

// NewPhotoStorage creates a new storage instance
func NewPhotoStorage(cfg StorageConfig) (*PhotoStorage, error) {
	// Create context for background workers
	ctx, cancel := context.WithCancel(context.Background())

	ps := &PhotoStorage{
		config:            cfg,
		cache:             NewLRUCache(cfg.CacheSize),
		metadata:          NewMetadataManager(cfg.MetadataDir),
		thumbnail:         NewThumbnailManager(cfg.ThumbnailsDir),
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
func (ps *PhotoStorage) UploadAsset(userID int, file multipart.File, header *multipart.FileHeader) (*model.PHAsset, error) {
	// Check file size
	if header.Size > ps.config.MaxUploadSize {
		return nil, ErrFileTooLarge
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create asset filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", ps.nextID(), ext)
	assetPath := filepath.Join(ps.config.AssetsDir, filename)

	// Save asset file
	if err := os.WriteFile(assetPath, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to save asset: %w", err)
	}

	// Initialize the MetadataExtractor with the path to exiftool
	extractor := NewMetadataExtractor("/usr/local/bin/exiftool")

	// Extract metadata
	width, height, camera, err := extractor.ExtractMetadata(assetPath)
	if err != nil {
		log.Printf("Metadata extraction failed: %v", err)
	}
	mediaType := GetMediaType(ext)

	// Create asset
	asset := &model.PHAsset{
		ID:           ps.lastID,
		UserID:       userID,
		Filename:     filename,
		CreationDate: time.Now(),
		MediaType:    mediaType,
		Width:        width,
		Height:       height,
		Camera:       camera,
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
func (ps *PhotoStorage) GetAsset(id int) (*model.PHAsset, error) {
	// Check cache first
	if asset, found := ps.cache.Get(id); found {
		return asset, nil
	}

	// Load from metadata
	asset, err := ps.metadata.LoadMetadata(id)
	if err != nil {
		return nil, err
	}

	// Add to cache
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
func (ps *PhotoStorage) UpdateAsset(id int, update model.AssetUpdate) (*model.PHAsset, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Load current asset
	asset, err := ps.metadata.LoadMetadata(id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if update.Filename != nil {
		asset.Filename = *update.Filename
	}
	if update.IsFavorite != nil {
		asset.IsFavorite = *update.IsFavorite
	}
	if update.IsHidden != nil {
		asset.IsHidden = *update.IsHidden
	}
	asset.ModificationDate = time.Now()

	// Save updated metadata
	if err := ps.metadata.SaveMetadata(asset); err != nil {
		return nil, err
	}

	// Update indexes
	ps.updateIndexesForAsset(asset)

	// Update cache
	ps.cache.Put(id, asset)

	return asset, nil
}

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

	// Remove from cache
	ps.cache.Remove(id)

	// Update stats
	ps.statsMu.Lock()
	ps.stats.TotalAssets--
	ps.statsMu.Unlock()

	return nil
}

// SearchAssets searches assets based on criteria
func (ps *PhotoStorage) SearchAssets(filters model.SearchFilters) ([]*model.PHAsset, int, error) {

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Start with all assets for user
	results := ps.userIndex[filters.UserID]
	total := len(results)

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
	if filters.StartDate != nil || filters.EndDate != nil {
		results = ps.filterByDateRange(results, filters.StartDate, filters.EndDate)
	}

	// Convert IDs to assets
	assets := make([]*model.PHAsset, 0, len(results))
	for _, id := range results {
		asset, err := ps.GetAsset(id)
		if err != nil {
			continue // Skip assets that can't be loaded
		}
		assets = append(assets, asset)
	}

	// Apply pagination
	start := filters.Offset
	if start > len(assets) {
		start = len(assets)
	}
	end := start + filters.Limit
	if end > len(assets) {
		end = len(assets)
	}

	return assets[start:end], total, nil
}

// GetSystemStats returns storage statistics
func (ps *PhotoStorage) GetSystemStats() StorageStats {
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

// ========================
// Internal Implementation
// ========================

// nextID generates the next asset ID
func (ps *PhotoStorage) nextID() int {
	ps.lastID++
	return ps.lastID
}

// addToIndexes adds an asset to all indexes
func (ps *PhotoStorage) addToIndexes(asset *model.PHAsset) {
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
	ps.hiddenIndex[asset.ID] = asset.IsHidden
	ps.mediaTypeIndex[string(asset.MediaType)] = append(ps.mediaTypeIndex[string(asset.MediaType)], asset.ID)

	if asset.Camera != "" {
		ps.cameraIndex[asset.Camera] = append(ps.cameraIndex[asset.Camera], asset.ID)
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

	ps.indexDirty = true
}

// updateIndexesForAsset updates indexes when an asset changes
func (ps *PhotoStorage) updateIndexesForAsset(asset *model.PHAsset) {
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
		LastID         int
		AssetIndex     map[int]string
		UserIndex      map[int][]int
		DateIndex      map[string][]int
		TextIndex      map[string][]int
		FavoriteIndex  map[int]bool
		HiddenIndex    map[int]bool
		MediaTypeIndex map[string][]int
		CameraIndex    map[string][]int
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
	ps.hiddenIndex = indexData.HiddenIndex
	ps.mediaTypeIndex = indexData.MediaTypeIndex
	ps.cameraIndex = indexData.CameraIndex

	ps.lastRebuild = time.Now()
	return nil
}

// saveIndex saves the index to disk
func (ps *PhotoStorage) saveIndex() error {
	indexData := struct {
		LastID         int
		AssetIndex     map[int]string
		UserIndex      map[int][]int
		DateIndex      map[string][]int
		TextIndex      map[string][]int
		FavoriteIndex  map[int]bool
		HiddenIndex    map[int]bool
		MediaTypeIndex map[string][]int
		CameraIndex    map[string][]int
	}{
		LastID:         ps.lastID,
		AssetIndex:     ps.assetIndex,
		UserIndex:      ps.userIndex,
		DateIndex:      ps.dateIndex,
		TextIndex:      ps.textIndex,
		FavoriteIndex:  ps.favoriteIndex,
		HiddenIndex:    ps.hiddenIndex,
		MediaTypeIndex: ps.mediaTypeIndex,
		CameraIndex:    ps.cameraIndex,
	}

	data, err := json.Marshal(indexData)
	if err != nil {
		return err
	}

	tmpFile := ps.config.IndexFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, ps.config.IndexFile)
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
			log.Printf("Asset file missing for %d: %s", id, asset.Filename)
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

// cleanupOrphanedAssets removes assets with missing files
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

// filterByText filters assets by search query
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

// filterByMediaType filters assets by media type
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

// filterByFavorite filters assets by favorite status
func (ps *PhotoStorage) filterByFavorite(assetIDs []int, favorite bool) []int {
	filtered := make([]int, 0, len(assetIDs))
	for _, id := range assetIDs {
		if ps.favoriteIndex[id] == favorite {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

// filterByDateRange filters assets by date range
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

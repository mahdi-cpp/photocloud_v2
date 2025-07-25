package storage

import (
	"encoding/json"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/pkg/asset_model"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MetadataManager handles asset metadata
type MetadataManager struct {
	dir   string
	mutex sync.RWMutex
}

func NewMetadataManager(dir string) *MetadataManager {
	return &MetadataManager{
		dir: dir,
	}
}

// SaveMetadata saves asset metadata
func (m *MetadataManager) SaveMetadata(asset *asset_model.PHAsset) error {
	path := m.getMetadataPath(asset.ID)

	data, err := json.MarshalIndent(asset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Atomic write
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return os.Rename(tmpPath, path)
}

// LoadMetadata loads asset metadata
func (m *MetadataManager) LoadMetadata(id int) (*asset_model.PHAsset, error) {
	path := m.getMetadataPath(id)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var asset asset_model.PHAsset
	if err := json.Unmarshal(data, &asset); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &asset, nil
}

func (m *MetadataManager) LoadUserAllMetadata() (map[int]*asset_model.PHAsset, error) {

	startTime := time.Now() // Capture start time

	var assets = make(map[int]*asset_model.PHAsset)

	// Scan metadata directory
	files, err := os.ReadDir(m.dir)
	if err != nil {
		fmt.Println("failed to read metadata directory: %w", err)
		return nil, err
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
		asset, err := m.LoadMetadata(id)
		if err != nil {
			log.Printf("Skipping invalid metadata %s: %v", filename, err)
			continue
		}

		//assets = append(assets, asset)
		assets[asset.ID] = asset
	}

	// Calculate and log execution duration
	duration := time.Since(startTime)
	log.Printf("Load Metadata in %v. Scanned %d assets", duration, len(assets))
	fmt.Println("")

	return assets, nil
}

// DeleteMetadata removes metadata file
func (m *MetadataManager) DeleteMetadata(id int) error {
	path := m.getMetadataPath(id)
	return os.Remove(path)
}

func (m *MetadataManager) getMetadataPath(id int) string {
	return filepath.Join(m.dir, fmt.Sprintf("%d.json", id))
}

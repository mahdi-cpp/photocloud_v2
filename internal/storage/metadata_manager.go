package storage

import (
	"encoding/json"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"os"
	"path/filepath"
)

// MetadataManager handles asset metadata
type MetadataManager struct {
	dir string
}

func NewMetadataManager(dir string) *MetadataManager {
	return &MetadataManager{dir: dir}
}

// SaveMetadata saves asset metadata
func (m *MetadataManager) SaveMetadata(asset *model.PHAsset) error {
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
func (m *MetadataManager) LoadMetadata(id int) (*model.PHAsset, error) {
	path := m.getMetadataPath(id)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var asset model.PHAsset
	if err := json.Unmarshal(data, &asset); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &asset, nil
}

// DeleteMetadata removes metadata file
func (m *MetadataManager) DeleteMetadata(id int) error {
	path := m.getMetadataPath(id)
	return os.Remove(path)
}

func (m *MetadataManager) getMetadataPath(id int) string {
	return filepath.Join(m.dir, fmt.Sprintf("%d.json", id))
}

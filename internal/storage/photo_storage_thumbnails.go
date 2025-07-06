package storage

import (
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/pkg/imaging"
	"log"
	"os"
	"path/filepath"
)

// SaveThumbnail saves a thumbnail to disk
func (ps *PhotoStorage) SaveThumbnail(assetID, width, height int, data []byte) error {
	filename := ps.getThumbnailFilename(assetID, width, height)
	path := filepath.Join(ps.config.ThumbnailsDir, filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return nil
}

// GetThumbnail retrieves or generates a thumbnail
func (ps *PhotoStorage) GetThumbnail(id int, width, height int) ([]byte, error) {

	// First try to get existing thumbnail
	if thumb, err := ps.thumbnail.GetThumbnail(id, width, height); err == nil {
		return thumb, nil
	}

	// Get asset
	asset, err := ps.GetAsset(id)
	if err != nil {
		return nil, err
	}

	// Get asset content
	content, err := ps.GetAssetContent(id)
	if err != nil {
		return nil, err
	}

	// Generate thumbnail
	thumbnailService := imaging.NewThumbnailService(width, height, 85, ps, true, "")
	thumbData, err := thumbnailService.GenerateThumbnail(asset, content)
	if err != nil {
		return nil, fmt.Errorf("thumbnail generation failed: %w", err)
	}

	// Save thumbnail for future use
	if err := ps.thumbnail.SaveThumbnail(id, width, height, thumbData); err != nil {
		log.Printf("Failed to save thumbnail: %v", err)
	}

	// Update stats
	ps.statsMu.Lock()
	ps.stats.ThumbnailsGen++
	ps.statsMu.Unlock()

	return thumbData, nil
}

func (ps *PhotoStorage) getThumbnailFilename(assetID, width, height int) string {
	return fmt.Sprintf("%d_%dx%d.jpg", assetID, width, height)
}

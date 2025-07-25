package thumbnail

import (
	"fmt"
	"os"
	"path/filepath"
)

// ThumbnailManager handles thumbnail storage
type ThumbnailManager struct {
	dir string
}

func NewThumbnailManager(dir string) *ThumbnailManager {
	return &ThumbnailManager{dir: dir}
}

// SaveThumbnail saves a thumbnail
func (m *ThumbnailManager) SaveThumbnail(id, width, height int, data []byte) error {
	path := m.getThumbnailPath(id, width, height)
	return os.WriteFile(path, data, 0644)
}

// GetThumbnail retrieves a thumbnail
func (m *ThumbnailManager) GetThumbnail(id, width, height int) ([]byte, error) {
	path := m.getThumbnailPath(id, width, height)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrThumbnailNotFound
		}
		return nil, err
	}
	return data, nil
}

// DeleteThumbnails removes all thumbnails for an asset
func (m *ThumbnailManager) DeleteThumbnails(id int) {
	pattern := filepath.Join(m.dir, fmt.Sprintf("%d_*.jpg", id))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	for _, file := range matches {
		os.Remove(file)
	}
}

func (m *ThumbnailManager) getThumbnailPath(id, width, height int) string {
	filename := fmt.Sprintf("%d_%dx%d.jpg", id, width, height)
	return filepath.Join(m.dir, filename)
}

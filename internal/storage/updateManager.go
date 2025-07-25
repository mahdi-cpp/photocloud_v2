package storage

import (
	"encoding/json"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"
	"os"
	"path/filepath"
	"sync"
)

// UpdateManager handles asset metadata
type UpdateManager struct {
	dir   string
	mutex sync.RWMutex
}

func NewUpdateManager(dir string) *UpdateManager {
	return &UpdateManager{dir: dir}
}

func (update *UpdateManager) updateCameraMake(id int, CameraMake string) error {
	update.mutex.RLock()
	defer update.mutex.RUnlock()

	path := update.getMetadataPath(id)
	err := updateJSONFile(path, func(cfg *happle_models.PHAsset) error {
		cfg.CameraMake = CameraMake
		return nil
	})
	if err != nil {
		fmt.Println("Failed updateCameraMake: ", err.Error())
		return err
	}
	return nil
}

func (update *UpdateManager) updateAlbum(id int, CameraMake string) error {
	update.mutex.RLock()
	defer update.mutex.RUnlock()

	path := update.getMetadataPath(id)
	err := updateJSONFile(path, func(cfg *happle_models.PHAsset) error {
		cfg.CameraMake = CameraMake
		return nil
	})
	if err != nil {
		fmt.Println("Failed updateCameraMake: ", err.Error())
		return err
	}
	return nil
}

// UpdateJSONFile Generic JSON updater function
// https://chat.deepseek.com/a/chat/s/35e13526-87de-4a6c-a43c-02555b5ec33d
func updateJSONFile[T any](filePath string, updateFunc func(*T) error) error {

	// Read existing data or initialize zero value
	var data T
	fileBytes, err := os.ReadFile(filePath)
	if err == nil {
		if err := json.Unmarshal(fileBytes, &data); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// Execute update operation
	if err := updateFunc(&data); err != nil {
		return err
	}

	// Marshal updated data
	newJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Handler temp file
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, "tmp-*.json")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	// Write to temp file
	if _, err := tmpFile.Write(newJSON); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Preserve permissions
	if fileInfo, err := os.Stat(filePath); err == nil {
		os.Chmod(tmpName, fileInfo.Mode())
	} else {
		os.Chmod(tmpName, 0644) // Default if new file
	}

	// Atomic replacement
	return os.Rename(tmpName, filePath)
}

// OverwriteJSONFile replaces the entire JSON file with the given struct
// https://chat.deepseek.com/a/chat/s/35e13526-87de-4a6c-a43c-02555b5ec33d
func (update *UpdateManager) OverwriteJSONFile(filePath string, data interface{}) error {
	update.mutex.RLock()
	defer update.mutex.RUnlock()

	// Marshal struct to JSON
	newJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Handler temp file in same directory
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, "tmp-*.json")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) // Cleanup if fails

	// Write JSON to temp file
	if _, err := tmpFile.Write(newJSON); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Preserve permissions if file exists
	if fileInfo, err := os.Stat(filePath); err == nil {
		os.Chmod(tmpName, fileInfo.Mode())
	} else {
		os.Chmod(tmpName, 0644) // Default for new files
	}

	// Atomically replace target file
	return os.Rename(tmpName, filePath)
}

func (update *UpdateManager) getMetadataPath(id int) string {
	return filepath.Join(update.dir, fmt.Sprintf("%d.json", id))
}

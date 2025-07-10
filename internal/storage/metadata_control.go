package storage

import (
	"encoding/json"
	"os"
	"sync"
)

// MetadataControl manages JSON file operations for any type T
type MetadataControl[T any] struct {
	filePath string
	mutex    sync.RWMutex // Protects concurrent access
}

// NewMetadataManagerV2 creates a new metadata for the specified JSON file
func NewMetadataManagerV2[T any](filePath string) *MetadataControl[T] {
	return &MetadataControl[T]{filePath: filePath}
}

// Read retrieves the current data from the JSON file
func (m *MetadataControl[T]) Read() (T, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var data T
	file, err := os.ReadFile(m.filePath)
	if err != nil {
		// Return empty data if file doesn't exist
		if os.IsNotExist(err) {
			return data, nil
		}
		return data, err
	}

	if len(file) == 0 {
		return data, nil
	}

	if err := json.Unmarshal(file, &data); err != nil {
		return data, err
	}

	return data, nil
}

// Update modifies the data using a callback function
func (m *MetadataControl[T]) Update(updateFunc func(*T) error) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Read current data
	var data T
	file, err := os.ReadFile(m.filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(file) > 0 {
		if err := json.Unmarshal(file, &data); err != nil {
			return err
		}
	}

	// Apply updates
	if err := updateFunc(&data); err != nil {
		return err
	}

	// Write updated data
	return m.writeData(data)
}

// Write replaces the entire file contents
func (m *MetadataControl[T]) Write(data T) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.writeData(data)
}

// writeData handles the actual file writing
func (m *MetadataControl[T]) writeData(data T) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first
	tempFile := m.filePath + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return err
	}

	// Atomic rename
	return os.Rename(tempFile, m.filePath)
}

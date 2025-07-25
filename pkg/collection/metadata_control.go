package collection

import (
	"encoding/json"
	"os"
	"sync"
)

// MetadataControl manages JSON file operations for any type T
type MetadataControl[T any] struct {
	filePath string
	mutex    sync.RWMutex
}

// NewMetadataControl creates a new metadata controller
func NewMetadataControl[T any](filePath string) *MetadataControl[T] {
	return &MetadataControl[T]{
		filePath: filePath,
	}
}

// Read retrieves the current data (returns pointer to allow modification)
func (m *MetadataControl[T]) Read() (*T, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	data := new(T)
	file, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil
		}
		return nil, err
	}

	if len(file) == 0 {
		return data, nil
	}

	if err := json.Unmarshal(file, data); err != nil {
		return nil, err
	}

	return data, nil
}

// Update modifies the data using a callback function (now accepts pointer)
func (m *MetadataControl[T]) Update(updateFunc func(*T) error) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Read current data into a pointer
	data, err := m.readData()
	if err != nil {
		return err
	}

	// Apply updates to the pointer
	if err := updateFunc(data); err != nil {
		return err
	}

	// Write updated data
	return m.writeData(data)
}

// readData helper function
func (m *MetadataControl[T]) readData() (*T, error) {
	data := new(T)
	file, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil
		}
		return nil, err
	}

	if len(file) > 0 {
		if err := json.Unmarshal(file, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// writeData handles the actual file writing (accepts pointer)
func (m *MetadataControl[T]) writeData(data *T) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	tempFile := m.filePath + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return err
	}

	return os.Rename(tempFile, m.filePath)
}

// Write replaces the entire file contents (accepts pointer)
func (m *MetadataControl[T]) Write(data *T) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.writeData(data)
}

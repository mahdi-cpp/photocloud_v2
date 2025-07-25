package storage

import "github.com/mahdi-cpp/photocloud_v2/pkg/asset_model"

// BoolIndexer indexes assets based on boolean fields
type BoolIndexer struct {
	BaseIndexer
	fieldName string // e.g., "IsFavorite", "IsHidden"
	trueSet   map[int]bool
	falseSet  map[int]bool
}

// NewBoolIndexer creates a new boolean indexer for a specific field
func NewBoolIndexer(fieldName string) *BoolIndexer {
	return &BoolIndexer{
		fieldName: fieldName,
		trueSet:   make(map[int]bool),
		falseSet:  make(map[int]bool),
	}
}

// Add adds an asset to the index
func (idx *BoolIndexer) Add(asset *asset_model.PHAsset) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	value := idx.getFieldValue(asset)
	if value {
		idx.trueSet[asset.ID] = true
		delete(idx.falseSet, asset.ID)
	} else {
		idx.falseSet[asset.ID] = true
		delete(idx.trueSet, asset.ID)
	}
}

// Remove removes an asset from the index
func (idx *BoolIndexer) Remove(id int) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.trueSet, id)
	delete(idx.falseSet, id)
}

// Search returns assets matching the boolean value
func (idx *BoolIndexer) Search(query interface{}) []int {
	value, ok := query.(bool)
	if !ok {
		return nil
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if value {
		return idx.getKeys(idx.trueSet)
	}
	return idx.getKeys(idx.falseSet)
}

// Filter filters assets based on boolean value
func (idx *BoolIndexer) Filter(ids []int, query interface{}) []int {
	value, ok := query.(bool)
	if !ok {
		return ids
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var set map[int]bool
	if value {
		set = idx.trueSet
	} else {
		set = idx.falseSet
	}

	var filtered []int
	for _, id := range ids {
		if set[id] {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

// getFieldValue extracts the boolean field value using reflection
func (idx *BoolIndexer) getFieldValue(asset *asset_model.PHAsset) bool {
	switch idx.fieldName {
	case "IsFavorite":
		return asset.IsFavorite
	case "IsHidden":
		return asset.IsHidden
	case "IsScreenshot":
		return asset.IsScreenshot
	default:
		return false
	}
}

// getKeys returns all keys from a map
func (idx *BoolIndexer) getKeys(set map[int]bool) []int {
	keys := make([]int, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}

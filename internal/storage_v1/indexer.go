package storage_v1

import (
	"github.com/mahdi-cpp/api-go-pkg/common_models"
	"strings"
	"sync"
	"time"
)

// Indexer defines the interface for indexing assets
type Indexer interface {
	Add(asset *common_models.PHAsset)
	Remove(id int)
	Search(query interface{}) []int
	Filter(ids []int, query interface{}) []int
}

// BaseIndexer provides common functionality for all indexers
type BaseIndexer struct {
	mu sync.RWMutex
}

// TextIndexer implements full-text search indexing
type TextIndexer struct {
	BaseIndexer
	index map[string]map[int]bool // word -> assetID -> exists
}

func NewTextIndexer() *TextIndexer {
	return &TextIndexer{
		index: make(map[string]map[int]bool),
	}
}

func (idx *TextIndexer) Add(asset *common_models.PHAsset) {
	words := tokenize(asset.Filename)

	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, word := range words {
		if idx.index[word] == nil {
			idx.index[word] = make(map[int]bool)
		}
		idx.index[word][asset.ID] = true
	}
}

func (idx *TextIndexer) Remove(id int) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for word := range idx.index {
		delete(idx.index[word], id)
		if len(idx.index[word]) == 0 {
			delete(idx.index, word)
		}
	}
}

func (idx *TextIndexer) Search(query interface{}) []int {
	queryStr, ok := query.(string)
	if !ok {
		return nil
	}

	words := tokenize(queryStr)
	results := make(map[int]int) // assetID -> match count

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	for _, word := range words {
		if assets, exists := idx.index[word]; exists {
			for id := range assets {
				results[id]++
			}
		}
	}

	// Return sorted by relevance (most matches first)
	var sorted []int
	for id, count := range results {
		if count == len(words) { // All words matched
			sorted = append(sorted, id)
		}
	}
	return sorted
}

func (idx *TextIndexer) Filter(ids []int, query interface{}) []int {
	queryStr, ok := query.(string)
	if !ok {
		return ids
	}

	words := tokenize(queryStr)
	matching := make(map[int]bool)

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Find all assets that match all words
	for _, word := range words {
		if assets, exists := idx.index[word]; exists {
			for id := range assets {
				if contains(ids, id) {
					matching[id] = true
				}
			}
		}
	}

	// Filter original list
	var filtered []int
	for _, id := range ids {
		if matching[id] {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

// DateIndexer implements date-based indexing
type DateIndexer struct {
	BaseIndexer
	index map[string]map[int]bool // date -> assetID -> exists
}

func NewDateIndexer() *DateIndexer {
	return &DateIndexer{
		index: make(map[string]map[int]bool),
	}
}

func (idx *DateIndexer) Add(asset *common_models.PHAsset) {
	dateKey := asset.CreationDate.Format("2006-01-02")

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if idx.index[dateKey] == nil {
		idx.index[dateKey] = make(map[int]bool)
	}
	idx.index[dateKey][asset.ID] = true
}

func (idx *DateIndexer) Remove(id int) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for date := range idx.index {
		delete(idx.index[date], id)
		if len(idx.index[date]) == 0 {
			delete(idx.index, date)
		}
	}
}

func (idx *DateIndexer) Search(query interface{}) []int {
	dateStr, ok := query.(string)
	if !ok {
		return nil
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if assets, exists := idx.index[dateStr]; exists {
		ids := make([]int, 0, len(assets))
		for id := range assets {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

func (idx *DateIndexer) Filter(ids []int, query interface{}) []int {
	dateRange, ok := query.([]time.Time)
	if !ok || len(dateRange) < 1 {
		return ids
	}

	start := dateRange[0]
	var end time.Time
	if len(dateRange) > 1 {
		end = dateRange[1]
	} else {
		end = time.Now()
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var filtered []int
	for dateStr, assets := range idx.index {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if (date.Equal(start) || date.After(start)) &&
			(date.Equal(end) || date.Before(end)) {
			for id := range assets {
				if contains(ids, id) {
					filtered = append(filtered, id)
				}
			}
		}
	}
	return filtered
}

// MediaTypeIndexer indexes assets by media type
type MediaTypeIndexer struct {
	BaseIndexer
	index map[string]map[int]bool // mediaType -> assetID -> exists
}

func NewMediaTypeIndexer() *MediaTypeIndexer {
	return &MediaTypeIndexer{
		index: make(map[string]map[int]bool),
	}
}

// ... similar implementation pattern for other indexers ...

// Helper functions
func tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.Fields(text)
	var tokens []string
	for _, word := range words {
		if len(word) > 2 {
			tokens = append(tokens, word)
		}
	}
	return tokens
}

func contains(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

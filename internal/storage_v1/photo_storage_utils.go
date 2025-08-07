package storage_v1

import (
	"sort"
	"strings"
	"time"
)

// Additional utility methods for PhotoStorage

func (ps *PhotoStorage) GetUserAssets(userID int) []int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	assets := make([]int, len(ps.userIndex[userID]))
	copy(assets, ps.userIndex[userID])

	// Sort by date (newest first)
	sort.Slice(assets, func(i, j int) bool {
		return ps.getAssetDate(assets[i]).After(ps.getAssetDate(assets[j]))
	})

	return assets
}

func (ps *PhotoStorage) getAssetDate(assetID int) time.Time {
	// Find in date index
	for dateStr, ids := range ps.dateIndex {
		for _, id := range ids {
			if id == assetID {
				t, _ := time.Parse("2006-01-02", dateStr)
				return t
			}
		}
	}
	return time.Time{}
}

func (ps *PhotoStorage) CountUserAssets(userID int) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.userIndex[userID])
}

func (ps *PhotoStorage) SuggestSearchTerms(userID int, prefix string, limit int) []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	prefix = strings.ToLower(prefix)
	var suggestions []string

	// Find matching terms in text index
	for term := range ps.textIndex {
		if strings.HasPrefix(term, prefix) {
			suggestions = append(suggestions, term)
			if len(suggestions) >= limit {
				break
			}
		}
	}

	return suggestions
}

func (ps *PhotoStorage) GetAssetsWithoutThumbnails() ([]int, error) {
	// Implementation to find assets without thumbnails
	// This would typically scan the thumbnails directory and compare with asset index
	return nil, nil
}

func (ps *PhotoStorage) CleanupOrphanedAssets() (int, error) {
	// Implementation already exists in periodicMaintenance
	// This would just trigger it and return count
	return 0, nil
}

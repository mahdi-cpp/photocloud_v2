package model

import "time"

// StorageStats represents storage system statistics
type StorageStats struct {
	TotalAssets   int   `json:"totalAssets"`
	CacheSize     int   `json:"cacheSize"`
	CacheHits     int64 `json:"cacheHits"`
	CacheMisses   int64 `json:"cacheMisses"`
	Uploads24h    int   `json:"uploads24h"`
	ThumbnailsGen int   `json:"thumbnailsGenerated"`
}

// IndexStatus represents index health information
type IndexStatus struct {
	LastRebuild       time.Time `json:"lastRebuild"`
	AssetCount        int       `json:"assetCount"`
	IndexSize         int       `json:"indexSize"`
	Dirty             bool      `json:"dirty"`
	RebuildInProgress bool      `json:"rebuildInProgress"`
}

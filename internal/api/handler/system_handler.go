package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage_v1"
	"log"
	"net/http"
	"time"
)

type SystemHandler struct {
	storage *storage_v1.PhotoStorage
}

func NewSystemHandler(storage *storage_v1.PhotoStorage) *SystemHandler {
	return &SystemHandler{storage: storage}
}

// GetSystemStatus godoc
// @Summary Get system health and statistics
// @Produce json
// @Success 200 {object} SystemStatusResponse
// @Router /system/status [get]
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	stats := h.storage.GetSystemStats()
	indexStatus := h.storage.GetIndexStatus()

	c.JSON(http.StatusOK, SystemStatusResponse{
		Status:    "operational",
		Timestamp: time.Now().Unix(),
		Stats:     storage.Stats(stats),
		Index:     indexStatus,
	})
}

// RebuildIndex godoc
// @Summary Trigger index rebuild
// @Produce json
// @Success 202 {object} map[string]string
// @Router /system/rebuild-index [post]
func (h *SystemHandler) RebuildIndex(c *gin.Context) {
	go func() {
		if err := h.storage.RebuildIndex(); err != nil {
			log.Printf("Index rebuild failed: %v", err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Index rebuild started",
	})
}

// SystemStatusResponse represents system health data
type SystemStatusResponse struct {
	Status    string                 `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	Stats     storage.Stats          `json:"stats"`
	Index     storage_v1.IndexStatus `json:"index"`
}

// StorageStats represents storage system statistics
type StorageStats struct {
	TotalAssets   int `json:"totalAssets"`
	CacheSize     int `json:"cacheSize"`
	CacheHits     int `json:"cacheHits"`
	CacheMisses   int `json:"cacheMisses"`
	Uploads24h    int `json:"uploads24h"`
	ThumbnailsGen int `json:"thumbnailsGenerated"`
}

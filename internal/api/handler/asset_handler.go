package handler

import (
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewAssetHandler(userStorageManager *storage.UserStorageManager) *AssetHandler {
	return &AssetHandler{userStorageManager: userStorageManager}
}

func (h *AssetHandler) UploadAsset(c *gin.Context) {

	userID := c.GetInt("userID")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload error"})
		return
	}
	defer file.Close()

	// Create asset metadata
	asset := &model.PHAsset{
		UserID:   userID,
		Filename: header.Filename,
	}

	asset, err = h.userStorageManager.UploadAsset(c, asset.UserID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
		return
	}

	//asset, err := h.userStorageManager.UploadAsset(c, userID, file, header)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
	//	return
	//}

	c.JSON(http.StatusCreated, asset)
}

func (h *AssetHandler) UpdateAssets(c *gin.Context) {

	startTime := time.Now()

	var update model.AssetUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	fmt.Println("UpdateAssets: ", update.AssetIds)

	asset, err := h.userStorageManager.UpdateAsset(c, update.AssetIds, update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log performance
	duration := time.Since(startTime)
	log.Printf("UpdateAssets: assets count: %d,  (in %v)", len(update.AssetIds), duration)

	c.JSON(http.StatusCreated, asset)
}

func (h *AssetHandler) GetAsset(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	asset, err := h.userStorageManager.GetAsset(c, 1, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

func (h *AssetHandler) SearchAssets(c *gin.Context) {

	userID := c.GetInt("userID")
	query := c.Query("query")
	mediaType := c.Query("type")

	var dateRange []time.Time
	if start := c.Query("start"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			dateRange = append(dateRange, t)
		}
	}
	if end := c.Query("end"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			dateRange = append(dateRange, t)
		}
	}

	filters := model.AssetSearchFilters{
		UserID:    userID,
		Query:     query,
		MediaType: model.MediaType(mediaType),
	}

	if len(dateRange) > 0 {
		filters.StartDate = &dateRange[0]
	}
	if len(dateRange) > 1 {
		filters.EndDate = &dateRange[1]
	}

	//assets, _, err := s.repo.SearchAssets(ctx, filters)
	//return assets, err

	//assets, _, err := h.userStorageManager.SearchAssets(c, filters)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
	//	return
	//}

	//c.JSON(http.StatusOK, assets)
}

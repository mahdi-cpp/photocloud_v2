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
	assetRepo *storage.AssetRepositoryImpl
}

func NewAssetHandler(assetRepo *storage.AssetRepositoryImpl) *AssetHandler {
	return &AssetHandler{assetRepo: assetRepo}
}

// UploadAsset godoc
// @Summary Upload a new asset
// @Accept  multipart/form-data
// @Param   file formData file true "Asset file"
// @Success 201 {object} model.PHAsset
// @Router /assets [post]
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

	asset, err = h.assetRepo.CreateAsset(c, asset, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
		return
	}
	//asset, err := h.assetRepo.UploadAsset(c, userID, file, header)
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

	asset, err := h.assetRepo.UpdateAsset(c, update.AssetIds, update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log performance
	duration := time.Since(startTime)
	log.Printf("UpdateAssets: assets count: %d,  (in %v)", len(update.AssetIds), duration)

	c.JSON(http.StatusCreated, asset)
}

// GetAsset godoc
// @Summary Get asset metadata
// @Param   id path int true "Asset ID"
// @Success 200 {object} model.PHAsset
// @Router /assets/{id} [get]
func (h *AssetHandler) GetAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	asset, err := h.assetRepo.GetAsset(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// SearchAssets godoc
// @Summary Search assets
// @Param   query query string false "Search query"
// @Param   type query string false "Media type" Enums(image,video)
// @Param   start query string false "Start date (YYYY-MM-DD)"
// @Param   end query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} model.PHAsset
// @Router /search [get]
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

	assets, _, err := h.assetRepo.SearchAssets(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, assets)
}

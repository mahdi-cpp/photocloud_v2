package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mahdi-cpp/photocloud_v2/internal/service"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	assetService *service.AssetService
}

func NewAssetHandler(assetService *service.AssetService) *AssetHandler {
	return &AssetHandler{assetService: assetService}
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

	asset, err := h.assetService.UploadAsset(c, userID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
		return
	}

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

	asset, err := h.assetService.GetAsset(c, id)
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

	assets, err := h.assetService.SearchAssets(c, userID, query, mediaType, dateRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, assets)
}

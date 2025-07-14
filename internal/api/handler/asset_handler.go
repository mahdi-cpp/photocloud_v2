package handler

import (
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewAssetHandler(userStorageManager *storage.UserStorageManager) *AssetHandler {
	return &AssetHandler{userStorageManager: userStorageManager}
}

func (h *AssetHandler) Upload(c *gin.Context) {

	userID := c.GetInt("userID")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload error"})
		return
	}
	defer file.Close()

	// Handler asset metadata
	asset := &model.PHAsset{
		UserID:   userID,
		Filename: header.Filename,
	}

	asset, err = h.userStorageManager.UploadAsset(c, asset.UserID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
		return
	}

	//asset, err := h.userStorageManager.Upload(c, userID, file, header)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
	//	return
	//}

	c.JSON(http.StatusCreated, asset)
}

func (h *AssetHandler) Update(c *gin.Context) {

	startTime := time.Now()

	var update model.AssetUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	fmt.Println("Update: ", update.AssetIds)

	asset, err := h.userStorageManager.UpdateAsset(c, update.AssetIds, update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log performance
	duration := time.Since(startTime)
	log.Printf("Update: assets count: %d,  (in %v)", len(update.AssetIds), duration)

	c.JSON(http.StatusCreated, asset)
}

func (h *AssetHandler) Get(c *gin.Context) {

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

func (h *AssetHandler) Search(c *gin.Context) {

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

	//assets, _, err := s.repo.Search(ctx, filters)
	//return assets, err

	//assets, _, err := h.userStorageManager.Search(c, filters)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
	//	return
	//}

	//c.JSON(http.StatusOK, assets)
}

func (h *AssetHandler) Delete(c *gin.Context) {

	var request model.AssetDelete
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.userStorageManager.Delete(c, request.UserID, request.AssetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "successful delete asset with id: "+strconv.Itoa(request.AssetID))
}

//----------------------------------------

func (h *AssetHandler) OriginalDownload(c *gin.Context) {

	filename := c.Param("filename")
	filepath2 := filepath.Join("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki", filename)

	fileSize, err := storage.GetFileSize(filepath2)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to get file size"})
		return
	}

	//c.Header("Content-Type", "mage/jpeg")
	//c.Header("Content-Encoding", "identity") // Disable compression
	//c.Next()
	c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Accept-Ranges", "bytes")
	c.File(filepath2)
}

func (h *AssetHandler) TinyImageDownload(c *gin.Context) {
	filename := c.Param("filename")

	if strings.Contains(filename, "png") {
		imgData, err := h.userStorageManager.RepositoryGetIcon(filename)
		if err != nil {
			fmt.Println("icon read error: ", err.Error())
		} else {
			c.Data(http.StatusOK, "image/png", imgData) // Adjust MIME type as necessary
		}
		return
	}

	filepathTiny := filepath.Join("mahdi_abdolmaleki/thumbnails", filename)

	imgData, err := h.userStorageManager.RepositoryGetImage(filepathTiny)
	if err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "File not found"})
	} else {
		c.Data(http.StatusOK, "image/jpeg", imgData)
	}
}

func (h *AssetHandler) IconDownload(c *gin.Context) {
	filename := c.Param("filename")
	imgData, err := h.userStorageManager.RepositoryGetImage(filename)
	if err != nil {
		c.Data(http.StatusOK, "image/png", imgData) // Adjust MIME type as necessary
	}
}

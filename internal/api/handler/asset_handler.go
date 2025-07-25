package handler

import (
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"
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

func (handler *AssetHandler) Create(c *gin.Context) {

}

func (handler *AssetHandler) Upload(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload error"})
		return
	}
	defer file.Close()

	// Handler asset metadata
	asset := &happle_models.PHAsset{
		UserID:   userID,
		Filename: header.Filename,
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	asset, err = userStorage.UploadAsset(asset.UserID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
		return
	}

	//asset, err := handler.userStorageManager.Upload(c, userID, file, header)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
	//	return
	//}

	c.JSON(http.StatusCreated, asset)
}

func (handler *AssetHandler) Update(c *gin.Context) {

	startTime := time.Now()

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var update happle_models.AssetUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	asset, err := userStorage.UpdateAsset(update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userStorage.UpdateCollections()

	// Log performance
	duration := time.Since(startTime)
	log.Printf("Update: assets count: %d,  (in %v)", len(update.AssetIds), duration)

	c.JSON(http.StatusCreated, asset)
}

func (handler *AssetHandler) UpdateAll(c *gin.Context) {
	startTime := time.Now()

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var update happle_models.AssetUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	allAssets := userStorage.GetAllAssets()
	for _, asset := range allAssets {
		update.AssetIds = append(update.AssetIds, asset.ID)
	}

	asset, err := userStorage.UpdateAsset(update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userStorage.UpdateCollections()

	// Log performance
	duration := time.Since(startTime)
	log.Printf("Update: assets count: %d,  (in %v)", len(update.AssetIds), duration)

	c.JSON(http.StatusCreated, asset)
}

func (handler *AssetHandler) Get(c *gin.Context) {

	userIDStr := c.Query("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	asset, exists := userStorage.GetAsset(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

func (handler *AssetHandler) Search(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

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

	filters := happle_models.PHFetchOptions{
		UserID:    userID,
		Query:     query,
		MediaType: happle_models.MediaType(mediaType),
	}

	if len(dateRange) > 0 {
		filters.StartDate = &dateRange[0]
	}
	if len(dateRange) > 1 {
		filters.EndDate = &dateRange[1]
	}

	//assets, _, err := s.repo.Search(ctx, filters)
	//return assets, err

	//assets, _, err := handler.userStorageManager.Search(c, filters)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
	//	return
	//}

	//c.JSON(http.StatusOK, assets)
}

func (handler *AssetHandler) Delete(c *gin.Context) {

	userIDStr := c.Query("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var request happle_models.AssetDelete
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	err = userStorage.DeleteAsset(request.AssetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "successful delete asset with id: "+strconv.Itoa(request.AssetID))
}

func (handler *AssetHandler) Filters(c *gin.Context) {

	fmt.Println("Filters 1")

	userID, err := getUserId(c)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID must be an integer"})
		return
	}

	fmt.Println("Filters")

	var with happle_models.PHFetchOptions
	if err := c.ShouldBindJSON(&with); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		fmt.Println("Invalid request")
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	items, total, err := userStorage.FetchAssets(with)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed user FetchAssets"})
		return
	}

	fmt.Println("Filters count: ", len(items))

	result := happle_models.PHFetchResult[*happle_models.PHAsset]{
		Items:  items,
		Total:  total,
		Limit:  100,
		Offset: 100,
	}

	c.JSON(http.StatusOK, result)
}

//----------------------------------------

func (handler *AssetHandler) OriginalDownload(c *gin.Context) {

	filename := c.Param("filename")
	filepath2 := filepath.Join("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets", filename)

	fileSize, err := storage.GetFileSize(filepath2)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to get file size"})
		return
	}

	filepathTiny := filepath.Join("mahdi_abdolmaleki/assets", filename)

	imgData, err := handler.userStorageManager.RepositoryGetOriginalImage(filepathTiny)
	if err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "File not found"})
	} else {
		c.Data(http.StatusOK, "image/jpeg", imgData)
	}

	//c.Header("Content-Type", "mage/jpeg")
	//c.Header("Content-Encoding", "identity") // Disable compression
	//c.Next()
	c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Accept-Ranges", "bytes")
	c.File(filepath2)
}

func (handler *AssetHandler) TinyImageDownload(c *gin.Context) {

	filename := c.Param("filename")
	if strings.Contains(filename, "png") {
		imgData, err := handler.userStorageManager.RepositoryGetIcon(filename)
		if err != nil {
			fmt.Println("icon read error: ", err.Error())
		} else {
			c.Data(http.StatusOK, "image/png", imgData) // Adjust MIME type as necessary
		}
		return
	}

	filepathTiny := filepath.Join("mahdi_abdolmaleki/thumbnails", filename)

	imgData, err := handler.userStorageManager.RepositoryGetTinyImage(filepathTiny)
	if err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "File not found"})
	} else {
		c.Data(http.StatusOK, "image/jpeg", imgData)
	}
}

func (handler *AssetHandler) IconDownload(c *gin.Context) {
	filename := c.Param("filename")
	imgData, err := handler.userStorageManager.RepositoryGetTinyImage(filename)
	if err != nil {
		c.Data(http.StatusOK, "image/png", imgData) // Adjust MIME type as necessary
	}
}

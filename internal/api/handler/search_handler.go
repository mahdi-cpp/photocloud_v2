package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"github.com/mahdi-cpp/photocloud_v2/pkg/asset_model"
	"net/http"
)

type SearchHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewSearchHandler(userStorageManager *storage.UserStorageManager) *SearchHandler {
	return &SearchHandler{userStorageManager: userStorageManager}
}

func (handler *SearchHandler) Filters(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var with asset_model.PHFetchOptions
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	fmt.Println("Filters count: ", len(items))

	result := asset_model.PHFetchResult[*asset_model.PHAsset]{
		Items:  items,
		Total:  total,
		Limit:  100,
		Offset: 100,
	}
	c.JSON(http.StatusOK, result)
}

func (handler *SearchHandler) Search(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var with asset_model.PHFetchOptions
	if err := c.ShouldBindJSON(&with); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	items, total, err := userStorage.FetchAssets(with)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	result := asset_model.PHFetchResult[*asset_model.PHAsset]{
		Items:  items,
		Total:  total,
		Limit:  100,
		Offset: 100,
	}
	c.JSON(http.StatusOK, result)
}

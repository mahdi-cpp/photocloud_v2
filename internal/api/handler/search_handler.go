package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
)

type SearchHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewSearchHandler(userStorageManager *storage.UserStorageManager) *SearchHandler {
	return &SearchHandler{userStorageManager: userStorageManager}
}

func (h *SearchHandler) Filters(c *gin.Context) {

	var filters model.PHFetchOptions
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		fmt.Println("Invalid request")
		return
	}

	fmt.Println("Filters userId: ", filters.UserID)

	items, total, err := h.userStorageManager.FetchAssets(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	fmt.Println("Filters count: ", len(items))

	result := model.PHFetchResult[*model.PHAsset]{
		Items:  items,
		Total:  total,
		Limit:  100,
		Offset: 100,
	}
	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) Search(c *gin.Context) {

	var filters model.PHFetchOptions
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	items, total, err := h.userStorageManager.FetchAssets(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	result := model.PHFetchResult[*model.PHAsset]{
		Items:  items,
		Total:  total,
		Limit:  100,
		Offset: 100,
	}
	c.JSON(http.StatusOK, result)
}

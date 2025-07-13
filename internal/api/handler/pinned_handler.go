package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
	"strconv"
)

type PinnedHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewPinnedHandler(userStorageManager *storage.UserStorageManager) *PinnedHandler {
	return &PinnedHandler{userStorageManager: userStorageManager}
}

func (handler *PinnedHandler) GetList(c *gin.Context) {

	userIDStr := c.Query("userID") // Get the string value
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		// Handle error (e.g., invalid input, empty value)
		// Example: Return an HTTP 400 Bad Request error
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	manager := handler.userStorageManager.GetPinnedManager(c, userID)
	data, err := manager.GetList(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, data)
}

func (handler *PinnedHandler) Create(c *gin.Context) {

	var pinned model.PinnedHandler
	if err := c.ShouldBindJSON(&pinned); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager := handler.userStorageManager.GetPinnedManager(c, 4)
	data, err := manager.Create(pinned.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, data)
}

func (handler *PinnedHandler) Update(c *gin.Context) {

	var album model.AlbumHandler
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager := handler.userStorageManager.GetPinnedManager(c, album.UserID)
	album2, err := manager.Update(album.ID, album.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, album2)
}

func (handler *PinnedHandler) Delete(c *gin.Context) {

	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager := handler.userStorageManager.GetAlbumManager(c, 4)
	err := manager.Delete(album.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

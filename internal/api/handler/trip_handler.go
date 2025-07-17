package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
	"strconv"
)

type TripHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewTripHandler(userStorageManager *storage.UserStorageManager) *TripHandler {
	return &TripHandler{userStorageManager: userStorageManager}
}

func (handler *TripHandler) GetList(c *gin.Context) {

	userIDStr := c.Query("userID") // Get the string value
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		// Handle error (e.g., invalid input, empty value)
		// Example: Return an HTTP 400 Bad Request error
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	manager, err := handler.userStorageManager.GetTripManager(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	data, err := manager.GetList(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, data)
}

func (handler *TripHandler) Create(c *gin.Context) {

	var trip model.TripHandler
	if err := c.ShouldBindJSON(&trip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager, err := handler.userStorageManager.GetTripManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	data, err := manager.Create(trip.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, data)

}

func (handler *TripHandler) Update(c *gin.Context) {

	var album model.AlbumHandler
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager, err := handler.userStorageManager.GetTripManager(c, album.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	album2, err := manager.Update(album.ID, album.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, album2)
}

func (handler *TripHandler) Delete(c *gin.Context) {

	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager, err := handler.userStorageManager.GetAlbumManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	err = manager.Delete(album.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

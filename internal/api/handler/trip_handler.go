package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
)

type TripHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewTripHandler(userStorageManager *storage.UserStorageManager) *TripHandler {
	return &TripHandler{userStorageManager: userStorageManager}
}

func (handler *TripHandler) GetList(c *gin.Context) {

	tripManager := handler.userStorageManager.GetAlbumManager(c, 4)
	data, err := tripManager.GetList(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, data)
}

func (handler *TripHandler) Create(c *gin.Context) {

	var trip model.Trip
	if err := c.ShouldBindJSON(&trip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	tripManager := handler.userStorageManager.GetTripManager(c, 4)
	data, err := tripManager.Create(trip.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, data)
}

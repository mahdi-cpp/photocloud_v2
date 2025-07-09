package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"log"
	"net/http"
)

type TripHandler struct {
	manager *storage.TripManager
}

func NewTripHandler(manager *storage.TripManager) *TripHandler {
	return &TripHandler{manager: manager}
}

func (handler *TripHandler) GetList(c *gin.Context) {
	list, err := handler.manager.GetList(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, list)
}

func (handler *TripHandler) Create(c *gin.Context) {

	var trip model.Trip
	if err := c.ShouldBindJSON(&trip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	createTrip, err := handler.manager.CreateTrip("رامسر 1403", "تور پاییزی", true)
	if err != nil {
		log.Fatal("Failed CreateTrip: ", err)
		return
	}

	c.JSON(http.StatusCreated, createTrip)
}

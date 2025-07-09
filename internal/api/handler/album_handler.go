package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"log"
	"net/http"
)

type AlbumHandler struct {
	manager *storage.AlbumManager
}

func NewAlbumHandler(manager *storage.AlbumManager) *AlbumHandler {
	return &AlbumHandler{manager: manager}
}

func (handler *AlbumHandler) GetList(c *gin.Context) {
	assets, err := handler.manager.List(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, assets)
}

func (handler *AlbumHandler) Create(c *gin.Context) {

	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	album2, err := handler.manager.CreateAlbum("Camera ", "favourite", true)
	if err != nil {
		log.Fatal("Failed Create: ", err)
		return
	}

	c.JSON(http.StatusCreated, album2)
}

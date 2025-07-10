package handler

import (
	"fmt"
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
	fmt.Println("Ip: ", c.ClientIP())

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

	album2, err := handler.manager.Create("Camera ", "favourite", true)
	if err != nil {
		log.Fatal("Failed Create: ", err)
		return
	}

	c.JSON(http.StatusCreated, album2)
}

func (handler *AlbumHandler) Update(c *gin.Context) {

	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	album2, err := handler.manager.Update(album.ID, album.Name)
	if err != nil {
		log.Println("failed to update: ", err)
		return
	}

	c.JSON(http.StatusCreated, album2)
}

func (handler *AlbumHandler) Delete(c *gin.Context) {

	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := handler.manager.Delete(album.ID)
	if err != nil {
		log.Println("failed to delete: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to delete"})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

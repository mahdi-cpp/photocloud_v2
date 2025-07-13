package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
)

type AlbumHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewAlbumHandler(userStorageManager *storage.UserStorageManager) *AlbumHandler {
	return &AlbumHandler{userStorageManager: userStorageManager}
}

func (handler *AlbumHandler) GetList(c *gin.Context) {
	fmt.Println("Ip: ", c.ClientIP())

	albumManager := handler.userStorageManager.GetAlbumManager(c, 4)
	albums, err := albumManager.GetList(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, albums)
}

func (handler *AlbumHandler) Create(c *gin.Context) {

	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	albumManager := handler.userStorageManager.GetAlbumManager(c, 4)
	album2, err := albumManager.Create("Camera ", "favourite", true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, album2)
}

func (handler *AlbumHandler) Update(c *gin.Context) {

	var album model.AlbumHandler
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	fmt.Println("Album Update: ", album.ID)

	albumManager := handler.userStorageManager.GetAlbumManager(c, album.UserID)
	album2, err := albumManager.Update(album.ID, album.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
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

	albumManager := handler.userStorageManager.GetAlbumManager(c, 4)
	err := albumManager.Delete(album.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

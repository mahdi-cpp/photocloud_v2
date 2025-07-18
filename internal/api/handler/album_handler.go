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
	return &AlbumHandler{
		userStorageManager: userStorageManager,
	}
}

func (handler *AlbumHandler) Create(c *gin.Context) {

	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	albumManager, err := handler.userStorageManager.GetAlbumManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	album2, err := albumManager.Create(&album)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, album2)
}

func (handler *AlbumHandler) Update(c *gin.Context) {

	var itemHandler model.AlbumHandler
	if err := c.ShouldBindJSON(&itemHandler); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	fmt.Println("Album Collection Update: ", itemHandler.ID)

	collectionManager, err := handler.userStorageManager.GetAlbumManager(c, itemHandler.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	item, err := collectionManager.Get(itemHandler.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	model.UpdateAlbum(item, itemHandler)

	album2, err := collectionManager.Update(item)
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

	albumManager, err := handler.userStorageManager.GetAlbumManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	err = albumManager.Delete(album.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

func (handler *AlbumHandler) GetCollectionList(c *gin.Context) {

	albumManager, err := handler.userStorageManager.GetAlbumManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	albums, err := albumManager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, albums)
}

func (handler *AlbumHandler) GetListV2(c *gin.Context) {

	albumManager, err := handler.userStorageManager.GetAlbumManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	albums, err := albumManager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	fmt.Println("GetListV2 albums count: ", len(albums))

	result := model.PHCollectionList[*model.Album]{
		Collections: make([]*model.PHCollection[*model.Album], len(albums)),
	}

	for i, album := range albums {
		assets, _ := albumManager.GetItemAssets(album.ID)
		result.Collections[i] = &model.PHCollection[*model.Album]{
			Item:   album,
			Assets: assets,
		}
	}

	c.JSON(http.StatusOK, result)
}

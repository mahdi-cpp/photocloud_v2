package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
	"strconv"
)

type CollectionHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewCollectionHandler(userStorageManager *storage.UserStorageManager) *CollectionHandler {
	return &CollectionHandler{
		userStorageManager: userStorageManager,
	}
}

func (handler *CollectionHandler) Create(c *gin.Context) {
	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetCollectionManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	album2, err := collectionManager.Create(&album)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, album2)
}

func (handler *CollectionHandler) Update(c *gin.Context) {

	var itemHandler model.AlbumHandler
	if err := c.ShouldBindJSON(&itemHandler); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetCollectionManager(c, itemHandler.UserID)
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

func (handler *CollectionHandler) Delete(c *gin.Context) {
	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetCollectionManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	err = collectionManager.Delete(album.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "Delete album with id:"+strconv.Itoa(album.ID))
}

func (handler *CollectionHandler) GetCollectionList(c *gin.Context) {

	collectionManager, err := handler.userStorageManager.GetCollectionManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	items, err := collectionManager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Create collection list without interface constraint
	result := model.PHCollectionList[*model.Album]{
		Collections: make([]*model.PHCollection[*model.Album], len(items)),
	}

	for i, item := range items {
		assets, _ := collectionManager.GetAlbumAssets(item.ID)
		result.Collections[i] = &model.PHCollection[*model.Album]{
			Item:   item,
			Assets: assets,
		}
	}

	c.JSON(http.StatusOK, result)
}

func (handler *CollectionHandler) GetCollectionListWith(c *gin.Context) {

	collectionManager, err := handler.userStorageManager.GetCollectionManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Get only visible albums
	albums, err := collectionManager.GetList(func(a *model.Album) bool {
		return !a.IsHidden
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, albums)
}

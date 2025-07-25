package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"github.com/mahdi-cpp/photocloud_v2/pkg/asset_model"
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

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var request model.CollectionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	newItem, err := userStorage.AlbumManager.Create(&model.Album{Title: request.Title})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	update := asset_model.AssetUpdate{AssetIds: request.AssetIds, AddAlbums: []int{newItem.ID}}
	_, err = userStorage.UpdateAsset(update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userStorage.UpdateCollections()

	c.JSON(http.StatusCreated, model.CollectionResponse{
		ID:    newItem.ID,
		Title: newItem.Title,
	})
}

func (handler *AlbumHandler) Update(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var itemHandler model.AlbumHandler
	if err := c.ShouldBindJSON(&itemHandler); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	item, err := userStorage.AlbumManager.Get(itemHandler.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	model.UpdateAlbum(item, itemHandler)

	item2, err := userStorage.AlbumManager.Update(item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *AlbumHandler) Delete(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var item model.Album
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	err = userStorage.AlbumManager.Delete(item.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

func (handler *AlbumHandler) GetCollectionList(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	item2, err := userStorage.AlbumManager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *AlbumHandler) GetListV2(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	items, err := userStorage.AlbumManager.GetAllSorted("creationDate", "1asc")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	result := model.PHCollectionList[*model.Album]{
		Collections: make([]*model.PHCollection[*model.Album], len(items)),
	}

	for i, item := range items {
		assets, _ := userStorage.AlbumManager.GetItemAssets(item.ID)
		result.Collections[i] = &model.PHCollection[*model.Album]{
			Item:   item,
			Assets: assets,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

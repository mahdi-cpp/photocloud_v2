package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
)

type CollectionHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewCollectionHandler(userStorageManager *storage.UserStorageManager) *CollectionHandler {
	return &CollectionHandler{
		userStorageManager: userStorageManager,
	}
}

func (handler *CollectionHandler) GetCollection(c *gin.Context) {

	albumManager, err := handler.userStorageManager.GetAlbumManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	albums, err := albumManager.GetList(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, albums)
}

func (handler *CollectionHandler) GetCollectionList(c *gin.Context) {

	albumManager, err := handler.userStorageManager.GetAlbumManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	items, err := albumManager.GetAlbumList()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Create collection list without interface constraint
	result := model.PHCollectionList[*model.Album]{
		Collections: make([]*model.PHCollection[*model.Album], len(items)),
	}

	//for i, col := range items {
	//	result.Collections[i] = &model.PHCollection[*model.Album]{
	//		Item: []*model.Album{&col},
	//	}
	//}

	//result := model.PHFetchResult[*model.PHCollectionAlbum]{
	//	Item:  items,
	//	Total:  len(items),
	//	Limit:  100,
	//	Offset: 100,
	//}

	c.JSON(http.StatusOK, result)
}

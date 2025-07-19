package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
	"strconv"
)

type PinnedHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewPinnedHandler(userStorageManager *storage.UserStorageManager) *PinnedHandler {
	return &PinnedHandler{
		userStorageManager: userStorageManager,
	}
}

func (handler *PinnedHandler) Create(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var item model.Pinned
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetPinnedManager(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	item2, err := collectionManager.Create(&item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *PinnedHandler) Update(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var itemHandler model.PinnedHandler
	if err := c.ShouldBindJSON(&itemHandler); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetPinnedManager(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	item, err := collectionManager.Get(itemHandler.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
	model.UpdatePinned(item, itemHandler)

	item2, err := collectionManager.Update(item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *PinnedHandler) Delete(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var item model.Pinned
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetPinnedManager(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	err = collectionManager.Delete(item.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "Delete item with id:"+strconv.Itoa(item.ID))
}

func (handler *PinnedHandler) GetList(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetPinnedManager(c, userID)
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
	result := model.PHCollectionList[*model.Pinned]{
		Collections: make([]*model.PHCollection[*model.Pinned], len(items)),
	}

	for i, item := range items {
		assets, _ := collectionManager.GetItemAssets(item.ID)
		result.Collections[i] = &model.PHCollection[*model.Pinned]{
			Item:   item,
			Assets: assets,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (handler *PinnedHandler) GetCollectionListWith(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetPinnedManager(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Get only visible items
	items, err := collectionManager.GetList(func(a *model.Pinned) bool {
		return a.Icon == ""
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

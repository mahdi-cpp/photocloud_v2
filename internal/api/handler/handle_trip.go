package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
	"strconv"
)

type TripHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewTripHandler(userStorageManager *storage.UserStorageManager) *TripHandler {
	return &TripHandler{
		userStorageManager: userStorageManager,
	}
}

func (handler *TripHandler) Create(c *gin.Context) {
	var item model.Trip
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetTripManager(c, 4)
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

func (handler *TripHandler) Update(c *gin.Context) {

	var itemHandler model.TripHandler
	if err := c.ShouldBindJSON(&itemHandler); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetTripManager(c, itemHandler.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	item, err := collectionManager.Get(itemHandler.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
	model.UpdateTrip(item, itemHandler)

	item2, err := collectionManager.Update(item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *TripHandler) Delete(c *gin.Context) {
	var item model.Trip
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetTripManager(c, 4)
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

func (handler *TripHandler) GetCollectionList(c *gin.Context) {

	collectionManager, err := handler.userStorageManager.GetTripManager(c, 4)
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
	result := model.PHCollectionList[*model.Trip]{
		Collections: make([]*model.PHCollection[*model.Trip], len(items)),
	}

	for i, item := range items {
		assets, _ := collectionManager.GetItemAssets(item.ID)
		result.Collections[i] = &model.PHCollection[*model.Trip]{
			Item:   item,
			Assets: assets,
		}
	}

	c.JSON(http.StatusOK, result)
}

func (handler *TripHandler) GetCollectionListWith(c *gin.Context) {

	collectionManager, err := handler.userStorageManager.GetTripManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Get only visible items
	items, err := collectionManager.GetList(func(a *model.Trip) bool {
		return !a.IsCollection
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, items)
}

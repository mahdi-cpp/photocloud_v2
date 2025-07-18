package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
	"strconv"
)

type PersonHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewPersonsHandler(userStorageManager *storage.UserStorageManager) *PersonHandler {
	return &PersonHandler{
		userStorageManager: userStorageManager,
	}
}

func (handler *PersonHandler) Create(c *gin.Context) {
	var item model.Person
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetPersonManager(c, 4)
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

func (handler *PersonHandler) Update(c *gin.Context) {

	var itemHandler model.PersonHandler
	if err := c.ShouldBindJSON(&itemHandler); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetPersonManager(c, itemHandler.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	item, err := collectionManager.Get(itemHandler.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
	model.UpdatePerson(item, itemHandler)

	item2, err := collectionManager.Update(item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *PersonHandler) Delete(c *gin.Context) {
	var item model.Person
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collectionManager, err := handler.userStorageManager.GetPersonManager(c, 4)
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

func (handler *PersonHandler) GetCollectionList(c *gin.Context) {

	collectionManager, err := handler.userStorageManager.GetPersonManager(c, 4)
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
	result := model.PHCollectionList[*model.Person]{
		Collections: make([]*model.PHCollection[*model.Person], len(items)),
	}

	for i, item := range items {
		assets, _ := collectionManager.GetItemAssets(item.ID)
		result.Collections[i] = &model.PHCollection[*model.Person]{
			Item:   item,
			Assets: assets,
		}
	}

	c.JSON(http.StatusOK, result)
}

func (handler *PersonHandler) GetCollectionListWith(c *gin.Context) {

	collectionManager, err := handler.userStorageManager.GetPersonManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Get only visible items
	items, err := collectionManager.GetList(func(a *model.Person) bool {
		return !a.IsCollection
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, items)
}

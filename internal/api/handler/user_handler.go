package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"
	"net/http"
)

type UserHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewUserHandler(userStorageManager *storage.UserStorageManager) *UserHandler {
	return &UserHandler{
		userStorageManager: userStorageManager,
	}
}

func (handler *UserHandler) Create(c *gin.Context) {

	var item happle_models.User
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager, err := handler.userStorageManager.GetUserManager()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	item2, err := manager.Create(&item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *UserHandler) Update(c *gin.Context) {

	var itemHandler happle_models.UserHandler
	if err := c.ShouldBindJSON(&itemHandler); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	fmt.Println("User Update: ", itemHandler.ID)

	collectionManager, err := handler.userStorageManager.GetUserManager()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	item, err := collectionManager.Get(itemHandler.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error: user not found": err})
		return
	}

	happle_models.UpdateUser(item, itemHandler)

	item2, err := collectionManager.Update(item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *UserHandler) Delete(c *gin.Context) {

	var item happle_models.User
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager, err := handler.userStorageManager.GetUserManager()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	err = manager.Delete(item.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

func (handler *UserHandler) GetCollectionList(c *gin.Context) {

	manager, err := handler.userStorageManager.GetUserManager()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	items, err := manager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, items)
}

func (handler *UserHandler) GetListV2(c *gin.Context) {

	manager, err := handler.userStorageManager.GetUserManager()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	items, err := manager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	result := happle_models.PHCollectionList[*happle_models.User]{
		Collections: make([]*happle_models.PHCollection[*happle_models.User], len(items)),
	}

	for i, item := range items {
		result.Collections[i] = &happle_models.PHCollection[*happle_models.User]{
			Item: item,
		}
	}

	c.JSON(http.StatusOK, result)
}

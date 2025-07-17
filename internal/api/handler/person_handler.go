package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
	"strconv"
)

type PersonHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewPersonHandler(userStorageManager *storage.UserStorageManager) *PersonHandler {
	return &PersonHandler{userStorageManager: userStorageManager}
}

func (handler *PersonHandler) Create(c *gin.Context) {

	var person model.Person
	if err := c.ShouldBindJSON(&person); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager, err := handler.userStorageManager.GetPersonManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	data, err := manager.Create(person.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, data)
}

func (handler *PersonHandler) GetList(c *gin.Context) {

	userID := c.GetInt("userID")
	if userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID: " + strconv.Itoa(userID)})
	}

	manager, err := handler.userStorageManager.GetPersonManager(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	data, err := manager.GetList(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, data)
}

func (handler *PersonHandler) Update(c *gin.Context) {

	var album model.AlbumHandler
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	fmt.Println("Album Update: ", album.ID)

	manager, err := handler.userStorageManager.GetPersonManager(c, album.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	album2, err := manager.Update(album.ID, album.Name, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, album2)
}

func (handler *PersonHandler) Delete(c *gin.Context) {

	var album model.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	manager, err := handler.userStorageManager.GetPersonManager(c, 4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	err = manager.Delete(album.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

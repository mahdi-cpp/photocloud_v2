package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"
	"net/http"
)

type VillageHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewVillageHandler(userStorageManager *storage.UserStorageManager) *VillageHandler {
	return &VillageHandler{
		userStorageManager: userStorageManager,
	}
}

func (handler *VillageHandler) GetList(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	items, err := userStorage.VillageManager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	fmt.Println("villages: ", len(items))

	result := happle_models.PHCollectionList[*model.Village]{
		Collections: make([]*happle_models.PHCollection[*model.Village], len(items)),
	}

	for i, item := range items {
		result.Collections[i] = &happle_models.PHCollection[*model.Village]{
			Item: item,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

package handler

import (
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	userStorageManager *storage.UserStorageManager
}

func NewSearchHandler(userStorageManager *storage.UserStorageManager) *SearchHandler {
	return &SearchHandler{userStorageManager: userStorageManager}
}

func (h *SearchHandler) Filters(c *gin.Context) {

	var filters model.AssetSearchFilters
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		fmt.Println("Invalid request")
		return
	}

	fmt.Println("Filters userId: ", filters.UserID)

	assets, total, err := h.userStorageManager.FilterAssets(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	fmt.Println("Filters count: ", len(assets))

	c.JSON(http.StatusOK, FilterResponse{
		Results: assets,
		Total:   total,
		Limit:   filters.FetchLimit,
		Offset:  filters.FetchOffset,
	})
}

type FilterResponse struct {
	Results []*model.PHAsset `json:"results"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

type AdvancedFilterRequest struct {
	Query       string     `json:"query"`
	MediaType   string     `json:"mediaType"`
	CameraModel string     `json:"cameraModel"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	Favorite    *bool      `json:"favorite"`
	Hidden      *bool      `json:"hidden"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

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

func (h *SearchHandler) AssetFilters(c *gin.Context) {

	var filters model.AssetSearchFilters
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	fmt.Println("AssetFilters userId: ", filters.UserID)

	assets, total, err := h.userStorageManager.FilterAssets(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, SearchResponse{
		Results: assets,
		Total:   total,
		Limit:   filters.Limit,
		Offset:  filters.Offset,
	})
}

// SearchResponse contains search results with pagination info
type SearchResponse struct {
	Results []*model.PHAsset `json:"results"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// AdvancedSearchRequest defines complex search parameters
type AdvancedSearchRequest struct {
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

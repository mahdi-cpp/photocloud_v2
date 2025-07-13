package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"net/http"
)

func (h *SearchHandler) Search(c *gin.Context) {

	var filters model.AssetSearchFilters
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	assets, total, err := h.userStorageManager.FilterAssets(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, FilterResponse{
		Results: assets,
		Total:   total,
		Limit:   filters.FetchLimit,
		Offset:  filters.FetchOffset,
	})
}

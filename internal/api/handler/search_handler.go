package handler

import (
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/service"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// SearchAssets godoc
// @Summary Search assets with filters
// @Description Search across all user's assets with flexible filters
// @Produce json
// @Param query query string false "Search keywords"
// @Param type query string false "Media type (image, video)" Enums(image,video)
// @Param favorite query bool false "Filter favorites"
// @Param hidden query bool false "Filter hidden assets"
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Param camera query string false "Camera model"
// @Param limit query int false "Results limit (default 100)"
// @Param offset query int false "Pagination offset"
// @Success 200 {array} model.PHAsset
// @Router /search [get]
func (h *SearchHandler) SearchAssets(c *gin.Context) {

	userID := c.GetInt("userID")
	query := c.Query("query")

	fmt.Println("query:", query)

	// Parse filters
	filters := model.SearchFilters{
		UserID: userID,
	}

	if mediaType := c.Query("type"); mediaType != "" {
		filters.MediaType = model.MediaType(mediaType)
	}

	if isFavorite := c.Query("isFavorite"); isFavorite != "" {
		log.Println("IsFavorite: ")
		if val, err := strconv.ParseBool(isFavorite); err == nil {
			filters.IsFavorite = &val
		}
	}
	if isScreenshot := c.Query("isScreenshot"); isScreenshot != "" {
		log.Println("isScreenshot: ")
		if val, err := strconv.ParseBool(isScreenshot); err == nil {
			filters.IsScreenshot = &val
		}
	}

	if isHidden := c.Query("isHidden"); isHidden != "" {
		if val, err := strconv.ParseBool(isHidden); err == nil {
			filters.IsHidden = &val
		}
	}

	if camera := c.Query("camera"); camera != "" {
		filters.CameraModel = camera
	}

	// Parse date range
	if start := c.Query("start"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			filters.StartDate = &t
		}
	}
	if end := c.Query("end"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			t = t.Add(24 * time.Hour) // Include entire end day
			filters.EndDate = &t
		}
	}

	// Parse pagination
	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 {
			filters.Limit = val
		}
	} else {
		filters.Limit = 100 // Default limit
	}

	if offset := c.Query("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			filters.Offset = val
		}
	}

	// Execute search
	assets, total, err := h.searchService.SearchAssets(c, filters)
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

// AdvancedSearch godoc
// @Summary Advanced search with multiple criteria
// @Description JSON-based search with complex filters
// @Accept json
// @Produce json
// @Param search body AdvancedSearchRequest true "Search criteria"
// @Success 200 {array} model.PHAsset
// @Router /search/advanced [post]
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	userID := c.GetInt("userID")

	var req AdvancedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	filters := model.SearchFilters{
		UserID:      userID,
		Query:       req.Query,
		MediaType:   model.MediaType(req.MediaType),
		CameraModel: req.CameraModel,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		IsFavorite:  req.Favorite,
		IsHidden:    req.Hidden,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	assets, total, err := h.searchService.SearchAssets(c, filters)
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

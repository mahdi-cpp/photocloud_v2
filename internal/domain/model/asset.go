package model

import (
	"time"
)

const (
	ImageType MediaType = "image"
	VideoType MediaType = "video"
)

type PHAsset struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Url       string    `json:"url"`
	Filename  string    `json:"filename"`
	MediaType MediaType `json:"mediaType"`

	PixelWidth  int `json:"pixelWidth"`
	PixelHeight int `json:"pixelHeight"`

	CameraMake  string    `json:"cameraMake"`
	CameraModel string    `json:"cameraModel"`
	Location    []float64 `json:"location"` // [latitude:longitude]

	IsFavorite   bool `json:"isFavorite"`
	IsScreenshot bool `json:"isScreenshot"`
	IsHidden     bool `json:"isHidden"`

	Albums  []int `json:"albums"`
	Trips   []int `json:"trips"`
	Persons []int `json:"persons"`

	// Video Properties
	Duration float64 `gorm:"default:0" json:"duration"`

	// Content Availability
	CanDelete           bool `gorm:"default:true" json:"canDelete"`
	CanEditContent      bool `gorm:"default:true" json:"canEditContent"`
	CanAddToSharedAlbum bool `gorm:"default:true" json:"CanAddToSharedAlbum"`

	// Advanced Properties
	IsUserLibraryAsset bool `gorm:"default:true" json:"IsUserLibraryAsset"`

	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

// AssetSearchFilters defines search parameters
type AssetSearchFilters struct {
	UserID      int `json:"userID"`
	Query       string
	MediaType   MediaType
	PixelWidth  int
	PixelHeight int

	CameraMake  string
	CameraModel string

	StartDate *time.Time
	EndDate   *time.Time

	IsFavorite   *bool
	IsScreenshot *bool
	IsHidden     *bool
	IsLandscape  *bool

	Albums  []int
	Trips   []int
	Persons []int

	NearPoint    []float64 `json:"nearPoint"`    // [latitude, longitude]
	WithinRadius float64   `json:"withinRadius"` // in kilometers
	BoundingBox  []float64 `json:"boundingBox"`  // [minLat, minLon, maxLat, maxLon]

	SortBy    string `json:"sortBy"`    // Field to sort by (e.g., "creationDate", "filename")
	SortOrder string `json:"sortOrder"` // "asc" or "desc"

	Limit  int
	Offset int
}

type AssetUpdate struct {
	AssetIds    []int   `json:"assetIds,omitempty"` // Asset Ids
	UserID      int     `json:"userID"`
	Filename    *string `json:"filename,omitempty"`
	CameraMake  *string `json:"cameraMake,omitempty"`
	CameraModel *string `json:"cameraModel,omitempty"`

	IsFavorite   *bool `json:"isFavorite,omitempty"`
	IsScreenshot *bool `json:"IsScreenshot,omitempty"`
	IsHidden     *bool `json:"isHidden,omitempty"`

	Albums       *[]int `json:"albums,omitempty"`       // Full album replacement
	AddAlbums    []int  `json:"addAlbums,omitempty"`    // Albums to add
	RemoveAlbums []int  `json:"removeAlbums,omitempty"` // Albums to remove

	Trips       *[]int `json:"trips,omitempty"`       // Full trip replacement
	AddTrips    []int  `json:"AddTrips,omitempty"`    // Trips to add
	RemoveTrips []int  `json:"RemoveTrips,omitempty"` // Trips to remove

	Persons       *[]int `json:"persons,omitempty"`       // Full Person replacement
	AddPersons    []int  `json:"AddPersons,omitempty"`    // Persons to add
	RemovePersons []int  `json:"RemovePersons,omitempty"` // Persons to remove
}

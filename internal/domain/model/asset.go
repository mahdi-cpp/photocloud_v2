package model

import (
	"time"
)

const (
	ImageType MediaType = "image"
	VideoType MediaType = "video"
)

type PHAsset struct {
	ID               int       `json:"id"`
	UserID           int       `json:"userId"`
	Url              string    `json:"url"`
	Filename         string    `json:"filename"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
	MediaType        MediaType `json:"mediaType"`
	PixelWidth       int       `json:"pixelWidth"`
	PixelHeight      int       `json:"pixelHeight"`
	CameraMake       string    `json:"cameraMake"`
	CameraModel      string    `json:"cameraModel"`
	Location         []float64 `json:"location"` // [latitude:longitude]
	IsFavorite       bool      `json:"isFavorite"`
	IsScreenshot     bool      `json:"isScreenshot"`
	IsHidden         bool      `json:"isHidden"`
	Albums           []int     `json:"albums"`
	Trips            []int     `json:"trips"`
	Persons          []int     `json:"persons"`
}

// AssetSearchFilters defines search parameters
type AssetSearchFilters struct {
	UserID      int
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

	Limit  int
	Offset int
}

type AssetUpdate struct {
	Filename     *string `json:"filename,omitempty"`
	IsFavorite   *bool   `json:"isFavorite,omitempty"`
	IsScreenshot *bool   `json:"IsScreenshot,omitempty"`
	IsHidden     *bool   `json:"isHidden,omitempty"`
}

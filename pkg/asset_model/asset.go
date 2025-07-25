package asset_model

import (
	"time"
)

const (
	ImageType MediaType = "image"
	VideoType MediaType = "video"
	SongType  MediaType = "song"
)

type PHAsset struct {
	ID          int       `json:"id"`
	UserID      int       `json:"userId"`
	Url         string    `json:"url"`
	Filename    string    `json:"filename"`
	Format      string    `json:"format"`
	MediaType   MediaType `json:"mediaType"`
	Orientation int       `json:"orientation"`

	PixelWidth  int `json:"pixelWidth"`
	PixelHeight int `json:"pixelHeight"`

	CameraMake  string    `json:"cameraMake"`
	CameraModel string    `json:"cameraModel"`
	Location    []float64 `json:"location"` // [latitude:longitude]

	IsCamera     bool `json:"isCamera"`
	IsFavorite   bool `json:"isFavorite"`
	IsScreenshot bool `json:"isScreenshot"`
	IsHidden     bool `json:"isHidden"`

	Albums  []int `json:"albums"`
	Trips   []int `json:"trips"`
	Persons []int `json:"persons"`

	// Video Properties
	Duration float64 `json:"duration"`

	// Content Availability
	CanDelete           bool `json:"canDelete"`
	CanEditContent      bool `json:"canEditContent"`
	CanAddToSharedAlbum bool `json:"canAddToSharedAlbum"`

	// Advanced Properties
	IsUserLibraryAsset bool `json:"IsUserLibraryAsset"`

	CapturedDate     time.Time `json:"capturedDate"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type PHFetchOptions struct {
	UserID      int `json:"userID"`
	Query       string
	MediaType   MediaType
	PixelWidth  int
	PixelHeight int

	CameraMake  string
	CameraModel string

	StartDate *time.Time
	EndDate   *time.Time

	IsCamera      *bool
	IsFavorite    *bool
	IsScreenshot  *bool
	IsHidden      *bool
	IsLandscape   *bool
	NotInOneAlbum *bool

	HideScreenshot *bool `json:"hideScreenshot"`

	Albums  []int
	Trips   []int
	Persons []int

	NearPoint    []float64 `json:"nearPoint"`    // [latitude, longitude]
	WithinRadius float64   `json:"withinRadius"` // in kilometers
	BoundingBox  []float64 `json:"boundingBox"`  // [minLat, minLon, maxLat, maxLon]

	SortBy    string `json:"sortBy"`    // Field to sort by (e.g., "creationDate", "filename")
	SortOrder string `json:"sortOrder"` // "asc" or "desc"

	FetchOffset int `json:"fetchOffset"`
	FetchLimit  int `json:"fetchLimit"`
}

type AssetUpdate struct {
	AssetIds []int `json:"assetIds,omitempty"` // Asset Ids

	Filename  *string   `json:"filename,omitempty"`
	MediaType MediaType `json:"mediaType,omitempty"`

	CameraMake  *string `json:"cameraMake,omitempty"`
	CameraModel *string `json:"cameraModel,omitempty"`

	IsCamera     *bool `json:"isCamera,omitempty"`
	IsFavorite   *bool `json:"isFavorite,omitempty"`
	IsScreenshot *bool `json:"IsScreenshot,omitempty"`
	IsHidden     *bool `json:"isHidden,omitempty"`

	Albums       *[]int `json:"albums,omitempty"`       // Full album replacement
	AddAlbums    []int  `json:"addAlbums,omitempty"`    // Albums to add
	RemoveAlbums []int  `json:"removeAlbums,omitempty"` // Albums to remove

	Trips       *[]int `json:"trips,omitempty"`       // Full trip replacement
	AddTrips    []int  `json:"addTrips,omitempty"`    // Trips to add
	RemoveTrips []int  `json:"removeTrips,omitempty"` // Trips to remove

	Persons       *[]int `json:"persons,omitempty"`       // Full Person replacement
	AddPersons    []int  `json:"addPersons,omitempty"`    // Persons to add
	RemovePersons []int  `json:"removePersons,omitempty"` // Persons to remove
}

type AssetDelete struct {
	AssetID int `json:"assetID"`
}

// https://chat.deepseek.com/a/chat/s/9b010f32-b23d-4f9b-ae0c-31a9b2c9408c

type PHFetchResult[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

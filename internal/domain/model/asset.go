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
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	Camera           string    `json:"camera"`
	IsFavorite       bool      `json:"isFavorite"`
	IsScreenshot     bool      `json:"isScreenshot"`
	IsHidden         bool      `json:"isHidden"`
}

// AssetSearchFilters defines search parameters
type AssetSearchFilters struct {
	UserID       int
	Query        string
	MediaType    MediaType
	CameraModel  string
	StartDate    *time.Time
	EndDate      *time.Time
	IsFavorite   *bool
	IsScreenshot *bool
	IsHidden     *bool
	IsLandscape  *bool
	Limit        int
	Offset       int
}

type AssetUpdate struct {
	Filename     *string `json:"filename,omitempty"`
	IsFavorite   *bool   `json:"isFavorite,omitempty"`
	IsScreenshot *bool   `json:"IsScreenshot,omitempty"`
	IsHidden     *bool   `json:"isHidden,omitempty"`
}

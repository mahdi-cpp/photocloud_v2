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
	Filename         string    `json:"filename"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
	MediaType        MediaType `json:"mediaType"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	Camera           string    `json:"camera"`
	IsFavorite       bool      `json:"isFavorite"`
	IsHidden         bool      `json:"isHidden"`
}

type AssetUpdate struct {
	Filename   *string `json:"filename,omitempty"`
	IsFavorite *bool   `json:"isFavorite,omitempty"`
	IsHidden   *bool   `json:"isHidden,omitempty"`
}

// SearchFilters defines search parameters
type SearchFilters struct {
	UserID      int
	Query       string
	MediaType   MediaType
	CameraModel string
	StartDate   *time.Time
	EndDate     *time.Time
	IsFavorite  *bool
	IsHidden    *bool
	Limit       int
	Offset      int
}

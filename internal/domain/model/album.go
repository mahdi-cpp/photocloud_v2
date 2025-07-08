package model

import "time"

type Album struct {
	ID               int       `json:"id"`
	Name             string    `json:"url"`
	IsCollection     bool      `json:"isFavorite"`
	IsScreenshot     bool      `json:"isScreenshot"`
	IsHidden         bool      `json:"isHidden"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type AlbumUpdate struct {
	Name         *string `json:"name,omitempty"`
	IsCollection *bool   `json:"IsCollection,omitempty"`
	IsHidden     *bool   `json:"isHidden,omitempty"`
}

package model

import (
	"time"
)

func (a *Album) SetID(id int)                    { a.ID = id }
func (a *Album) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Album) SetModificationDate(t time.Time) { a.ModificationDate = t }
func (a *Album) GetID() int                      { return a.ID }
func (a *Album) GetCreationDate() time.Time      { return a.CreationDate }
func (a *Album) GetModificationDate() time.Time  { return a.ModificationDate }

type Album struct {
	ID               int       `json:"id"`
	Title            string    `json:"title"`
	Subtitle         string    `json:"subtitle"`
	AlbumType        string    `json:"albumType"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	IsHidden         bool      `json:"isHidden"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type AlbumHandler struct {
	ID           int    `json:"id"`
	Title        string `json:"title,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	AlbumType    string `json:"albumType,omitempty"`
	IsCollection *bool  `json:"isCollection,omitempty"`
	IsHidden     *bool  `json:"isHidden,omitempty"`
}

func UpdateAlbum(item *Album, handler AlbumHandler) *Album {

	if handler.Title != "" {
		item.Title = handler.Title
	}
	if handler.Subtitle != "" {
		item.Subtitle = handler.Subtitle
	}

	if handler.AlbumType != "" {
		item.AlbumType = handler.AlbumType
	}

	if handler.IsCollection != nil {
		item.IsCollection = *handler.IsCollection
	}

	if handler.IsHidden != nil {
		item.IsHidden = *handler.IsHidden
	}

	return item
}

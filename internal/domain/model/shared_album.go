package model

import "time"

func (a *SharedAlbum) GetID() int                      { return a.ID }
func (a *SharedAlbum) SetID(id int)                    { a.ID = id }
func (a *SharedAlbum) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *SharedAlbum) SetModificationDate(t time.Time) { a.ModificationDate = t }

type SharedAlbum struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	AlbumType        string    `json:"albumType"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	IsHidden         bool      `json:"isHidden"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type SharedAlbumHandler struct {
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	AlbumType    string `json:"albumType,omitempty"`
	IsCollection *bool  `json:"isCollection,omitempty"`
	IsHidden     *bool  `json:"isHidden,omitempty"`
}

func UpdateSharedAlbum(album *SharedAlbum, handler SharedAlbumHandler) *SharedAlbum {

	if handler.Name != "" {
		album.Name = handler.Name
	}

	if handler.AlbumType != "" {
		album.AlbumType = handler.AlbumType
	}

	if handler.IsCollection != nil {
		album.IsCollection = *handler.IsCollection
	}

	if handler.IsHidden != nil {
		album.IsHidden = *handler.IsHidden
	}

	return album
}

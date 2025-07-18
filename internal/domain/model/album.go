package model

import "time"

func (a *Album) GetID() int                      { return a.ID }
func (a *Album) SetID(id int)                    { a.ID = id }
func (a *Album) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Album) SetModificationDate(t time.Time) { a.ModificationDate = t }

type Album struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	AlbumType        string    `json:"albumType,omitempty"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	IsHidden         bool      `json:"isHidden"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type AlbumHandler struct {
	UserID       int    `json:"userID"`
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	AlbumType    string `json:"albumType,omitempty"`
	IsCollection *bool  `json:"isCollection,omitempty"`
	IsHidden     *bool  `json:"isHidden,omitempty"`
}

func UpdateAlbum(album *Album, handler AlbumHandler) *Album {

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

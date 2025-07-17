package model

import "time"

//type AlbumCollection struct {
//	Albums []Album `json:"albums"`
//}

//type PHCollectionListAlbum struct {
//	Albums []PHCollectionAlbum `json:"albums"`
//}

//type PHCollectionAlbum struct {
//	Album  *Album     `json:"album"`
//	Assets []*PHAsset `json:"assets"`
//}

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
	IsCollection bool   `json:"IsCollection,omitempty"`
	IsHidden     bool   `json:"isHidden,omitempty"`
}

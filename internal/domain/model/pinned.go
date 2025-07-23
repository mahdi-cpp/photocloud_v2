package model

import "time"

func (a *Pinned) GetID() int                      { return a.ID }
func (a *Pinned) SetID(id int)                    { a.ID = id }
func (a *Pinned) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Pinned) SetModificationDate(t time.Time) { a.ModificationDate = t }
func (a *Pinned) GetCreationDate() time.Time      { return a.CreationDate }
func (a *Pinned) GetModificationDate() time.Time  { return a.ModificationDate }

type Pinned struct {
	ID               int       `json:"id"`
	Title            string    `json:"title"`
	Subtitle         string    `json:"subtitle"`
	Type             string    `json:"type"`
	AlbumID          int       `json:"albumID"`
	Icon             string    `json:"icon"`
	Count            int       `json:"count"`
	Index            int       `json:"index"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type PinnedHandler struct {
	ID           int    `json:"id"`
	Title        string `json:"title,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	Icon         string `json:"icon,omitempty"`
	IsCollection *bool  `json:"isCollection,omitempty"`
	IsHidden     *bool  `json:"isHidden,omitempty"`
}

func UpdatePinned(item *Pinned, handler PinnedHandler) *Pinned {

	if handler.Title != "" {
		item.Title = handler.Title
	}
	if handler.Subtitle != "" {
		item.Subtitle = handler.Subtitle
	}

	if handler.Icon != "" {
		item.Icon = handler.Icon
	}

	return item
}

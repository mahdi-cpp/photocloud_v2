package model

import "time"

func (a *Pinned) GetID() int                      { return a.ID }
func (a *Pinned) SetID(id int)                    { a.ID = id }
func (a *Pinned) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Pinned) SetModificationDate(t time.Time) { a.ModificationDate = t }

type Pinned struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type PinnedHandler struct {
	UserID       int    `json:"userID"`
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	IsCollection *bool  `json:"isCollection,omitempty"`
	IsHidden     *bool  `json:"isHidden,omitempty"`
}

func UpdatePinned(pinned *Pinned, handler PinnedHandler) *Pinned {

	if handler.Name != "" {
		pinned.Name = handler.Name
	}

	if handler.IsCollection != nil {
		pinned.IsCollection = *handler.IsCollection
	}

	return pinned
}

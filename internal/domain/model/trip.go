package model

import "time"

func (a *Trip) GetID() int                      { return a.ID }
func (a *Trip) SetID(id int)                    { a.ID = id }
func (a *Trip) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Trip) SetModificationDate(t time.Time) { a.ModificationDate = t }
func (a *Trip) GetCreationDate() time.Time      { return a.CreationDate }
func (a *Trip) GetModificationDate() time.Time  { return a.ModificationDate }

type Trip struct {
	ID               int       `json:"id"`
	Title            string    `json:"title,omitempty"`
	Subtitle         string    `json:"subtitle,omitempty"`
	TripType         string    `json:"tripType,omitempty"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type TripHandler struct {
	ID           int    `json:"id"`
	Title        string `json:"title,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	TripType     string `json:"trip,omitempty"`
	IsCollection *bool  `json:"isCollection,omitempty"`
}

func UpdateTrip(item *Trip, handler TripHandler) *Trip {

	if handler.Title != "" {
		item.Title = handler.Title
	}
	if handler.Subtitle != "" {
		item.Subtitle = handler.Subtitle
	}

	if handler.TripType != "" {
		item.TripType = handler.TripType
	}

	if handler.IsCollection != nil {
		item.IsCollection = *handler.IsCollection
	}

	return item
}

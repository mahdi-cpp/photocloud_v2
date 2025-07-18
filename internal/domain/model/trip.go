package model

import "time"

type TripCollection struct {
	Trips []Trip `json:"trips,omitempty"`
}

func (a *Trip) GetID() int                      { return a.ID }
func (a *Trip) SetID(id int)                    { a.ID = id }
func (a *Trip) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Trip) SetModificationDate(t time.Time) { a.ModificationDate = t }

type Trip struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	TripType         string    `json:"tripType,omitempty"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type TripHandler struct {
	UserID       int    `json:"userID"`
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	IsCollection *bool  `json:"IsCollection,omitempty"`
	IsHidden     *bool  `json:"isHidden,omitempty"`
}

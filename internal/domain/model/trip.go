package model

import "time"

type TripCollection struct {
	Trips []Trip `json:"trips,omitempty"`
}

type Trip struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	TripType         string    `json:"tripType,omitempty"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	IsHidden         bool      `json:"isHidden"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type TripUpdate struct {
	Name         *string `json:"name,omitempty"`
	IsCollection *bool   `json:"IsCollection,omitempty"`
	IsHidden     *bool   `json:"isHidden,omitempty"`
}

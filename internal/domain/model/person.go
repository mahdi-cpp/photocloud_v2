package model

import "time"

type PersonCollection struct {
	Persons []Person `json:"persons,omitempty"`
}

type Person struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type PersonHandler struct {
	UserID       int     `json:"userID"`
	ID           int     `json:"id"`
	Name         *string `json:"name,omitempty"`
	IsCollection *bool   `json:"IsCollection,omitempty"`
	IsHidden     *bool   `json:"isHidden,omitempty"`
}

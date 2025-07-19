package model

import "time"

func (a *Person) GetID() int                      { return a.ID }
func (a *Person) SetID(id int)                    { a.ID = id }
func (a *Person) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Person) SetModificationDate(t time.Time) { a.ModificationDate = t }

type Person struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type PersonHandler struct {
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	IsCollection *bool  `json:"IsCollection,omitempty"`
	IsHidden     *bool  `json:"isHidden,omitempty"`
}

func UpdatePerson(person *Person, handler PersonHandler) *Person {

	if handler.Name != "" {
		person.Name = handler.Name
	}

	if handler.IsCollection != nil {
		person.IsCollection = *handler.IsCollection
	}
	return person
}

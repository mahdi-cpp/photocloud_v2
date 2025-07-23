package model

import "time"

func (a *Person) GetID() int                      { return a.ID }
func (a *Person) SetID(id int)                    { a.ID = id }
func (a *Person) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Person) SetModificationDate(t time.Time) { a.ModificationDate = t }
func (a *Person) GetCreationDate() time.Time      { return a.CreationDate }
func (a *Person) GetModificationDate() time.Time  { return a.ModificationDate }

type Person struct {
	ID               int       `json:"id"`
	Title            string    `json:"title,omitempty"`
	Subtitle         string    `json:"subtitle,omitempty"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type PersonHandler struct {
	ID           int    `json:"id"`
	Title        string `json:"title,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	IsCollection *bool  `json:"IsCollection,omitempty"`
	IsHidden     *bool  `json:"isHidden,omitempty"`
}

func UpdatePerson(item *Person, handler PersonHandler) *Person {

	if handler.Title != "" {
		item.Title = handler.Title
	}
	if handler.Subtitle != "" {
		item.Subtitle = handler.Subtitle
	}

	if handler.IsCollection != nil {
		item.IsCollection = *handler.IsCollection
	}
	return item
}

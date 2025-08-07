package model

import "time"

func (a *Village) GetID() int                      { return a.ID }
func (a *Village) SetID(id int)                    { a.ID = id }
func (a *Village) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Village) SetModificationDate(t time.Time) { a.ModificationDate = t }
func (a *Village) GetCreationDate() time.Time      { return a.CreationDate }
func (a *Village) GetModificationDate() time.Time  { return a.ModificationDate }

type Village struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type Polygon struct {
	Name string        `json:"name"`
	Data [][][]float64 `json:"data"`
}

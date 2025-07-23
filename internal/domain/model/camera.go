package model

import "time"

func (a *Camera) GetID() int                      { return a.ID }
func (a *Camera) SetID(id int)                    { a.ID = id }
func (a *Camera) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Camera) SetModificationDate(t time.Time) { a.ModificationDate = t }
func (a *Camera) GetCreationDate() time.Time      { return a.CreationDate }
func (a *Camera) GetModificationDate() time.Time  { return a.ModificationDate }

type Camera struct {
	ID               int       `json:"id"`
	CameraMake       string    `json:"cameraMake"`
	CameraModel      string    `json:"cameraModel"`
	Count            int       `json:"count"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type CameraHandler struct {
	ID          int    `json:"id"`
	CameraMake  string `json:"cameraMake,omitempty"`
	CameraModel string `json:"cameraModel,omitempty"`
}

func UpdateCamera(album *Camera, handler CameraHandler) *Camera {

	if handler.CameraMake != "" {
		album.CameraMake = handler.CameraMake
	}

	if handler.CameraModel != "" {
		album.CameraModel = handler.CameraModel
	}

	return album
}

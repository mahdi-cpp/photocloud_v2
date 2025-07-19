package model

import "time"

func (c *Camera) GetID() int                      { return c.ID }
func (c *Camera) SetID(id int)                    { c.ID = id }
func (c *Camera) SetCreationDate(t time.Time)     { c.CreationDate = t }
func (c *Camera) SetModificationDate(t time.Time) { c.ModificationDate = t }

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

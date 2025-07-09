package storage

import (
	"errors"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"time"
)

type AlbumManager struct {
	manager *MetadataControl[model.AlbumCollection]
}

func NewAlbumManager(path string) *AlbumManager {
	return &AlbumManager{
		manager: NewMetadataManagerV2[model.AlbumCollection](path),
	}
}

// CreateAlbum adds a new album with auto-generated ID
func (am *AlbumManager) CreateAlbum(name string, albumType string, isCollection bool) (*model.Album, error) {
	var newAlbum *model.Album

	err := am.manager.Update(func(albums *model.AlbumCollection) error {
		// Generate ID
		maxID := 0
		for _, album := range albums.Albums {
			if album.ID > maxID {
				maxID = album.ID
			}
		}

		// Create new album
		newAlbum = &model.Album{
			ID:               maxID + 1,
			Name:             name,
			AlbumType:        albumType,
			IsCollection:     isCollection,
			IsHidden:         false,
			CreationDate:     time.Now(),
			ModificationDate: time.Now(),
		}

		// Add to collection
		albums.Albums = append(albums.Albums, *newAlbum)
		return nil
	})

	return newAlbum, err
}

// UpdateAlbum modifies an existing album
func (am *AlbumManager) UpdateAlbum(id int, name string, albumType string, isHidden bool) (*model.Album, error) {
	var updatedAlbum *model.Album

	err := am.manager.Update(func(albums *model.AlbumCollection) error {
		for i, album := range albums.Albums {
			if album.ID == id {
				// Update fields
				albums.Albums[i].Name = name
				albums.Albums[i].AlbumType = albumType
				albums.Albums[i].IsHidden = isHidden
				albums.Albums[i].ModificationDate = time.Now()

				updatedAlbum = &albums.Albums[i]
				return nil
			}
		}
		return errors.New("album not found")
	})

	return updatedAlbum, err
}

// DeleteAlbum removes an album by ID
func (am *AlbumManager) DeleteAlbum(id int) error {
	return am.manager.Update(func(albums *model.AlbumCollection) error {
		for i, album := range albums.Albums {
			if album.ID == id {
				// Remove album from slice
				albums.Albums = append(albums.Albums[:i], albums.Albums[i+1:]...)
				return nil
			}
		}
		return errors.New("album not found")
	})
}

// GetAlbum retrieves an album by ID
func (am *AlbumManager) GetAlbum(id int) (*model.Album, error) {
	albums, err := am.manager.Read()
	if err != nil {
		return nil, err
	}

	for _, album := range albums.Albums {
		if album.ID == id {
			return &album, nil
		}
	}
	return nil, errors.New("album not found")
}

// List returns all albums with optional filters
func (am *AlbumManager) List(includeHidden bool) ([]model.Album, error) {
	albums, err := am.manager.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Album
	for _, album := range albums.Albums {
		if !album.IsHidden || includeHidden {
			result = append(result, album)
		}
	}
	return result, nil
}

// GetAlbumsByType returns albums of a specific type
func (am *AlbumManager) GetAlbumsByType(albumType string) ([]model.Album, error) {
	albums, err := am.manager.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Album
	for _, album := range albums.Albums {
		if album.AlbumType != "" && album.AlbumType == albumType {
			result = append(result, album)
		}
	}
	return result, nil
}

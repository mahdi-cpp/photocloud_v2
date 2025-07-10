package storage

import (
	"errors"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/registery"
	"strconv"
	"time"
)

type AlbumManager struct {
	albumRegistry *registery.Registry[model.Album]
	metadata      *MetadataControl[model.AlbumCollection]
}

func NewAlbumManager(path string) (*AlbumManager, error) {

	manager := &AlbumManager{
		albumRegistry: registery.NewRegistry[model.Album](),
		metadata:      NewMetadataManagerV2[model.AlbumCollection](path),
	}

	albums, err := manager.load()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Albums: %w", err)
	}

	for _, album := range albums {
		manager.albumRegistry.Register(strconv.Itoa(album.ID), album)
	}

	return manager, nil
}

func (manager *AlbumManager) load() ([]model.Album, error) {
	albumCollection, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Album
	for _, album := range albumCollection.Albums {
		result = append(result, album)
	}

	return result, nil
}

// Create adds a new album with auto-generated ID
func (manager *AlbumManager) Create(name string, albumType string, isCollection bool) (*model.Album, error) {
	var newAlbum *model.Album

	err := manager.metadata.Update(func(albums *model.AlbumCollection) error {

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

		manager.albumRegistry.Register(strconv.Itoa(newAlbum.ID), *newAlbum)

		// Add to collection
		albums.Albums = append(albums.Albums, *newAlbum)
		return nil
	})

	return newAlbum, err
}

// Update modifies an existing album
func (manager *AlbumManager) Update(id int, name string) (*model.Album, error) {
	var updatedAlbum *model.Album

	err := manager.metadata.Update(func(albums *model.AlbumCollection) error {
		for i, album := range albums.Albums {
			if album.ID == id {
				// Update fields
				albums.Albums[i].Name = name
				//albums.Albums[i].AlbumType = albumType
				//albums.Albums[i].IsHidden = isHidden
				albums.Albums[i].ModificationDate = time.Now()

				updatedAlbum = &albums.Albums[i]
				manager.albumRegistry.Update(getKey(updatedAlbum.ID), *updatedAlbum)
				return nil
			}
		}
		return errors.New("album not found")
	})

	return updatedAlbum, err
}

// Delete removes an album by ID
func (manager *AlbumManager) Delete(id int) error {
	return manager.metadata.Update(func(albums *model.AlbumCollection) error {
		for i, album := range albums.Albums {
			if album.ID == id {
				// Remove album from slice
				albums.Albums = append(albums.Albums[:i], albums.Albums[i+1:]...)
				manager.albumRegistry.Delete(getKey(album.ID))
				return nil
			}
		}
		return errors.New("album not found")
	})
}

// Get retrieves an album by ID
func (manager *AlbumManager) Get(id int) (*model.Album, error) {
	album, err := manager.albumRegistry.Get(getKey(id))
	if err != nil {
		return nil, errors.New("album not found")
	}
	return &album, nil
}

// List returns all albums with optional filters
func (manager *AlbumManager) List(includeHidden bool) ([]model.Album, error) {

	albums := manager.albumRegistry.GetAllValues()
	var result []model.Album
	for _, album := range albums {
		if !album.IsHidden || includeHidden {
			result = append(result, album)
		}
	}

	return result, nil
}

// GetByType returns albums of a specific type
func (manager *AlbumManager) GetByType(albumType string) ([]model.Album, error) {

	albums, err := manager.metadata.Read()
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

func getKey(id int) string {
	return strconv.Itoa(id)
}

package storage

import (
	"errors"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/registery"
	"sort"
	"strconv"
	"time"
)

type AlbumManager struct {
	parent     *UserStorage
	metadata   *MetadataControl[model.PHCollectionList[*model.Album]]
	items      *registery.Registry[model.Album]
	itemAssets map[int][]*model.PHAsset
}

func NewAlbumManager(parent *UserStorage, path string) (*AlbumManager, error) {

	manager := &AlbumManager{
		parent:     parent,
		items:      registery.NewRegistry[model.Album](),
		metadata:   NewMetadataControl[model.PHCollectionList[*model.Album]](path),
		itemAssets: make(map[int][]*model.PHAsset),
	}

	albums, err := manager.load()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Album: %w", err)
	}

	for _, album := range albums {
		fmt.Println(album.Name)
		manager.items.Register(strconv.Itoa(album.ID), album)
	}

	return manager, nil
}

func (manager *AlbumManager) load() ([]model.Album, error) {
	collectionList, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Album
	for _, collection := range collectionList.Collections {
		result = append(result, *collection.Item)
	}

	return result, nil
}

func (manager *AlbumManager) Create(newItem *model.Album) (*model.Album, error) {
	//var newAlbum *model.Album

	err := manager.metadata.Update(func(collectionList *model.PHCollectionList[*model.Album]) error {

		// Generate ID
		maxID := 0
		for _, collection := range collectionList.Collections {
			if collection.Item.ID > maxID {
				maxID = collection.Item.ID
			}
		}

		newItem.ID = maxID + 1
		newItem.CreationDate = time.Now()
		newItem.ModificationDate = time.Now()

		manager.items.Register(strconv.Itoa(newItem.ID), *newItem)

		// Add to collection
		collectionList.Collections = append(collectionList.Collections,
			&model.PHCollection[*model.Album]{
				Item:   newItem,
				Assets: nil,
			})

		return nil
	})

	if err == nil {
		manager.parent.prepareAlbums()
	}

	return newItem, err
}

func (manager *AlbumManager) Update(itemUpdate *model.Album) (*model.Album, error) {

	var updatedAlbum *model.Album

	err := manager.metadata.Update(func(collectionList *model.PHCollectionList[*model.Album]) error {

		for i, collection := range collectionList.Collections {
			if collection.Item.ID == itemUpdate.ID {

				collectionList.Collections[i].Item = itemUpdate
				collectionList.Collections[i].Item.ModificationDate = time.Now()

				updatedAlbum = collectionList.Collections[i].Item
				manager.items.Update(getKey(updatedAlbum.ID), *updatedAlbum)
				return nil
			}
		}

		return errors.New("collection not found")
	})

	return updatedAlbum, err
}

func (manager *AlbumManager) Delete(id int) error {
	return manager.metadata.Update(func(collectionList *model.PHCollectionList[*model.Album]) error {
		for i, collection := range collectionList.Collections {
			if collection.Item.ID == id {
				// Remove collection from slice
				collectionList.Collections = append(collectionList.Collections[:i], collectionList.Collections[i+1:]...)
				manager.items.Delete(getKey(collection.Item.ID))
				return nil
			}
		}
		return errors.New("collection not found")
	})
}

func (manager *AlbumManager) Get(id int) (*model.Album, error) {
	album, err := manager.items.Get(getKey(id))
	if err != nil {
		return nil, errors.New("album not found")
	}
	return &album, nil
}

func (manager *AlbumManager) GetList(includeHidden bool) ([]*model.Album, error) {

	albums := manager.items.GetAllValues()
	var result []*model.Album
	for _, album := range albums {
		if !album.IsHidden || includeHidden {
			result = append(result, &album)
		}
	}
	SortAlbums(result, "creationDate", "asc")
	return result, nil
}

func (manager *AlbumManager) GetAlbumList() ([]model.Album, error) {
	return manager.items.GetAllValues(), nil
}

func (manager *AlbumManager) GetByType(albumType string) ([]model.Album, error) {

	collectionList, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Album
	for _, collection := range collectionList.Collections {
		if collection.Item.AlbumType != "" && collection.Item.AlbumType == albumType {
			result = append(result, *collection.Item)
		}
	}
	return result, nil
}

func (manager *AlbumManager) GetAlbumAssets(albumID int) ([]*model.PHAsset, error) {
	return manager.itemAssets[albumID], nil
}

func getKey(id int) string {
	return strconv.Itoa(id)
}

func SortAlbums(assets []*model.Album, sortBy, sortOrder string) {

	if sortBy == "" {
		return // No sorting requested
	}

	sort.Slice(assets, func(i, j int) bool {
		a := assets[i]
		b := assets[j]

		switch sortBy {
		case "id":
			if sortOrder == "asc" {
				return a.ID < b.ID
			}
			return a.ID > b.ID
		case "creationDate":
			if sortOrder == "asc" {
				return a.CreationDate.Before(b.CreationDate)
			}
			return a.CreationDate.After(b.CreationDate)
		case "modificationDate":
			if sortOrder == "asc" {
				return a.ModificationDate.Before(b.ModificationDate)
			}
			return a.ModificationDate.After(b.ModificationDate)

		default:
			return false // No sorting for unknown fields
		}
	})
}

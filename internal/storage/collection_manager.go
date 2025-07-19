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

// CollectionItem defines the required interface for managed items
type CollectionItem interface {
	GetID() int
	SetID(int)
	SetCreationDate(time.Time)
	SetModificationDate(time.Time)
}

// CollectionManager manages any type of collection items
type CollectionManager[T CollectionItem] struct {
	metadata   *MetadataControl[[]T]
	items      *registery.Registry[T]
	itemAssets map[int][]*model.PHAsset
}

func NewCollectionManager[T CollectionItem](path string) (*CollectionManager[T], error) {

	manager := &CollectionManager[T]{
		items:      registery.NewRegistry[T](),
		metadata:   NewMetadataControl[[]T](path),
		itemAssets: make(map[int][]*model.PHAsset),
	}

	items, err := manager.load()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize collection manager: %w", err)
	}

	for _, item := range items {
		manager.items.Register(strconv.Itoa(item.GetID()), item)
	}

	return manager, nil
}

func (manager *CollectionManager[T]) load() ([]T, error) {
	dataPtr, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	if dataPtr == nil {
		return []T{}, nil
	}

	// Dereference the pointer to get the actual slice
	return *dataPtr, nil
}

func (manager *CollectionManager[T]) Create(newItem T) (T, error) {

	err := manager.metadata.Update(func(items *[]T) error {
		// Generate ID
		maxID := 0
		for _, item := range *items {
			if item.GetID() > maxID {
				maxID = item.GetID()
			}
		}

		newItem.SetID(maxID + 1)
		newItem.SetCreationDate(time.Now())
		newItem.SetModificationDate(time.Now())

		// Add to collection
		*items = append(*items, newItem)
		manager.items.Register(strconv.Itoa(newItem.GetID()), newItem)

		return nil
	})

	return newItem, err
}

func (manager *CollectionManager[T]) Update(updatedItem T) (T, error) {
	err := manager.metadata.Update(func(items *[]T) error {
		for i, item := range *items {
			if item.GetID() == updatedItem.GetID() {
				updatedItem.SetModificationDate(time.Now())
				(*items)[i] = updatedItem
				manager.items.Update(strconv.Itoa(updatedItem.GetID()), updatedItem)
				return nil
			}
		}
		return errors.New("item not found")
	})

	return updatedItem, err
}

func (manager *CollectionManager[T]) Delete(id int) error {
	return manager.metadata.Update(func(items *[]T) error {
		for i, item := range *items {
			if item.GetID() == id {
				// Remove item from slice
				*items = append((*items)[:i], (*items)[i+1:]...)
				manager.items.Delete(strconv.Itoa(id))
				return nil
			}
		}
		return errors.New("item not found")
	})
}

func (manager *CollectionManager[T]) Get(id int) (T, error) {
	item, err := manager.items.Get(strconv.Itoa(id))
	if err != nil {
		var zero T
		return zero, errors.New("item not found")
	}
	return item, nil
}

func (manager *CollectionManager[T]) GetList(filterFunc func(T) bool) ([]T, error) {
	allItems := manager.items.GetAllValues()
	var result []T
	for _, item := range allItems {
		if filterFunc == nil || filterFunc(item) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (manager *CollectionManager[T]) GetAll() ([]T, error) {
	return manager.items.GetAllValues(), nil
}

func (manager *CollectionManager[T]) GetBy(filterFunc func(T) bool) ([]T, error) {
	allItems := manager.items.GetAllValues()
	var result []T
	for _, item := range allItems {
		if filterFunc(item) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (manager *CollectionManager[T]) GetItemAssets(id int) ([]*model.PHAsset, error) {
	return manager.itemAssets[id], nil
}

// SortItems sorts items using a custom comparison function
func SortItems[T any](items []T, less func(i, j int) bool) {
	sort.Slice(items, less)
}

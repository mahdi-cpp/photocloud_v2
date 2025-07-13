package storage

import (
	"errors"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/registery"
	"strconv"
	"time"
)

type PinnedManager struct {
	registry *registery.Registry[model.Pinned]
	metadata *MetadataControl[model.PinnedCollection]
}

func NewPinnedManager(path string) (*PinnedManager, error) {

	manager := &PinnedManager{
		registry: registery.NewRegistry[model.Pinned](),
		metadata: NewMetadataControl[model.PinnedCollection](path),
	}

	pins, err := manager.load()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Albums: %w", err)
	}

	for _, pinned := range pins {
		manager.registry.Register(strconv.Itoa(pinned.ID), pinned)
	}
	return manager, nil
}

func (manager *PinnedManager) Create(name string) (*model.Pinned, error) {
	var newPinned *model.Pinned

	err := manager.metadata.Update(func(collection *model.PinnedCollection) error {
		// Generate ID
		maxID := 0
		for _, pin := range collection.Pins {
			if pin.ID > maxID {
				maxID = pin.ID
			}
		}

		// Handler new pin
		newPinned = &model.Pinned{
			ID:               maxID + 1,
			Name:             name,
			CreationDate:     time.Now(),
			ModificationDate: time.Now(),
		}

		// Add to collection
		collection.Pins = append(collection.Pins, *newPinned)
		return nil
	})

	return newPinned, err
}

func (manager *PinnedManager) Update(id int, name string) (*model.Pinned, error) {
	var updated *model.Pinned

	err := manager.metadata.Update(func(collection *model.PinnedCollection) error {
		for i, pin := range collection.Pins {
			if pin.ID == id {
				// Update fields
				collection.Pins[i].Name = name
				collection.Pins[i].ModificationDate = time.Now()
				updated = &collection.Pins[i]
				return nil
			}
		}
		return errors.New("pin not found")
	})

	return updated, err
}

func (manager *PinnedManager) Delete(id int) error {
	return manager.metadata.Update(func(collection *model.PinnedCollection) error {
		for i, pin := range collection.Pins {
			if pin.ID == id {
				// Remove pin from slice
				collection.Pins = append(collection.Pins[:i], collection.Pins[i+1:]...)
				return nil
			}
		}
		return errors.New("pin not found")
	})
}

func (manager *PinnedManager) Get(id int) (*model.Pinned, error) {

	collection, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	for _, pin := range collection.Pins {
		if pin.ID == id {
			return &pin, nil
		}
	}
	return nil, errors.New("pin not found")
}

func (manager *PinnedManager) GetList(includeHidden bool) ([]model.Pinned, error) {
	collection, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Pinned
	for _, pin := range collection.Pins {
		//if !pin.IsHidden || includeHidden {
		result = append(result, pin)
		//}
	}
	return result, nil
}

func (manager *PinnedManager) load() ([]model.Pinned, error) {
	collection, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Pinned
	for _, pin := range collection.Pins {
		result = append(result, pin)
	}

	return result, nil
}

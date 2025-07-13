package storage

import (
	"errors"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/registery"
	"strconv"
	"time"
)

type PersonManager struct {
	registry *registery.Registry[model.Person]
	metadata *MetadataControl[model.PersonCollection]
}

func NewPersonManager(path string) (*PersonManager, error) {

	manager := &PersonManager{
		registry: registery.NewRegistry[model.Person](),
		metadata: NewMetadataControl[model.PersonCollection](path),
	}

	albums, err := manager.load()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Albums: %w", err)
	}

	for _, album := range albums {
		manager.registry.Register(strconv.Itoa(album.ID), album)
	}

	return manager, nil
}

func (manager *PersonManager) Create(name string) (*model.Person, error) {
	var newTrip *model.Person

	err := manager.metadata.Update(func(trips *model.PersonCollection) error {
		// Generate ID
		maxID := 0
		for _, trip := range trips.Persons {
			if trip.ID > maxID {
				maxID = trip.ID
			}
		}

		// Handler new trip
		newTrip = &model.Person{
			ID:   maxID + 1,
			Name: name,
			//TripType:         tripType,
			//IsCollection:     isCollection,
			CreationDate:     time.Now(),
			ModificationDate: time.Now(),
		}

		// Add to collection
		trips.Persons = append(trips.Persons, *newTrip)
		return nil
	})

	return newTrip, err
}

func (manager *PersonManager) Update(id int, name string, isHidden bool) (*model.Person, error) {
	var updated *model.Person

	err := manager.metadata.Update(func(collection *model.PersonCollection) error {
		for i, person := range collection.Persons {
			if person.ID == id {
				// Update fields
				collection.Persons[i].Name = name
				collection.Persons[i].ModificationDate = time.Now()

				updated = &collection.Persons[i]
				return nil
			}
		}
		return errors.New("person not found")
	})

	return updated, err
}

func (manager *PersonManager) Delete(id int) error {
	return manager.metadata.Update(func(collection *model.PersonCollection) error {
		for i, trip := range collection.Persons {
			if trip.ID == id {
				// Remove trip from slice
				collection.Persons = append(collection.Persons[:i], collection.Persons[i+1:]...)
				return nil
			}
		}
		return errors.New("trip not found")
	})
}

func (manager *PersonManager) Get(id int) (*model.Person, error) {

	persons, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	for _, person := range persons.Persons {
		if person.ID == id {
			return &person, nil
		}
	}
	return nil, errors.New("person not found")
}

func (manager *PersonManager) GetList(includeHidden bool) ([]model.Person, error) {
	collection, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Person
	for _, trip := range collection.Persons {
		//if !trip.IsHidden || includeHidden {
		result = append(result, trip)
		//}
	}
	return result, nil
}

func (manager *PersonManager) load() ([]model.Person, error) {
	collection, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Person
	for _, person := range collection.Persons {
		result = append(result, person)
	}

	return result, nil
}

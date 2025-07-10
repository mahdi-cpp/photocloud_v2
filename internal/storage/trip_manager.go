package storage

import (
	"errors"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"time"
)

type TripManager struct {
	metadata *MetadataControl[model.TripCollection]
}

func NewTripManager(path string) *TripManager {
	return &TripManager{
		metadata: NewMetadataManagerV2[model.TripCollection](path),
	}
}

// CreateTrip adds a new trip with auto-generated ID
func (manager *TripManager) CreateTrip(name string, tripType string, isCollection bool) (*model.Trip, error) {
	var newTrip *model.Trip

	err := manager.metadata.Update(func(trips *model.TripCollection) error {
		// Generate ID
		maxID := 0
		for _, trip := range trips.Trips {
			if trip.ID > maxID {
				maxID = trip.ID
			}
		}

		// Create new trip
		newTrip = &model.Trip{
			ID:               maxID + 1,
			Name:             name,
			TripType:         tripType,
			IsCollection:     isCollection,
			IsHidden:         false,
			CreationDate:     time.Now(),
			ModificationDate: time.Now(),
		}

		// Add to collection
		trips.Trips = append(trips.Trips, *newTrip)
		return nil
	})

	return newTrip, err
}

// Update modifies an existing trip
func (manager *TripManager) Update(id int, name string, tripType string, isHidden bool) (*model.Trip, error) {
	var updated *model.Trip

	err := manager.metadata.Update(func(trips *model.TripCollection) error {
		for i, trip := range trips.Trips {
			if trip.ID == id {
				// Update fields
				trips.Trips[i].Name = name
				trips.Trips[i].TripType = tripType
				trips.Trips[i].IsHidden = isHidden
				trips.Trips[i].ModificationDate = time.Now()

				updated = &trips.Trips[i]
				return nil
			}
		}
		return errors.New("trip not found")
	})

	return updated, err
}

// Delete removes trip by ID
func (manager *TripManager) Delete(id int) error {
	return manager.metadata.Update(func(trips *model.TripCollection) error {
		for i, trip := range trips.Trips {
			if trip.ID == id {
				// Remove trip from slice
				trips.Trips = append(trips.Trips[:i], trips.Trips[i+1:]...)
				return nil
			}
		}
		return errors.New("trip not found")
	})
}

// Get retrieves trip by ID
func (manager *TripManager) Get(id int) (*model.Trip, error) {

	trips, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	for _, trip := range trips.Trips {
		if trip.ID == id {
			return &trip, nil
		}
	}
	return nil, errors.New("trip not found")
}

// GetList returns all trips with optional filters
func (manager *TripManager) GetList(includeHidden bool) ([]model.Trip, error) {
	trips, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Trip
	for _, trip := range trips.Trips {
		if !trip.IsHidden || includeHidden {
			result = append(result, trip)
		}
	}
	return result, nil
}

// GetByType returns trips of a specific type
func (manager *TripManager) GetByType(tripType string) ([]model.Trip, error) {
	trips, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.Trip
	for _, trip := range trips.Trips {
		if trip.TripType != "" && trip.TripType == tripType {
			result = append(result, trip)
		}
	}
	return result, nil
}

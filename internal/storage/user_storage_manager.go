package storage

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/pkg/collection"
	"github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"
	"github.com/mahdi-cpp/photocloud_v2/pkg/image_loader"
	"github.com/mahdi-cpp/photocloud_v2/pkg/thumbnail"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type UserStorageManager struct {
	mu                  sync.RWMutex
	config              Config
	storages            map[int]*UserStorage // Maps user IDs to their UserStorage
	userManager         *collection.CollectionManager[*happle_models.User]
	originalImageLoader *image_loader.ImageLoader
	tinyImageLoader     *image_loader.ImageLoader
	iconLoader          *image_loader.ImageLoader
	ctx                 context.Context
}

func NewUserStorageManager(cfg Config) (*UserStorageManager, error) {

	// Handler the manager
	manager := &UserStorageManager{
		storages: make(map[int]*UserStorage),
		config:   cfg,
		ctx:      context.Background(),
	}

	var err error
	manager.userManager, err = collection.NewCollectionManager[*happle_models.User]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/users.json")
	if err != nil {
		panic(err)
	}

	manager.originalImageLoader = image_loader.NewImageLoader(50, "/media/mahdi/Cloud/apps/Photos", 5*time.Minute)
	manager.tinyImageLoader = image_loader.NewImageLoader(30000, "/media/mahdi/Cloud/apps/Photos", 60*time.Minute)
	manager.iconLoader = image_loader.NewImageLoader(1000, "/var/cloud/icons", 0)

	manager.loadAllIcons()

	return manager, nil
}

func (us *UserStorageManager) loadAllIcons() {
	us.iconLoader.GetLocalBasePath()

	// Scan metadata directory
	files, err := os.ReadDir(us.iconLoader.GetLocalBasePath())
	if err != nil {
		fmt.Println("failed to read metadata directory: %w", err)
	}

	var images []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".png") {
			images = append(images, "/media/mahdi/Cloud/apps/Photos/parsa_nasiri/assets/"+file.Name())
		}
	}
}

func (us *UserStorageManager) GetAssetManager(c *gin.Context, userID int) (*collection.CollectionManager[*model.Person], error) {
	userStorage, err := us.GetUserStorage(c, userID)
	if err != nil {
		return nil, err
	}

	return userStorage.PersonManager, nil
}

// periodicMaintenance runs background tasks
func (us *UserStorageManager) periodicMaintenance() {

	saveTicker := time.NewTicker(10 * time.Second)
	statsTicker := time.NewTicker(30 * time.Minute)
	rebuildTicker := time.NewTicker(24 * time.Hour)
	cleanupTicker := time.NewTicker(1 * time.Hour)

	for {
		select {
		case <-saveTicker.C:
			fmt.Println("saveTicker")
		case <-rebuildTicker.C:
			fmt.Println("rebuildTicker")
		case <-statsTicker.C:
			fmt.Println("statsTicker")
		case <-cleanupTicker.C:
			fmt.Println("cleanupTicker")
		}
	}
}

//----------------------------------------------------------------------------------------

func (us *UserStorageManager) RepositoryGetOriginalImage(filename string) ([]byte, error) {
	return us.originalImageLoader.LoadImage(us.ctx, filename)
}

func (us *UserStorageManager) RepositoryGetTinyImage(filename string) ([]byte, error) {
	return us.tinyImageLoader.LoadImage(us.ctx, filename)
}

func (us *UserStorageManager) RepositoryGetIcon(filename string) ([]byte, error) {
	return us.iconLoader.LoadImage(us.ctx, filename)
}

//-----------------------------------------------------------------------------------------

func (us *UserStorageManager) GetUserStorage(c *gin.Context, userID int) (*UserStorage, error) {

	us.mu.Lock()
	defer us.mu.Unlock()

	var err error

	if userID <= 0 {
		return nil, fmt.Errorf("user id is Invalid")
	}

	user, err := us.userManager.Get(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if userStorage already exists for this user
	if storage, exists := us.storages[userID]; exists {
		return storage, nil
	}

	// Handler context for background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Handler user-specific subdirectories
	userAssetDir := filepath.Join(us.config.AppDir, user.Username, us.config.AssetsDir)
	userMetadataDir := filepath.Join(us.config.AppDir, user.Username, us.config.MetadataDir)
	userThumbnailsDir := filepath.Join(us.config.AppDir, user.Username, us.config.ThumbnailsDir)
	//albumFile := filepath.Join(us.config.AppDir, user.Username, us.config.AlbumCollectionFile)
	//tripFile := filepath.Join(us.config.AppDir, user.Username, us.config.TripCollectionFile)
	//personFile := filepath.Join(us.config.AppDir, user.Username, us.config.PersonCollectionFile)
	//pinnedCollectionFile := filepath.Join(us.config.AppDir, user.Username, "pinnedCollectionFile.json")

	// Ensure user directories exist
	userDirs := []string{userAssetDir, userMetadataDir, userThumbnailsDir}
	for _, dir := range userDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create user directory %s: %w", dir, err)
		}
	}

	// Handler user-specific config
	userConfig := us.config
	userConfig.MetadataDir = userMetadataDir
	userConfig.ThumbnailsDir = userThumbnailsDir

	// Handler new userStorage for this user
	userStorage := &UserStorage{
		config:            userConfig,
		user:              *user,
		metadata:          NewMetadataManager(userMetadataDir),
		thumbnail:         thumbnail.NewThumbnailManager(userThumbnailsDir),
		maintenanceCtx:    ctx,
		cancelMaintenance: cancel,
	}

	userStorage.assets, err = userStorage.metadata.LoadUserAllMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata for user %s: %w", userID, err)
	}

	userStorage.AlbumManager, err = collection.NewCollectionManager[*model.Album]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/albums.json")
	if err != nil {
		panic(err)
	}

	userStorage.SharedAlbumManager, err = collection.NewCollectionManager[*model.SharedAlbum]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/shared_albums.json")
	if err != nil {
		panic(err)
	}

	userStorage.TripManager, err = collection.NewCollectionManager[*model.Trip]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/trips.json")
	if err != nil {
		panic(err)
	}

	userStorage.PersonManager, err = collection.NewCollectionManager[*model.Person]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/persons.json")
	if err != nil {
		panic(err)
	}

	userStorage.PinnedManager, err = collection.NewCollectionManager[*model.Pinned]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/pinned.json")
	if err != nil {
		panic(err)
	}

	//userStorage.CameraManager, err = NewCollectionManager[*model.Camera]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/camera.json")
	//if err != nil {
	//	panic(err)
	//}

	userStorage.prepareAlbums()
	userStorage.prepareTrips()
	userStorage.preparePersons()
	userStorage.prepareCameras()
	userStorage.preparePinned()

	// Store the new userStorage
	us.storages[userID] = userStorage

	return userStorage, nil
}

func (us *UserStorageManager) RemoveStorageForUser(userID int) {
	us.mu.Lock()
	defer us.mu.Unlock()

	if storage, exists := us.storages[userID]; exists {
		// Cancel any background operations
		storage.cancelMaintenance()
		// Remove from map
		delete(us.storages, userID)
	}
}

func (us *UserStorageManager) GetUserManager() (*collection.CollectionManager[*happle_models.User], error) {
	return us.userManager, nil
}

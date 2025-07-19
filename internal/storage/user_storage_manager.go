package storage

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/image_loader"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
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
	userManager         *CollectionManager[*model.User]
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
	manager.userManager, err = NewCollectionManager[*model.User]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/users.json")
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

//func (us *UserStorageManager) UploadAsset(c *gin.Context, userID int, file multipart.File, header *multipart.FileHeader) (*model.PHAsset, error) {
//
//	userStorage, err := us.GetUserStorage(c, userID)
//	if err != nil {
//		return nil, err
//	}
//
//	return userStorage.UploadAsset(userID, file, header)
//}

//func (us *UserStorageManager) FetchAssets(c *gin.Context, with model.PHFetchOptions) ([]*model.PHAsset, int, error) {
//
//	userStorage, err := us.GetUserStorage(c, with.UserID)
//	if err != nil {
//		return nil, 0, err
//	}
//
//	return userStorage.FetchAssets(with)
//}

func (us *UserStorageManager) GetAssetManager(c *gin.Context, userID int) (*CollectionManager[*model.Person], error) {
	userStorage, err := us.GetUserStorage(c, userID)
	if err != nil {
		return nil, err
	}

	return userStorage.PersonManager, nil
}

//func (us *UserStorageManager) UpdateAsset(c *gin.Context, assetIds []int, update model.AssetUpdate) (string, error) {
//	userStorage, err := us.GetUserStorage(c, update.UserID)
//	if err != nil {
//		return "", err
//	}
//
//	return userStorage.UpdateAsset(assetIds, update)
//}

//func (us *UserStorageManager) Prepare(c *gin.Context, update model.AssetUpdate) {
//	userStorage, err := us.GetUserStorage(c, update.UserID)
//	if err != nil {
//		return
//	}
//
//	userStorage.UpdateCollections()
//}

//func (us *UserStorageManager) GetAsset(c *gin.Context, userId int, assetId int) (*model.PHAsset, bool) {
//	userStorage, err := us.GetUserStorage(c, userId)
//	if err != nil {
//		return nil, false
//	}
//
//	return userStorage.GetAsset(assetId)
//}

//func (us *UserStorageManager) Delete(c *gin.Context, userId int, assetId int) error {
//	userStorage, err := us.GetUserStorage(c, userId)
//	if err != nil {
//		return err
//	}
//
//	return userStorage.DeleteAsset(assetId)
//}

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
		thumbnail:         NewThumbnailManager(userThumbnailsDir),
		maintenanceCtx:    ctx,
		cancelMaintenance: cancel,
	}

	userStorage.assets, err = userStorage.metadata.LoadUserAllMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata for user %s: %w", userID, err)
	}

	userStorage.AlbumManager, err = NewCollectionManager[*model.Album]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/collection.json")
	if err != nil {
		panic(err)
	}

	userStorage.SharedAlbumManager, err = NewCollectionManager[*model.SharedAlbum]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/shared_albums.json")
	if err != nil {
		panic(err)
	}

	userStorage.TripManager, err = NewCollectionManager[*model.Trip]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/trips.json")
	if err != nil {
		panic(err)
	}

	userStorage.PersonManager, err = NewCollectionManager[*model.Person]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/persons.json")
	if err != nil {
		panic(err)
	}

	userStorage.PinnedManager, err = NewCollectionManager[*model.Pinned]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/pinned.json")
	if err != nil {
		panic(err)
	}

	userStorage.CameraManager, err = NewCollectionManager[*model.Camera]("/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/camera.json")
	if err != nil {
		panic(err)
	}

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

func (us *UserStorageManager) GetUserManager() (*CollectionManager[*model.User], error) {
	return us.userManager, nil
}

//func (us *UserStorageManager) GetAlbumManager(c *gin.Context, userID int) (*CollectionManager[*model.Album], error) {
//	userStorage, err := us.GetUserStorage(c, userID)
//	if err != nil {
//		return nil, err
//	}
//
//	return userStorage.AlbumManager, nil
//}

//func (us *UserStorageManager) GetTripManager(c *gin.Context, userID int) (*CollectionManager[*model.Trip], error) {
//	userStorage, err := us.GetUserStorage(c, userID)
//	if err != nil {
//		return nil, err
//	}
//
//	return userStorage.TripManager, nil
//}

func (us *UserStorageManager) GetPersonManager(c *gin.Context, userID int) (*CollectionManager[*model.Person], error) {
	userStorage, err := us.GetUserStorage(c, userID)
	if err != nil {
		return nil, err
	}

	return userStorage.PersonManager, nil
}

func (us *UserStorageManager) GetPinnedManager(c *gin.Context, userID int) (*CollectionManager[*model.Pinned], error) {
	userStorage, err := us.GetUserStorage(c, userID)
	if err != nil {
		return nil, err
	}

	return userStorage.PinnedManager, nil
}

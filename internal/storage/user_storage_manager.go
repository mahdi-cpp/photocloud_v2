package storage

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/api-go-pkg/collection"
	"github.com/mahdi-cpp/api-go-pkg/common_models"
	"github.com/mahdi-cpp/api-go-pkg/image_loader"
	"github.com/mahdi-cpp/api-go-pkg/metadata"
	"github.com/mahdi-cpp/api-go-pkg/network"
	"github.com/mahdi-cpp/api-go-pkg/thumbnail"
	"github.com/mahdi-cpp/photocloud_v2/config"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"log"
	"sync"
	"time"
)

type UserStorageManager struct {
	mu           sync.RWMutex
	users        map[int]*common_models.User
	userStorages map[int]*UserStorage // Maps user IDs to their UserStorage
	iconLoader   *image_loader.ImageLoader
	ctx          context.Context
}

func NewUserStorageManager() (*UserStorageManager, error) {

	// Handler the manager
	manager := &UserStorageManager{
		userStorages: make(map[int]*UserStorage),
		users:        make(map[int]*common_models.User),
		ctx:          context.Background(),
	}

	userControl := network.NewNetworkControl[[]common_models.User]("http://localhost:8080/api/v1/user/")

	// Make request (nil body if not needed)
	users, err := userControl.Read("list", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Use the data
	for _, user := range *users {
		fmt.Printf("%d: %s (%s %s)\n",
			user.ID,
			user.Username,
			user.FirstName,
			user.LastName)
		manager.users[user.ID] = &user
	}

	manager.iconLoader = image_loader.NewImageLoader(1000, config.GetPath("/data/icons"), 0)
	manager.loadAllIcons()

	return manager, nil
}

func (us *UserStorageManager) loadAllIcons() {
	us.iconLoader.GetLocalBasePath()

	// Scan metadata directory
	//files, err := os.ReadDir(us.iconLoader.GetLocalBasePath())
	//if err != nil {
	//	fmt.Println("failed to read metadata directory: %w", err)
	//}

	//var images []string
	//for _, file := range files {
	//	if strings.HasSuffix(file.Name(), ".png") {
	//		images = append(images, "/media/mahdi/Cloud/apps/Photos/parsa_nasiri/assets/"+file.Name())
	//	}
	//}
}

func (us *UserStorageManager) GetAssetManager(c *gin.Context, userID int) (*collection.Manager[*model.Person], error) {
	userStorage, err := us.GetUserStorage(c, userID)
	if err != nil {
		return nil, err
	}

	return userStorage.PersonManager, nil
}

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

func (us *UserStorageManager) RepositoryGetOriginalImage(userID int, filename string) ([]byte, error) {
	return us.userStorages[userID].originalImageLoader.LoadImage(us.ctx, filename)
}

func (us *UserStorageManager) RepositoryGetTinyImage(userID int, filename string) ([]byte, error) {
	return us.userStorages[userID].tinyImageLoader.LoadImage(us.ctx, filename)
}

func (us *UserStorageManager) RepositoryGetIcon(filename string) ([]byte, error) {
	return us.iconLoader.LoadImage(us.ctx, filename)
}

func (us *UserStorageManager) GetUserStorage(c *gin.Context, userID int) (*UserStorage, error) {

	us.mu.Lock()
	defer us.mu.Unlock()

	var err error

	if userID <= 0 {
		return nil, fmt.Errorf("user id is Invalid")
	}

	var user = us.users[userID]

	// Check if userStorage already exists for this user
	if storage, exists := us.userStorages[userID]; exists {
		return storage, nil
	}

	fmt.Println("ali1")
	// Handler context for background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Handler user-specific subdirectories
	//userAssetDir := filepath.Join(us.config.AppDir, user.PhoneNumber, us.config.AssetsDir)
	//userMetadataDir := filepath.Join(us.config.AppDir, user.PhoneNumber, us.config.MetadataDir)
	//userThumbnailsDir := filepath.Join(us.config.AppDir, user.PhoneNumber, us.config.ThumbnailsDir)
	//albumFile := filepath.Join(us.config.AppDir, user.Username, us.config.AlbumCollectionFile)
	//tripFile := filepath.Join(us.config.AppDir, user.Username, us.config.TripCollectionFile)
	//personFile := filepath.Join(us.config.AppDir, user.Username, us.config.PersonCollectionFile)
	//pinnedCollectionFile := filepath.Join(us.config.AppDir, user.Username, "pinnedCollectionFile.json")

	// Ensure user directories exist
	//userDirs := []string{userAssetDir, userMetadataDir, userThumbnailsDir}
	//for _, dir := range userDirs {
	//	if err := os.MkdirAll(dir, 0755); err != nil {
	//		return nil, fmt.Errorf("failed to create user directory %s: %w", dir, err)
	//	}
	//}

	// Handler user-specific config
	//userConfig := us.config
	//userConfig.MetadataDir = config.GetUserPath(user.PhoneNumber, "assets")
	//userConfig.ThumbnailsDir = config.GetUserPath(user.PhoneNumber, "thumbnails")

	// Handler new userStorage for this user
	userStorage := &UserStorage{
		user:              *user,
		metadata:          metadata.NewMetadataManager(config.GetUserPath(user.PhoneNumber, "metadata")),
		thumbnail:         thumbnail.NewThumbnailManager(config.GetUserPath(user.PhoneNumber, "thumbnails")),
		maintenanceCtx:    ctx,
		cancelMaintenance: cancel,
	}

	userStorage.originalImageLoader = image_loader.NewImageLoader(50, config.GetUserPath(user.PhoneNumber, "assets"), 5*time.Minute)
	userStorage.tinyImageLoader = image_loader.NewImageLoader(30000, config.GetUserPath(user.PhoneNumber, "thumbnails"), 60*time.Minute)

	userStorage.assets, err = userStorage.metadata.LoadUserAllMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata for user %s: %w", userID, err)
	}

	userStorage.AlbumManager, err = collection.NewCollectionManager[*model.Album](config.GetUserPath(user.PhoneNumber, "data/albums.json"))
	if err != nil {
		panic(err)
	}

	userStorage.SharedAlbumManager, err = collection.NewCollectionManager[*model.SharedAlbum](config.GetUserPath(user.PhoneNumber, "data/shared_albums.json"))
	if err != nil {
		panic(err)
	}

	userStorage.TripManager, err = collection.NewCollectionManager[*model.Trip](config.GetUserPath(user.PhoneNumber, "data/trips.json"))
	if err != nil {
		panic(err)
	}

	userStorage.PersonManager, err = collection.NewCollectionManager[*model.Person](config.GetUserPath(user.PhoneNumber, "data/persons.json"))
	if err != nil {
		panic(err)
	}

	userStorage.PinnedManager, err = collection.NewCollectionManager[*model.Pinned](config.GetUserPath(user.PhoneNumber, "data/pinned.json"))
	if err != nil {
		panic(err)
	}

	userStorage.VillageManager, err = collection.NewCollectionManager[*model.Village](config.GetPath("/data/villages.json"))
	if err != nil {
		panic(err)
	}

	userStorage.prepareAlbums()
	userStorage.prepareTrips()
	userStorage.preparePersons()
	userStorage.prepareCameras()
	userStorage.preparePinned()

	// Store the new userStorage
	us.userStorages[userID] = userStorage

	return userStorage, nil
}

func (us *UserStorageManager) RemoveStorageForUser(userID int) {
	us.mu.Lock()
	defer us.mu.Unlock()

	if storage, exists := us.userStorages[userID]; exists {
		// Cancel any background operations
		storage.cancelMaintenance()
		// Remove from map
		delete(us.userStorages, userID)
	}
}

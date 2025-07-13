package storage

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config defines storage system configuration
type Config struct {
	AppDir               string
	AssetsDir            string
	MetadataDir          string
	ThumbnailsDir        string
	IndexFile            string
	CacheSize            int
	MaxUploadSize        int64
	AlbumCollectionFile  string
	TripCollectionFile   string
	PersonCollectionFile string
}

// Stats holds storage system statistics
type Stats struct {
	TotalAssets   int
	CacheHits     int64
	CacheMisses   int64
	Uploads24h    int
	ThumbnailsGen int
}

type UserStorageManager struct {
	mu              sync.RWMutex
	storages        map[int]*UserStorage // Maps user IDs to their UserStorage
	config          Config
	userManager     *UserManager
	imageRepository *ImageRepository
}

func NewUserStorageManager(cfg Config) (*UserStorageManager, error) {

	// Handler the manager
	manager := &UserStorageManager{
		storages:        make(map[int]*UserStorage),
		config:          cfg,
		userManager:     NewUserManager("/media/mahdi/Cloud/apps/system/users.json"),
		imageRepository: NewImageRepository(),
	}

	// Ensure base directories exist
	dirs := []string{cfg.AssetsDir, cfg.MetadataDir, cfg.ThumbnailsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return manager, nil
}

func (us *UserStorageManager) GetStorageForUser(c *gin.Context, userID int) (*UserStorage, error) {
	us.mu.Lock()
	defer us.mu.Unlock()

	if userID <= 0 {
		return nil, fmt.Errorf("user id is Invalid !!??? ")
	}

	var user = us.userManager.GetById(userID)

	// Check if storage already exists for this user
	if storage, exists := us.storages[userID]; exists {
		return storage, nil
	}

	// Handler context for background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Handler user-specific subdirectories
	userAssetDir := filepath.Join(us.config.AppDir, user.Username, us.config.AssetsDir)
	userMetadataDir := filepath.Join(us.config.AppDir, user.Username, us.config.MetadataDir)
	userThumbnailsDir := filepath.Join(us.config.AppDir, user.Username, us.config.ThumbnailsDir)
	albumFile := filepath.Join(us.config.AppDir, user.Username, us.config.AlbumCollectionFile)
	tripFile := filepath.Join(us.config.AppDir, user.Username, us.config.TripCollectionFile)
	personFile := filepath.Join(us.config.AppDir, user.Username, us.config.PersonCollectionFile)
	pinnedCollectionFile := filepath.Join(us.config.AppDir, user.Username, "pinnedCollectionFile.json")

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

	// Handler new storage for this user
	storage := &UserStorage{
		config:            userConfig,
		user:              user,
		metadata:          NewMetadataManager(userMetadataDir),
		thumbnail:         NewThumbnailManager(userThumbnailsDir),
		maintenanceCtx:    ctx,
		cancelMaintenance: cancel,
	}

	storage.pinnedManager, _ = NewPinnedManager(pinnedCollectionFile)
	storage.albumManager, _ = NewAlbumManager(albumFile)
	storage.tripManager, _ = NewTripManager(tripFile)
	storage.personManager, _ = NewPersonManager(personFile)

	// Load user assets
	var err error
	storage.assets, err = storage.metadata.LoadUserAllMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata for user %s: %w", userID, err)
	}

	// Store the new storage
	us.storages[userID] = storage

	return storage, nil
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

func (us *UserStorageManager) GetPinnedManager(c *gin.Context, userID int) *PinnedManager {
	userStorage, err := us.GetStorageForUser(c, userID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.pinnedManager
}

func (us *UserStorageManager) GetAlbumManager(c *gin.Context, userID int) *AlbumManager {
	userStorage, err := us.GetStorageForUser(c, userID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.albumManager
}

func (us *UserStorageManager) GetTripManager(c *gin.Context, userID int) *TripManager {
	userStorage, err := us.GetStorageForUser(c, userID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.tripManager
}

func (us *UserStorageManager) GetPersonManager(c *gin.Context, userID int) *PersonManager {
	userStorage, err := us.GetStorageForUser(c, userID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.personManager
}

func (us *UserStorageManager) UploadAsset(c *gin.Context, userID int, file multipart.File, header *multipart.FileHeader) (*model.PHAsset, error) {

	userStorage, err := us.GetStorageForUser(c, userID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.UploadAsset(userID, file, header)
}

func (us *UserStorageManager) FilterAssets(c *gin.Context, filters model.AssetSearchFilters) ([]*model.PHAsset, int, error) {

	userStorage, err := us.GetStorageForUser(c, filters.UserID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.FilterAssets(filters)
}

func (us *UserStorageManager) UpdateAsset(c *gin.Context, assetIds []int, update model.AssetUpdate) (string, error) {
	userStorage, err := us.GetStorageForUser(c, update.UserID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.UpdateAsset(assetIds, update)
}

func (us *UserStorageManager) GetAsset(c *gin.Context, userId int, assetId int) (*model.PHAsset, error) {
	userStorage, err := us.GetStorageForUser(c, userId)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.GetAsset(assetId)
}

func (us *UserStorageManager) Delete(c *gin.Context, userId int, assetId int) error {
	userStorage, err := us.GetStorageForUser(c, userId)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.DeleteAsset(assetId)
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

func (us *UserStorageManager) RepositorySearch(fileName string) (string, error) {
	return us.imageRepository.SearchFile(fileName)
}

//func (us *UserStorageManager) RepositoryGetIcon(filename string) ([]byte, bool) {
//	return us.imageRepository.GetIconCash(filename)
//}

func (us *UserStorageManager) RepositoryGetImage(filename string) ([]byte, bool) {
	return us.imageRepository.GetImage(filename)
}

func (us *UserStorageManager) AddTinyImage(filepath string, filename string) {
	us.imageRepository.AddTinyImage(filepath, filename)
}

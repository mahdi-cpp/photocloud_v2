package storage

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/registery"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
)

type UserStorageManager struct {
	mu          sync.RWMutex
	storages    map[int]*UserStorage // Maps user IDs to their UserStorage
	config      Config
	userManager *UserManager
}

func NewUserStorageManager(cfg Config) (*UserStorageManager, error) {

	// Create the manager
	manager := &UserStorageManager{
		storages:    make(map[int]*UserStorage),
		config:      cfg,
		userManager: NewUserManager("/media/mahdi/Cloud/apps/system/users.json"),
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

func (m *UserStorageManager) GetStorageForUser(c *gin.Context, userID int) (*UserStorage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if userID <= 0 {
		return nil, fmt.Errorf("user id is Invalid !!??? ")
	}

	var user = m.userManager.GetById(userID)

	// Check if storage already exists for this user
	if storage, exists := m.storages[userID]; exists {
		return storage, nil
	}

	// Create context for background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Create user-specific subdirectories
	userAssetDir := filepath.Join(m.config.AppDir, user.Username, m.config.AssetsDir)
	userMetadataDir := filepath.Join(m.config.AppDir, user.Username, m.config.MetadataDir)
	userThumbnailsDir := filepath.Join(m.config.AppDir, user.Username, m.config.ThumbnailsDir)

	// Ensure user directories exist
	userDirs := []string{userAssetDir, userMetadataDir, userThumbnailsDir}
	for _, dir := range userDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create user directory %s: %w", dir, err)
		}
	}

	// Create user-specific config
	userConfig := m.config
	userConfig.MetadataDir = userMetadataDir
	userConfig.ThumbnailsDir = userThumbnailsDir

	// Create new storage for this user
	storage := &UserStorage{
		config:            userConfig,
		user:              user,
		metadata:          NewMetadataManager(userMetadataDir),
		thumbnail:         NewThumbnailManager(userThumbnailsDir),
		albumRegistry:     registery.NewRegistry[model.Album](),
		maintenanceCtx:    ctx,
		cancelMaintenance: cancel,
	}

	// Load user assets
	var err error
	storage.assets, err = storage.metadata.LoadUserAllMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata for user %s: %w", userID, err)
	}

	// Store the new storage
	m.storages[userID] = storage

	return storage, nil
}

func (m *UserStorageManager) RemoveStorageForUser(userID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if storage, exists := m.storages[userID]; exists {
		// Cancel any background operations
		storage.cancelMaintenance()
		// Remove from map
		delete(m.storages, userID)
	}
}

func (m *UserStorageManager) UploadAsset(c *gin.Context, userID int, file multipart.File, header *multipart.FileHeader) (*model.PHAsset, error) {

	userStorage, err := m.GetStorageForUser(c, userID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.UploadAsset(userID, file, header)
}

func (m *UserStorageManager) FilterAssets(c *gin.Context, filters model.AssetSearchFilters) ([]*model.PHAsset, int, error) {

	userStorage, err := m.GetStorageForUser(c, filters.UserID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.FilterAssets(filters)
}

func (m *UserStorageManager) UpdateAsset(c *gin.Context, assetIds []int, update model.AssetUpdate) (string, error) {
	userStorage, err := m.GetStorageForUser(c, update.UserID)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.UpdateAsset(assetIds, update)
}

func (m *UserStorageManager) GetAsset(c *gin.Context, userId int, assetId int) (*model.PHAsset, error) {
	userStorage, err := m.GetStorageForUser(c, userId)
	if err != nil {
		log.Fatal(err)
	}

	return userStorage.GetAsset(assetId)
}

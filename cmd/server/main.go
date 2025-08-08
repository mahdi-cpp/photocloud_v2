package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/internal/api/handler"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// Load configuration
	//cfg, err := loadConfig()
	//if err != nil {
	//	log.Fatalf("Failed to load configuration: %v", err)
	//}

	//storageCfg := storage.Config{
	//	AppDir:               cfg.Storage.AppDir,
	//	AssetsDir:            cfg.Storage.AssetsDir,
	//	MetadataDir:          cfg.Storage.MetadataDir,
	//	ThumbnailsDir:        cfg.Storage.ThumbnailsDir,
	//	IndexFile:            cfg.Storage.IndexFile,
	//	CacheSize:            cfg.Storage.Cache.Size,
	//	AlbumCollectionFile:  cfg.Storage.AlbumCollectionFile,
	//	TripCollectionFile:   cfg.Storage.TripCollectionFile,
	//	PersonCollectionFile: cfg.Storage.PersonCollectionFile,
	//}

	userStorageManager, err := storage.NewUserStorageManager()
	if err != nil {
		log.Fatal(err)
	}

	assetHandler := handler.NewAssetHandler(userStorageManager)
	searchHandler := handler.NewSearchHandler(userStorageManager)
	albumHandler := handler.NewAlbumHandler(userStorageManager)
	tripHandler := handler.NewTripHandler(userStorageManager)
	personHandler := handler.NewPersonsHandler(userStorageManager)
	pinnedHandler := handler.NewPinnedHandler(userStorageManager)
	cameraHandler := handler.NewCameraHandler(userStorageManager)
	sharedAlbumHandler := handler.NewSharedAlbumHandler(userStorageManager)
	villageHandler := handler.NewVillageHandler(userStorageManager)

	// Handler Gin router
	router := createRouter(
		assetHandler,
		albumHandler,
		villageHandler,
		sharedAlbumHandler,
		tripHandler,
		personHandler,
		searchHandler,
		pinnedHandler,
		cameraHandler)

	// Start server
	startServer(router)
}

//func loadConfig() (error) {
//
//	// Initialize Viper
//	v := viper.New()
//	v.SetConfigName("config")
//	v.SetConfigType("yaml")
//	v.AddConfigPath("./config")
//	v.AutomaticEnv()
//	v.SetEnvPrefix("PHOTOCLOUD")
//
//	// Set default values
//	v.SetDefault("server.host", "0.0.0.0")
//	v.SetDefault("server.port", 8080)
//	v.SetDefault("server.mode", "release")
//	v.SetDefault("storage.cache.size", 1000)
//	v.SetDefault("media.thumbnails.default_width", 300)
//	v.SetDefault("media.thumbnails.default_height", 300)
//	v.SetDefault("auth.jwt.expiration", "720h") // 30 days
//
//	// Read configuration
//	if err := v.ReadInConfig(); err != nil {
//		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
//			return nil, fmt.Errorf("failed to read config: %w", err)
//		}
//		log.Println("Config file not found, using environment variables and defaults")
//	}
//
//	var cfg config.Config
//	if err := v.Unmarshal(&cfg); err != nil {
//		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
//	}
//
//	// Validate configuration
//	if err := cfg.Validate(); err != nil {
//		return nil, fmt.Errorf("invalid configuration: %w", err)
//	}
//
//	return &cfg, nil
//}

func createRouter(
	assetHandler *handler.AssetHandler,
	albumHandler *handler.AlbumHandler,
	villageHandler *handler.VillageHandler,
	sharedAlbumHandler *handler.SharedAlbumHandler,
	tripHandler *handler.TripHandler,
	personHandler *handler.PersonHandler,
	searchHandler *handler.SearchHandler,
	pinnedHandler *handler.PinnedHandler,
	cameraHandler *handler.CameraHandler,
) *gin.Engine {

	// Set Gin mode
	gin.SetMode("release")

	// Handler router with default middleware
	router := gin.Default()

	// API routes
	api := router.Group("/api/v1")
	{

		// Search routes
		api.GET("/search", searchHandler.Search)
		api.POST("/search/filters", searchHandler.Filters)

		// Asset routes
		api.POST("/assets/create", assetHandler.Create)
		api.POST("/assets", assetHandler.Upload)
		api.GET("/assets/:id", assetHandler.Get)
		api.POST("/assets/update", assetHandler.Update)
		api.POST("/assets/update_all", assetHandler.UpdateAll)
		api.POST("/assets/delete", assetHandler.Delete)
		api.POST("/assets/filters", assetHandler.Filters)

		//http://localhost:8080/api/v1/assets/download/thumbnail/map_270.jpg
		api.GET("/assets/download/:filename", assetHandler.OriginalDownload)
		api.GET("/assets/download/thumbnail/:filename", assetHandler.TinyImageDownload)
		api.GET("/assets/download/icons/:filename", assetHandler.IconDownload)

		api.POST("/village/list", villageHandler.GetList)

		api.POST("/album/create", albumHandler.Create)
		api.POST("/album/update", albumHandler.Update)
		api.POST("/album/delete", albumHandler.Delete)
		api.POST("/album/list", albumHandler.GetListV2)

		api.POST("/shared_album/create", sharedAlbumHandler.Create)
		api.POST("/shared_album/update", sharedAlbumHandler.Update)
		api.POST("/shared_album/delete", sharedAlbumHandler.Delete)
		api.POST("/shared_album/list", sharedAlbumHandler.GetList)

		api.POST("/trip/create", tripHandler.Create)
		api.POST("/trip/update", tripHandler.Update)
		api.POST("/trip/delete", tripHandler.Delete)
		api.POST("/trip/list", tripHandler.GetCollectionList)

		api.POST("/person/create", personHandler.Create)
		api.POST("/person/update", personHandler.Update)
		api.POST("/person/delete", personHandler.Delete)
		api.POST("/person/list", personHandler.GetCollectionList)

		api.POST("/pinned/create", pinnedHandler.Create)
		api.POST("/pinned/update", pinnedHandler.Update)
		api.POST("/pinned/delete", pinnedHandler.Delete)
		api.POST("/pinned/list", pinnedHandler.GetList)

		//api.POST("/camera/create", cameraHandler.Create)
		//api.POST("/camera/update", cameraHandler.Update)
		//api.POST("/camera/delete", cameraHandler.Delete)
		api.POST("/camera/list", cameraHandler.GetList)

	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func startServer(router *gin.Engine) {

	// Handler HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", "0.0.0.0", 8081),
		Handler: router,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

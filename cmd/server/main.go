package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/photocloud_v2/config"
	"github.com/mahdi-cpp/photocloud_v2/internal/api/handler"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// Initialize loader (with local image directory)
	//loader := image_loader.NewImageLoader(5000, "", 10*time.Minute)

	//// Scan metadata directory
	//files, err := os.ReadDir("/media/mahdi/Cloud/apps/Photos/parsa_nasiri/assets")
	//if err != nil {
	//	fmt.Println("failed to read metadata directory: %w", err)
	//}
	//
	//var images []string
	//for _, file := range files {
	//	fmt.Println(file.Name())
	//	images = append(images, "/media/mahdi/Cloud/apps/Photos/parsa_nasiri/assets/"+file.Name())
	//}
	//
	//// Load various image types
	//images := []string{
	//	//"/var/cloud/upload/upload5/20190809_000407.jpg",
	//	//"Screenshot_20240113_180718_Instagram.jpg",
	//	//"Screenshot_20240120_020041_Instagram.jpg",
	//	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/18.jpg",
	//	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/17.jpg",
	//	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/25.jpg",
	//	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/26.jpg",
	//	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/27.jpg",
	//	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/28.jpg",
	//	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/42.jpg",
	//
	//	//"https://mahdiali.s3.ir-thr-at1.arvanstorage.ir/%D9%86%D9%82%D8%B4%D9%87-%D8%AA%D8%A7%DB%8C%D9%85%D8%B1-%D8%B1%D8%A7%D9%87-%D9%BE%D9%84%D9%87-%D8%B3%D9%87-%D8%B3%DB%8C%D9%85.jpg?versionId=", // Network URL
	//	//"https://mahdicpp.s3.ir-thr-at1.arvanstorage.ir/0f470b87c13e25bc4211683711e71e2a.jpg?versionId=",
	//}
	//
	//ctx := context.Background()
	//for _, img := range images {
	//	data, err := loader.LoadImage(ctx, img)
	//	if err != nil {
	//		log.Printf("Failed to load %s: %v", img, err)
	//		continue
	//	}
	//	fmt.Printf("Loaded %s (%d kB)\n", img, len(data)/1024)
	//}

	//Print metrics
	//f, n, g, e, avg := loader.Metrics()
	//fmt.Printf("\nLoader Metrics:\n")
	//fmt.Printf("File loads: %d\n", f)
	//fmt.Printf("Network loads: %d\n", n)
	//fmt.Printf("Generated images: %d\n", g)
	//fmt.Printf("Errors: %d\n", e)
	//fmt.Printf("Avg load time: %s\n", avg)

	//Get metrics
	//loadMetric := loader.Metrics()
	//fmt.Printf("CurrentCacheBytes: %s\n", image_loader.FormatBytes(loadMetric.CurrentCacheBytes))

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	storageCfg := storage.Config{
		AppDir:               cfg.Storage.AppDir,
		AssetsDir:            cfg.Storage.AssetsDir,
		MetadataDir:          cfg.Storage.MetadataDir,
		ThumbnailsDir:        cfg.Storage.ThumbnailsDir,
		IndexFile:            cfg.Storage.IndexFile,
		CacheSize:            cfg.Storage.Cache.Size,
		AlbumCollectionFile:  cfg.Storage.AlbumCollectionFile,
		TripCollectionFile:   cfg.Storage.TripCollectionFile,
		PersonCollectionFile: cfg.Storage.PersonCollectionFile,
	}

	userStorageManager, err := storage.NewUserStorageManager(storageCfg)
	if err != nil {
		log.Fatal(err)
	}

	userHandler := handler.NewUserHandler(userStorageManager)
	assetHandler := handler.NewAssetHandler(userStorageManager)
	searchHandler := handler.NewSearchHandler(userStorageManager)

	albumHandler := handler.NewAlbumHandler(userStorageManager)
	tripHandler := handler.NewTripHandler(userStorageManager)
	personHandler := handler.NewPersonsHandler(userStorageManager)
	pinnedHandler := handler.NewPinnedHandler(userStorageManager)
	cameraHandler := handler.NewCameraHandler(userStorageManager)
	sharedAlbumHandler := handler.NewSharedAlbumHandler(userStorageManager)

	// Handler Gin router
	router := createRouter(cfg,
		userHandler,
		assetHandler,
		albumHandler,
		sharedAlbumHandler,
		tripHandler,
		personHandler,
		searchHandler,
		pinnedHandler,
		cameraHandler)

	// Start server
	startServer(cfg, router)
}

func loadConfig() (*config.Config, error) {

	// Initialize Viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AutomaticEnv()
	v.SetEnvPrefix("PHOTOCLOUD")

	// Set default values
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("storage.cache.size", 1000)
	v.SetDefault("media.thumbnails.default_width", 300)
	v.SetDefault("media.thumbnails.default_height", 300)
	v.SetDefault("auth.jwt.expiration", "720h") // 30 days

	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		log.Println("Config file not found, using environment variables and defaults")
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func createRouter(
	cfg *config.Config,
	userHandler *handler.UserHandler,
	assetHandler *handler.AssetHandler,
	albumHandler *handler.AlbumHandler,
	sharedAlbumHandler *handler.SharedAlbumHandler,
	tripHandler *handler.TripHandler,
	personHandler *handler.PersonHandler,
	searchHandler *handler.SearchHandler,
	pinnedHandler *handler.PinnedHandler,
	cameraHandler *handler.CameraHandler,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Handler router with default middleware
	router := gin.Default()

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/user/create", userHandler.Create)
		api.POST("/user/update", userHandler.Update)
		api.POST("/user/delete", userHandler.Delete)
		api.POST("/user/list", userHandler.GetCollectionList)

		// Search routes
		api.GET("/search", searchHandler.Search)
		api.POST("/search/filters", searchHandler.Filters)

		// Asset routes
		api.POST("/assets", assetHandler.Upload)
		api.GET("/assets/:id", assetHandler.Get)
		api.POST("/assets/update", assetHandler.Update)
		api.POST("/assets/delete", assetHandler.Delete)
		api.POST("/assets/filters", assetHandler.Filters)

		api.GET("/assets/download/:filename", assetHandler.OriginalDownload)
		api.GET("/assets/download/thumbnail/:filename", assetHandler.TinyImageDownload)
		api.GET("/assets/download/icons/:filename", assetHandler.IconDownload)

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

		api.POST("/camera/create", cameraHandler.Create)
		api.POST("/camera/update", cameraHandler.Update)
		api.POST("/camera/delete", cameraHandler.Delete)
		api.POST("/camera/list", cameraHandler.GetList)

		//api.PUT("/assets/:id", assetHandler.Update)
		//api.DELETE("/assets/:id", assetHandler.DeleteAsset)
		//api.GET("/assets/:id/content", assetHandler.GetAssetContent)
		//api.GET("/assets/:id/thumbnail", assetHandler.GetAssetThumbnail)

		//api.GET("/search/suggest", searchHandler.SuggestSearchTerms)

		// System routes
		//api.GET("/system/status", systemHandler.GetSystemStatus)
		//api.POST("/system/rebuild-index", systemHandler.RebuildIndex)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func startServer(cfg *config.Config, router *gin.Engine) {

	// Handler HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

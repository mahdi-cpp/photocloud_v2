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

// curl -X GET http://localhost:8080/api/v1/search  -H "Content-Type: application/json" -d '{"userId": 3327,"isFavorite": true,"isHidden": false,"limit": 10}'
// curl -X GET "http://localhost:8080/api/v1/search?userId=3327&isFavorite=true,isScreenshot=true&isHidden=false&limit=10"

// http://localhost:8080/api/v1/search?userId=3327&isFavorite=true&isHidden=false&limit=10

// curl -X POST http://localhost:8080/api/v1/search/advanced_v2 -H "Content-Type: application/json" -d '{ "mediaType": "image", "isFavorite": true, "albums": [1, 5, 10] }'

func main() {

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage system
	photoStorage, err := initStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create repository
	assetRepo := storage.NewAssetRepository(photoStorage)

	// Initialize services
	//assetService := service.NewAssetService(assetRepo)
	//searchService := service.NewSearchService(assetRepo)

	//thumbnailService := imaging.NewThumbnailService(
	//	cfg.Media.Thumbnails.DefaultWidth,
	//	cfg.Media.Thumbnails.DefaultHeight,
	//	cfg.Media.Thumbnails.JPEGQuality,
	//	photoStorage,
	//)

	albumManager := storage.NewAlbumManager(cfg.Storage.AlbumCollectionFile)
	tripManager := storage.NewTripManager(cfg.Storage.TripCollectionFile)

	// Create handlers
	assetHandler := handler.NewAssetHandler(assetRepo)
	albumHandler := handler.NewAlbumHandler(albumManager)
	tripHandler := handler.NewTripHandler(tripManager)
	searchHandler := handler.NewSearchHandler(assetRepo)
	systemHandler := handler.NewSystemHandler(photoStorage)

	// Create Gin router
	router := createRouter(cfg, assetHandler, albumHandler, tripHandler, searchHandler, systemHandler)

	// Create logger
	//logger, _ := zap.NewProduction()
	//defer logger.Sync()

	// Register middlewares
	//middleware.RegisterMiddlewares(
	//	router,
	//	logger,
	//	cfg.Auth.JWT.Secret,
	//	cfg.Server.CORS.AllowOrigins,
	//	cfg.Server.CORS.AllowMethods,
	//	cfg.Server.CORS.AllowHeaders,
	//	cfg.Auth.RateLimiting.Requests,
	//	cfg.Auth.RateLimiting.Burst,
	//)

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

func initStorage(cfg *config.Config) (*storage.PhotoStorage, error) {

	// Create storage configuration
	storageCfg := storage.Config{
		AssetsDir:     cfg.Storage.AssetsDir,
		MetadataDir:   cfg.Storage.MetadataDir,
		ThumbnailsDir: cfg.Storage.ThumbnailsDir,
		IndexFile:     cfg.Storage.IndexFile,
		CacheSize:     cfg.Storage.Cache.Size,
	}

	// Create storage
	photoStorage, _ := storage.NewPhotoStorage(storageCfg)

	//// Create metadata extractor
	//metadataExtractor := imaging.NewMetadataExtractor(cfg.Integrations.ExifTool.Path)
	//
	//// Create thumbnail service
	//thumbnailService := imaging.NewThumbnailService(
	//	cfg.Media.Thumbnails.DefaultWidth,
	//	cfg.Media.Thumbnails.DefaultHeight,
	//	cfg.Media.Thumbnails.JPEGQuality,
	//	photoStorage, // implements ThumbnailStorage
	//	cfg.Media.Video.Enabled,
	//	cfg.Integrations.Ffmpeg.Path,
	//)

	// Create thumbnail service
	//thumbnailService := imaging.NewThumbnailService(
	//	cfg.Media.Thumbnails.DefaultWidth,
	//	cfg.Media.Thumbnails.DefaultHeight,
	//	cfg.Media.Thumbnails.JPEGQuality,
	//	photoStorage, // implements ThumbnailStorage
	//	cfg.Media.Video.Enabled,
	//	cfg.Integrations.Ffmpeg.Path,
	//)

	// Start background workers
	//go photoStorage.StartMaintenanceWorkers(
	//	cfg.Storage.Index.RebuildInterval,
	//	cfg.Storage.Index.SaveInterval,
	//	cfg.Storage.Index.IntegrityCheck,
	//)

	return photoStorage, nil
}

func createRouter(
	cfg *config.Config,
	assetHandler *handler.AssetHandler,
	albumHandler *handler.AlbumHandler,
	tripHandler *handler.TripHandler,
	searchHandler *handler.SearchHandler,
	systemHandler *handler.SystemHandler,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router with default middleware
	router := gin.Default()

	//// Apply CORS middleware if enabled
	//if cfg.Server.CORS.Enabled {
	//	router.Use(middleware.CORSMiddleware(
	//		cfg.Server.CORS.AllowOrigins,
	//		cfg.Server.CORS.AllowMethods,
	//		cfg.Server.CORS.AllowHeaders,
	//	))
	//}
	//
	//// Apply authentication middleware
	//router.Use(middleware.AuthMiddleware(cfg.Auth.JWT.Secret))
	//
	//// Apply rate limiting
	//if cfg.Auth.RateLimiting.Enabled {
	//	router.Use(middleware.RateLimitMiddleware(
	//		cfg.Auth.RateLimiting.Requests,
	//		cfg.Auth.RateLimiting.Burst,
	//	))
	//}

	// API routes
	api := router.Group("/api/v1")
	{
		// Asset routes
		api.POST("/assets", assetHandler.UploadAsset)
		api.GET("/assets/:id", assetHandler.GetAsset)
		api.POST("/assets/update", assetHandler.UpdateAssets)

		api.GET("/album/getList", albumHandler.GetList)
		api.POST("/album/create", albumHandler.Create)

		api.GET("/trip/getList", tripHandler.GetList)
		api.POST("/trip/create", tripHandler.Create)

		//api.PUT("/assets/:id", assetHandler.UpdateAssets)
		//api.DELETE("/assets/:id", assetHandler.DeleteAsset)
		//api.GET("/assets/:id/content", assetHandler.GetAssetContent)
		//api.GET("/assets/:id/thumbnail", assetHandler.GetAssetThumbnail)

		// Search routes
		api.GET("/search", searchHandler.SearchAssets)

		api.POST("/search/advanced", searchHandler.AdvancedSearch)
		api.POST("/search/advanced_v2", searchHandler.AdvancedSearchV2)

		//api.GET("/search/suggest", searchHandler.SuggestSearchTerms)

		// System routes
		api.GET("/system/status", systemHandler.GetSystemStatus)
		api.POST("/system/rebuild-index", systemHandler.RebuildIndex)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	//// Metrics endpoint
	//if cfg.Monitoring.Metrics.Enabled {
	//	router.GET(cfg.Monitoring.Metrics.Path, middleware.MetricsHandler())
	//}
	//
	//// Add request ID middleware and logging
	//router.Use(middleware.RequestIDMiddleware())
	//router.Use(middleware.LoggingMiddleware(cfg.Monitoring.Logging))

	return router
}

func startServer(cfg *config.Config, router *gin.Engine) {
	// Create HTTP server
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

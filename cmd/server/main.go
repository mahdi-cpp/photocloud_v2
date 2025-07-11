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

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	storageCfg := storage.Config{
		AppDir:        cfg.Storage.AppDir,
		AssetsDir:     cfg.Storage.AssetsDir,
		MetadataDir:   cfg.Storage.MetadataDir,
		ThumbnailsDir: cfg.Storage.ThumbnailsDir,
		IndexFile:     cfg.Storage.IndexFile,
		CacheSize:     cfg.Storage.Cache.Size,
	}

	userStorageManager, err := storage.NewUserStorageManager(storageCfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create handlers
	assetHandler := handler.NewAssetHandler(userStorageManager)
	searchHandler := handler.NewSearchHandler(userStorageManager)
	//systemHandler := handler.NewSystemHandler(userStorageManager)

	albumHandler := handler.NewAlbumHandler(userStorageManager)
	tripHandler := handler.NewTripHandler(userStorageManager)

	// Create Gin router
	router := createRouter(cfg, assetHandler, albumHandler, tripHandler, searchHandler)

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
	assetHandler *handler.AssetHandler,
	albumHandler *handler.AlbumHandler,
	tripHandler *handler.TripHandler,
	searchHandler *handler.SearchHandler,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router with default middleware
	router := gin.Default()

	// API routes
	api := router.Group("/api/v1")
	{
		// Search routes
		api.GET("/search", searchHandler.Search)
		api.POST("/search/filters", searchHandler.Filters)

		// Asset routes
		api.POST("/assets", assetHandler.Upload)
		api.GET("/assets/:id", assetHandler.Get)
		api.POST("/assets/update", assetHandler.Update)
		api.POST("/assets/delete", assetHandler.Delete)

		api.GET("/album/getList", albumHandler.GetList)
		api.POST("/album/create", albumHandler.Create)
		api.POST("/album/update", albumHandler.Update)
		api.POST("/album/delete", albumHandler.Delete)

		api.GET("/trip/getList", tripHandler.GetList)
		api.POST("/trip/create", tripHandler.Create)

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

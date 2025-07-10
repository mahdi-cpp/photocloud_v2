package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Media      MediaConfig      `mapstructure:"media"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Workers    WorkersConfig    `mapstructure:"workers"`
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("storage config: %w", err)
	}
	if err := c.Media.Validate(); err != nil {
		return fmt.Errorf("media config: %w", err)
	}
	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}
	return nil
}

// ServerConfig defines server settings
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"`
	GracefulTimeout time.Duration `mapstructure:"graceful_timeout"`
	MaxBodySize     string        `mapstructure:"max_body_size"`
	CORS            CORSConfig    `mapstructure:"cors"`
}

func (c *ServerConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("invalid port number")
	}
	if c.Mode != "debug" && c.Mode != "release" && c.Mode != "test" {
		return errors.New("invalid server mode, must be debug/release/test")
	}
	return nil
}

// CORSConfig defines CORS settings
type CORSConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	AllowOrigins []string `mapstructure:"allow_origins"`
	AllowMethods []string `mapstructure:"allow_methods"`
	AllowHeaders []string `mapstructure:"allow_headers"`
}

// StorageConfig defines storage settings
type StorageConfig struct {
	AppDir              string       `mapstructure:"app_dir"`
	AssetsDir           string       `mapstructure:"assets_dir"`
	MetadataDir         string       `mapstructure:"metadata_dir"`
	ThumbnailsDir       string       `mapstructure:"thumbnails_dir"`
	IndexFile           string       `mapstructure:"index_file"`
	AlbumCollectionFile string       `mapstructure:"albumCollection_file"`
	TripCollectionFile  string       `mapstructure:"tripCollection_file"`
	Cache               CacheConfig  `mapstructure:"cache"`
	Index               IndexConfig  `mapstructure:"index"`
	Upload              UploadConfig `mapstructure:"upload"`
}

func (c *StorageConfig) Validate() error {
	// Validate directories
	dirs := []struct {
		name string
		path string
	}{
		{"app_dir", c.AppDir},

		//{"assets_dir", c.AssetsDir},
		//{"metadata_dir", c.MetadataDir},
		//{"thumbnails_dir", c.ThumbnailsDir},
		//{"index_file", filepath.Dir(c.IndexFile)},
		//{"albumCollection_file", filepath.Dir(c.AlbumCollectionFile)},
		//{"tripCollection_file", filepath.Dir(c.TripCollectionFile)},
		//{"upload_temp_dir", c.Upload.TempDir},
	}

	for _, dir := range dirs {
		if dir.path == "" {
			return fmt.Errorf("%s must be specified", dir.name)
		}

		// Check if directory exists and is writable
		if err := checkDirectory(dir.path); err != nil {
			return fmt.Errorf("%s: %w", dir.name, err)
		}
	}

	// Validate cache settings
	if c.Cache.Size < 10 {
		return errors.New("cache size must be at least 10")
	}

	// Validate index intervals
	if c.Index.RebuildInterval < time.Hour {
		return errors.New("index rebuild interval must be at least 1 hour")
	}
	if c.Index.SaveInterval < time.Minute {
		return errors.New("index save interval must be at least 1 minute")
	}

	// Parse upload max size
	maxBytes, err := parseSize(c.Upload.MaxSize)
	if err != nil {
		return fmt.Errorf("invalid upload max_size: %w", err)
	}
	c.Upload.maxSizeBytes = maxBytes

	return nil
}

// GetUploadMaxBytes returns the parsed max upload size in bytes
func (c *StorageConfig) GetUploadMaxBytes() int64 {
	return c.Upload.maxSizeBytes
}

// CacheConfig defines cache settings
type CacheConfig struct {
	Size          int           `mapstructure:"size"`
	StatsInterval time.Duration `mapstructure:"stats_interval"`
}

// IndexConfig defines index management settings
type IndexConfig struct {
	RebuildInterval time.Duration `mapstructure:"rebuild_interval"`
	SaveInterval    time.Duration `mapstructure:"save_interval"`
	IntegrityCheck  time.Duration `mapstructure:"integrity_check"`
}

// UploadConfig defines upload handling settings
type UploadConfig struct {
	TempDir   string        `mapstructure:"temp_dir"`
	MaxSize   string        `mapstructure:"max_size"`
	Retention time.Duration `mapstructure:"retention"`

	// Computed values
	maxSizeBytes int64
}

// MediaConfig defines media processing settings
type MediaConfig struct {
	Thumbnails ThumbnailsConfig `mapstructure:"thumbnails"`
	Image      ImageConfig      `mapstructure:"image"`
	Video      VideoConfig      `mapstructure:"video"`
}

func (c *MediaConfig) Validate() error {
	if c.Thumbnails.JPEGQuality < 1 || c.Thumbnails.JPEGQuality > 100 {
		return errors.New("JPEG quality must be between 1-100")
	}
	if c.Thumbnails.PNGCompression < 0 || c.Thumbnails.PNGCompression > 9 {
		return errors.New("PNG compression must be between 0-9")
	}
	return nil
}

// ThumbnailsConfig defines thumbnail generation settings
type ThumbnailsConfig struct {
	DefaultWidth     int  `mapstructure:"default_width"`
	DefaultHeight    int  `mapstructure:"default_height"`
	JPEGQuality      int  `mapstructure:"jpeg_quality"`
	PNGCompression   int  `mapstructure:"png_compression"`
	GenerateOnUpload bool `mapstructure:"generate_on_upload"`
}

// ImageConfig defines image processing settings
type ImageConfig struct {
	MaxWidth  int `mapstructure:"max_width"`
	MaxHeight int `mapstructure:"max_height"`
}

// VideoConfig defines video processing settings
type VideoConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	MaxDuration     time.Duration `mapstructure:"max_duration"`
	ThumbnailOffset string        `mapstructure:"thumbnail_offset"`
}

// AuthConfig defines authentication settings
type AuthConfig struct {
	JWT          JWTConfig        `mapstructure:"jwt"`
	RateLimiting RateLimitConfig  `mapstructure:"rate_limiting"`
	Encryption   EncryptionConfig `mapstructure:"encryption"`
}

func (c *AuthConfig) Validate() error {
	if c.JWT.Secret == "" {
		return errors.New("JWT secret must be specified")
	}
	if len(c.JWT.Secret) < 32 {
		return errors.New("JWT secret must be at least 32 characters")
	}
	return nil
}

// JWTConfig defines JWT settings
type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	Issuer            string        `mapstructure:"issuer"`
	Expiration        time.Duration `mapstructure:"expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`
}

// RateLimitConfig defines rate limiting settings
type RateLimitConfig struct {
	Enabled  bool `mapstructure:"enabled"`
	Requests int  `mapstructure:"requests"` // requests per minute
	Burst    int  `mapstructure:"burst"`
}

// EncryptionConfig defines encryption settings
type EncryptionConfig struct {
	Assets bool   `mapstructure:"assets"`
	Key    string `mapstructure:"key"`
}

// MonitoringConfig defines monitoring settings
type MonitoringConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// MetricsConfig defines metrics settings
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// TracingConfig defines tracing settings
type TracingConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"` // Number of backups
	MaxAge     int    `mapstructure:"max_age"`     // Days
}

// WorkersConfig defines background workers settings
type WorkersConfig struct {
	ThumbnailGenerator WorkerConfig `mapstructure:"thumbnail_generator"`
	IndexMaintenance   WorkerConfig `mapstructure:"index_maintenance"`
}

// WorkerConfig defines a worker's settings
type WorkerConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	Concurrency int           `mapstructure:"concurrency"`
	BatchSize   int           `mapstructure:"batch_size"`
	Interval    time.Duration `mapstructure:"interval"`
}

// Helper functions

// parseSize converts size strings (e.g., "100MB") to bytes
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	if sizeStr == "" {
		return 0, errors.New("empty size string")
	}

	// Extract numeric part and unit
	var num int64
	var unit string
	_, err := fmt.Sscanf(sizeStr, "%d%s", &num, &unit)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %w", err)
	}

	// Convert to bytes
	switch unit {
	case "B", "":
		return num, nil
	case "KB":
		return num * 1024, nil
	case "MB":
		return num * 1024 * 1024, nil
	case "GB":
		return num * 1024 * 1024 * 1024, nil
	case "TB":
		return num * 1024 * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
}

// checkDirectory verifies a directory exists and is writable
func checkDirectory(path string) error {
	// Check if directory exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}

	// Check it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check write permission
	file, err := os.CreateTemp(path, "write-test-")
	if err != nil {
		return fmt.Errorf("directory not writable: %w", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	return nil
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	// Set environment variable prefix
	v.SetEnvPrefix("PHOTOCLOUD")
	v.AutomaticEnv()

	// Set default values
	setDefaults(v)

	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults defines default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.graceful_timeout", "30s")
	v.SetDefault("server.max_body_size", "100MB")
	v.SetDefault("server.cors.enabled", true)
	v.SetDefault("server.cors.allow_origins", []string{"*"})
	v.SetDefault("server.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("server.cors.allow_headers", []string{"Origin", "Content-Type", "Authorization"})

	// Storage defaults
	v.SetDefault("storage.assets_dir", "/var/photocloud/assets")
	v.SetDefault("storage.metadata_dir", "/var/photocloud/metadata")
	v.SetDefault("storage.thumbnails_dir", "/var/photocloud/thumbnails")
	v.SetDefault("storage.index_file", "/var/photocloud/index.dat")
	v.SetDefault("storage.cache.size", 1000)
	v.SetDefault("storage.cache.stats_interval", "5m")
	v.SetDefault("storage.index.rebuild_interval", "24h")
	v.SetDefault("storage.index.save_interval", "5m")
	v.SetDefault("storage.index.integrity_check", "12h")
	v.SetDefault("storage.upload.temp_dir", "/tmp/photocloud/uploads")
	v.SetDefault("storage.upload.max_size", "500MB")
	v.SetDefault("storage.upload.retention", "1h")

	// Media defaults
	v.SetDefault("media.thumbnails.default_width", 300)
	v.SetDefault("media.thumbnails.default_height", 300)
	v.SetDefault("media.thumbnails.jpeg_quality", 85)
	v.SetDefault("media.thumbnails.png_compression", 6)
	v.SetDefault("media.thumbnails.generate_on_upload", true)
	v.SetDefault("media.image.max_width", 10000)
	v.SetDefault("media.image.max_height", 10000)
	v.SetDefault("media.video.enabled", true)
	v.SetDefault("media.video.max_duration", "10m")
	v.SetDefault("media.video.thumbnail_offset", "00:00:05")

	// Auth defaults
	v.SetDefault("auth.jwt.secret", "change-me-to-secure-random-value")
	v.SetDefault("auth.jwt.issuer", "photocloud")
	v.SetDefault("auth.jwt.expiration", "720h")          // 30 days
	v.SetDefault("auth.jwt.refresh_expiration", "2160h") // 90 days
	v.SetDefault("auth.rate_limiting.enabled", true)
	v.SetDefault("auth.rate_limiting.requests", 100) // per minute
	v.SetDefault("auth.rate_limiting.burst", 25)
	v.SetDefault("auth.encryption.assets", false)
	v.SetDefault("auth.encryption.key", "")

	// Monitoring defaults
	v.SetDefault("monitoring.metrics.enabled", true)
	v.SetDefault("monitoring.metrics.path", "/metrics")
	v.SetDefault("monitoring.tracing.enabled", false)
	v.SetDefault("monitoring.tracing.endpoint", "localhost:4317")
	v.SetDefault("monitoring.logging.level", "info")
	v.SetDefault("monitoring.logging.format", "json")
	v.SetDefault("monitoring.logging.file", "/var/log/photocloud/app.log")
	v.SetDefault("monitoring.logging.max_size", 100) // MB
	v.SetDefault("monitoring.logging.max_backups", 3)
	v.SetDefault("monitoring.logging.max_age", 30) // days

	// Workers defaults
	v.SetDefault("workers.thumbnail_generator.enabled", true)
	v.SetDefault("workers.thumbnail_generator.concurrency", 4)
	v.SetDefault("workers.thumbnail_generator.batch_size", 10)
	v.SetDefault("workers.thumbnail_generator.interval", "5m")
	v.SetDefault("workers.index_maintenance.enabled", true)
	v.SetDefault("workers.index_maintenance.interval", "1h")
}

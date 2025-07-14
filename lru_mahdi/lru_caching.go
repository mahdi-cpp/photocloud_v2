package lru_mahdi

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

// Configurable paths (should be set via initialization)
var (
	ImageFolders []string
	IconFolder   string
)

type ImageRepository struct {
	imageCache *lru.Cache[string, []byte] // LRU cache for images
	iconCache  *lru.Cache[string, []byte] // Separate cache for icons
	mu         sync.RWMutex               // Protects cache initialization
}

func InitializePaths(folders []string, iconPath string) {
	ImageFolders = folders
	IconFolder = iconPath
}

func NewImageRepository(maxImages, maxIcons int) (*ImageRepository, error) {
	r := &ImageRepository{}

	// Create LRU caches with size limits
	var err error
	r.imageCache, err = lru.New[string, []byte](maxImages)
	if err != nil {
		return nil, fmt.Errorf("failed to create image cache: %w", err)
	}

	r.iconCache, err = lru.New[string, []byte](maxIcons)
	if err != nil {
		return nil, fmt.Errorf("failed to create icon cache: %w", err)
	}

	return r, nil
}

func (r *ImageRepository) GetImage(fullPath string) ([]byte, error) {
	// Check cache first
	if data, ok := r.imageCache.Get(fullPath); ok {
		fmt.Println("image of cache")
		return data, nil
	}

	// Load and cache on miss
	data, err := r.loadAndCacheImage(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}
	return data, nil
}

func (r *ImageRepository) GetIcon(iconName string) ([]byte, error) {
	// Check cache first
	if data, ok := r.iconCache.Get(iconName); ok {
		return data, nil
	}

	// Load and cache on miss
	data, err := r.loadAndCacheIcon(iconName)
	if err != nil {
		return nil, fmt.Errorf("failed to load icon: %w", err)
	}
	return data, nil
}

func (r *ImageRepository) loadAndCacheImage(fullPath string) ([]byte, error) {

	// Validate path security
	if !isSafePath(fullPath, ImageFolders) {
		return nil, errors.New("invalid file path")
	}

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Encode to bytes
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, nil); err != nil {
			return nil, err
		}
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	data := buf.Bytes()

	// Cache the result
	r.imageCache.Add(fullPath, data)
	return data, nil
}

func (r *ImageRepository) loadAndCacheIcon(iconName string) ([]byte, error) {
	iconPath := filepath.Join(IconFolder, iconName)

	// Validate path security
	if !isSafePath(iconPath, []string{IconFolder}) {
		return nil, errors.New("invalid icon path")
	}

	// Open file
	file, err := os.Open(iconPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read directly for icons (no processing needed)
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read icon: %w", err)
	}

	// Cache the result
	r.iconCache.Add(iconName, data)
	return data, nil
}

// CacheMetrics --------------------------------
type CacheMetrics struct {
	ImageHits   uint64
	ImageMisses uint64
	IconHits    uint64
	IconMisses  uint64
	LastReset   time.Time
}

func isSafePath(path string, allowedPaths []string) bool {
	for _, safePath := range allowedPaths {
		if strings.HasPrefix(filepath.Clean(path), filepath.Clean(safePath)) {
			return true
		}
	}
	return false
}

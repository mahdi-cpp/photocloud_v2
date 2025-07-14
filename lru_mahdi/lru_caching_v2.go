package lru_mahdi

//https://chat.deepseek.com/a/chat/s/5f0242a7-4635-4f65-9841-5e2a1d67920c

import (
	"bytes"
	"context"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ImageLoader handles loading and processing images
type ImageLoader struct {
	cache         *lru.Cache[string, *imageItem]
	localBasePath string
	httpClient    *http.Client
	generateLock  sync.Mutex
	metrics       LoaderMetrics
}

// New struct to store image data with access time
type imageItem struct {
	data       []byte
	lastAccess int64 // UnixNano timestamp
	size       int64 // Only for metrics tracking
}

type LoaderMetrics struct {
	mu                sync.Mutex
	FileLoads         int
	NetworkLoads      int
	Generations       int
	LoadErrors        int
	expiredItems      int
	evictedItems      int
	CurrentCacheBytes int64 // Current memory used by cache
	totalOriginal     int64 // Total original bytes processed
	totalFinal        int64 // Total final bytes processed

	loadDurations []time.Duration
}

func NewImageLoader(count int, localBasePath string) *ImageLoader {

	cache, _ := lru.New[string, *imageItem](count)

	loader := &ImageLoader{
		cache:         cache,
		localBasePath: localBasePath,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 10,
			},
		},
	}

	// Start background cleaner (runs every 60 minute, expires after 60 minutes)
	go loader.StartCleaner(time.Second, 10*time.Second)

	return loader
}

// LoadImage handles all image requests (main entry point)
func (il *ImageLoader) LoadImage(ctx context.Context, imageID string) ([]byte, error) {

	start := time.Now()

	//// Check cache first
	//if data, ok := il.cache.Get(imageID); ok {
	//	fmt.Println("Check cache first")
	//	return data, nil
	//}

	if item, ok := il.cache.Get(imageID); ok {
		return item.data, nil

		//if time.Since(time.Unix(0, item.lastAccess)) <= 30*time.Minute {
		//	// Update access time and return
		//	atomic.StoreInt64(&item.lastAccess, time.Now().UnixNano())
		//
		//}
		//
		//// Remove expired item
		//il.cache.Remove(imageID)
		//il.metrics.mu.Lock()
		//il.metrics.CurrentCacheBytes -= item.size
		//il.metrics.mu.Unlock()
	}

	// Determine source and load
	var data []byte
	var err error

	switch {
	case strings.HasPrefix(imageID, "http://") || strings.HasPrefix(imageID, "https://"):
		data, err = il.loadNetworkImage(ctx, imageID)
	case strings.HasPrefix(imageID, "gen:"):
		//data, err = il.generateImage(ctx, imageID)
	case strings.HasPrefix(imageID, "placeholder:"):
		//data, err = il.createPlaceholder(ctx, imageID)
	default:
		data, err = il.loadLocalImage(imageID)
	}

	// Track metrics
	duration := time.Since(start)
	fmt.Println(duration)

	il.metrics.mu.Lock()
	//il.metrics.loadDurations = append(il.metrics.loadDurations, duration)
	if err != nil {
		il.metrics.LoadErrors++
	} else {
		// After loading image, add to cache
		newItem := &imageItem{
			data:       data,
			lastAccess: time.Now().UnixNano(),
			size:       int64(len(data)),
		}

		il.cache.Add(imageID, newItem)

		// Update current cache size
		il.metrics.CurrentCacheBytes += newItem.size
	}
	il.metrics.mu.Unlock()

	return data, err
}

// StartCleaner background cleaner routine
func (il *ImageLoader) StartCleaner(interval, ttl time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		il.cleanExpired(ttl)
	}
}

func (il *ImageLoader) cleanExpired(ttl time.Duration) {
	keys := il.cache.Keys()
	now := time.Now()

	for _, key := range keys {
		if item, ok := il.cache.Peek(key); ok {

			if now.Sub(time.Unix(0, item.lastAccess)) > ttl {

				// Get size before removal
				sizeBefore := atomic.LoadInt64(&il.metrics.CurrentCacheBytes)
				//finalSize := int64(len(item.data))

				il.cache.Remove(key)

				il.metrics.mu.Lock()
				il.metrics.expiredItems++
				il.metrics.CurrentCacheBytes -= item.size
				sizeAfter := il.metrics.CurrentCacheBytes
				il.metrics.mu.Unlock()

				fmt.Printf("Expired item removed: %s\n", key)
				fmt.Printf("Size: %s\n", formatBytes(item.size))
				fmt.Printf("Cache memory: %s → %s\n", formatBytes(sizeBefore), formatBytes(sizeAfter))
			}
		}
	}
}

func (il *ImageLoader) loadLocalImage(imageID string) ([]byte, error) {

	path := filepath.Join(il.localBasePath, imageID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Get original size BEFORE conversion
	originalSize := len(data)
	var result []byte

	// Validate image format
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid image format for %s: %w", imageID, err)
	}

	// Convert to standard format if needed
	switch strings.ToLower(format) {
	case "jpeg", "jpg", "png":
		result = data
	default:
		result, err = il.convertImageFormat(data, "jpeg")
		if err != nil {
			return nil, err
		}
	}

	// Get final size AFTER conversion
	finalSize := len(result)

	// Update metrics
	il.metrics.mu.Lock()
	il.metrics.FileLoads++
	il.metrics.totalOriginal += int64(originalSize)
	il.metrics.totalFinal += int64(finalSize)
	il.metrics.mu.Unlock()

	return result, nil
}

// loadNetworkImage downloads from URL
func (il *ImageLoader) loadNetworkImage(ctx context.Context, url string) ([]byte, error) {
	il.metrics.mu.Lock()
	il.metrics.NetworkLoads++
	il.metrics.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp, err := il.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Validate image
	if _, _, err := image.DecodeConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("invalid image from %s: %w", url, err)
	}

	return data, nil
}

// convertImageFormat converts between image formats
func (il *ImageLoader) convertImageFormat(data []byte, targetFormat string) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	switch strings.ToLower(targetFormat) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(&buf, img)
	default:
		return nil, fmt.Errorf("unsupported format: %s", targetFormat)
	}

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (il *ImageLoader) Metrics() *LoaderMetrics {

	il.metrics.mu.Lock()
	defer il.metrics.mu.Unlock()

	// Calculate compression ratio
	if il.metrics.CurrentCacheBytes > 0 {
		//compressionRatio = float64(il.metrics.localBytesFinal) / float64(il.metrics.localBytesOriginal)
	}

	// Calculate average duration
	totalDuration := time.Duration(0)
	for _, d := range il.metrics.loadDurations {
		totalDuration += d
	}

	//if count := len(il.metrics.loadDurations); count > 0 {
	//	avgDuration = totalDuration / time.Duration(count)
	//}

	return &il.metrics
}

//// Usage Example
//func main() {
//
//	// Create cache (1000 items capacity)
//	cache, _ := lru.New[string, []byte](1000)
//
//	// Initialize loader (with local image directory)
//	loader := NewImageLoader(cache, "")
//
//	// Load various image types
//	images := []string{
//		//"/var/cloud/upload/upload5/20190809_000407.jpg",
//		//"Screenshot_20240113_180718_Instagram.jpg",
//		//"Screenshot_20240120_020041_Instagram.jpg",
//		"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/18.jpg",
//		"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/17.jpg",
//		"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/25.jpg",
//		"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/26.jpg",
//		"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/27.jpg",
//		"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets/28.jpg",
//
//		//"https://mahdiali.s3.ir-thr-at1.arvanstorage.ir/%D9%86%D9%82%D8%B4%D9%87-%D8%AA%D8%A7%DB%8C%D9%85%D8%B1-%D8%B1%D8%A7%D9%87-%D9%BE%D9%84%D9%87-%D8%B3%D9%87-%D8%B3%DB%8C%D9%85.jpg?versionId=", // Network URL
//		//"https://mahdicpp.s3.ir-thr-at1.arvanstorage.ir/0f470b87c13e25bc4211683711e71e2a.jpg?versionId=",
//	}
//
//	ctx := context.Background()
//	for _, img := range images {
//		data, err := loader.LoadImage(ctx, img)
//		if err != nil {
//			log.Printf("Failed to load %s: %v", img, err)
//			continue
//		}
//		fmt.Printf("Loaded %s (%d kB)\n", img, len(data)/1024)
//	}
//
//	// Print metrics
//	f, n, g, e, avg := loader.Metrics()
//	fmt.Printf("\nLoader Metrics:\n")
//	fmt.Printf("File loads: %d\n", f)
//	fmt.Printf("Network loads: %d\n", n)
//	fmt.Printf("Generated images: %d\n", g)
//	fmt.Printf("Errors: %d\n", e)
//	fmt.Printf("Avg load time: %s\n", avg)
//}

// formatBytes converts bytes to human-readable string
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

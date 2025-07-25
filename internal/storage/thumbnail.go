package storage

import (
	"bytes"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/pkg/asset_model"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"sync"

	"github.com/disintegration/imaging"
)

// ThumbnailService handles thumbnail generation
type ThumbnailService struct {
	defaultWidth   int
	defaultHeight  int
	jpegQuality    int
	pngCompression int
	storage        ThumbnailStorage
	videoEnabled   bool
	ffmpegPath     string
	mu             sync.Mutex
}

// NewThumbnailService creates a new thumbnail service
func NewThumbnailService(
	defaultWidth, defaultHeight int,
	jpegQuality int,
	storage ThumbnailStorage,
	videoEnabled bool,
	ffmpegPath string,
) *ThumbnailService {
	return &ThumbnailService{
		defaultWidth:  defaultWidth,
		defaultHeight: defaultHeight,
		jpegQuality:   jpegQuality,
		storage:       storage,
		videoEnabled:  videoEnabled,
		ffmpegPath:    ffmpegPath,
	}
}

// GenerateThumbnail generates a thumbnail for an asset
func (s *ThumbnailService) GenerateThumbnail(asset *asset_model.PHAsset, content []byte) ([]byte, error) {
	switch asset.MediaType {
	case asset_model.ImageTypeJPEG, asset_model.ImageTypePNG, asset_model.ImageTypeGIF:
		return s.generateImageThumbnail(content, asset.MediaType)
	case asset_model.VideoTypeMP4, asset_model.VideoTypeMOV:
		if s.videoEnabled {
			return s.generateVideoThumbnail(asset, content)
		}
		return nil, fmt.Errorf("video processing disabled")
	default:
		return nil, fmt.Errorf("unsupported media type: %s", asset.MediaType)
	}
}

// generateImageThumbnail creates a thumbnail from image content
func (s *ThumbnailService) generateImageThumbnail(content []byte, mediaType asset_model.MediaType) ([]byte, error) {
	// Decode image
	img, err := imaging.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image using Lanczos filter for high quality
	dst := imaging.Resize(img, s.defaultWidth, s.defaultHeight, imaging.Lanczos)

	// Encode to buffer
	var buf bytes.Buffer
	switch mediaType {
	case asset_model.ImageTypePNG:
		err = png.Encode(&buf, dst)
	default:
		// Default to JPEG for all other types
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: s.jpegQuality})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// generateVideoThumbnail creates a thumbnail from video content
func (s *ThumbnailService) generateVideoThumbnail(asset *asset_model.PHAsset, content []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Handler temp file
	tmpFile, err := os.CreateTemp("", "video-*.mp4")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write video content
	if _, err := tmpFile.Write(content); err != nil {
		return nil, fmt.Errorf("failed to write video content: %w", err)
	}

	// Handler output buffer
	buf := bytes.NewBuffer(nil)

	// Use ffmpeg to extract thumbnail
	ffmpegArgs := []string{
		"-i", tmpFile.Name(),
		"-ss", "00:00:01", // Capture at 1 second
		"-vframes", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-",
	}

	cmd := exec.Command(s.ffmpegPath, ffmpegArgs...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr // Capture ffmpeg errors

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w", err)
	}

	// If we got no output, fail
	if buf.Len() == 0 {
		return nil, fmt.Errorf("no thumbnail generated from video")
	}

	// Resize the generated thumbnail
	return s.resizeThumbnail(buf.Bytes())
}

// resizeThumbnail resizes an existing thumbnail image
func (s *ThumbnailService) resizeThumbnail(thumbData []byte) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(thumbData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode thumbnail: %w", err)
	}

	// Resize to desired dimensions
	resized := imaging.Resize(img, s.defaultWidth, s.defaultHeight, imaging.Lanczos)

	// Encode as JPEG
	buf := bytes.NewBuffer(nil)
	if err := jpeg.Encode(buf, resized, &jpeg.Options{Quality: s.jpegQuality}); err != nil {
		return nil, fmt.Errorf("failed to encode resized thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// GetThumbnail retrieves or generates a thumbnail for an asset
func (s *ThumbnailService) GetThumbnail(
	assetID int,
	asset *asset_model.PHAsset,
	width, height int,
) ([]byte, error) {
	// Try to get from storage
	if thumb, err := s.storage.GetThumbnail(assetID, width, height); err == nil {
		return thumb, nil
	}

	// Get asset content
	content, err := s.storage.GetAssetContent(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset content: %w", err)
	}

	// Generate thumbnail
	thumbData, err := s.GenerateThumbnail(asset, content)
	if err != nil {
		return nil, fmt.Errorf("thumbnail generation failed: %w", err)
	}

	// Save to storage
	if err := s.storage.SaveThumbnail(assetID, width, height, thumbData); err != nil {
		// Log error but return the thumbnail
		fmt.Printf("Failed to save thumbnail: %v\n", err)
	}

	return thumbData, nil
}

// ProcessUpload generates thumbnails during upload
func (s *ThumbnailService) ProcessUpload(
	file multipart.File,
	header *multipart.FileHeader,
	asset *asset_model.PHAsset,
) ([]byte, error) {
	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	// Reset file reader for storage
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Generate thumbnail
	return s.GenerateThumbnail(asset, content)
}

// GenerateMissingThumbnails processes assets without thumbnails
func (s *ThumbnailService) GenerateMissingThumbnails() (int, error) {
	assetIDs, err := s.storage.GetAssetsWithoutThumbnails()
	if err != nil {
		return 0, fmt.Errorf("failed to get assets without thumbnails: %w", err)
	}

	successCount := 0
	for _, id := range assetIDs {
		asset, err := s.storage.GetAsset(id)
		if err != nil {
			continue // Skip if asset not found
		}

		content, err := s.storage.GetAssetContent(id)
		if err != nil {
			continue // Skip if content missing
		}

		if _, err := s.GenerateThumbnail(asset, content); err == nil {
			successCount++
		}
	}

	return successCount, nil
}

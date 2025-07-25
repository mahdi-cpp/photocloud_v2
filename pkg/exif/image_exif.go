package asset_create

import (
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ImageExtractor extracts metadata from media files
type ImageExtractor struct {
	exifToolPath string
}

// NewMetadataExtractor creates a new metadata extractor
func NewMetadataExtractor(exifToolPath string) *ImageExtractor {
	return &ImageExtractor{exifToolPath: exifToolPath}
}

// ExtractMetadata extracts metadata from a file
func (e *ImageExtractor) ExtractMetadata(filePath string) (width, height int, camera string, err error) {
	// First try with exifTool if available
	if e.exifToolPath != "" {
		if width, height, camera, err = e.extractWithExifTool(filePath); err == nil {
			return width, height, camera, nil
		}
	}

	// Fallback to basic image decoding
	return e.extractBasicMetadata(filePath)
}

// extractWithExifTool uses exiftool for metadata extraction
func (e *ImageExtractor) extractWithExifTool(filePath string) (int, int, string, error) {
	cmd := exec.Command(e.exifToolPath,
		"-ImageWidth",
		"-ImageHeight",
		"-Model",
		"-T", // Tab separated output
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, "", fmt.Errorf("exiftool failed: %w", err)
	}

	// Parse output: "width\theight\tmodel"
	parts := strings.Split(strings.TrimSpace(string(output)), "\t")
	if len(parts) < 3 {
		return 0, 0, "", fmt.Errorf("unexpected exiftool output")
	}

	width, _ := strconv.Atoi(parts[0])
	height, _ := strconv.Atoi(parts[1])
	camera := parts[2]

	return width, height, camera, nil
}

// extractBasicMetadata uses standard image decoding
func (e *ImageExtractor) extractBasicMetadata(filePath string) (int, int, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to decode image: %w", err)
	}

	return config.Width, config.Height, format, nil
}

// GetMediaType determines media type from filename
func GetMediaType(filename string) happle_models.MediaType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return happle_models.ImageTypeJPEG
	case ".png":
		return happle_models.ImageTypePNG
	case ".gif":
		return happle_models.ImageTypeGIF
	case ".mp4":
		return happle_models.VideoTypeMP4
	case ".mov":
		return happle_models.VideoTypeMOV
	default:
		return happle_models.UnknownType
	}
}

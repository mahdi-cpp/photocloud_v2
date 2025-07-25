package asset_create

import (
	"bytes"
	"fmt"
	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"strings"
	"time"
)

func GetCaptureDate(filePath string) (string, error) {

	// Extract raw EXIF data
	rawExif, err := exif.SearchFileAndExtractExif(filePath)
	if err != nil {
		return "", fmt.Errorf("error extracting EXIF: %w", err)
	}

	// Initialize mapping and tag index
	im := exifcommon.NewIfdMapping()
	if err := exifcommon.LoadStandardIfds(im); err != nil {
		return "", fmt.Errorf("error loading IFDs: %w", err)
	}
	ti := exif.NewTagIndex()

	// Parse EXIF data
	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return "", fmt.Errorf("error parsing EXIF: %w", err)
	}

	// Get root IFD (IFD0)
	rootIfd := index.RootIfd
	if rootIfd == nil {
		return "", fmt.Errorf("root IFD not found")
	}

	// Get EXIF SubIFD (where DateTimeOriginal lives)
	exifIfd, err := rootIfd.ChildWithIfdPath(exifcommon.IfdExifStandardIfdIdentity)
	if err != nil {
		return "", fmt.Errorf("EXIF sub-IFD not found: %w", err)
	}

	// Find DateTimeOriginal tag (0x9003)
	tagId := uint16(0x9003)
	tag, err := exifIfd.FindTagWithId(tagId)
	if err != nil {
		return "", fmt.Errorf("DateTimeOriginal tag not found: %w", err)
	}

	// Get tag value
	value, err := tag[0].Value()
	if err != nil {
		return "", fmt.Errorf("error reading tag value: %w", err)
	}

	// Type assertion to string
	captureDate, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("DateTimeOriginal is not a string")
	}

	return captureDate, nil
}

func GetCaptureTime(filePath string) (time.Time, error) {

	// Extract raw EXIF data
	rawExif, err := exif.SearchFileAndExtractExif(filePath)
	if err != nil {
		return time.Time{}, fmt.Errorf("error extracting EXIF: %w", err)
	}

	// Initialize mapping and tag index
	im := exifcommon.NewIfdMapping()
	if err := exifcommon.LoadStandardIfds(im); err != nil {
		return time.Time{}, fmt.Errorf("error loading IFDs: %w", err)
	}
	ti := exif.NewTagIndex()

	// Parse EXIF data
	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing EXIF: %w", err)
	}

	// Get root IFD
	rootIfd := index.RootIfd
	if rootIfd == nil {
		return time.Time{}, fmt.Errorf("root IFD not found")
	}

	// Get EXIF SubIFD
	exifIfd, err := rootIfd.ChildWithIfdPath(exifcommon.IfdExifStandardIfdIdentity)
	if err != nil {
		return time.Time{}, fmt.Errorf("EXIF sub-IFD not found: %w", err)
	}

	// Find DateTimeOriginal tag (0x9003)
	tagId := uint16(0x9003)
	tag, err := exifIfd.FindTagWithId(tagId)
	if err != nil {
		return time.Time{}, fmt.Errorf("DateTimeOriginal tag not found: %w", err)
	}

	// Get tag value
	value, err := tag[0].Value()
	if err != nil {
		return time.Time{}, fmt.Errorf("error reading tag value: %w", err)
	}

	// Type assertion to string
	dateStr, ok := value.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("DateTimeOriginal is not a string")
	}

	// Parse to time.Time using EXIF format
	return parseExifDateTime(dateStr)
}

// parseExifDateTime converts EXIF date string to time.Time
func parseExifDateTime(dateStr string) (time.Time, error) {
	// EXIF format: "2006:01:02 15:04:05"
	// Some implementations include null terminators - trim them
	cleanStr := trimNulls(dateStr)

	// Parse with EXIF layout
	t, err := time.Parse("2006:01:02 15:04:05", cleanStr)
	if err == nil {
		return t, nil
	}

	// Try alternative layouts if standard format fails
	layouts := []string{
		"2006:01:02 15:04:05 -0700", // With timezone
		"2006:01:02 15:04:05",       // Standard
		"2006-01-02 15:04:05",       // Hyphen-separated date
		"2006:01:02",                // Date only
	}

	for _, layout := range layouts {
		t, err = time.Parse(layout, cleanStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse date: %q", dateStr)
}

// trimNulls removes null terminators and extra spaces
func trimNulls(s string) string {
	// Trim trailing null bytes and spaces
	return trimSpace(string(bytes.TrimRight([]byte(s), "\x00")))
}

// trimSpace handles multi-space trimming
func trimSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

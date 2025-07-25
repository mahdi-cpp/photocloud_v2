package asset_create

import (
	"fmt"
	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"time"
)

// ImageMetadata holds all camera-related EXIF data
type ImageMetadata struct {
	CameraMake  string
	CameraModel string
	CaptureTime time.Time
}

func GetImageMetadata(filePath string) (ImageMetadata, error) {

	result := ImageMetadata{}

	// Extract raw EXIF data
	rawExif, err := exif.SearchFileAndExtractExif(filePath)
	if err != nil {
		return result, fmt.Errorf("error extracting EXIF: %w", err)
	}

	// Initialize mapping and tag index
	im := exifcommon.NewIfdMapping()
	if err := exifcommon.LoadStandardIfds(im); err != nil {
		return result, fmt.Errorf("error loading IFDs: %w", err)
	}
	ti := exif.NewTagIndex()

	// Parse EXIF data
	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return result, fmt.Errorf("error parsing EXIF: %w", err)
	}

	// Get root IFD (where CameraMake/CameraModel are stored)
	rootIfd := index.RootIfd
	if rootIfd == nil {
		return result, fmt.Errorf("root IFD not found")
	}

	// Get Camera CameraMake (tag 0x010F)
	if makeTag, err := rootIfd.FindTagWithId(0x010F); err == nil {
		if makeVal, err := makeTag[0].Value(); err == nil {
			if s, ok := makeVal.(string); ok {
				result.CameraMake = trimNulls(s)
			}
		}
	}

	// Get Camera CameraModel (tag 0x0110)
	if modelTag, err := rootIfd.FindTagWithId(0x0110); err == nil {
		if modelVal, err := modelTag[0].Value(); err == nil {
			if s, ok := modelVal.(string); ok {
				result.CameraModel = trimNulls(s)
			}
		}
	}

	// Get Capture Time (from Exif SubIFD)
	if exifIfd, err := rootIfd.ChildWithIfdPath(exifcommon.IfdExifStandardIfdIdentity); err == nil {
		if dtTag, err := exifIfd.FindTagWithId(0x9003); err == nil { // DateTimeOriginal
			if dtVal, err := dtTag[0].Value(); err == nil {
				if s, ok := dtVal.(string); ok {
					result.CaptureTime, _ = parseExifDateTime(trimNulls(s))
				}
			}
		}
	}

	return result, nil
}

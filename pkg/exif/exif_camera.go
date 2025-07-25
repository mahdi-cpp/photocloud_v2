package asset_create

import (
	"github.com/dsoprea/go-exif/v3"
	"io/ioutil"
	"strings"
)

// GetCameraModel returns the camera make and model from the EXIF data of an image file
// Returns:
// - make: Camera manufacturer (empty string if not found)
// - model: Camera model (empty string if not found)
// - err: Error if any occurred
func GetCameraModel(filepath string) (make, model string, err error) {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", "", err
	}

	// Parse the EXIF data
	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		return "", "", err
	}

	// Get the EXIF tags
	tags, _, err := exif.GetFlatExifData(rawExif, nil)
	if err != nil {
		return "", "", err
	}

	// Function to find a tag by name
	findTag := func(name string) (value string, found bool) {
		for _, tag := range tags {
			if tag.TagName == name {
				if tag.Value != nil {
					return tag.Value.(string), true
				}
				return "", true
			}
		}
		return "", false
	}

	// Get camera make and model
	make, _ = findTag("CameraMake")
	model, _ = findTag("CameraModel")

	make = sanitizeString(make)
	model = sanitizeString(model)

	return make, model, nil
}

func sanitizeString(s string) string {
	// Remove non-printable characters
	s = strings.Map(func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}, s)
	return strings.TrimSpace(s)
}

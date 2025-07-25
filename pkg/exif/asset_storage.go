package asset_create

import (
	"encoding/json"
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/pkg/asset_model"
	"io"
	"os"
	"path/filepath"
)

// Ensure directory exists
func init() {
	//if _, err := os.Stat(rootPath); os.IsNotExist(err) {
	//	os.MkdirAll(rootPath, 0755)
	//}
}

func GetMetadataPath(id int) string {
	return filepath.Join(AppDir, username, MetadataDir, fmt.Sprintf("%d.json", id))
}

// SaveAssetMetadata saves a PHAsset to a JSON file
func SaveAssetMetadata(asset asset_model.PHAsset) error {

	// Create filename based on ID and creation date
	//filename := filepath.Join(AppDir+username+MetadataDir, asset.ID+".json")
	filename := GetMetadataPath(asset.ID)

	// Convert to JSON
	data, err := json.MarshalIndent(asset, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filename, data, 0644)
}

func CopyFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

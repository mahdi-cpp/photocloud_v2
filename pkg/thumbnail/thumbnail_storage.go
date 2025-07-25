package thumbnail

import "github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"

// ThumbnailStorage defines thumbnail persistence operations
type ThumbnailStorage interface {
	SaveThumbnail(assetID, width, height int, data []byte) error
	GetThumbnail(assetID, width, height int) ([]byte, error)
	GetAssetsWithoutThumbnails() ([]int, error)
	GetAsset(assetID int) (*happle_models.PHAsset, error)
	GetAssetContent(assetID int) ([]byte, error)
}

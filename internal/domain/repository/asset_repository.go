package repository

import (
	"context"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"mime/multipart"
)

// AssetRepository defines the interface for asset persistence operations
type AssetRepository interface {

	// Asset operations
	CreateAsset(ctx context.Context, asset *model.PHAsset, file multipart.File, header *multipart.FileHeader) (*model.PHAsset, error)
	GetAsset(ctx context.Context, assetID int) (*model.PHAsset, error)
	GetAssetContent(ctx context.Context, assetID int) ([]byte, error)
	UpdateAsset(ctx context.Context, assetID int, update *model.AssetUpdate) (*model.PHAsset, error)
	DeleteAsset(ctx context.Context, assetID int) error
	GetAssetThumbnail(ctx context.Context, assetID int, width, height int) ([]byte, error)

	// Batch operations
	GetAssetsByUser(ctx context.Context, userID int, limit, offset int) ([]*model.PHAsset, int, error)
	GetRecentAssets(ctx context.Context, userID int, days int) ([]*model.PHAsset, error)
	CountUserAssets(ctx context.Context, userID int) (int, error)

	// Search operations
	SearchAssets(ctx context.Context, filters model.AssetSearchFilters) ([]*model.PHAsset, int, error)
	SuggestSearchTerms(ctx context.Context, userID int, prefix string, limit int) ([]string, error)

	// System operations
	RebuildIndex(ctx context.Context) error
	GetStorageStats(ctx context.Context) (*model.StorageStats, error)
	GetIndexStatus(ctx context.Context) (*model.IndexStatus, error)

	// Maintenance operations
	DeleteOrphanedAssets(ctx context.Context) (int, error)
	GenerateMissingThumbnails(ctx context.Context) (int, error)
	CleanupExpiredUploads(ctx context.Context) error
}

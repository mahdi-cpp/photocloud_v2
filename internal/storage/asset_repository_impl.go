package storage

import (
	"context"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"mime/multipart"
	"time"
)

// AssetRepositoryImpl implements the AssetRepository interface
type AssetRepositoryImpl struct {
	storage *PhotoStorage
}

// NewAssetRepository creates a new repository instance
func NewAssetRepository(storage *PhotoStorage) *AssetRepositoryImpl {
	return &AssetRepositoryImpl{storage: storage}
}

func (r *AssetRepositoryImpl) UploadAsset(ctx context.Context, asset *model.PHAsset, file multipart.File, header *multipart.FileHeader) (*model.PHAsset, error) {

	// Use the storage to upload the asset
	createdAsset, err := r.storage.UploadAsset(asset.UserID, file, header)
	if err != nil {
		return nil, err
	}

	// Copy system-generated fields to the returned asset
	asset.ID = createdAsset.ID
	asset.CreationDate = createdAsset.CreationDate
	asset.MediaType = createdAsset.MediaType
	asset.PixelWidth = createdAsset.PixelWidth
	asset.PixelHeight = createdAsset.PixelHeight
	asset.CameraModel = createdAsset.CameraModel

	return asset, nil
}

func (r *AssetRepositoryImpl) GetAsset(ctx context.Context, assetID int) (*model.PHAsset, error) {
	return r.storage.GetAsset(assetID)
}

func (r *AssetRepositoryImpl) GetAssetContent(ctx context.Context, assetID int) ([]byte, error) {
	return r.storage.GetAssetContent(assetID)
}

func (r *AssetRepositoryImpl) UpdateAsset(ctx context.Context, assetIds []int, update model.AssetUpdate) (string, error) {
	return r.storage.UpdateAsset(assetIds, update)
}

//func (r *AssetRepositoryImpl) Create(ctx context.Context, createAlbum model.Album) (string, error) {
//	r.storage.Create()
//	return "", nil
//}

func (r *AssetRepositoryImpl) DeleteAsset(ctx context.Context, assetID int) error {
	return r.storage.DeleteAsset(assetID)
}

func (r *AssetRepositoryImpl) GetAssetThumbnail(ctx context.Context, assetID int, width, height int) ([]byte, error) {
	return r.storage.GetThumbnail(assetID, width, height)
}

func (r *AssetRepositoryImpl) GetAssetsByUser(ctx context.Context, userID int, limit, offset int) ([]*model.PHAsset, int, error) {

	// Get all asset IDs for user
	assetIDs := r.storage.GetUserAssets(userID)
	total := len(assetIDs)

	// Apply pagination
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	// Fetch assets
	assets := make([]*model.PHAsset, 0, end-start)
	for _, id := range assetIDs[start:end] {
		asset, err := r.storage.GetAsset(id)
		if err != nil {
			return nil, 0, err
		}
		assets = append(assets, asset)
	}

	return assets, total, nil
}

func (r *AssetRepositoryImpl) GetRecentAssets(ctx context.Context, userID int, days int) ([]*model.PHAsset, error) {

	end := time.Now()
	start := end.AddDate(0, 0, -days)

	assets, _, err := r.storage.SearchAssets(model.AssetSearchFilters{
		UserID:    userID,
		StartDate: &start,
		EndDate:   &end,
	})
	return assets, err
}

func (r *AssetRepositoryImpl) CountUserAssets(ctx context.Context, userID int) (int, error) {
	return r.storage.CountUserAssets(userID), nil
}

func (r *AssetRepositoryImpl) SearchAssets(ctx context.Context, filters model.AssetSearchFilters) ([]*model.PHAsset, int, error) {
	return r.storage.SearchAssets(filters)
}

func (r *AssetRepositoryImpl) FilterAssets(ctx context.Context, filters model.AssetSearchFilters) ([]*model.PHAsset, int, error) {
	return r.storage.FilterAssets(filters)
}

func (r *AssetRepositoryImpl) SuggestSearchTerms(ctx context.Context, userID int, prefix string, limit int) ([]string, error) {

	return r.storage.SuggestSearchTerms(userID, prefix, limit), nil
}

func (r *AssetRepositoryImpl) RebuildIndex(ctx context.Context) error {
	return r.storage.RebuildIndex()
}

func (r *AssetRepositoryImpl) GetStorageStats(ctx context.Context) (*model.StorageStats, error) {
	stats := r.storage.GetSystemStats()

	return &model.StorageStats{
		TotalAssets:   stats.TotalAssets,
		CacheSize:     r.storage.cache.Len(),
		CacheHits:     stats.CacheHits,
		CacheMisses:   stats.CacheMisses,
		Uploads24h:    stats.Uploads24h,
		ThumbnailsGen: stats.ThumbnailsGen,
	}, nil
}

func (r *AssetRepositoryImpl) GetIndexStatus(ctx context.Context) (*model.IndexStatus, error) {
	status := r.storage.GetIndexStatus()

	return &model.IndexStatus{
		LastRebuild:       status.LastRebuild,
		AssetCount:        status.AssetCount,
		IndexSize:         status.TextIndexSize + status.DateIndexSize,
		Dirty:             status.Dirty,
		RebuildInProgress: false, // Not implemented in storage
	}, nil
}

func (r *AssetRepositoryImpl) DeleteOrphanedAssets(ctx context.Context) (int, error) {
	return r.storage.CleanupOrphanedAssets()
}

//func (r *AssetRepositoryImpl) GenerateMissingThumbnails(ctx context.Context) (int, error) {
//	// Get assets without thumbnails
//	assetIDs, err := r.storage.GetAssetsWithoutThumbnails()
//	if err != nil {
//		return 0, err
//	}
//
//	successCount := 0
//	for _, id := range assetIDs {
//		// Get asset
//		asset, err := r.storage.Get(id)
//		if err != nil {
//			continue
//		}
//
//		// Generate default thumbnail
//		if _, err := r.storage.GetThumbnail(id,
//			r.storage.config.Thumbnails.DefaultWidth,
//			r.storage.config.Thumbnails.DefaultHeight,
//		); err == nil {
//			successCount++
//		}
//	}
//
//	return successCount, nil
//}

func (r *AssetRepositoryImpl) CleanupExpiredUploads(ctx context.Context) error {
	// Not implemented in this storage system
	return nil
}

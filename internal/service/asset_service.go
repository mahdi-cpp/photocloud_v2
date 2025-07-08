package service

import (
	"context"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
	"mime/multipart"
	"time"
)

type AssetService struct {
	repo *storage.AssetRepositoryImpl
}

func NewAssetService(assetRepository *storage.AssetRepositoryImpl) *AssetService {
	return &AssetService{repo: assetRepository}
}

func (s *AssetService) UploadAsset(
	ctx context.Context,
	userID int,
	file multipart.File,
	header *multipart.FileHeader,
) (*model.PHAsset, error) {

	// Create asset metadata
	asset := &model.PHAsset{
		UserID:   userID,
		Filename: header.Filename,
	}

	return s.repo.CreateAsset(ctx, asset, file, header)
}

func (s *AssetService) SearchAssets(
	ctx context.Context,
	userID int,
	query string,
	mediaType string,
	dateRange []time.Time,
) ([]*model.PHAsset, error) {

	filters := model.AssetSearchFilters{
		UserID:    userID,
		Query:     query,
		MediaType: model.MediaType(mediaType),
	}

	if len(dateRange) > 0 {
		filters.StartDate = &dateRange[0]
	}
	if len(dateRange) > 1 {
		filters.EndDate = &dateRange[1]
	}

	assets, _, err := s.repo.SearchAssets(ctx, filters)
	return assets, err
}

func (s *AssetService) GetAsset(ctx context.Context, id int) (*model.PHAsset, error) {
	return s.repo.GetAsset(ctx, id)
}

package service

import (
	"context"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"github.com/mahdi-cpp/photocloud_v2/internal/storage"
)

type SearchService struct {
	repo *storage.AssetRepositoryImpl
}

func NewSearchService(repo *storage.AssetRepositoryImpl) *SearchService {
	return &SearchService{repo: repo}
}

func (s *SearchService) SearchAssets(ctx context.Context, filters model.SearchFilters) ([]*model.PHAsset, int, error) {

	assets, total, err := s.repo.SearchAssets(ctx, filters)

	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	start := filters.Offset
	if start > len(assets) {
		start = len(assets)
	}

	end := start + filters.Limit
	if end > len(assets) {
		end = len(assets)
	}

	return assets[start:end], total, nil
}

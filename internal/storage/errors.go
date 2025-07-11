package storage

import "errors"

var (
	ErrUnsupportedFormat = errors.New("unsupported image format")
	ErrVideoProcessing   = errors.New("video processing disabled")
	ErrThumbnailFailed   = errors.New("thumbnail generation failed")
)

var (
	ErrAssetNotFound     = errors.New("asset not found")
	ErrThumbnailNotFound = errors.New("thumbnail not found")
	ErrFileTooLarge      = errors.New("file size exceeds limit")
	ErrInvalidUpdate     = errors.New("invalid asset update")
	ErrMetadataCorrupted = errors.New("metadata corrupted")
	ErrIndexCorrupted    = errors.New("index corrupted")
)

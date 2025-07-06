package imaging

import "errors"

var (
	ErrUnsupportedFormat = errors.New("unsupported image format")
	ErrVideoProcessing   = errors.New("video processing disabled")
	ErrThumbnailFailed   = errors.New("thumbnail generation failed")
)

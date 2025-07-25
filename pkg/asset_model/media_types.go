package asset_model

type MediaType string

const (
	UnknownType   MediaType = "unknown"
	ImageTypeJPEG MediaType = "image/jpeg"
	ImageTypePNG  MediaType = "image/png"
	ImageTypeGIF  MediaType = "image/gif"
	VideoTypeMP4  MediaType = "video/mp4"
	VideoTypeMOV  MediaType = "video/quicktime"
)

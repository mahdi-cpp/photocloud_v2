package model

type PHCollectionList[T any] struct {
	Collections []*PHCollection[T] `json:"collections"`
}

type PHCollection[T any] struct {
	Item   T          `json:"item"`   // Generic items
	Assets []*PHAsset `json:"assets"` // Specific assets
}

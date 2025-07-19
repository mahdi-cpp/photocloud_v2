package model

// https://chat.deepseek.com/a/chat/s/9b010f32-b23d-4f9b-ae0c-31a9b2c9408c

type PHCollectionList[T any] struct {
	Collections []*PHCollection[T] `json:"collections"`
}

type PHCollection[T any] struct {
	Item   T          `json:"item"`   // Generic items
	Assets []*PHAsset `json:"assets"` // Specific assets
}

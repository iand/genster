package model

type FeatureImage struct {
	MediaObject MediaObject
	Crop        Crop
}

type Crop struct {
	Left   int
	Top    int
	Width  int
	Height int
}

type MediaObject interface{}

type JpegImage struct {
	FileSize int
	Width    int
	Height   int
}

var _ MediaObject = (*JpegImage)(nil)

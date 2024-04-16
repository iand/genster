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

type MediaObject struct {
	ID          string
	SrcFilePath string
	FileName    string
	FileSize    int
	MediaType   string
	Width       int
	Height      int
	Citations   []*GeneralCitation
}

package model

type FeatureImage struct {
	MediaObject MediaObject
	Crop        *Region
}

type Crop struct {
	Left   int
	Top    int
	Width  int
	Height int
}

type MediaObject struct {
	ID          string
	Title       string
	SrcFilePath string
	FileName    string
	FileSize    int
	MediaType   string
	Width       int
	Height      int
	Citations   []*GeneralCitation
}

type Region struct {
	Left   int
	Bottom int
	Width  int
	Height int
}

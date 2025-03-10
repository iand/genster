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
	Redacted    bool // true if the object's details should be redacted
}

type Region struct {
	Left   int
	Bottom int
	Width  int
	Height int
}

type CitedMediaObject struct {
	Object    *MediaObject
	Highlight *Region
}

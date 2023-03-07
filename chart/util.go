package chart

import (
	"io/ioutil"

	"github.com/golang/freetype/truetype"
)

func LoadFont(path string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	return font, nil
}

func max(f1, f2 float64) float64 {
	if f1 > f2 {
		return f1
	}

	return f2
}

func min(f1, f2 float64) float64 {
	if f1 < f2 {
		return f1
	}

	return f2
}

package narrative

import (
	"sort"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

func RenderText[T render.EncodedText](t model.Text, enc render.ContentBuilder[T]) error {
	if t.Title != "" {
		enc.Heading3(enc.EncodeText(t.Title), t.ID)
	}
	if t.Formatted {
		enc.Pre(t.Text)
		enc.Pre("")
	} else if t.Markdown {
		txt := EncodeText(t, enc)
		enc.Markdown(txt)
		enc.EmptyPara()
	} else {
		enc.Para(enc.EncodeText(text.FormatSentence(t.Text)))
		enc.EmptyPara()
	}

	return nil
}

func EncodeText[T render.EncodedText](t model.Text, enc render.TextEncoder[T]) string {
	if len(t.Links) == 0 {
		return t.Text
	}

	text := []rune(t.Text)

	// Ensure links are ordered by start position
	// Overlapping links are not supported
	sort.Slice(t.Links, func(i, j int) bool {
		return t.Links[i].Start < t.Links[j].Start
	})
	formatted := ""
	cursor := 0
	for _, l := range t.Links {
		formatted += string(text[cursor:l.Start])
		linktext := string(text[l.Start:l.End])
		formatted += enc.EncodeModelLink(enc.EncodeText(linktext), l.Object).String()
		cursor = l.End
	}
	formatted += string(text[cursor:])
	return formatted
}

package book

import "github.com/iand/genster/render/pandoc"

type Chapter struct {
	Content *pandoc.Content
	Title   string
}

func NewChapter(title string) *Chapter {
	cs := &pandoc.Content{}
	cs.Heading2(pandoc.Text(title), "")

	return &Chapter{
		Content: cs,
		Title:   title,
	}
}

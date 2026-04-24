package render_test

import (
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/render"
)

var _ render.EncodedText = (*md.Text)(nil)

var (
	_ render.Document[md.Text]       = (*md.Document)(nil)
	_ render.ContentBuilder[md.Text] = (*md.Document)(nil)
	_ render.TextEncoder[md.Text]    = (*md.Document)(nil)
)

var (
	_ render.ContentBuilder[md.Text] = (*md.Content)(nil)
	_ render.TextEncoder[md.Text]    = (*md.Document)(nil)
)

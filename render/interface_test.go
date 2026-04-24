package render_test

import (
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
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

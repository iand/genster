package render_test

import "github.com/iand/genster/render/md"
import "github.com/iand/genster/render"

var _ render.EncodedText = (*md.Text)(nil)

var _ render.Document[md.Text] = (*md.Document)(nil)
var _ render.ContentBuilder[md.Text] = (*md.Document)(nil)
var _ render.TextEncoder[md.Text] = (*md.Document)(nil)

var _ render.ContentBuilder[md.Text] = (*md.Content)(nil)
var _ render.TextEncoder[md.Text] = (*md.Document)(nil)

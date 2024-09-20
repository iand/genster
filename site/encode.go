package site

import (
	"net/url"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/text"
)

func EncodeRawLink[T render.EncodedText](u string, enc render.TextEncoder[T]) string {
	text := u

	pu, err := url.Parse(u)
	if err == nil && pu != nil && pu.Host != "" {
		text = pu.Host
	}

	return enc.EncodeLink(enc.EncodeText(text), u).String()
}

// EncodePeopleListInline encodes a list of people as a comma separated list
func EncodePeopleListInline[T render.EncodedText](ps []*model.Person, formatter func(*model.Person) string, enc render.TextEncoder[T]) string {
	ss := make([]string, len(ps))
	for i := range ps {
		ss[i] = enc.EncodeModelLink(enc.EncodeText(formatter(ps[i])), ps[i]).String()
	}
	return text.JoinList(ss)
}

type CitationSkippingEncoder[T render.EncodedText] struct {
	render.PageBuilder[T]
}

var (
	_ render.PageBuilder[md.Text] = (*CitationSkippingEncoder[md.Text])(nil)
	_ render.TextEncoder[md.Text] = (*CitationSkippingEncoder[md.Text])(nil)
)

func (e *CitationSkippingEncoder[T]) EncodeModelLinkDedupe(firstText T, subsequentText T, m any) T {
	return firstText
}

func (e *CitationSkippingEncoder[T]) EncodeWithCitations(s T, citations []*model.GeneralCitation) T {
	return s
}

// A PersonLinkingTextEncoder is a [TextEncoder] that only generates links for people.
type PersonLinkingTextEncoder[T render.EncodedText] struct {
	render.TextEncoder[T]
}

var _ render.TextEncoder[md.Text] = (*PersonLinkingTextEncoder[md.Text])(nil)

func (e *PersonLinkingTextEncoder[T]) EncodeModelLink(text T, m any) T {
	if _, ok := m.(*model.Person); ok {
		return e.TextEncoder.EncodeModelLink(text, m)
	}
	return text
}

func (e *PersonLinkingTextEncoder[T]) EncodeModelLinkDedupe(firstText T, subsequentText T, m any) T {
	if _, ok := m.(*model.Person); ok {
		return e.TextEncoder.EncodeModelLinkDedupe(firstText, subsequentText, m)
	}
	return firstText
}

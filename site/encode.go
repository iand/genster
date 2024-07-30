package site

import (
	"net/url"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
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
	_ render.PageBuilder[render.Markdown] = (*CitationSkippingEncoder[render.Markdown])(nil)
	_ render.TextEncoder[render.Markdown] = (*CitationSkippingEncoder[render.Markdown])(nil)
)

func (e *CitationSkippingEncoder[T]) EncodeModelLinkDedupe(firstText T, subsequentText T, m any) T {
	return firstText
}

func (e *CitationSkippingEncoder[T]) EncodeWithCitations(s T, citations []*model.GeneralCitation) T {
	return s
}

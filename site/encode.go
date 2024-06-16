package site

import (
	"net/url"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

func EncodeRawLink(u string, enc render.InlineMarkdownEncoder) string {
	text := u

	pu, err := url.Parse(u)
	if err == nil && pu != nil && pu.Host != "" {
		text = pu.Host
	}

	return enc.EncodeLink(text, u)
}

// EncodePeopleListInline encodes a list of people as a comma separated list
func EncodePeopleListInline(ps []*model.Person, formatter func(*model.Person) string, enc render.InlineMarkdownEncoder) string {
	ss := make([]string, len(ps))
	for i := range ps {
		ss[i] = enc.EncodeModelLink(formatter(ps[i]), ps[i])
	}
	return text.JoinList(ss)
}

type CitationSkippingEncoder struct {
	render.MarkupBuilder
}

var _ render.InlineMarkdownEncoder = (*CitationSkippingEncoder)(nil)

func (e *CitationSkippingEncoder) EncodeModelLinkDedupe(firstText string, subsequentText string, m any) string {
	return firstText
}

func (e *CitationSkippingEncoder) EncodeCitationDetail(c *model.GeneralCitation) string {
	return ""
}

func (e *CitationSkippingEncoder) EncodeWithCitations(s string, citations []*model.GeneralCitation) string {
	return s
}

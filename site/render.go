package site

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"golang.org/x/exp/slog"
)

// cleanCitationDetail removes some redundant information that isn't necessary when a source is included
func cleanCitationDetail(page string) string {
	page = strings.TrimPrefix(page, "The National Archives of the UK (TNA); Kew, Surrey, England; Census Returns of England and Wales, 1891;")
	page = strings.TrimPrefix(page, "The National Archives; Kew, London, England; 1871 England Census; ")

	page = text.FinishSentence(page)
	return page
}

func RenderFacts(facts []model.Fact, pov *model.POV, enc ExtendedMarkdownBuilder) error {
	enc.EmptyPara()

	categories := make([]string, 0)
	factsByCategory := make(map[string][]*model.Fact)

	for _, f := range facts {
		f := f // avoid shadowing
		fl, ok := factsByCategory[f.Category]
		if ok {
			fl = append(fl, &f)
			factsByCategory[f.Category] = fl
			continue
		}

		categories = append(categories, f.Category)
		factsByCategory[f.Category] = []*model.Fact{&f}
	}

	sort.Strings(categories)

	factlist := make([][2]string, 0, len(categories))
	for _, cat := range categories {
		fl, ok := factsByCategory[cat]
		if !ok {
			continue
		}
		if len(fl) == 0 {
			factlist = append(factlist, [2]string{cat, fl[0].Detail})
			continue
		}
		buf := new(strings.Builder)
		for i, f := range fl {
			if i > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(EncodeWithCitations(f.Detail, f.Citations, enc))
		}
		factlist = append(factlist, [2]string{cat, buf.String()})
	}

	enc.DefinitionList(factlist)
	return nil
}

func CleanTags(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	tags := make([]string, 0, len(ss))
	for _, s := range ss {
		tag := Tagify(s)
		if seen[tag] {
			continue
		}
		tags = append(tags, tag)
		seen[tag] = true
	}
	sort.Strings(tags)
	return tags
}

func Tagify(s string) string {
	s = strings.ToLower(s)
	parts := strings.Fields(s)
	s = strings.Join(parts, "-")
	return s
}

type Paginator struct {
	Index      string
	DirPattern string

	Entries []PaginatorEntry
}

func NewPaginator() *Paginator {
	return &Paginator{
		Index:      "index.md",
		DirPattern: "%02d",
	}
}

type PaginatorEntry struct {
	Key     string
	Content string
}

func (p *Paginator) AddEntry(key string, content string) {
	p.Entries = append(p.Entries, PaginatorEntry{
		Key:     key,
		Content: content,
	})
}

func (p *Paginator) WritePages(s *Site, baseDir string, section string, title string) error {
	sort.Slice(p.Entries, func(i, j int) bool {
		return p.Entries[i].Key < p.Entries[j].Key
	})

	type Page struct {
		FirstKey string
		LastKey  string
		Content  string
		Dir      string
	}

	var pages []*Page

	maxSize := 4096
	pg := &Page{
		Dir: fmt.Sprintf(p.DirPattern, 1),
	}
	for _, e := range p.Entries {
		if len(pg.Content)+len(e.Content) > maxSize {
			if len(pg.Content) == 0 {
				slog.Warn("skipping item since it is larger than maximum allowed for a single page")
				continue
			}
			pages = append(pages, pg)
			pg = &Page{
				Dir: fmt.Sprintf(p.DirPattern, len(pages)+1),
			}
		}
		if len(pg.Content) == 0 {
			pg.FirstKey = e.Key
		}
		pg.Content += e.Content
		pg.LastKey = e.Key
	}
	pages = append(pages, pg)

	for i, pg := range pages {
		idx := i + 1
		doc := s.NewDocument()
		doc.Title(fmt.Sprintf("%s (page %d of %d)", title, idx, len(pages)))
		doc.Section(section)

		doc.SetFrontMatterField(md.MarkdownTagIndexPage, section)
		if idx > 1 {
			doc.SetFrontMatterField(md.MarkdownTagFirstPage, filepath.Join(section, fmt.Sprintf(p.DirPattern, 1)))
			if idx > 2 {
				doc.SetFrontMatterField(md.MarkdownTagPrevPage, filepath.Join(section, fmt.Sprintf(p.DirPattern, idx-1)))
			}
		}
		if idx < len(pages) {
			doc.SetFrontMatterField(md.MarkdownTagLastPage, filepath.Join(section, fmt.Sprintf(p.DirPattern, len(pages))))
			if idx < len(pages)-1 {
				doc.SetFrontMatterField(md.MarkdownTagNextPage, filepath.Join(section, fmt.Sprintf(p.DirPattern, idx+1)))
			}
		}

		doc.SetBody(pg.Content)

		if err := writePage(doc, baseDir, filepath.Join(pg.Dir, p.Index)); err != nil {
			return fmt.Errorf("failed to write paginated page: %w", err)
		}
	}

	doc := s.NewDocument()
	doc.Title(title)

	var list []string
	for _, pg := range pages {
		if pg.FirstKey == pg.LastKey {
			list = append(list, doc.EncodeLink(pg.FirstKey, pg.Dir))
		} else {
			list = append(list, doc.EncodeLink(fmt.Sprintf("%s to %s", pg.FirstKey, pg.LastKey), pg.Dir))
		}
	}
	doc.UnorderedList(list)
	if err := writePage(doc, baseDir, p.Index); err != nil {
		return fmt.Errorf("failed to write paginated index: %w", err)
	}

	return nil
}

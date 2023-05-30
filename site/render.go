package site

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"golang.org/x/exp/slog"
)

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
			buf.WriteString(enc.EncodeWithCitations(f.Detail, f.Citations))
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
	HugoStyle bool

	Entries []PaginatorEntry
}

func NewPaginator() *Paginator {
	return &Paginator{}
}

type PaginatorEntry struct {
	Key           string
	Title         string
	Group         string
	GroupPriority int
	Content       string
}

func (p *Paginator) AddEntry(key string, title string, content string) {
	p.Entries = append(p.Entries, PaginatorEntry{
		Key:     key,
		Title:   title,
		Content: content,
	})
}

func (p *Paginator) AddEntryWithGroup(key string, title string, content string, group string, groupPriority int) {
	p.Entries = append(p.Entries, PaginatorEntry{
		Key:           key,
		Title:         title,
		Group:         group,
		GroupPriority: groupPriority,
		Content:       content,
	})
}

func (p *Paginator) WritePages(s *Site, baseDir string, layout PageLayout, title string, summary string) error {
	indexPage := "index.md"
	if p.HugoStyle {
		indexPage = "_index.md"
	}
	sort.Slice(p.Entries, func(i, j int) bool {
		if p.Entries[i].Group != p.Entries[j].Group {
			return p.Entries[i].GroupPriority < p.Entries[j].GroupPriority
		}
		return p.Entries[i].Key < p.Entries[j].Key
	})

	type Page struct {
		FirstKey   string
		LastKey    string
		FirstTitle string
		LastTitle  string
		Content    string
		Name       string
		Group      string
	}

	var pages []*Page

	maxSize := 4096
	pg := &Page{
		Name: fmt.Sprintf("%02d", 1),
	}
	if len(p.Entries) > 0 {
		pg.Group = p.Entries[0].Group
	}
	for _, e := range p.Entries {
		if e.Group != pg.Group {
			// start a new page and group
			pages = append(pages, pg)
			pg = &Page{
				Name:  fmt.Sprintf("%02d", len(pages)+1),
				Group: e.Group,
			}
		}
		if len(pg.Content)+len(e.Content) > maxSize {
			if len(pg.Content) == 0 {
				slog.Warn("skipping item since it is larger than maximum allowed for a single page")
				continue
			}
			pages = append(pages, pg)
			pg = &Page{
				Name:  fmt.Sprintf("%02d", len(pages)+1),
				Group: e.Group,
			}
		}
		if len(pg.Content) == 0 {
			pg.FirstKey = e.Key
			pg.FirstTitle = e.Title
		}
		pg.Content += e.Content
		pg.LastKey = e.Key
		pg.LastTitle = e.Title
	}
	pages = append(pages, pg)

	if len(pages) > 1 {

		for i, pg := range pages {
			idx := i + 1
			doc := s.NewDocument()
			doc.Title(fmt.Sprintf("%s (page %d of %d)", title, idx, len(pages)))
			doc.Layout(layout.String())

			if idx > 1 {
				doc.SetFrontMatterField(md.MarkdownTagFirstPage, fmt.Sprintf("%02d", 1))
				if idx > 2 {
					doc.SetFrontMatterField(md.MarkdownTagPrevPage, fmt.Sprintf("%02d", idx-1))
				}
			}
			if idx < len(pages) {
				doc.SetFrontMatterField(md.MarkdownTagLastPage, fmt.Sprintf("%02d", len(pages)))
				if idx < len(pages)-1 {
					doc.SetFrontMatterField(md.MarkdownTagNextPage, fmt.Sprintf("%02d", idx+1))
				}
			}

			doc.SetBody(pg.Content)

			var fname string
			if p.HugoStyle {
				fname = pg.Name + ".md"
			} else {
				fname = filepath.Join(pg.Name, indexPage)
			}

			if err := writePage(doc, baseDir, fname); err != nil {
				return fmt.Errorf("failed to write paginated page: %w", err)
			}
		}

		doc := s.NewDocument()
		doc.Title(title)
		if summary != "" {
			doc.Summary(summary)
		}
		doc.Layout(layout.String())

		var group string
		var list []string
		for _, pg := range pages {
			if pg.Group != group {
				if len(list) > 0 {
					doc.Heading3(group)
					doc.UnorderedList(list)
					list = list[:0]
				}
				group = pg.Group
			}
			if pg.FirstKey == pg.LastKey {
				list = append(list, doc.EncodeLink(pg.FirstTitle, pg.Name))
			} else {
				list = append(list, doc.EncodeLink(fmt.Sprintf("%s to %s", pg.FirstTitle, pg.LastTitle), pg.Name))
			}
		}

		if len(list) > 0 {
			doc.Heading3(group)
			doc.UnorderedList(list)
		}

		if err := writePage(doc, baseDir, indexPage); err != nil {
			return fmt.Errorf("failed to write paginated index: %w", err)
		}
	} else {
		doc := s.NewDocument()
		doc.Title(title)
		if summary != "" {
			doc.Summary(summary)
		}
		doc.Layout(layout.String())
		doc.SetBody(pages[0].Content)
		if err := writePage(doc, baseDir, indexPage); err != nil {
			return fmt.Errorf("failed to write paginated index: %w", err)
		}
	}
	return nil
}

package site

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/iand/genster/md"
)

type Paginator struct {
	HugoStyle   bool
	MaxPageSize int

	Entries []PaginatorEntry
}

func NewPaginator() *Paginator {
	return &Paginator{
		MaxPageSize: 4096,
	}
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
		if p.MaxPageSize > 0 && len(pg.Content)+len(e.Content) > p.MaxPageSize {
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
		var list []md.Text
		for _, pg := range pages {
			if pg.Group != group {
				if len(list) > 0 {
					doc.Heading3(md.Text(group), "")
					doc.UnorderedList(list)
					list = list[:0]
				}
				group = pg.Group
			}
			if pg.FirstKey == pg.LastKey {
				list = append(list, doc.EncodeLink(doc.EncodeText(pg.FirstTitle), pg.Name))
			} else {
				list = append(list, doc.EncodeLink(doc.EncodeText(fmt.Sprintf("%s to %s", pg.FirstTitle, pg.LastTitle)), pg.Name))
			}
		}

		if len(list) > 0 {
			doc.Heading3(md.Text(group), "")
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

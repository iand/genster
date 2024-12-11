package book

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/pandoc"
	"github.com/iand/genster/site"
	"github.com/iand/genster/tree"
)

func NewBook(t *tree.Tree) *Book {
	b := &Book{
		Tree: t,
	}
	return b
}

type Book struct {
	Tree             *tree.Tree
	IncludePrivate   bool
	IncludeDebugInfo bool
	PublishSet       *site.PublishSet
	Doc              *pandoc.Document
	Chapters         []*Chapter
}

func (b *Book) BuildPublishSet(m model.PersonMatcher) error {
	subset, err := site.NewPublishSet(b.Tree, m)
	if err != nil {
		return fmt.Errorf("build publish set: %w", err)
	}

	b.PublishSet = subset
	return nil
}

func (s *Book) BuildDocument() error {
	s.Doc = &pandoc.Document{}

	for _, f := range s.PublishSet.Families {
		err := s.AddFamilyChapter(f)
		if err != nil {
			return fmt.Errorf("family page: %w", err)
		}

		// if err := writePage(d, contentDir, fmt.Sprintf(s.PersonFilePattern, p.ID)); err != nil {
		// 	return fmt.Errorf("write person page: %w", err)
		// }

	}
	// for _, p := range s.PublishSet.People {
	// 	if s.LinkFor(p) == "" {
	// 		continue
	// 	}
	// 	d, err := RenderPersonPage(s, p)
	// 	if err != nil {
	// 		return fmt.Errorf("render person page: %w", err)
	// 	}

	// 	if err := writePage(d, contentDir, fmt.Sprintf(s.PersonFilePattern, p.ID)); err != nil {
	// 		return fmt.Errorf("write person page: %w", err)
	// 	}

	// }

	// for _, p := range s.PublishSet.Places {
	// 	if s.LinkFor(p) == "" {
	// 		continue
	// 	}
	// 	d, err := RenderPlacePage(s, p)
	// 	if err != nil {
	// 		return fmt.Errorf("render place page: %w", err)
	// 	}

	// 	if err := writePage(d, contentDir, fmt.Sprintf(s.PlaceFilePattern, p.ID)); err != nil {
	// 		return fmt.Errorf("write place page: %w", err)
	// 	}
	// }

	// for _, c := range s.PublishSet.Citations {
	// 	if s.LinkFor(c) == "" {
	// 		continue
	// 	}
	// 	d, err := RenderCitationPage(s, c)
	// 	if err != nil {
	// 		return fmt.Errorf("render citation page: %w", err)
	// 	}
	// 	if err := writePage(d, contentDir, fmt.Sprintf(s.CitationFilePattern, c.ID)); err != nil {
	// 		return fmt.Errorf("write citation page: %w", err)
	// 	}
	// }

	// for _, mo := range s.PublishSet.MediaObjects {
	// 	// TODO: redaction

	// 	// var ext string
	// 	// switch mo.MediaType {
	// 	// case "image/jpeg":
	// 	// 	ext = "jpg"
	// 	// case "image/png":
	// 	// 	ext = "png"
	// 	// case "image/gif":
	// 	// 	ext = "gif"
	// 	// default:
	// 	// 	return fmt.Errorf("unsupported media type: %v", mo.MediaType)
	// 	// }

	// 	fname := filepath.Join(mediaDir, fmt.Sprintf("%s/%s", s.MediaDir, mo.FileName))

	// 	if err := CopyFile(fname, mo.SrcFilePath); err != nil {
	// 		return fmt.Errorf("copy media object: %w", err)
	// 	}
	// }

	// s.BuildCalendar()

	// for month, c := range s.Calendars {
	// 	d, err := c.RenderPage(s)
	// 	if err != nil {
	// 		return fmt.Errorf("generate markdown: %w", err)
	// 	}

	// 	fname := fmt.Sprintf(s.CalendarFilePattern, month)

	// 	f, err := CreateFile(filepath.Join(contentDir, fname))
	// 	if err != nil {
	// 		return fmt.Errorf("create calendar file: %w", err)
	// 	}
	// 	if _, err := d.WriteTo(f); err != nil {
	// 		return fmt.Errorf("write calendar markdown: %w", err)
	// 	}
	// 	f.Close()
	// }

	// if err := s.WritePersonListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write people list pages: %w", err)
	// }

	// if err := s.WritePlaceListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write place list pages: %w", err)
	// }

	// // Not publishing sources at this time
	// // if err := s.WriteSourceListPages(contentDir); err != nil {
	// // 	return fmt.Errorf("write source list pages: %w", err)
	// // }

	// if err := s.WriteSurnameListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write surname list pages: %w", err)
	// }

	// if err := s.WriteInferenceListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write inferences pages: %w", err)
	// }

	// if err := s.WriteAnomalyListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write anomalies pages: %w", err)
	// }

	// if err := s.WriteTodoListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write todo pages: %w", err)
	// }

	// if err := s.WriteTreeOverview(contentDir); err != nil {
	// 	return fmt.Errorf("write tree overview: %w", err)
	// }

	// if err := s.WriteChartAncestors(contentDir); err != nil {
	// 	return fmt.Errorf("write ancestor chart: %w", err)
	// }

	// if err := s.WriteChartTrees(root); err != nil {
	// 	return fmt.Errorf("write chart trees: %w", err)
	// }

	// TODO: order chapters
	for _, c := range s.Chapters {
		s.Doc.AppendText(c.Content.Text())
	}

	return nil
}

func (s *Book) WriteDocument(fname string) error {
	path := filepath.Dir(fname)

	if err := os.MkdirAll(path, 0o777); err != nil {
		return fmt.Errorf("create path: %w", err)
	}

	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if _, err := s.Doc.WriteTo(f); err != nil {
		return fmt.Errorf("write content: %w", err)
	}
	return f.Close()
}

func (b *Book) AddFamilyChapter(f *model.Family) error {
	ch := NewChapter(f.PreferredUniqueName)
	b.Chapters = append(b.Chapters, ch)
	return nil
}

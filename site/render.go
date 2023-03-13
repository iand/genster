package site

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/werr"
	"golang.org/x/exp/slog"
)

func RenderPersonPage(s *Site, p *model.Person) (*md.Document, error) {
	pov := &model.POV{Person: p}

	d := s.NewDocument()
	d.Section(md.PageLayoutPerson)
	d.ID(p.ID)
	d.Title(p.PreferredUniqueName)
	d.SetFrontMatterField("gender", p.Gender.Noun())

	if p.Redacted {
		d.Summary("information withheld to preserve privacy")
		return d, nil
	}

	if p.Olb != "" {
		d.Summary(p.Olb)
	}
	d.AddTags(CleanTags(p.Tags))

	b := d.Body()

	// Render narrative
	n := &Narrative{
		Statements: make([]Statement, 0),
	}

	// Everyone has an intro
	intro := &IntroStatement{
		Principal: p,
	}
	for _, ev := range p.Timeline {
		switch tev := ev.(type) {
		case *model.BaptismEvent:
			if tev != p.BestBirthlikeEvent {
				intro.Baptisms = append(intro.Baptisms, tev)
			}
		}
	}
	if len(intro.Baptisms) > 0 {
		sort.Slice(intro.Baptisms, func(i, j int) bool {
			return intro.Baptisms[i].GetDate().SortsBefore(intro.Baptisms[j].GetDate())
		})
	}
	n.Statements = append(n.Statements, intro)

	// If death is known, add it
	if p.BestDeathlikeEvent != nil {
		n.Statements = append(n.Statements, &DeathStatement{
			Principal: p,
		})
	}

	for _, f := range p.Families {
		n.Statements = append(n.Statements, &FamilyStatement{
			Principal: p,
			Family:    f,
		})
	}

	n.Render(pov, b)

	if p.EditLink != nil {
		d.SetFrontMatterField("editlink", p.EditLink.URL)
		d.SetFrontMatterField("editlinktitle", p.EditLink.Title)
	}

	t := &model.Timeline{
		Events: make([]model.TimelineEvent, 0, len(p.Timeline)),
	}
	for _, ev := range p.Timeline {
		if ev.GetDate().IsUnknown() && ev.GetPlace().IsUnknown() {
			p.MiscFacts = append(p.MiscFacts, model.Fact{
				Category: ev.GetTitle(),
				Detail:   ev.GetDetail(),
			})
		} else {
			t.Events = append(t.Events, ev)
		}
	}

	if len(p.Timeline) > 0 {
		b.EmptyPara()
		b.Heading2("Timeline")

		b.ResetSeenLinks()
		if err := RenderTimeline(t, pov, b); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	if len(p.MiscFacts) > 0 {
		b.EmptyPara()
		b.Heading2("Other Information")
		if err := RenderFacts(p.MiscFacts, pov, b); err != nil {
			return nil, werr.Wrap(err)
		}
	}

	links := make([]string, 0, len(p.Links))
	for _, l := range p.Links {
		links = append(links, b.EncodeLink(l.Title, l.URL))
	}

	if len(links) > 0 {
		b.Heading2("Links")
		b.UnorderedList(links)
	}

	return d, nil
}

func RenderSourcePage(s *Site, sr *model.Source) (*md.Document, error) {
	d := s.NewDocument()

	d.Title(sr.Title)
	d.Section(md.PageLayoutSource)
	d.ID(sr.ID)
	d.AddTags(CleanTags(sr.Tags))

	return d, nil
}

func RenderPlacePage(s *Site, p *model.Place) (*md.Document, error) {
	pov := &model.POV{Place: p}

	d := s.NewDocument()
	b := d.Body()

	d.Title(p.PreferredName)
	d.Section(md.PageLayoutPlace)
	d.ID(p.ID)
	d.AddTags(CleanTags(p.Tags))

	desc := p.PreferredName + " is a" + text.MaybeAn(p.PlaceType.String())

	if !p.Parent.IsUnknown() {
		desc += " in " + b.EncodeModelLinkDedupe(p.Parent.PreferredUniqueName, p.Parent.PreferredName, p.Parent)
	}

	b.Para(text.FinishSentence(desc))

	t := &model.Timeline{
		Events: p.Timeline,
	}

	if len(p.Timeline) > 0 {
		b.EmptyPara()
		b.Heading2("Timeline")

		if err := RenderTimeline(t, pov, b); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	if len(p.Links) > 0 {
		b.Heading2("Links")
		for _, l := range p.Links {
			b.Para(b.EncodeLink(l.Title, l.URL))
		}
	}

	return d, nil
}

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
		d := s.NewDocument()
		d.Title(fmt.Sprintf("%s (page %d of %d)", title, idx, len(pages)))
		d.Section(section)

		d.SetFrontMatterField(md.MarkdownTagIndexPage, section)
		if idx > 1 {
			d.SetFrontMatterField(md.MarkdownTagFirstPage, filepath.Join(section, fmt.Sprintf(p.DirPattern, 1)))
			if idx > 2 {
				d.SetFrontMatterField(md.MarkdownTagPrevPage, filepath.Join(section, fmt.Sprintf(p.DirPattern, idx-1)))
			}
		}
		if idx < len(pages) {
			d.SetFrontMatterField(md.MarkdownTagLastPage, filepath.Join(section, fmt.Sprintf(p.DirPattern, len(pages))))
			if idx < len(pages)-1 {
				d.SetFrontMatterField(md.MarkdownTagNextPage, filepath.Join(section, fmt.Sprintf(p.DirPattern, idx+1)))
			}
		}

		d.Body().Set(pg.Content)

		if err := writePage(d, baseDir, filepath.Join(pg.Dir, p.Index)); err != nil {
			return fmt.Errorf("failed to write paginated page: %w", err)
		}
	}

	d := s.NewDocument()
	d.Title(title)
	b := d.Body()

	var list []string
	for _, pg := range pages {
		if pg.FirstKey == pg.LastKey {
			list = append(list, b.EncodeLink(pg.FirstKey, pg.Dir))
		} else {
			list = append(list, b.EncodeLink(fmt.Sprintf("%s to %s", pg.FirstKey, pg.LastKey), pg.Dir))
		}
	}
	b.UnorderedList(list)
	if err := writePage(d, baseDir, p.Index); err != nil {
		return fmt.Errorf("failed to write paginated index: %w", err)
	}

	return nil
}

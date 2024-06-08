package site

import (
	"container/heap"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gosimple/slug"
	"github.com/iand/genster/chart"
	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
	"github.com/iand/gtree"
)

func (s *Site) WritePages(contentDir string, mediaDir string) error {
	for _, p := range s.PublishSet.People {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderPersonPage(s, p)
		if err != nil {
			return fmt.Errorf("render person page: %w", err)
		}

		if err := writePage(d, contentDir, fmt.Sprintf(s.PersonFilePattern, p.ID)); err != nil {
			return fmt.Errorf("write person page: %w", err)
		}

		if s.GenerateWikiTree {
			d, err := RenderWikiTreePage(s, p)
			if err != nil {
				return fmt.Errorf("render wikitree page: %w", err)
			}

			if err := writePage(d, contentDir, fmt.Sprintf(s.WikiTreeFilePattern, p.ID)); err != nil {
				return fmt.Errorf("write wikitree page: %w", err)
			}
		}
	}

	for _, p := range s.PublishSet.Places {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderPlacePage(s, p)
		if err != nil {
			return fmt.Errorf("render place page: %w", err)
		}

		if err := writePage(d, contentDir, fmt.Sprintf(s.PlaceFilePattern, p.ID)); err != nil {
			return fmt.Errorf("write place page: %w", err)
		}
	}

	for _, c := range s.PublishSet.Citations {
		if s.LinkFor(c) == "" {
			continue
		}
		d, err := RenderCitationPage(s, c)
		if err != nil {
			return fmt.Errorf("render citation page: %w", err)
		}
		if err := writePage(d, contentDir, fmt.Sprintf(s.CitationFilePattern, c.ID)); err != nil {
			return fmt.Errorf("write citation page: %w", err)
		}
	}

	// Not publishing sources at this time
	// for _, so := range s.PublishSet.Sources {
	// 	if s.LinkFor(so) == "" {
	// 		continue
	// 	}
	// 	d, err := RenderSourcePage(s, so)
	// 	if err != nil {
	// 		return fmt.Errorf("render source page: %w", err)
	// 	}
	// 	if err := writePage(d, contentDir, fmt.Sprintf(s.SourceFilePattern, so.ID)); err != nil {
	// 		return fmt.Errorf("write source page: %w", err)
	// 	}
	// }

	for _, mo := range s.PublishSet.MediaObjects {
		// TODO: redaction
		if len(mo.Citations) == 0 {
			continue
		}

		// var ext string
		// switch mo.MediaType {
		// case "image/jpeg":
		// 	ext = "jpg"
		// case "image/png":
		// 	ext = "png"
		// case "image/gif":
		// 	ext = "gif"
		// default:
		// 	return fmt.Errorf("unsupported media type: %v", mo.MediaType)
		// }

		fname := filepath.Join(mediaDir, fmt.Sprintf("%s/%s", s.MediaDir, mo.FileName))

		if err := CopyFile(fname, mo.SrcFilePath); err != nil {
			return fmt.Errorf("copy media object: %w", err)
		}
	}

	s.BuildCalendar()

	for month, c := range s.Calendars {
		d, err := c.RenderPage(s)
		if err != nil {
			return fmt.Errorf("generate markdown: %w", err)
		}

		fname := fmt.Sprintf(s.CalendarFilePattern, month)

		f, err := CreateFile(filepath.Join(contentDir, fname))
		if err != nil {
			return fmt.Errorf("create calendar file: %w", err)
		}
		if err := d.WriteMarkdown(f); err != nil {
			return fmt.Errorf("write calendar markdown: %w", err)
		}
		f.Close()
	}

	if err := s.WritePersonListPages(contentDir); err != nil {
		return fmt.Errorf("write people list pages: %w", err)
	}

	if err := s.WritePlaceListPages(contentDir); err != nil {
		return fmt.Errorf("write place list pages: %w", err)
	}

	// Not publishing sources at this time
	// if err := s.WriteSourceListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write source list pages: %w", err)
	// }

	if err := s.WriteSurnameListPages(contentDir); err != nil {
		return fmt.Errorf("write surname list pages: %w", err)
	}

	if err := s.WriteInferenceListPages(contentDir); err != nil {
		return fmt.Errorf("write inferences pages: %w", err)
	}

	if err := s.WriteAnomalyListPages(contentDir); err != nil {
		return fmt.Errorf("write anomalies pages: %w", err)
	}

	if err := s.WriteTodoListPages(contentDir); err != nil {
		return fmt.Errorf("write todo pages: %w", err)
	}

	if err := s.WriteTreeOverview(contentDir); err != nil {
		return fmt.Errorf("write tree overview: %w", err)
	}

	if err := s.WriteChartAncestors(contentDir); err != nil {
		return fmt.Errorf("write ancestor chart: %w", err)
	}

	// if err := s.WriteChartTrees(root); err != nil {
	// 	return fmt.Errorf("write chart trees: %w", err)
	// }

	return nil
}

func (s *Site) NewDocument() *md.Document {
	doc := &md.Document{}
	doc.BasePath(s.BaseURL)
	doc.SetLinkBuilder(s)
	return doc
}

func (s *Site) NewMarkdownBuilder() MarkdownBuilder {
	enc := &md.Encoder{}
	enc.SetLinkBuilder(s)

	return enc
}

func (s *Site) WriteTreeOverview(root string) error {
	doc := s.NewDocument()
	if s.Tree.Name != "" {
		doc.Title(s.Tree.Name)
	} else {
		doc.Title("Tree Overview")
	}
	doc.Layout(PageLayoutTreeOverview.String())

	fname := "index.md"
	if s.GenerateHugo {
		fname = "_index.md"
	}

	desc := s.Tree.Description

	if desc != "" {
		doc.Para(text.FormatSentence(desc))
	}

	peopleDesc := ""

	numberOfPeople := s.PublishSet.NumberOfPeople()
	if numberOfPeople > 0 {
		doc.EmptyPara()
		peopleDesc = text.FormatSentence(fmt.Sprintf("There are %d people in this tree", numberOfPeople))
	}

	ancestorSurnames := FlattenMapByValueDesc(s.PublishSet.AncestorSurnameDistribution())
	if len(ancestorSurnames) > 0 {
		list := make([]string, 12)
		for i := range ancestorSurnames {
			if i > 11 {
				break
			}
			list[i] = doc.EncodeLink(ancestorSurnames[i].K, path.Join(s.BaseURL, s.ListSurnamesDir, slug.Make(ancestorSurnames[i].K)))
		}
		detail := text.JoinSentenceParts("The principle surnames are ", text.JoinList(list))
		peopleDesc = text.JoinSentences(peopleDesc, text.FormatSentence(detail))
		peopleDesc = text.JoinSentences(peopleDesc, text.FormatSentence(text.JoinSentenceParts("See", doc.EncodeLink("all surnames...", s.ListSurnamesDir))))
	}

	if peopleDesc != "" {
		doc.EmptyPara()
		doc.Para(peopleDesc)
	}

	doc.EmptyPara()
	doc.Para(text.JoinSentenceParts("See a", doc.EncodeLink("full list of ancestors", s.ChartAncestorsDir), "for", doc.EncodeModelLink(s.Tree.KeyPerson.PreferredFamiliarFullName, s.Tree.KeyPerson)))

	// Featured people
	featuredPeople := s.Tree.ListPeopleMatching(func(p *model.Person) bool {
		if s.LinkFor(p) == "" {
			return false
		}
		return p.Featured
	}, 8)
	if len(featuredPeople) > 0 {
		model.SortPeopleByName(featuredPeople)
		doc.EmptyPara()
		doc.Heading2("Featured")
		items := make([]string, len(featuredPeople))
		for i, p := range featuredPeople {
			items[i] = text.AppendRelated(doc.EncodeModelLink(p.PreferredUniqueName, p), p.Olb)
		}
		doc.UnorderedList(items)
	}

	// Currently puzzling over
	puzzlePeople := s.Tree.ListPeopleMatching(func(p *model.Person) bool {
		if s.LinkFor(p) == "" {
			return false
		}
		return p.Puzzle && !p.Featured
	}, 8)
	if len(puzzlePeople) > 0 {
		model.SortPeopleByName(puzzlePeople)
		doc.EmptyPara()
		doc.Heading2("Currently puzzling over")
		items := make([]string, len(puzzlePeople))
		for i, p := range puzzlePeople {
			items[i] = text.AppendRelated(doc.EncodeModelLink(p.PreferredUniqueName, p), p.Olb)
		}
		doc.UnorderedList(items)
	}

	// People with research notes
	rnPeople := s.Tree.ListPeopleMatching(func(p *model.Person) bool {
		if s.LinkFor(p) == "" {
			return false
		}
		if len(p.ResearchNotes) == 0 {
			return false
		}
		for _, pp := range puzzlePeople {
			if pp.SameAs(p) {
				return false
			}
		}
		for _, pp := range featuredPeople {
			if pp.SameAs(p) {
				return false
			}
		}
		return true
	}, 3)
	if len(rnPeople) > 0 {
		model.SortPeopleByName(rnPeople)
		doc.EmptyPara()
		detail := text.JoinSentenceParts("Other people with research notes:", EncodePeopleListInline(rnPeople, func(p *model.Person) string {
			return p.PreferredFamiliarFullName
		}, doc))
		doc.Para(text.FormatSentence(detail))
	}

	// Oldest people
	oldestPeople := s.PublishSet.OldestPeople(3)
	if len(oldestPeople) > 0 {
		doc.EmptyPara()
		detail := text.JoinSentenceParts("Oldest people:", EncodePeopleListInline(oldestPeople, func(p *model.Person) string {
			age, _ := p.AgeInYearsAtDeath()
			return fmt.Sprintf("%s (%d years)", p.PreferredFamiliarFullName, age)
		}, doc))
		doc.Para(text.FormatSentence(detail))

	}

	var notes string
	if !s.Tree.KeyPerson.IsUnknown() {
		doc.EmptyPara()

		detail := text.JoinSentenceParts("In this family tree,", doc.EncodeModelLink(s.Tree.KeyPerson.PreferredFamiliarFullName, s.Tree.KeyPerson), "acts as the primary reference point, with all relationships defined in relation to", s.Tree.KeyPerson.Gender.ObjectPronoun())
		notes = text.JoinSentences(notes, text.FormatSentence(detail))
		notes = text.JoinSentences(notes, text.FormatSentence(text.JoinSentenceParts("Names suffixed by the", md.DirectAncestorMarker, "symbol indicate direct ancestors")))
	}

	if !s.IncludePrivate {
		detail := text.JoinSentenceParts("The tree excludes information on people who are possibly alive or who have died within the past twenty years")
		notes = text.JoinSentences(notes, text.FormatSentence(detail))
	}

	if len(notes) > 0 {
		doc.EmptyPara()
		doc.Heading3("Notes")
		doc.Para(text.FormatSentence(notes))
	}

	if err := writePage(doc, root, fname); err != nil {
		return fmt.Errorf("write page: %w", err)
	}

	return nil
}

func (s *Site) WriteChartAncestors(root string) error {
	fname := "index.md"
	if s.GenerateHugo {
		fname = "_index.md"
	}

	generations := 8
	ancestors := s.PublishSet.Ancestors(s.Tree.KeyPerson, generations)

	doc := s.NewDocument()
	doc.Title("Ancestors of " + s.Tree.KeyPerson.PreferredFamiliarFullName)
	doc.Summary(text.JoinSentenceParts("This is a full list of ancestors of", doc.EncodeModelLink(s.Tree.KeyPerson.PreferredFamiliarFullName, s.Tree.KeyPerson)))
	doc.Layout(PageLayoutChartAncestors.String())

	g := 0
	doc.Heading3("Generation 1")

	doc.Para(text.JoinSentenceParts("1.", doc.EncodeLink(s.Tree.KeyPerson.PreferredFamiliarFullName, doc.LinkBuilder.LinkFor(s.Tree.KeyPerson))))
	for i := range ancestors {
		ig := -1
		idx := i + 2
		for idx > 0 {
			idx >>= 1
			ig++
		}
		if ig != g {
			g = ig
			if g == 1 {
				doc.Heading3("Generation 2: Parents")
			} else if g == 2 {
				doc.Heading3("Generation 3: Grandparents")
			} else if g == 3 {
				doc.Heading3("Generation 4: Great-Grandparents")
			} else if g == 4 {
				doc.Heading3("Generation 5: Great-Great-Grandparents")
			} else {
				doc.Heading3(fmt.Sprintf("Generation %d: %dx Great-Grandparents", g+1, g-2))
			}
		}
		if ancestors[i] != nil {
			detail := text.JoinSentenceParts(fmt.Sprintf("%d.", i+2), doc.EncodeBold(doc.EncodeLink(ancestors[i].PreferredFullName, doc.LinkBuilder.LinkFor(ancestors[i]))))

			var adds []string
			if ancestors[i].PrimaryOccupation != "" {
				adds = append(adds, ancestors[i].PrimaryOccupation)
			}
			if ancestors[i].BestBirthlikeEvent != nil && !ancestors[i].BestBirthlikeEvent.GetDate().IsUnknown() {
				adds = append(adds, EventWhatWhenWhere(ancestors[i].BestBirthlikeEvent, doc))
			}
			if ancestors[i].BestDeathlikeEvent != nil && !ancestors[i].BestDeathlikeEvent.GetDate().IsUnknown() {
				adds = append(adds, EventWhatWhenWhere(ancestors[i].BestDeathlikeEvent, doc))
			}

			detail = text.AppendClause(detail, text.JoinList(adds))
			doc.Para(detail)
		} else {

			name := "Not known"
			// Odd numbers are female, even numbers are male.
			// The child of entry n is found at (n-2)/2 if n is even and (n-3)/2 if n is odd.

			if i%2 == 0 {
				// male
				lb := (i - 2) / 2
				if lb >= 0 && ancestors[lb] != nil {
					name += " (father of " + ancestors[lb].PreferredFullName + ")"
				} else {
					lb = (lb - 2) / 2
					if lb >= 0 && ancestors[lb] != nil {
						name += " (grandfather of " + ancestors[lb].PreferredFullName + ")"
					} else {
						name += " (male)"
					}
				}
			} else {
				// female
				lb := (i - 3) / 2
				if lb >= 0 && ancestors[lb] != nil {
					name += " (mother of " + ancestors[lb].PreferredFullName + ")"
				} else {
					lb = (lb - 2) / 2
					if lb >= 0 && ancestors[lb] != nil {
						name += " (grandmother of " + ancestors[lb].PreferredFullName + ")"
					} else {
						name += " (female)"
					}
				}
			}

			doc.Para(text.JoinSentenceParts(fmt.Sprintf("%d.", i+2), name))
		}
	}

	baseDir := filepath.Join(root, s.ChartAncestorsDir)
	if err := writePage(doc, baseDir, fname); err != nil {
		return fmt.Errorf("failed to write ancestor overview: %w", err)
	}

	return nil
}

func (s *Site) WriteChartTrees(root string) error {
	fname := "index.md"
	if s.GenerateHugo {
		fname = "_index.md"
	}

	generations := 8
	ancestors := s.PublishSet.Ancestors(s.Tree.KeyPerson, generations)

	doc := s.NewDocument()
	doc.Title("Family Trees")
	doc.Summary(text.JoinSentenceParts("This is a list of family trees generated for various people"))
	doc.Layout(PageLayoutChartTrees.String())

	// index 14-29 are great-great grandparents, only produce chart if they have no known parents
	for i := 14; i <= 29; i++ {
		if ancestors[i] == nil {
			continue
		}

		if ancestors[i].Father != nil || ancestors[i].Mother != nil {
			continue
		}

		fname := filepath.Join(s.ChartTreesDir, ancestors[i].ID+".svg")
		if err := s.WriteDescendantTree(filepath.Join(root, fname), ancestors[i], 2); err != nil {
			return fmt.Errorf("failed to write descendant tree: %w", err)
		}
		doc.Para(doc.EncodeBold(doc.EncodeLink(ancestors[i].PreferredUniqueName, ancestors[i].ID+".svg")))
	}

	// index 30-61 are great-great-great grandparents, only produce chart if they have no known parents, but at a greater depth
	for i := 30; i <= 61; i++ {
		if ancestors[i] == nil {
			continue
		}

		if ancestors[i].Father != nil || ancestors[i].Mother != nil {
			continue
		}

		fname := filepath.Join(s.ChartTreesDir, ancestors[i].ID+".svg")
		if err := s.WriteDescendantTree(filepath.Join(root, fname), ancestors[i], 3); err != nil {
			return fmt.Errorf("failed to write descendant tree: %w", err)
		}
		doc.Para(doc.EncodeBold(doc.EncodeLink(ancestors[i].PreferredUniqueName, ancestors[i].ID+".svg")))
	}

	// produce chart for each member of a later generation
	for i := 62; i < len(ancestors); i++ {
		if ancestors[i] == nil {
			continue
		}
		fname := filepath.Join(s.ChartTreesDir, ancestors[i].ID+".svg")
		if err := s.WriteDescendantTree(filepath.Join(root, fname), ancestors[i], 4); err != nil {
			return fmt.Errorf("failed to write descendant tree: %w", err)
		}
		doc.Para(doc.EncodeBold(doc.EncodeLink(ancestors[i].PreferredUniqueName, ancestors[i].ID+".svg")))
	}

	baseDir := filepath.Join(root, s.ChartTreesDir)
	if err := writePage(doc, baseDir, fname); err != nil {
		return fmt.Errorf("failed to write chart trees index: %w", err)
	}

	return nil
}

func (s *Site) WriteDescendantTree(fname string, p *model.Person, depth int) error {
	ch, err := chart.BuildDescendantChart(s.Tree, p, 3, depth, true)
	if err != nil {
		return fmt.Errorf("build lineage: %w", err)
	}

	ch.Title = "Descendants of " + p.PreferredUniqueName
	ch.Notes = []string{}
	ch.Notes = append(ch.Notes, time.Now().Format("Generated _2 January 2006"))
	if !s.Tree.KeyPerson.IsUnknown() {
		ch.Notes = append(ch.Notes, "(★ denotes a direct ancestor of "+s.Tree.KeyPerson.PreferredFamiliarFullName+")")
	}

	opts := gtree.DefaultLayoutOptions()
	lay := ch.Layout(opts)

	svg, err := gtree.SVG(lay)
	if err != nil {
		return fmt.Errorf("render SVG: %w", err)
	}
	f, err := CreateFile(fname)
	if err != nil {
		return fmt.Errorf("create SVG file: %w", err)
	}
	defer f.Close()
	if err = os.WriteFile(fname, []byte(svg), 0o666); err != nil {
		return fmt.Errorf("write svg: %w", err)
	}
	return nil
}

func groupRelation(rel *model.Relation) (string, int) {
	var group string
	var groupPriority int
	distance := rel.Distance()
	if distance < 5 {
		group = "Close relations"
		groupPriority = 1
	} else if rel.IsDirectAncestor() {
		group = "Direct ancestors"
		groupPriority = 2
	} else if rel.IsCloseToDirectAncestor() {
		group = "Family of ancestors"
		groupPriority = 3
	} else if distance < 12 {
		group = "Distant relations"
		groupPriority = 4
	} else {
		group = "Others"
		groupPriority = 5
	}

	return group, groupPriority
}

func writePage(doc *md.Document, root string, fname string) error {
	f, err := CreateFile(filepath.Join(root, fname))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if err := doc.WriteMarkdown(f); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	return f.Close()
}

func (s *Site) BuildPublishSet(m model.PersonMatcher) error {
	subset, err := NewPublishSet(s.Tree, m)
	if err != nil {
		return fmt.Errorf("build publish set: %w", err)
	}

	s.PublishSet = subset
	return nil
}

type PublishSet struct {
	People       map[string]*model.Person
	Citations    map[string]*model.GeneralCitation
	Sources      map[string]*model.Source
	Repositories map[string]*model.Repository
	Places       map[string]*model.Place
	Families     map[string]*model.Family
	MediaObjects map[string]*model.MediaObject
	Events       map[model.TimelineEvent]bool
}

func NewPublishSet(t *tree.Tree, include model.PersonMatcher) (*PublishSet, error) {
	if t.KeyPerson != nil && !include(t.KeyPerson) {
		return nil, fmt.Errorf("key person must be included in subset")
	}

	ps := &PublishSet{
		People:       make(map[string]*model.Person),
		Citations:    make(map[string]*model.GeneralCitation),
		Sources:      make(map[string]*model.Source),
		Repositories: make(map[string]*model.Repository),
		Places:       make(map[string]*model.Place),
		Families:     make(map[string]*model.Family),
		MediaObjects: make(map[string]*model.MediaObject),
		Events:       make(map[model.TimelineEvent]bool),
	}

	for _, p := range t.People {
		if !include(p) {
			continue
		}
		ps.People[p.ID] = p
		for _, ev := range p.Timeline {
			ps.Events[ev] = true
		}
		if p.BestBirthlikeEvent != nil {
			ps.Events[p.BestBirthlikeEvent] = true
		}
		if p.BestDeathlikeEvent != nil {
			ps.Events[p.BestDeathlikeEvent] = true
		}

		if p.CauseOfDeath != nil {
			for _, c := range p.CauseOfDeath.Citations {
				ps.Citations[c.ID] = c
			}
		}

		for _, n := range p.KnownNames {
			for _, c := range n.Citations {
				ps.Citations[c.ID] = c
			}
		}
		for _, f := range p.MiscFacts {
			for _, c := range f.Citations {
				ps.Citations[c.ID] = c
			}
		}
		for _, o := range p.Occupations {
			for _, c := range o.Citations {
				ps.Citations[c.ID] = c
			}
		}
		for _, a := range p.Associations {
			for _, c := range a.Citations {
				ps.Citations[c.ID] = c
			}
		}
	}

	includedEvents := func(ev model.TimelineEvent) bool {
		return ps.Events[ev]
	}

	for ev := range ps.Events {
		pl := ev.GetPlace()
		if pl != nil {
			if _, ok := ps.Places[pl.ID]; !ok {
				pl.Timeline = model.FilterEventList(pl.Timeline, includedEvents)
				if len(pl.Timeline) > 0 {
					ps.Places[pl.ID] = pl
					for pl.Parent != nil {
						ps.Places[pl.Parent.ID] = pl.Parent
						pl = pl.Parent
					}
				}
			}
		}

		for _, c := range ev.GetCitations() {
			ps.Citations[c.ID] = c
		}
	}

	for _, c := range ps.Citations {
		for _, mo := range c.MediaObjects {
			ps.MediaObjects[mo.ID] = mo
		}
		if c.Source != nil {
			ps.Sources[c.Source.ID] = c.Source
		}

		c.PeopleCited = model.FilterPersonList(c.PeopleCited, include)
		c.EventsCited = model.FilterEventList(c.EventsCited, includedEvents)
	}

	includedCitations := func(c *model.GeneralCitation) bool {
		_, ok := ps.Citations[c.ID]
		return ok
	}

	for _, mo := range ps.MediaObjects {
		mo.Citations = model.FilterCitationList(mo.Citations, includedCitations)
	}

	for _, so := range ps.Sources {
		so.EventsCiting = model.FilterEventList(so.EventsCiting, includedEvents)
		for _, rr := range so.RepositoryRefs {
			if rr.Repository != nil {
				ps.Repositories[rr.Repository.ID] = rr.Repository
			}
		}
	}

	return ps, nil
}

func (ps *PublishSet) Includes(v any) bool {
	switch vt := v.(type) {
	case *model.Person:
		_, ok := ps.People[vt.ID]
		return ok
	case *model.GeneralCitation:
		_, ok := ps.Citations[vt.ID]
		return ok
	case *model.Source:
		_, ok := ps.Sources[vt.ID]
		return ok
	case *model.Family:
		_, ok := ps.Families[vt.ID]
		return ok
	case *model.Place:
		_, ok := ps.Places[vt.ID]
		return ok
	case *model.MediaObject:
		_, ok := ps.MediaObjects[vt.ID]
		return ok
	case *model.Repository:
		_, ok := ps.Repositories[vt.ID]
		return ok
	case model.TimelineEvent:
		_, ok := ps.Events[vt]
		return ok
	default:
		return false
	}
}

// Metrics

// In general all metrics exclude redacted people

// NumberOfPeople returns the number of people in the tree.
// It excludes redacted people.
func (t *PublishSet) NumberOfPeople() int {
	num := 0
	for _, p := range t.People {
		if !p.Redacted {
			num++
		}
	}
	return num
}

// AncestorSurnameDistribution returns a map of surnames and the number
// of direct ancestors with that surname
// It excludes redacted people.
func (ps *PublishSet) AncestorSurnameDistribution() map[string]int {
	dist := make(map[string]int)
	for _, p := range ps.People {
		if p.Redacted {
			continue
		}
		if p.PreferredFamilyName == model.UnknownNamePlaceholder {
			continue
		}
		if p.RelationToKeyPerson.IsDirectAncestor() {
			dist[p.PreferredFamilyName]++
		}
	}
	return dist
}

// TreeSurnameDistribution returns a map of surnames and the number
// of people in the tree with that surname
// It excludes redacted people.
func (ps *PublishSet) TreeSurnameDistribution() map[string]int {
	dist := make(map[string]int)
	for _, p := range ps.People {
		if p.Redacted {
			continue
		}
		if p.PreferredFamilyName == model.UnknownNamePlaceholder {
			continue
		}
		dist[p.PreferredFamilyName]++
	}
	return dist
}

// OldestPeople returns a list of the oldest people in the tree, sorted by descending age
// It excludes redacted people.
func (ps *PublishSet) OldestPeople(limit int) []*model.Person {
	h := new(PersonWithAgeHeap)
	heap.Init(h)

	for _, p := range ps.People {
		if p.Redacted {
			continue
		}
		age, ok := p.AgeInYearsAtDeath()
		if !ok {
			continue
		}
		heap.Push(h, &PersonWithAge{Person: p, Age: age})
		if h.Len() > limit {
			heap.Pop(h)
		}
	}

	list := make([]*model.Person, h.Len())
	for i := len(list) - 1; i >= 0; i-- {
		pa := heap.Pop(h).(*PersonWithAge)
		list[i] = pa.Person
	}
	return list
}

// Ancestors returns the ancestors of p. The returned list is ordered such that the
// father of entry n is found at (n+2)*2-2, the mother of entry n is found at (n+2)*2-1
// The list will always contain 2^n entries, with unknown ancestors left as nil at the
// appropriate index.
// Odd numbers are female, even numbers are male.
// The child of entry n is found at (n-2)/2 if n is even and (n-3)/2 if n is odd.
// 0: father
// 1: mother
// 2: father's father
// 3: father's mother
// 4: mother's father
// 5: mother's mother
// 6: father's father's father
// 7: father's father's mother
// 8: father's mother's father
// 9: father's mother's mother
// 10: mother's father's father
// 11: mother's father's mother
// 12: mother's mother's father
// 13: mother's mother's mother
// 14: father's father's father's father
// 15: father's father's father's mother
// 16: father's father's mother's father
// 17: father's father's mother's mother
// 18: father's mother's father's father
// 19: father's mother's father's mother
// 20: father's mother's mother's father
// 21: father's mother's mother's mother
// 22: mother's father's father's father
// 23: mother's father's father's mother
// 24: mother's father's mother's father
// 25: mother's father's mother's mother
// 26: mother's mother's father's father
// 27: mother's mother's father's mother
// 28: mother's mother's mother's father
// 29: mother's mother's mother's mother
func (ps *PublishSet) Ancestors(p *model.Person, generations int) []*model.Person {
	n := 0
	f := 2
	for i := 0; i < generations; i++ {
		n += f
		f *= 2
	}
	a := make([]*model.Person, n)

	a[0] = p.Father
	a[1] = p.Mother
	for idx := 0; idx < n; idx++ {
		if a[idx] == nil {
			continue
		}
		if a[idx].Father != nil {
			if (idx+2)*2-2 < n {
				a[(idx+2)*2-2] = a[idx].Father
			}
		}
		if a[idx].Mother != nil {
			if (idx+2)*2-1 < n {
				a[(idx+2)*2-1] = a[idx].Mother
			}
		}
	}

	return a
}

type PersonWithAge struct {
	Person *model.Person
	Age    int
}

type PersonWithAgeHeap []*PersonWithAge

func (h PersonWithAgeHeap) Len() int           { return len(h) }
func (h PersonWithAgeHeap) Less(i, j int) bool { return h[i].Age < h[j].Age }
func (h PersonWithAgeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PersonWithAgeHeap) Push(x interface{}) {
	*h = append(*h, x.(*PersonWithAge))
}

func (h *PersonWithAgeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

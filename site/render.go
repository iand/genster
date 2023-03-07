package site

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/iand/gdate"
	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/werr"
)

func RenderPersonPage(s *Site, p *model.Person) (*md.Document, error) {
	pov := &model.POV{Person: p}

	d := s.NewDocument()
	d.Layout(md.PageLayoutPerson)
	d.ID(p.ID)
	d.Title(p.PreferredUniqueName)

	if p.Redacted {
		d.Summary("information withheld to preserve privacy")
		return d, nil
	}

	if p.Olb != "" {
		d.Summary(p.Olb)
	}
	d.AddTags(CleanTags(p.Tags))

	// Render narrative
	n := &Narrative{
		Statements: make([]Statement, 0),
	}

	// Everyone has an intro
	n.Statements = append(n.Statements, &IntroStatement{
		Principal: p,
	})
	// If death is known, add it
	if p.BestDeathlikeEvent != nil {
		n.Statements = append(n.Statements, &DeathStatement{
			Principal: p,
		})
	}

	eventsInNarrative := make(map[model.TimelineEvent]bool)

	baptisms := []*model.BaptismEvent{}
	for _, ev := range p.Timeline {
		switch tev := ev.(type) {
		case *model.BaptismEvent:
			if tev != p.BestBirthlikeEvent {
				baptisms = append(baptisms, tev)
				eventsInNarrative[ev] = true
			}
		case *model.ProbateEvent:
			if tev != p.BestDeathlikeEvent {
				eventsInNarrative[ev] = true
				n.Statements = append(n.Statements, &WillAndProbateStatement{
					Principal: p,
					Event:     ev,
				})
			}
		case *model.CensusEvent:
			eventsInNarrative[ev] = true
			n.Statements = append(n.Statements, &CensusStatement{
				Principal: p,
				Event:     tev,
			})
			// TODO: wills
			// case *model.WillEvent:
			// 	if tev != p.BestDeathlikeEvent {
			// 		willsAndProbate = append(willsAndProbate, tev)
			// 	}
		case *model.IndividualNarrativeEvent:
			eventsInNarrative[ev] = true
			n.Statements = append(n.Statements, &VerbatimStatement{
				Principal: p,
				Detail:    ev.GetDetail(),
				Date:      ev.GetDate(),
				Citations: ev.GetCitations(),
			})
		case *model.ArrivalEvent:
			eventsInNarrative[ev] = true
			n.Statements = append(n.Statements, &ArrivalDepartureStatement{
				Principal: p,
				Event:     ev,
			})

		case *model.DepartureEvent:
			eventsInNarrative[ev] = true
			n.Statements = append(n.Statements, &ArrivalDepartureStatement{
				Principal: p,
				Event:     ev,
			})

		}
	}
	if len(baptisms) > 0 {
		sort.Slice(baptisms, func(i, j int) bool {
			return gdate.SortsBefore(baptisms[i].GetDate(), baptisms[j].GetDate())
		})
		n.Statements = append(n.Statements, &BaptismsStatement{
			Principal: p,
			Events:    baptisms,
		})
	}

	for _, f := range p.Families {
		n.Statements = append(n.Statements, &FamilyStatement{
			Principal: p,
			Family:    f,
		})
	}

	n.Render(pov, d)

	if p.EditLink != nil {
		d.Para(d.EncodeLink(p.EditLink.Title, p.EditLink.URL, false))
	}

	t := &model.Timeline{
		Events: make([]model.TimelineEvent, 0, len(p.Timeline)),
	}
	for _, ev := range p.Timeline {
		if gdate.IsUnknown(ev.GetDate()) && ev.GetPlace().IsUnknown() {
			p.MiscFacts = append(p.MiscFacts, model.Fact{
				Category: ev.GetTitle(),
				Detail:   ev.GetDetail(),
			})
		} else {
			t.Events = append(t.Events, ev)
		}
	}

	if len(p.Timeline) > 0 {
		d.EmptyPara()
		d.Heading2("Timeline")

		if err := RenderTimeline(t, pov, eventsInNarrative, d); err != nil {
			return nil, werr.Wrap(err)
		}
	}

	if len(p.MiscFacts) > 0 {
		d.EmptyPara()
		d.Heading2("Other Information")
		if err := RenderFacts(p.MiscFacts, pov, d); err != nil {
			return nil, werr.Wrap(err)
		}
	}

	links := make([]string, 0, len(p.Links))
	for _, l := range p.Links {
		links = append(links, d.EncodeLink(l.Title, l.URL, false))
	}

	if len(links) > 0 {
		d.Heading2("Links")
		d.UnorderedList(links)
	}

	return d, nil
}

func RenderSourcePage(s *Site, sr *model.Source) (*md.Document, error) {
	d := s.NewDocument()

	d.Title(sr.Title)
	d.Layout(md.PageLayoutSource)
	d.ID(sr.ID)
	d.AddTags(CleanTags(sr.Tags))

	return d, nil
}

func RenderPlacePage(s *Site, p *model.Place) (*md.Document, error) {
	pov := &model.POV{Place: p}

	d := s.NewDocument()

	d.Title(p.PreferredName)
	d.Layout(md.PageLayoutPlace)
	d.ID(p.ID)
	d.AddTags(CleanTags(p.Tags))

	desc := p.PreferredName + " is a" + text.MaybeAn(p.PlaceType.String())

	if !p.Parent.IsUnknown() {
		desc += " in " + d.EncodeModelLink(p.Parent.PreferredUniqueName, p.Parent, false)
	}

	d.Para(text.FinishSentence(desc))

	t := &model.Timeline{
		Events: p.Timeline,
	}

	if len(p.Timeline) > 0 {
		d.EmptyPara()
		d.Heading2("Timeline")

		if err := RenderTimeline(t, pov, nil, d); err != nil {
			return nil, werr.Wrap(err)
		}
	}

	if len(p.Links) > 0 {
		d.Heading2("Links")
		for _, l := range p.Links {
			d.Para(d.EncodeLink(l.Title, l.URL, false))
		}
	}

	return d, nil
}

func RenderInferencesPage(s *Site) (*md.Document, error) {
	d := s.NewDocument()
	d.Layout(md.PageLayoutInferences)
	d.Title("Inferences Made")

	d.Heading2("Inferences Made")

	wroteInferences := false
	for _, p := range s.Tree.People {
		items := make([]string, 0)
		for _, inf := range p.Inferences {
			items = append(items, inf.Type+" "+inf.Value)
		}

		if len(items) > 0 {
			d.Heading3(d.EncodeModelLink(p.PreferredUniqueName, p, true))
			d.UnorderedList(items)
			wroteInferences = true
		}

	}

	if !wroteInferences {
		d.Para("None made")
	}

	return d, nil
}

func RenderAnomaliesPage(s *Site) (*md.Document, error) {
	d := s.NewDocument()
	d.Layout(md.PageLayoutInferences)
	d.Title("Anomalies Found")

	d.Heading2("People")

	wroteAnomalies := false
	for _, p := range s.Tree.People {

		categories := make([]string, 0)
		anomaliesByCategory := make(map[string][]*model.Anomaly)

		for _, a := range p.Anomalies {
			a := a // avoid shadowing
			al, ok := anomaliesByCategory[a.Category]
			if ok {
				al = append(al, a)
				anomaliesByCategory[a.Category] = al
				continue
			}

			categories = append(categories, a.Category)
			anomaliesByCategory[a.Category] = []*model.Anomaly{a}
		}
		sort.Strings(categories)

		if len(anomaliesByCategory) > 0 {
			d.Heading4(p.PreferredUniqueName)
			if p.EditLink != nil {
				d.Para(d.EncodeModelLink("View "+p.PreferredFullName, p, false) + " or " + d.EncodeLink(p.EditLink.Title, p.EditLink.URL, false))
			} else {
				d.Para(d.EncodeModelLink("View "+p.PreferredFullName, p, false))
			}
			for _, cat := range categories {
				d.Heading4(cat + " anomalies")
				al := anomaliesByCategory[cat]
				items := make([][2]string, 0, len(al))

				for _, a := range al {
					items = append(items, [2]string{
						a.Context,
						a.Text,
					})
				}
				d.DefinitionList(items)
			}
			wroteAnomalies = true
		}

	}

	if !wroteAnomalies {
		d.Para("None found")
	}

	return d, nil
}

func EventTitle(ev model.TimelineEvent, enc BlockEncoder, obs *model.POV) string {
	// if gev, ok := ev.(*model.GeneralEvent); ok && obs != nil && obs.Person != nil {
	// 	fmt.Printf(" - [%s](%s) - General (%s)\n", obs.Person.PreferredFullName, obs.Person.Page, gev.Title)
	// }
	// if _, ok := ev.(*model.CensusEvent); ok && obs != nil && obs.Person != nil {
	// 	fmt.Printf(" - [%s](%s) - Census\n", obs.Person.PreferredFullName, obs.Person.Page)
	// }
	titleTemplate := ""

	simpleIndividualEventTemplate := `
{{- if .Name -}}
	{{- if .Relation -}}
		{{- .Relation }}, {{ .Name }}, {{ if .Uncertain }}probably {{ end }}{{ .Event }}
	{{- else -}}
		{{ .Name }} {{ if .Uncertain }}probably {{ end }}{{ .Event }}
	{{- end -}}
{{- else -}}
	{{- if .Relation -}}
		{{- .Relation }} {{ if .Uncertain }}probably {{ end }}{{ .Event }}
	{{- else -}}
		{{ if .Uncertain }}probably {{ end }}{{ .Event }}
	{{- end -}}
{{- end -}}`

	simplePartyEventTemplate := `
{{- if .Party1 -}}
	{{- if .Party2 -}}
		{{ .Party2 }} {{ if .Uncertain }}probably {{ end }}{{ .Event }} {{ .Party1 }}
	{{- else -}}
		{{ if .Uncertain }}probably {{ end }}{{ .Event }} {{ .Party1 }}
	{{- end -}}
{{- else -}}
	{{- if .Party2 -}}
		{{ if .Uncertain }}probably {{ end }}{{ .Event }} {{ .Party2 }}
	{{- else -}}
		{{ if .Uncertain }}probably {{ end }}{{ .Event }}
	{{- end -}}
{{- end -}}`

	inAtFrom := ""
	if !ev.GetPlace().IsUnknown() {
		inAtFrom = ev.GetPlace().PlaceType.InAt()
	}

	switch ev.(type) {

	case *model.BirthEvent:
		titleTemplate = simpleIndividualEventTemplate

	case *model.BaptismEvent:
		titleTemplate = simpleIndividualEventTemplate

	case *model.DeathEvent:
		titleTemplate = simpleIndividualEventTemplate

	case *model.BurialEvent:
		titleTemplate = simpleIndividualEventTemplate

	case *model.CremationEvent:
		titleTemplate = simpleIndividualEventTemplate

	case *model.ProbateEvent:
		titleTemplate = `
{{- if .Name -}}
	{{- if .Relation -}}
		probate {{ if .Uncertain }}probably {{ end }}granted for {{ .Relation }}, {{ .Name }}
	{{- else -}}
		probate {{ if .Uncertain }}probably {{ end }}granted for {{ .Name }}
	{{- end -}}
{{- else -}}
	{{- if .Relation -}}
		probate {{ if .Uncertain }}probably {{ end }}granted for {{ .Relation }}
	{{- else -}}
		probate {{ if .Uncertain }}probably {{ end }}granted
	{{- end -}}
{{- end -}}`

	case *model.CensusEvent:
		titleTemplate = `
{{- if .Name -}}
	{{- if .Relation -}}
		{{ .Relation }}, {{ .Name }} {{ if .Uncertain }}probably {{ end }}recorded in the census
	{{- else -}}
		{{ .Name }} {{ if .Uncertain }}probably {{ end }}recorded in the census
	{{- end -}}
{{- else -}}
	{{- if .Relation -}}
		{{ .Relation }} {{ if .Uncertain }}probably {{ end }}recorded in the census
	{{- else -}}
		{{ if .Uncertain }}probably {{ end }}recorded in the census
	{{- end -}}
{{- end -}}`

	case *model.MarriageEvent:
		titleTemplate = simplePartyEventTemplate

	case *model.MarriageLicenseEvent:
		titleTemplate = `
{{- if .Party1 -}}
	{{- if .Party2 -}}
		{{ .Party1 }} and {{ .Party2 }}{{ if .Uncertain }}probably {{ end }}obtained a license to marry
	{{- else -}}
		{{ if .Uncertain }}probably {{ end }}obtained a license to marry {{ .Party1 }}
	{{- end -}}
{{- else -}}
	{{- if .Party2 -}}
		{{ if .Uncertain }}probably {{ end }}obtained a license to marry {{ .Party2 }}
	{{- else -}}
		{{ if .Uncertain }}probably {{ end }}obtained a license to marry
	{{- end -}}
{{- end -}}`

	case *model.MarriageBannsEvent:
		titleTemplate = `
{{- if .Party1 -}}
	{{- if .Party2 -}}
		banns were {{ if .Uncertain }}probably {{ end }}read to announce the marriage of {{ .Party1 }} and {{ .Party2 }}
	{{- else -}}
		banns were {{ if .Uncertain }}probably {{ end }}read to announce marriage to {{ .Party1 }}
	{{- end -}}
{{- else -}}
	{{- if .Party2 -}}
		banns were {{ if .Uncertain }}probably {{ end }}read to announce marriage to {{ .Party2 }}
	{{- else -}}
		banns were {{ if .Uncertain }}probably {{ end }}read
	{{- end -}}
{{- end -}}`

	case *model.DivorceEvent:
		titleTemplate = `
{{- if .Party1 -}}
	{{- if .Party2 -}}
		{{ .Party1 }} and {{ .Party2 }} {{ if .Uncertain }}probably {{ end }}divorced
	{{- else -}}
		was {{ if .Uncertain }}probably {{ end }}divorced from {{ .Party1 }}
	{{- end -}}
{{- else -}}
	{{- if .Party2 -}}
		was {{ if .Uncertain }}probably {{ end }}divorced from {{ .Party2 }}
	{{- else -}}
		{{ if .Uncertain }}probably {{ end }}divorced
	{{- end -}}
{{- end -}}`

	case *model.DepartureEvent:
		titleTemplate = simpleIndividualEventTemplate
		inAtFrom = "from"

	case *model.ArrivalEvent:
		titleTemplate = simpleIndividualEventTemplate
		inAtFrom = "at"

	case *model.IndividualNarrativeEvent:
		titleTemplate = ""
	case *model.PlaceholderIndividualEvent:
		titleTemplate = ""
	case *model.PlaceholderPartyEvent:
		titleTemplate = ""
	case *model.ResidenceRecordedEvent:
		titleTemplate = `
{{- if .Name -}}
	{{- if .Relation -}}
		{{ .Relation }}, {{ .Name }} {{ if .Uncertain }}probable {{ end }}residence
	{{- else -}}
		{{ .Name }} {{ if .Uncertain }}probable {{ end }}residence
	{{- end -}}
{{- else -}}
	{{- if .Relation -}}
		{{ .Relation }} {{ if .Uncertain }}probable {{ end }}residence
	{{- else -}}
		{{ if .Uncertain }}probable {{ end }}residence
	{{- end -}}
{{- end -}}`

	// case *model.FamilyStartEvent:
	// 	eventName = "began a family"
	// 	onePartyTemplate = "began a family with {name}"
	// 	twoPartyTemplate = "{name1} and {name2} began a family"
	// case *model.FamilyEndEvent:
	// 	eventName = "ended a family"
	// 	onePartyTemplate = "ended a family with {name}"
	// 	twoPartyTemplate = "{name1} and {name2} ended a family"
	default:
		panic(fmt.Sprintf("unhandled event type: %T", ev))
	}

	title := ev.GetTitle()
	if titleTemplate != "" {
		data := map[string]any{
			"Event":     ev.Action(),
			"Uncertain": ev.IsInferred(),
		}
		if iev, ok := ev.(model.IndividualTimelineEvent); ok {
			principal := iev.GetPrincipal()
			if !principal.IsUnknown() {
				if !principal.SameAs(obs.Person) {
					data["Name"] = enc.EncodeModelLink(principal.PreferredFullName, principal, true)
					if !obs.Person.IsUnknown() {
						data["Relation"] = principal.RelationTo(obs.Person, ev.GetDate())
					}
				}
			}

		} else if pev, ok := ev.(model.PartyTimelineEvent); ok {
			p1 := pev.GetParty1()
			p2 := pev.GetParty2()
			if !p1.IsUnknown() && !obs.Person.SameAs(p1) {
				data["Party1"] = enc.EncodeModelLink(p1.PreferredFullName, p1, true)
			}
			if !p2.IsUnknown() && !obs.Person.SameAs(p2) {
				data["Party2"] = enc.EncodeModelLink(p2.PreferredFullName, p2, true)
			}
		}

		tmpl, err := template.New("title").Parse(titleTemplate)
		if err != nil {
			panic(err.Error())
		}
		b := new(strings.Builder)
		err = tmpl.Execute(b, data)
		if err != nil {
			panic(err.Error())
		}
		title = b.String()
	}

	if ev.GetDateType() == model.EventDateTypeRecorded {
		title += " recorded"
	}
	title += " " + ev.GetDate().Occurrence()

	pl := ev.GetPlace()
	if !pl.IsUnknown() && !pl.SameAs(obs.Place) {
		title += " " + inAtFrom + " " + enc.EncodeModelLink(pl.PreferredName, pl, true)
	}
	return EncodeWithCitations(text.UpperFirst(title), ev.GetCitations(), enc)
}

// cleanCitationDetail removes some redundant information that isn't necessary when a source is included
func cleanCitationDetail(page string) string {
	page = strings.TrimPrefix(page, "The National Archives of the UK (TNA); Kew, Surrey, England; Census Returns of England and Wales, 1891;")
	page = strings.TrimPrefix(page, "The National Archives; Kew, London, England; 1871 England Census; ")

	page = text.FinishSentence(page)
	return page
}

func RenderTimeline(t *model.Timeline, pov *model.POV, eventsInNarrative map[model.TimelineEvent]bool, enc StructuredMarkdownEncoder) error {
	enc.EmptyPara()
	if len(t.Events) == 0 {
		return nil
	}
	model.SortTimelineEvents(t.Events)

	events := make([][2]string, 0, len(t.Events))
	for i := range t.Events {
		var detail string
		if !eventsInNarrative[t.Events[i]] {
			detail = t.Events[i].GetDetail()
		}
		events = append(events, [2]string{
			EventTitle(t.Events[i], enc, pov),
			detail,
		})
	}

	enc.DefinitionList(events)
	return nil // werr.Wrap(fmt.Errorf("RenderTimeline error test (REMOVEME)"))
}

func RenderFacts(facts []model.Fact, pov *model.POV, enc StructuredMarkdownEncoder) error {
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
		if seen[s] {
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

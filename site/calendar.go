package site

import (
	"fmt"
	"sort"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

type Calendar struct {
	Events []model.TimelineEvent
}

func (c *Calendar) RenderPage(s *Site) (render.Page[md.Text], error) {
	monthNames := []string{
		1:  "January",
		2:  "February",
		3:  "March",
		4:  "April",
		5:  "May",
		6:  "June",
		7:  "July",
		8:  "August",
		9:  "September",
		10: "October",
		11: "November",
		12: "December",
	}

	doc := s.NewDocument()

	type eventDay struct {
		day  int
		year int
		ev   model.TimelineEvent
		text string
	}

	var eventDays []eventDay

	month := 0
	items := []md.Text{}
	for _, ev := range c.Events {
		y, m, d, ok := ev.GetDate().YMD()
		if !ok {
			continue
		}
		if month == 0 {
			month = m
		}

		evd := eventDay{
			day:  d,
			year: y,
			ev:   ev,
		}

		switch tev := ev.(type) {
		case *model.BirthEvent:
			evd.text = doc.EncodeModelLink(doc.EncodeText(tev.Principal.PreferredUniqueName), tev.Principal).String() + " was born."
		case *model.BaptismEvent:
			evd.text = doc.EncodeModelLink(doc.EncodeText(tev.Principal.PreferredUniqueName), tev.Principal).String() + " was baptised."
		case *model.DeathEvent:
			evd.text = doc.EncodeModelLink(doc.EncodeText(tev.Principal.PreferredUniqueName), tev.Principal).String() + " died."
		case *model.BurialEvent:
			evd.text = doc.EncodeModelLink(doc.EncodeText(tev.Principal.PreferredUniqueName), tev.Principal).String() + " was buried."
		case *model.MarriageEvent:
			evd.text = doc.EncodeModelLink(doc.EncodeText(tev.Husband.PreferredUniqueName), tev.Husband).String() + " and " + doc.EncodeModelLink(doc.EncodeText(tev.Wife.PreferredUniqueName), tev.Wife).String() + " were married."
		default:
			evd.text = EventWhatWhenWhere(tev, doc)
		}

		if tev, ok := ev.(model.IndividualTimelineEvent); ok {
			p := tev.GetPrincipal()
			if p.RelationToKeyPerson.IsDirectAncestor() {
				evd.text += " " + text.UpperFirst(p.Gender.SubjectPronoun()) + " is a direct ancestor."
			}
		} else if tev, ok := ev.(model.UnionTimelineEvent); ok {
			p1 := tev.GetHusband()
			p2 := tev.GetHusband()
			if p1.RelationToKeyPerson.IsDirectAncestor() && p2.RelationToKeyPerson.IsDirectAncestor() {
				evd.text += " Both are direct ancestors."
			} else if p1.RelationToKeyPerson.IsDirectAncestor() {
				evd.text += " " + text.UpperFirst(p1.Gender.SubjectPronoun()) + " is a direct ancestor."
			} else if p2.RelationToKeyPerson.IsDirectAncestor() {
				evd.text += " " + text.UpperFirst(p2.Gender.SubjectPronoun()) + " is a direct ancestor."
			}
		}

		eventDays = append(eventDays, evd)
	}
	doc.Title(fmt.Sprintf("On this day in %s", monthNames[month]))
	doc.Layout(PageLayoutCalendar.String())
	doc.SetFrontMatterField("month", monthNames[month])

	sort.Slice(eventDays, func(i, j int) bool {
		if eventDays[i].day == eventDays[j].day {
			return eventDays[i].year < eventDays[j].year
		}
		return eventDays[i].day < eventDays[j].day
	})

	day := 0
	for _, evd := range eventDays {
		if day != evd.day {
			if len(items) > 0 {
				doc.UnorderedList(items)
				doc.EmptyPara()
				items = items[:0]
			}

			day = evd.day
			doc.Heading2(md.Text(fmt.Sprintf("%d%s", day, text.CardinalSuffix(day))), fmt.Sprintf("day%d", day))
		}

		items = append(items, md.Text(evd.text))

	}

	if len(items) > 0 {
		doc.UnorderedList(items)
	}

	return doc, nil
}

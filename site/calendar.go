package site

import (
	"fmt"
	"sort"

	"github.com/iand/gdate"
	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

type Calendar struct {
	Events []model.TimelineEvent
}

func (c *Calendar) RenderPage(s *Site) (*md.Document, error) {
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

	b := s.NewDocument()

	type eventDay struct {
		day  int
		year int
		ev   model.TimelineEvent
		text string
	}

	var eventDays []eventDay

	month := 0
	items := []string{}
	for _, ev := range c.Events {
		dt, ok := gdate.AsPrecise(ev.GetDate())
		if !ok {
			continue
		}
		if month == 0 {
			month = dt.M
		}

		evd := eventDay{
			day:  dt.D,
			year: dt.Y,
			ev:   ev,
		}

		switch tev := ev.(type) {
		case *model.BirthEvent:
			evd.text = b.EncodeModelLink(tev.Principal.PreferredUniqueName, tev.Principal, false) + " was born."
		case *model.BaptismEvent:
			evd.text = b.EncodeModelLink(tev.Principal.PreferredUniqueName, tev.Principal, false) + " was baptised."
		case *model.DeathEvent:
			evd.text = b.EncodeModelLink(tev.Principal.PreferredUniqueName, tev.Principal, false) + " died."
		case *model.BurialEvent:
			evd.text = b.EncodeModelLink(tev.Principal.PreferredUniqueName, tev.Principal, false) + " was buried."
		case *model.MarriageEvent:
			evd.text = b.EncodeModelLink(tev.Party1.PreferredUniqueName, tev.Party1, false) + " and " + b.EncodeModelLink(tev.Party2.PreferredUniqueName, tev.Party2, false) + " were married."
		default:
			evd.text = EventTitle(tev, b, &model.POV{})
		}

		if tev, ok := ev.(model.IndividualTimelineEvent); ok {
			p := tev.GetPrincipal()
			if p.RelationToKeyPerson.IsDirectAncestor() {
				evd.text += " " + text.UpperFirst(p.Gender.SubjectPronoun()) + " is a direct ancestor."
			}
		} else if tev, ok := ev.(model.PartyTimelineEvent); ok {
			p1 := tev.GetParty1()
			p2 := tev.GetParty1()
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
	b.Title(fmt.Sprintf("On this day in %s", monthNames[month]))
	b.Layout(md.PageLayoutCalendar)
	b.SetFrontMatterField("month", monthNames[month])

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
				b.UnorderedList(items)
				b.EmptyPara()
				items = items[:0]
			}

			day = evd.day
			b.Heading2(fmt.Sprintf("%d%s", day, text.CardinalSuffix(day)))
		}

		items = append(items, evd.text)

	}

	if len(items) > 0 {
		b.UnorderedList(items)
	}

	return b, nil
}

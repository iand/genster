package narrative

import (
	"testing"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// A spouse's timeline entry for a union event uses the event's transitive phrasing.
func TestTimelineUnionEventTitleForSpouse(t *testing.T) {
	testCases := []struct {
		name string
		ev   func(g model.GeneralUnionEvent) model.TimelineEvent
		want string
	}{
		{
			name: "marriage",
			ev:   func(g model.GeneralUnionEvent) model.TimelineEvent { return &model.MarriageEvent{GeneralUnionEvent: g} },
			want: "Married Mary Frost.",
		},
		{
			name: "banns",
			ev: func(g model.GeneralUnionEvent) model.TimelineEvent {
				return &model.MarriageBannsEvent{GeneralUnionEvent: g}
			},
			want: "Had banns read to marry Mary Frost.",
		},
		{
			name: "licence",
			ev: func(g model.GeneralUnionEvent) model.TimelineEvent {
				return &model.MarriageLicenseEvent{GeneralUnionEvent: g}
			},
			want: "Obtained a licence to marry Mary Frost.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "p", PreferredUniqueName: "James Howes", Gender: model.GenderMale}
			s := &model.Person{ID: "s", PreferredUniqueName: "Mary Frost", Gender: model.GenderFemale}
			ev := tc.ev(model.GeneralUnionEvent{
				Husband: p,
				Wife:    s,
				Participants: []*model.EventParticipant{
					{Person: p, Role: model.EventRoleHusband},
					{Person: s, Role: model.EventRoleWife},
				},
			})

			enc := &md.Content{}
			fmtr := NewNarrativeTimelineEntryFormatter[md.Text](&model.POV{Person: p}, enc, logging.Default(), &TimelineNameChooser{}, false)
			checkContains(t, fmtr.Title(0, ev), tc.want, "to marry to marry", "Marriage banns read")
		})
	}
}

func TestPersonMarriageSummaryStartEventVerb(t *testing.T) {
	type fam struct {
		spouse string
		event  func(d *model.Date) model.TimelineEvent
		year   int
	}
	marriage := func(d *model.Date) model.TimelineEvent {
		return &model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: d}}
	}
	banns := func(d *model.Date) model.TimelineEvent {
		return &model.MarriageBannsEvent{GeneralEvent: model.GeneralEvent{Date: d}}
	}
	licence := func(d *model.Date) model.TimelineEvent {
		return &model.MarriageLicenseEvent{GeneralEvent: model.GeneralEvent{Date: d}}
	}

	testCases := []struct {
		name    string
		fams    []fam
		want    string
		notWant []string
	}{
		{
			name: "single banns",
			fams: []fam{{"Mary Frost", banns, 1764}},
			want: "James had banns read to marry Mary Frost in 1764.",
		},
		{
			name: "single licence",
			fams: []fam{{"Mary Frost", licence, 1764}},
			want: "James obtained a licence to marry Mary Frost in 1764.",
		},
		{
			name: "single marriage",
			fams: []fam{{"Mary Frost", marriage, 1764}},
			want: "James married Mary Frost in 1764.",
		},
		{
			name: "banns then marriage",
			fams: []fam{{"Mary Frost", banns, 1764}, {"Susannah Furness", marriage, 1769}},
			want: "James had banns read to marry Mary Frost in 1764 and married Susannah Furness in 1769.",
		},
		{
			name: "two banns share the verb",
			fams: []fam{{"Mary Frost", banns, 1764}, {"Susannah Furness", banns, 1769}},
			want: "James had banns read to marry Mary Frost in 1764 and Susannah Furness in 1769.",
		},
		{
			name: "first marriage dated before a year keeps its date",
			fams: []fam{{"Alfred Mayson", marriage, 0}, {"John Beardsley", marriage, 1912}},
			want: "James married Alfred Mayson before 1884 and John Beardsley in 1912.",
		},
		{
			name: "two marriages share the verb",
			fams: []fam{{"Mary Frost", marriage, 1764}, {"Susannah Furness", marriage, 1769}},
			want: "James married Mary Frost in 1764 and Susannah Furness in 1769.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "p", PreferredFamiliarName: "James", Gender: model.GenderMale}
			for i, fm := range tc.fams {
				spouse := &model.Person{ID: "s" + string(rune('a'+i)), PreferredFamiliarFullName: fm.spouse, Gender: model.GenderFemale}
				d := model.Year(fm.year)
				if fm.year == 0 {
					d = model.BeforeYear(1884)
				}
				ev := fm.event(d)
				f := &model.Family{
					ID: "f" + string(rune('a'+i)), Bond: model.FamilyBondMarried, Father: p, Mother: spouse,
					BestStartDate: ev.GetDate(), BestStartEvent: ev,
				}
				p.Families = append(p.Families, f)
			}

			enc := &md.Content{}
			got := PersonMarriageSummary(p, enc, &DefaultNameChooser{}, enc.EncodeText("James"), false, true).String()
			checkContains(t, got, tc.want, tc.notWant...)
		})
	}
}

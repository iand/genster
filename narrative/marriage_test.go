package narrative

import (
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

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
				ev := fm.event(model.Year(fm.year))
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

package gedcom

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/iand/genster/model"
)

func TestFillCensusEntry(t *testing.T) {
	testCases := []struct {
		v    string
		want *model.CensusEntry
	}{
		{
			v:    "",
			want: &model.CensusEntry{},
		},

		{
			v: "Relationship to Head: Daughter",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationDaughter,
			},
		},

		{
			v: "Relationship to Head: Servant",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationServant,
			},
		},

		{
			v: "Relationship to Head: Son",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationSon,
			},
		},

		{
			v: "Relation to Head: Mate",
			want: &model.CensusEntry{
				Detail: "Relation to Head: Mate",
			},
		},

		{
			v: "Relation to Head: Nephew  Staying with his aunt Mary A. Scrivener's family.",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationNephew,
				Detail:         "Staying with his aunt Mary A. Scrivener's family.",
			},
		},

		{
			v: "Relation to Head: Niece",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationNiece,
			},
		},

		{
			v: "Relation to Head: Servant",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationServant,
			},
		},

		{
			v: "Relation to Head: Servant at the Black Lion",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationServant,
				Detail:         "at the Black Lion",
			},
		},

		{
			v: "Relation to Head: Servant. Widow",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationServant,
				MaritalStatus:  model.CensusEntryMaritalStatusWidowed,
			},
		},
		{
			v: "Relation to Head of House: Boarder",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationBoarder,
			},
		},

		{
			v: "Relation to Head of House: Brother",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationBrother,
			},
		},

		{
			v: "Relation to Head of House: Daughter, Name recorded as Ellen",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationDaughter,
				Detail:         "Name recorded as Ellen",
			},
		},

		{
			v: "Relation to Head of House: Head",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationHead,
			},
		},

		{
			v: "Wellington Street. Relation to Head of House: Boarder",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationBoarder,
				Detail:         "Wellington Street.",
			},
		},

		{
			v: "Marital Status: Married; Relation to Head: Wife",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationWife,
				MaritalStatus:  model.CensusEntryMaritalStatusMarried,
			},
		},

		{
			v: "Marital Status: Married; Relationship to Head: Head",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationHead,
				MaritalStatus:  model.CensusEntryMaritalStatusMarried,
			},
		},

		{
			v: "Marital Status: Married; Relationship to Head: Wife",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationWife,
				MaritalStatus:  model.CensusEntryMaritalStatusMarried,
			},
		},

		{
			v: "Marital Status: Married; Relationship: Head",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationHead,
				MaritalStatus:  model.CensusEntryMaritalStatusMarried,
			},
		},

		{
			v: "Marital Status: Married; Relationship: Wife",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationWife,
				MaritalStatus:  model.CensusEntryMaritalStatusMarried,
			},
		},

		{
			v: "Marital Status: Single; Relation to Head: Daughter",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationDaughter,
				MaritalStatus:  model.CensusEntryMaritalStatusUnmarried,
			},
		},

		{
			v: "Marital Status: Single; Relation to Head: Son",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationSon,
				MaritalStatus:  model.CensusEntryMaritalStatusUnmarried,
			},
		},

		{
			v: "Relation to Head: Son; Residence Marital Status: Unmarried",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationSon,
				MaritalStatus:  model.CensusEntryMaritalStatusUnmarried,
				Detail:         "Residence",
			},
		},

		{
			v: "Relation to Head: Wife; Residence Marital Status: Married",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationWife,
				MaritalStatus:  model.CensusEntryMaritalStatusMarried,
				Detail:         "Residence",
			},
		},

		{
			v: "Relationship: Lod Daur",
			want: &model.CensusEntry{
				Detail: "Relationship: Lod Daur",
			},
		},

		{
			v: "Relationship: Lod Son",
			want: &model.CensusEntry{
				Detail: "Relationship: Lod Son",
			},
		},

		{
			v: "Relationship: Daughter",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationDaughter,
			},
		},

		{
			v: "Relationship: Daughter-in-law",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationDaughterInLaw,
			},
		},

		{
			v: "Relationship: Granddaughter",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationGranddaughter,
			},
		},

		{
			v: "Relationship: Niece",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationNiece,
			},
		},
		{
			v: "Relation to Head of House: Visitor",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationVisitor,
			},
		},
		{
			v: "Relationship: Soldier",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationSoldier,
			},
		},
		{
			v: "Relation to Head of House: Father-in-law",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationFatherInLaw,
			},
		},
		{
			v: "Relation to Head of House: Sister-in-law",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationSisterInLaw,
			},
		},
		{
			v: "Marital Status: Widowed; Relation to Head: Father-in-law",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationFatherInLaw,
				MaritalStatus:  model.CensusEntryMaritalStatusWidowed,
			},
		},
		{
			v: "Relation to Head: Lodger  Occupation: Railway porter",
			want: &model.CensusEntry{
				RelationToHead: model.CensusEntryRelationLodger,
				Detail:         "Occupation: Railway porter",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.v, func(t *testing.T) {
			ce := &model.CensusEntry{}

			fillCensusEntry(tc.v, ce)

			if diff := cmp.Diff(tc.want, ce); diff != "" {
				t.Errorf("fillCensusEntry(%q) mismatch (-want +got):\n%s", tc.v, diff)
			}
		})
	}
}

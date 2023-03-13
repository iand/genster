package model

import (
	"github.com/iand/gdate"
)

const UnknownNamePlaceholder = "–?–" // en dashes, sorts after other names

type Person struct {
	ID string // canonical identifier
	// Page                      string    // path to page in site
	Tags                      []string  // tags to add to the person's page
	PreferredFullName         string    // full legal name
	PreferredGivenName        string    // name that can be used in prose, usually just the first name
	PreferredFamiliarName     string    // name that can be used in prose, usually just the first name or a nickname
	PreferredFamiliarFullName string    // full name using just the first name or a nickname
	PreferredFamilyName       string    // family name, or surname
	PreferredSortName         string    // name organised for sorting, generally as surname, forenames
	PreferredUniqueName       string    // a name with additional uniquely identifying information such as years of birth and death or a numeric identifier
	NickName                  string    // a name other than their given name that the are known by
	Olb                       string    // One line bio
	Gender                    Gender    // male, female or unknown
	RelationToKeyPerson       *Relation // optional relation to the key person in the tree
	Father                    *Person
	Mother                    *Person
	Spouses                   []*Person     // list of people this person was in a relationship with
	Children                  []*Person     // list of people this person was considered a parent to
	Families                  []*Family     // list of families this person participated in as a parent
	VitalYears                string        // best guess at year of birth and death in yyyy-yyyy format
	BestBirthlikeEvent        TimelineEvent // event that best represents the person's birth
	BestDeathlikeEvent        TimelineEvent // event that best represents the person's death
	Timeline                  []TimelineEvent
	BeingTense                string // tense to use when refering to person: 'is' if they are possibly alive, 'was' if they are dead

	PossiblyAlive   bool         // true if this person is possibly still alive
	Unknown         bool         // true if this person is known to have existed but no other information is known
	Unmarried       bool         // true if it is known that the person did not marry
	Childless       bool         // true if it is known that the person did not have any children
	Illegitimate    bool         // true if it is known that the person was born illlegitimately
	BornInWorkhouse bool         // true if the birth place of the person was a workhouse
	DiedInWorkhouse bool         // true if the death place of the person was a workhouse
	Pauper          bool         // true if the person was, at some stage, noted as a pauper
	CauseOfDeath    CauseOfDeath // cause of death, if known

	Occupations        []*Occupation // list of occupations
	PrimaryOccupation  string        // simple description of main occupation
	EditLink           *Link         // link to a page that can be used to edit the details of this person
	Links              []Link        // list of links to more information relevant to this person
	Redacted           bool          // true if the person's details should be redacted
	RedactionKeepsName bool          // true if this person's name should be kept during redaction
	Inferences         []Inference   // list of inferences made
	Anomalies          []*Anomaly    // list of anomalies detected
	MiscFacts          []Fact        // miscellaneous facts
	FeatureImage       *FeatureImage // an image that can be used to represent the person
}

func (p *Person) IsUnknown() bool {
	if p == nil {
		return true
	}
	return p.Unknown
}

func (p *Person) SameAs(other *Person) bool {
	if p == nil || other == nil {
		return false
	}
	return p == other || (p.ID != "" && p.ID == other.ID)
}

func (p *Person) AgeInYearsAt(dt *Date) (int, bool) {
	if p.BestBirthlikeEvent == nil || p.BestBirthlikeEvent.GetDate().IsUnknown() || dt.IsUnknown() {
		return 0, false
	}

	return p.BestBirthlikeEvent.GetDate().WholeYearsUntil(dt)
}

func (p *Person) PreciseAgeAt(dt *Date) (*gdate.PreciseInterval, bool) {
	if p.BestBirthlikeEvent == nil || p.BestBirthlikeEvent.GetDate().IsUnknown() || dt.IsUnknown() {
		return nil, false
	}

	in := p.BestBirthlikeEvent.GetDate().IntervalUntil(dt)
	if in.IsUnknown() {
		return nil, false
	}

	if pi, ok := gdate.AsPreciseInterval(in.Interval); ok {
		return pi, true
	}

	return nil, false
}

// RelationTo returns a textual description of the relation of p to other.
// Returns an empty string if no relation was determined
func (p *Person) RelationTo(other *Person, dt *Date) string {
	if p.SameAs(other) {
		return "Self"
	}

	for _, ch := range p.Children {
		if ch.SameAs(other) {
			// other person is a child of this one
			return p.Gender.RelationToChildrenNoun()
		}
	}
	for _, ch := range other.Children {
		if ch.SameAs(p) {
			// this person is a child of the other one
			return p.Gender.RelationToParentNoun()
		}
	}

	for _, f := range p.Families {
		if f.Father.SameAs(other) {
			// The other person is the husband, so this person is the wife of other
			if f.BestEndEvent != nil && f.BestEndEvent.GetDate().SortsBefore(dt) {
				return "Former wife"
			}
			if f.BestStartEvent != nil && f.BestStartEvent.GetDate().SortsBefore(dt) {
				return "Wife"
			}
			if f.BestStartEvent == nil {
				return "Wife"
			}
		} else if f.Mother.SameAs(other) {
			// The other person is the wife, so this person is the husband of other
			if f.BestEndEvent != nil && f.BestEndEvent.GetDate().SortsBefore(dt) {
				return "Former husband"
			}

			if f.BestStartEvent != nil && f.BestStartEvent.GetDate().SortsBefore(dt) {
				return "Husband"
			}

			if f.BestStartEvent == nil {
				return "Husband"
			}
		}
	}

	return ""
}

func (p *Person) BestBirthDate() *Date {
	if p.BestBirthlikeEvent == nil {
		return UnknownDate()
	}

	return p.BestBirthlikeEvent.GetDate()
}

func (p *Person) BestDeathDate() *Date {
	if p.BestDeathlikeEvent == nil {
		return UnknownDate()
	}

	return p.BestDeathlikeEvent.GetDate()
}

func (p *Person) RemoveDuplicateFamilies() {
	unique := make(map[string]*Family)
	for _, f := range p.Families {
		unique[f.ID] = f
	}
	if len(unique) == len(p.Families) {
		return
	}
	p.Families = p.Families[:0]
	for _, f := range unique {
		p.Families = append(p.Families, f)
	}
}

func (p *Person) RemoveDuplicateChildren() {
	unique := make(map[string]*Person)
	for _, c := range p.Children {
		unique[c.ID] = c
	}
	if len(unique) == len(p.Children) {
		return
	}
	p.Children = p.Children[:0]
	for _, c := range unique {
		p.Children = append(p.Children, c)
	}
}

func (p *Person) RemoveDuplicateSpouses() {
	unique := make(map[string]*Person)
	for _, c := range p.Spouses {
		unique[c.ID] = c
	}
	if len(unique) == len(p.Spouses) {
		return
	}
	p.Spouses = p.Spouses[:0]
	for _, c := range unique {
		p.Spouses = append(p.Spouses, c)
	}
}

func UnknownPerson() *Person {
	return &Person{
		PreferredFullName:         "an unknown person",
		PreferredGivenName:        "unknown",
		PreferredFamiliarName:     "unknown",
		PreferredFamiliarFullName: "unknown",
		PreferredFamilyName:       "unknown",
		PreferredSortName:         "unknown person",
		PreferredUniqueName:       "an unknown person",
		Unknown:                   true,
	}
}

type PersonActionFunc func(*Person)

func RecurseDescendants(p *Person, fn PersonActionFunc) {
	for _, c := range p.Children {
		fn(c)
		RecurseDescendants(c, fn)
	}
}

func YearsSinceDeath(p *Person) (int, bool) {
	if p.BestDeathlikeEvent == nil || p.BestDeathlikeEvent.GetDate().IsUnknown() {
		return 0, false
	}

	in := IntervalSince(p.BestDeathlikeEvent.GetDate())
	return in.WholeYears()
}

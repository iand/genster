package model

const (
	FamilyBondMarried         = "married"
	FamilyBondUnmarried       = "unmarried"
	FamilyBondLikelyMarried   = "likely married"
	FamilyBondLikelyUnmarried = "likely unmarried"
	FamilyBondUnknown         = "unknown"

	FamilyEndReasonUnknown  = "unknown"
	FamilyEndReasonDivorce  = "divorce"
	FamilyEndReasonDeath    = "death"
	FamilyEndReasonAnulment = "anulment"
)

type Family struct {
	ID string // canonical id
	// Page                string   // path to page in site
	Tags []string // tags to add to the person's page

	// TODO: populate
	PreferredUniqueName string // name used to identify the family
	Father              *Person
	Mother              *Person
	Children            []*Person
	BestStartEvent      TimelineEvent // event that best represents the start of the family unit if the bond is a marriage type
	BestEndEvent        TimelineEvent // event that best represents the end of the family unit if the bond is a marriage type

	BestStartDate *Date // date that best represents the start of the family unit
	BestEndDate   *Date // date that best represents the end of the family unit

	Timeline       []TimelineEvent
	Bond           string  // the kind of bond between the parents in the family
	EndReason      string  // the reason the family unit ended
	EndDeathPerson *Person // the person whose death ended the family unit, if any

	PublishChildren bool // true if this family's children should always be included in the publish set
}

func (f *Family) OtherParent(p *Person) *Person {
	if p.SameAs(f.Father) {
		return f.Mother
	}
	if p.SameAs(f.Mother) {
		return f.Father
	}
	return UnknownPerson()
}

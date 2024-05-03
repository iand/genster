package model

type EventDateType string

const (
	EventDateTypeOccurred EventDateType = "occurred" // the date represents the date the event occurred
	EventDateTypeRecorded EventDateType = "recorded" // the date represents the date the event was recorded
)

type CensusEntryRelation string

// These can all be followed by " of the head of the household." (except head and any impersonal ones)
const (
	CensusEntryRelationUnknown       CensusEntryRelation = ""
	CensusEntryRelationHead          CensusEntryRelation = "head"
	CensusEntryRelationWife          CensusEntryRelation = "wife"
	CensusEntryRelationHusband       CensusEntryRelation = "husband"
	CensusEntryRelationSon           CensusEntryRelation = "son"
	CensusEntryRelationDaughter      CensusEntryRelation = "daughter"
	CensusEntryRelationChild         CensusEntryRelation = "child"
	CensusEntryRelationFather        CensusEntryRelation = "father"
	CensusEntryRelationMother        CensusEntryRelation = "mother"
	CensusEntryRelationUncle         CensusEntryRelation = "uncle"
	CensusEntryRelationAunt          CensusEntryRelation = "aunt"
	CensusEntryRelationLodger        CensusEntryRelation = "lodger"
	CensusEntryRelationBoarder       CensusEntryRelation = "boarder"
	CensusEntryRelationInmate        CensusEntryRelation = "inmate"
	CensusEntryRelationPatient       CensusEntryRelation = "patient"
	CensusEntryRelationServant       CensusEntryRelation = "servant"
	CensusEntryRelationNephew        CensusEntryRelation = "nephew"
	CensusEntryRelationNiece         CensusEntryRelation = "niece"
	CensusEntryRelationBrother       CensusEntryRelation = "brother"
	CensusEntryRelationSister        CensusEntryRelation = "sister"
	CensusEntryRelationSonInLaw      CensusEntryRelation = "son-in-law"
	CensusEntryRelationDaughterInLaw CensusEntryRelation = "daughter-in-law"
	CensusEntryRelationFatherInLaw   CensusEntryRelation = "father-in-law"
	CensusEntryRelationMotherInLaw   CensusEntryRelation = "mother-in-law"
	CensusEntryRelationBrotherInLaw  CensusEntryRelation = "brother-in-law"
	CensusEntryRelationSisterInLaw   CensusEntryRelation = "sister-in-law"
	CensusEntryRelationGrandson      CensusEntryRelation = "grandson"
	CensusEntryRelationGranddaughter CensusEntryRelation = "granddaughter"
	CensusEntryRelationVisitor       CensusEntryRelation = "visitor"
	CensusEntryRelationSoldier       CensusEntryRelation = "soldier"
	CensusEntryRelationFosterChild   CensusEntryRelation = "foster child"
)

// IsImpersonal reports whether the relation is to the place rather than the head
func (r CensusEntryRelation) IsImpersonal() bool {
	switch r {
	case CensusEntryRelationLodger,
		CensusEntryRelationBoarder,
		CensusEntryRelationInmate,
		CensusEntryRelationPatient,
		CensusEntryRelationServant,
		CensusEntryRelationSoldier,
		CensusEntryRelationVisitor:
		return true
	default:
		return false
	}
}

func (r CensusEntryRelation) String() string {
	return string(r)
}

type CensusEntryMaritalStatus string

const (
	CensusEntryMaritalStatusUnknown   CensusEntryMaritalStatus = ""
	CensusEntryMaritalStatusMarried   CensusEntryMaritalStatus = "married"
	CensusEntryMaritalStatusUnmarried CensusEntryMaritalStatus = "unmarried"
	CensusEntryMaritalStatusWidowed   CensusEntryMaritalStatus = "widowed"
	CensusEntryMaritalStatusDivorced  CensusEntryMaritalStatus = "divorced"
)

func (c CensusEntryMaritalStatus) String() string {
	return string(c)
}

type ModeOfDeath string

const (
	ModeOfDeathNatural        ModeOfDeath = ""
	ModeOfDeathSuicide        ModeOfDeath = "suicide"
	ModeOfDeathLostAtSea      ModeOfDeath = "lost at sea"
	ModeOfDeathKilledInAction ModeOfDeath = "killed in action"
	ModeOfDeathDrowned        ModeOfDeath = "drowned"
	ModeOfDeathExecuted       ModeOfDeath = "executed"
)

func (m ModeOfDeath) What() string {
	switch m {

	case ModeOfDeathNatural:
		return "died"
	case ModeOfDeathSuicide:
		return "died by own hand"
	case ModeOfDeathLostAtSea:
		return ""
	case ModeOfDeathKilledInAction:
		return ""
	case ModeOfDeathDrowned:
		return ""
	case ModeOfDeathExecuted:
		return ""

	default:
		return string(m)
	}
}

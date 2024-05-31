package model

import "github.com/iand/genster/text"

type OccupationStatus string

const (
	OccupationStatusUnknown    OccupationStatus = ""
	OccupationStatusRetired    OccupationStatus = "retired"
	OccupationStatusFormer     OccupationStatus = "former"
	OccupationStatusUnemployed OccupationStatus = "unemployed"
	OccupationStatusApprentice OccupationStatus = "apprentice"
	OccupationStatusJourneyman OccupationStatus = "journeyman"
	OccupationStatusMaster     OccupationStatus = "master"
)

func (o OccupationStatus) String() string {
	return string(o)
}

// OccupationGroup is the general class of the occupation
type OccupationGroup string

const (
	OccupationGroupUnknown    OccupationGroup = ""
	OccupationGroupLabouring  OccupationGroup = "labouring" // ag lab
	OccupationGroupIndustrial OccupationGroup = "industrial"
	OccupationGroupMaritime   OccupationGroup = "maritime"   // seaman, mariner
	OccupationGroupCrafts     OccupationGroup = "crafts"     // carpenter, shoemaker etc
	OccupationGroupClerical   OccupationGroup = "clerical"   // clerk
	OccupationGroupCommercial OccupationGroup = "commercial" // baker, victualer, grocer, publican, dealer
	OccupationGroupMilitary   OccupationGroup = "military"
	OccupationGroupPolice     OccupationGroup = "police"  // policeman, prison warder
	OccupationGroupService    OccupationGroup = "service" // nurse, servant, valet, groom
)

type Occupation struct {
	Date        *Date
	StartDate   *Date
	EndDate     *Date
	Place       *Place
	Name        string // the name of the occupation, to be used in a sentence as `a {name}`
	Comment     string // an explanatatory comment to be used alongside or as a footnote to the title
	Detail      string
	Status      OccupationStatus
	Group       OccupationGroup
	Citations   []*GeneralCitation
	Occurrences int
	Unknown     bool
}

func (o *Occupation) IsUnknown() bool {
	if o == nil {
		return true
	}
	return o.Unknown
}

func (o *Occupation) String() string {
	if o.Name == "" {
		return ""
	}
	str := ""
	if o.Status != OccupationStatusUnknown {
		str = o.Status.String() + " "
	}
	str += o.Name
	return "a" + text.MaybeAn(str)
}

func UnknownOccupation() *Occupation {
	return &Occupation{
		Date:    UnknownDate(),
		Name:    "worker at an unknown trade",
		Status:  OccupationStatusUnknown,
		Unknown: true,
	}
}

// plate layer -  Laid and maintained railway tracks. The name predates railways, and is derived from the original 'plateways' which existed hundreds of years ago, laid for coal mines (https://www.genesreunited.co.nz/boards/board/ancestors/thread/788539)

// func ParseOccupationText(text string, citations []*GeneralCitation) *Fact {

// }

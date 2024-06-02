package model

import (
	"github.com/iand/genster/place"
)

type Place struct {
	ID                  string   // canonical identifier
	Tags                []string // tags to add to the place's page
	OriginalText        string   // the original text that was used to fill in the place information
	Hints               []place.Hint
	PreferredName       string       // fully parsed name but with just the minimum amount of context, such as "locality, region"
	PreferredUniqueName string       // fully parsed name but with just enough extra context to make it unique
	PreferredFullName   string       // the fully parsed name
	PreferredSortName   string       // name organised for sorting, generally as a reverse hierarchy of country, region, locality
	Parent              *Place       // the parent of this place in the administrative hierarchy
	PlaceType           PlaceType    // the type of place, such as "village", "town", "parish"
	Numbered            bool         // whether the place is a numbered building
	Singular            bool         // whether the place is a singular member of a group such as the register office or the barracks, not a named church.
	BuildingKind        BuildingKind // the kind of building, such as "church", "workhouse" or "register office"
	Timeline            []TimelineEvent
	Unknown             bool    // true if this place is known to have existed but no other information is known
	Links               []Link  // list of links to more information relevant to this place
	Latitude            float64 // latitude of the place in decimal degrees, +ve is east of meridian, -ve is west
	Longitude           float64 // longitude of the place in decimal degrees, +ve is north of equator, -ve is south

	CountryName  *place.PlaceName
	UKNationName *place.PlaceName

	Kind place.PlaceKind // the kind of place - DEPRECATED

	ResearchNotes []Text // research notes associated with this place
	Comments      []Text // comments associated with this place
}

func (p *Place) IsUnknown() bool {
	if p == nil {
		return true
	}
	return p.Unknown
}

func (p *Place) SameAs(other *Place) bool {
	if p == nil || other == nil {
		return false
	}
	return p == other || (p.ID != "" && p.ID == other.ID)
}

func (p *Place) Where() string {
	if p == nil {
		return "an unknown place"
	}

	return p.InAt() + " " + p.PreferredFullName
}

func (p *Place) InAt() string {
	switch p.PlaceType {
	case PlaceTypeAddress, PlaceTypeBuilding:
		return "at"
	case PlaceTypeStreet:
		return "at"
	case PlaceTypeShip:
		return "aboard"
	default:
		return "in"
	}
}

func UnknownPlace() *Place {
	return &Place{
		PreferredName:       "unknown",
		PreferredFullName:   "an unknown place",
		PreferredUniqueName: "an unknown place",
		PreferredSortName:   "unknown place",
		Unknown:             true,
		PlaceType:           PlaceTypeUnknown,
		BuildingKind:        BuildingKindNone,
	}
}

type PlaceType string

const (
	PlaceTypeUnknown      = "place"
	PlaceTypeAddress      = "address"
	PlaceTypeCountry      = "country"
	PlaceTypeBuilding     = "building"
	PlaceTypeBurialGround = "burial ground"
	PlaceTypeStreet       = "street"
	PlaceTypeShip         = "ship"
	PlaceTypeCategory     = "category" // used ony for grouping related places
	PlaceTypeCity         = "city"
	PlaceTypeTown         = "town"
	PlaceTypeVillage      = "village"
	PlaceTypeHamlet       = "hamlet"
	PlaceTypeParish       = "parish"
	PlaceTypeCounty       = "county"
)

func (p PlaceType) String() string {
	return string(p)
}

type BuildingKind string

const (
	BuildingKindNone           = ""
	BuildingKindChurch         = "church"
	BuildingKindWorkhouse      = "workhouse"
	BuildingKindRegisterOffice = "register office"
	BuildingKindFarm           = "farm"
	BuildingKindHospital       = "hospital"
)

func (p BuildingKind) String() string {
	return string(p)
}

type PlaceMatcher func(*Place) bool

func PlaceHasTag(tag string) PlaceMatcher {
	return func(p *Place) bool {
		for _, t := range p.Tags {
			if t == tag {
				return true
			}
		}
		return false
	}
}

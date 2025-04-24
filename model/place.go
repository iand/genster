package model

import (
	"strings"
	"time"

	"github.com/iand/genster/place"
)

const UnknownPlaceName = "unknown"

type Place struct {
	ID           string   // canonical identifier
	Tags         []string // tags to add to the place's page
	OriginalText string   // the original text that was used to fill in the place information
	Hints        []place.Hint
	// TODO: consolidate these name fields
	// Need names that are:
	//  - just the name of the place
	//  - the name of the place with all levels before parish/town
	//  - the name of the place with all levels up to parish/town
	//  - the name of the place with all levels up to county/state
	//  - the full name including all levels up to country
	//  - the sort name including all levels up to country, reverse order
	//  - parish/town level up to county/state
	//  - parish/town level including all levels up to country
	// Area naming:
	//  - region: county/state
	//  - district: town/parish

	Name             string // the bare name of the place, could be street name or country name
	FullName         string // the name including the hierarchy of places it belongs to, comma separated
	ProseName        string // the name including the hierarchy of places it belongs to, comma separated with parts qualified by their type such as "St. Martins Church in the parish of Fressingfield, in Suffolk, England"
	NameWithDistrict string // the name of the place with context such as town or parish
	NameWithRegion   string // the name of the place with locality and region, such as county
	NameWithCountry  string // the name of the place with locality, region and country

	FullContext     string // the full hierarchy of places this place belongs to, comma separated, no leading comma
	DistrictContext string // an edited hierarchy of places this place belongs to up to the district, comma separated, no leading comma
	RegionContext   string // an edited hierarchy of places this place belongs to up to the region, comma separated, no leading comma
	CountryContext  string // an edited hierarchy of places this place belongs to up to the country, comma separated, no leading comma

	PreferredSortName string // name organised for sorting, generally as a reverse hierarchy of country, region, locality

	Parent       *Place       // the parent of this place in the administrative hierarchy
	PlaceType    PlaceType    // the type of place, such as "village", "town", "parish"
	Numbered     bool         // whether the place is a numbered building
	Singular     bool         // whether the place is a singular member of a group such as the register office or the barracks, not a named church.
	BuildingKind BuildingKind // the kind of building, such as "church", "workhouse" or "register office"
	Timeline     []TimelineEvent
	Unknown      bool         // true if this place is known to have existed but no other information is known
	Links        []Link       // list of links to more information relevant to this place
	GeoLocation  *GeoLocation // geographic location of the place

	Country  *Place // Country is the place lowest in the parent hierarchy that has the type of country
	Region   *Place // Region is the regional place lowest in the parent hierarchy below the country, such as state, county, nation, island, archipelago etc.
	District *Place // District is the populated place lowest in the parent hierarchy below the region level, such as a parish, town, village etc.

	CountryName  *place.PlaceName
	UKNationName *place.PlaceName

	ResearchNotes []Text              // research notes associated with this place
	Comments      []Text              // comments associated with this place
	Gallery       []*CitedMediaObject // images and documents associated with the place

	UpdateTime *time.Time // time of last update, if known
	CreateTime *time.Time // time of creation
}

type GeoLocation struct {
	Latitude  float64 // latitude of the centre in decimal degrees, +ve is east of meridian, -ve is west
	Longitude float64 // longitude of the centre in decimal degrees, +ve is north of equator, -ve is south
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

func (p *Place) SameCountry(other *Place) bool {
	if p == nil || other == nil {
		return false
	}
	return p.Country.SameAs(other.Country)
}

func (p *Place) SameRegion(other *Place) bool {
	if p == nil || other == nil {
		return false
	}
	return p.Region.SameAs(other.Region)
}

func (p *Place) Where() string {
	if p == nil {
		return "an unknown place"
	}

	return p.InAt() + " " + p.FullName
}

func (p *Place) InAt() string {
	if p == nil {
		return ""
	}
	switch p.PlaceType {
	case PlaceTypeBuilding:
		switch p.BuildingKind {
		case BuildingKindChurch:
			if strings.HasPrefix(strings.ToLower(p.Name), "church") {
				return "at the"
			}
		}
		return "at"

	case PlaceTypeAddress:
		return "at"
	case PlaceTypeStreet:
		return "at"
	case PlaceTypeShip:
		return "aboard the"
	// case PlaceTypeParish:
	// 	return "in the parish of"
	// case PlaceTypeRegistrationDistrict:
	// 	return "in the registration district of"
	default:
		return "in"
	}
}

func (p *Place) DescriptivePrefix() string {
	if p == nil {
		return ""
	}
	switch p.PlaceType {
	case PlaceTypeParish:
		return "the parish of"
	case PlaceTypeRegistrationDistrict:
		return "the registration district of"
	default:
		return ""
	}
}

func (p *Place) Created() (time.Time, bool) {
	if p.CreateTime == nil {
		return time.Time{}, false
	}
	return *p.CreateTime, true
}

func (p *Place) Updated() (time.Time, bool) {
	if p.UpdateTime == nil {
		return time.Time{}, false
	}
	return *p.UpdateTime, true
}

// Hierarchy returns the list of places that form the full
// hierachy for this place. The first entry is p, the next
// entry is p's parent, the next entry is p's grandparent
// and so on. The list walways contains at least one element,
// which will be op itself.
func (p *Place) Hierarchy() []*Place {
	hierarchy := []*Place{p}
	par := p.Parent
	for par != nil {
		hierarchy = append(hierarchy, par)
		par = par.Parent
	}
	return hierarchy
}

func UnknownPlace() *Place {
	return &Place{
		Name:              UnknownPlaceName,
		FullName:          "an unknown place",
		PreferredSortName: "unknown place",
		Unknown:           true,
		PlaceType:         PlaceTypeUnknown,
		BuildingKind:      BuildingKindNone,
	}
}

type PlaceType string

const (
	PlaceTypeUnknown = "place"
	PlaceTypeAddress = "address"
	PlaceTypeCountry = "country"
	PlaceTypeNation  = "nation" // nation is a place with national identity within a country
	// PlaceTypeTerritory            = "territory" // maybe territory is a place associated with a country but is not goverened by it (early New Mexico/Texas?)
	PlaceTypeArchipelago          = "archipelago" // a group of islands that is part of a country
	PlaceTypeBuilding             = "building"
	PlaceTypeBurialGround         = "burial ground"
	PlaceTypeStreet               = "street"
	PlaceTypeShip                 = "ship"
	PlaceTypeCategory             = "category" // used ony for grouping related places
	PlaceTypeCity                 = "city"
	PlaceTypeTown                 = "town"
	PlaceTypeVillage              = "village"
	PlaceTypeHamlet               = "hamlet"
	PlaceTypeParish               = "parish"
	PlaceTypeCounty               = "county"
	PlaceTypeState                = "county"
	PlaceTypeRegistrationDistrict = "registration district"
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

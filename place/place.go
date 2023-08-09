package place

// // TODO: Give places a Familiar Name, Sort Name, Full Name, Unique Name, just like Person
// // Remember when we mention it more than once and switch from Full Name to FamiliarName
// // Also remember the original unparsed text

// type Place interface {
// 	// String returns the most detailed description of the place available
// 	String() string

// 	Country() string

// 	// Locality returns just the locality of the place such as the town or village
// 	Locality() string

// 	// FullLocality returns the locality of the place within it's national hierarchy such as town, county, country
// 	FullLocality() string
// }

type PlaceKind string

const (
	PlaceKindUnknown  PlaceKind = "unknown"
	PlaceKindCountry  PlaceKind = "country"
	PlaceKindUKNation PlaceKind = "uknation"
	PlaceKindAddress  PlaceKind = "address"
)

// See https://www.visionofbritain.org.uk/types/level/11 for parish etc

type PlaceNameHierarchy struct {
	Name                    PlaceName
	Kind                    PlaceKind
	NormalizedWithHierarchy string
	Parent                  *PlaceNameHierarchy
	Child                   *PlaceNameHierarchy
}

func (p *PlaceNameHierarchy) String() string {
	if p.Parent == nil {
		return p.Name.Name
	}
	return p.Name.Name + "," + p.Parent.String()
}

func (p *PlaceNameHierarchy) HasParent() bool {
	return p.Parent != nil
}

type PlaceName struct {
	ID         string
	Kind       PlaceKind
	Name       string
	Adjective  string
	Aliases    []string
	Unknown    bool
	ChildHints []Hint
	PartOf     []*PlaceName
}

func (c *PlaceName) IsUnknown() bool {
	if c == nil {
		return true
	}
	return c.Unknown
}

func (c *PlaceName) SameAs(other *PlaceName) bool {
	if c == nil || other == nil {
		return false
	}
	if c.Unknown || other.Unknown {
		return false
	}
	return c == other || (c.Name != "" && c.Name == other.Name)
}

func (pn *PlaceName) FindContainerKind(kind PlaceKind) (*PlaceName, bool) {
	if pn.Kind == kind {
		return pn, true
	}

	for _, po := range pn.PartOf {
		if po.Kind == kind {
			return po, true
		}

		c, ok := po.FindContainerKind(kind)
		if ok {
			return c, true
		}
	}

	return nil, false
}

func UnknownPlaceName() *PlaceName {
	return &PlaceName{
		Name:      "unknown",
		Adjective: "unknown",
		Unknown:   true,
	}
}

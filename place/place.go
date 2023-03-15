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
	PlaceKindCountry  PlaceKind = "country"
	PlaceKindUKNation PlaceKind = "uknation"
	PlaceKindAddress  PlaceKind = "address"
)

// See https://www.visionofbritain.org.uk/types/level/11 for parish etc

type PlaceHierarchy struct {
	Name                    PlaceName
	Kind                    PlaceKind
	NormalizedWithHierarchy string
	Parent                  *PlaceHierarchy
	Child                   *PlaceHierarchy
}

func (p *PlaceHierarchy) String() string {
	if p.Parent == nil {
		return p.Name.Name
	}
	return p.Name.Name + "," + p.Parent.String()
}

func (p *PlaceHierarchy) HasParent() bool {
	return p.Parent != nil
}

type PlaceName struct {
	ID         string
	Name       string
	Adjective  string
	Aliases    []string
	Unknown    bool
	ChildHints []Hint
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
	return c == other || (c.Name != "" && c.Name == other.Name)
}

func UnknownPlaceName() PlaceName {
	return PlaceName{
		Name:      "unknown",
		Adjective: "unknown",
		Unknown:   true,
	}
}

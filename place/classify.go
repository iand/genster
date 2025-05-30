package place

import (
	"strings"
	"sync"
	"unicode"
)

type Hint interface {
	ID() string // a unique identifier for the hint
}

type PlaceKind string

const (
	PlaceKindUnknown  PlaceKind = "unknown"
	PlaceKindCountry  PlaceKind = "country"
	PlaceKindUKNation PlaceKind = "uknation"
	PlaceKindAddress  PlaceKind = "address"
)

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

func ClassifyName(name string, hints ...Hint) *PlaceName {
	pn, ok := LookupPlaceName(name)
	if ok {
		return pn
	}

	parts := splitPlaceName(name)
	if len(parts) == 0 {
		return UnknownPlaceName()
	}

	pn = &PlaceName{
		Name: Clean(name),
		Kind: PlaceKindUnknown,
	}

	for i := len(parts) - 1; i >= 0; i-- {
		po, ok := LookupPlaceName(parts[i])
		if ok {
			pn.PartOf = append(pn.PartOf, po)
		}
	}

	return pn
}

var (
	placeNameLookup     = map[string]*PlaceName{}
	placeNameLookupOnce sync.Once
)

func LookupPlaceName(v string) (*PlaceName, bool) {
	placeNameLookupOnce.Do(func() {
		for _, c := range ukNationNames {
			indexPlaceName(c)
		}
		indexPlaceName(countryUK)

		for _, c := range countryNames {
			indexPlaceName(c)
		}
	})
	pn, ok := placeNameLookup[strings.ToLower(v)]
	return pn, ok
}

func indexPlaceName(pn *PlaceName) {
	placeNameLookup[strings.ToLower(pn.Name)] = pn
	for _, al := range pn.Aliases {
		placeNameLookup[strings.ToLower(al)] = pn
	}
}

var countryUK = &PlaceName{Kind: PlaceKindCountry, Name: "United Kingdom", Adjective: "British", Aliases: []string{"UK", "Britain"}, ChildHints: []Hint{}}

// TODO: not all country names, or historic country names are here
var countryNames = []*PlaceName{
	{Kind: PlaceKindCountry, Name: "Abkhazia", Adjective: "Abkhaz"},
	{Kind: PlaceKindCountry, Name: "Afghanistan", Adjective: "Afghan"},
	{Kind: PlaceKindCountry, Name: "Albania", Adjective: "Albanian"},
	{Kind: PlaceKindCountry, Name: "Algeria", Adjective: "Algerian"},
	{Kind: PlaceKindCountry, Name: "American Samoa", Adjective: "American Samoan"},
	{Kind: PlaceKindCountry, Name: "Andorra", Adjective: "Andorran"},
	{Kind: PlaceKindCountry, Name: "Angola", Adjective: "Angolan"},
	{Kind: PlaceKindCountry, Name: "Anguilla", Adjective: "Anguillan"},
	{Kind: PlaceKindCountry, Name: "Antigua and Barbuda", Adjective: "Antiguan"},
	{Kind: PlaceKindCountry, Name: "Argentina", Adjective: "Argentine"},
	{Kind: PlaceKindCountry, Name: "Armenia", Adjective: "Armenian"},
	{Kind: PlaceKindCountry, Name: "Aruba", Adjective: "Aruban"},
	{Kind: PlaceKindCountry, Name: "Australia", Adjective: "Australian"},
	{Kind: PlaceKindCountry, Name: "Austria", Adjective: "Austrian"},
	{Kind: PlaceKindCountry, Name: "Azerbaijan", Adjective: "Azerbaijani"},
	{Kind: PlaceKindCountry, Name: "Bahrain", Adjective: "Bahraini"},
	{Kind: PlaceKindCountry, Name: "Bangladesh", Adjective: "Bangladeshi"},
	{Kind: PlaceKindCountry, Name: "Barbados", Adjective: "Barbadian"},
	{Kind: PlaceKindCountry, Name: "Belarus", Adjective: "Belarusian"},
	{Kind: PlaceKindCountry, Name: "Belgium", Adjective: "Belgian"},
	{Kind: PlaceKindCountry, Name: "Belize", Adjective: "Belizean"},
	{Kind: PlaceKindCountry, Name: "Benin", Adjective: "Beninese"},
	{Kind: PlaceKindCountry, Name: "Bermuda", Adjective: "Bermudian"},
	{Kind: PlaceKindCountry, Name: "Bhutan", Adjective: "Bhutanese"},
	{Kind: PlaceKindCountry, Name: "Bolivia", Adjective: "Bolivian"},
	{Kind: PlaceKindCountry, Name: "Cambodia", Adjective: "Cambodian"},
	{Kind: PlaceKindCountry, Name: "Canada", Adjective: "Canadian"},
	{Kind: PlaceKindCountry, Name: "Chile", Adjective: "Chilean"},
	{Kind: PlaceKindCountry, Name: "China", Adjective: "Chinese"},
	{Kind: PlaceKindCountry, Name: "Colombia", Adjective: "Colombian"},
	{Kind: PlaceKindCountry, Name: "Comoros", Adjective: "Comoran"},
	{Kind: PlaceKindCountry, Name: "Cook Islands", Adjective: "Cook Island"},
	{Kind: PlaceKindCountry, Name: "Costa Rica", Adjective: "Costa Rican"},
	{Kind: PlaceKindCountry, Name: "Croatia", Adjective: "Croatian"},
	{Kind: PlaceKindCountry, Name: "Cuba", Adjective: "Cuban"},
	{Kind: PlaceKindCountry, Name: "Cyprus", Adjective: "Cypriot"},
	{Kind: PlaceKindCountry, Name: "Czech Republic", Adjective: "Czech"},
	{Kind: PlaceKindCountry, Name: "Denmark", Adjective: "Danish"},
	{Kind: PlaceKindCountry, Name: "Djibouti", Adjective: "Djiboutian"},
	{Kind: PlaceKindCountry, Name: "Dominica", Adjective: "Dominican"},
	{Kind: PlaceKindCountry, Name: "Dominican Republic", Adjective: "Dominican"},
	{Kind: PlaceKindCountry, Name: "Ecuador", Adjective: "Ecuadorian"},
	{Kind: PlaceKindCountry, Name: "Egypt", Adjective: "Egyptian"},
	{Kind: PlaceKindCountry, Name: "El Salvador", Adjective: "Salvadoran"},
	{Kind: PlaceKindCountry, Name: "Equatorial Guinea", Adjective: "Equatoguinean"},
	{Kind: PlaceKindCountry, Name: "Eritrea", Adjective: "Eritrean"},
	{Kind: PlaceKindCountry, Name: "Estonia", Adjective: "Estonian"},
	{Kind: PlaceKindCountry, Name: "Ethiopia", Adjective: "Ethiopian"},
	{Kind: PlaceKindCountry, Name: "Faroe Islands", Adjective: "Faroese"},
	{Kind: PlaceKindCountry, Name: "Fiji", Adjective: "Fijian"},
	{Kind: PlaceKindCountry, Name: "Finland", Adjective: "Finnish"},
	{Kind: PlaceKindCountry, Name: "France", Adjective: "French"},
	{Kind: PlaceKindCountry, Name: "French Guiana", Adjective: "French Guianese"},
	{Kind: PlaceKindCountry, Name: "French Polynesia", Adjective: "French Polynesian"},
	{Kind: PlaceKindCountry, Name: "Gabon", Adjective: "Gabonese"},
	{Kind: PlaceKindCountry, Name: "Georgia", Adjective: "Georgian"},
	{Kind: PlaceKindCountry, Name: "Germany", Adjective: "German"},
	{Kind: PlaceKindCountry, Name: "Ghana", Adjective: "Ghanaian"},
	{Kind: PlaceKindCountry, Name: "Gibraltar", Adjective: "Gibraltan"},
	{Kind: PlaceKindCountry, Name: "Greece", Adjective: "Greek"},
	{Kind: PlaceKindCountry, Name: "Guatemala", Adjective: "Guatemalan"},
	{Kind: PlaceKindCountry, Name: "Guernsey", Adjective: "Guernsey"},
	{Kind: PlaceKindCountry, Name: "Guinea-Bissau", Adjective: "Bissau-Guinean"},
	{Kind: PlaceKindCountry, Name: "Guinea", Adjective: "Guinean"},
	{Kind: PlaceKindCountry, Name: "Guyana", Adjective: "Guyanese"},
	{Kind: PlaceKindCountry, Name: "Haiti", Adjective: "Haitian"},
	{Kind: PlaceKindCountry, Name: "Honduras", Adjective: "Honduran"},
	{Kind: PlaceKindCountry, Name: "Hong Kong", Adjective: "Hong Kong"},
	{Kind: PlaceKindCountry, Name: "Hungary", Adjective: "Hungarian"},
	{Kind: PlaceKindCountry, Name: "Iceland", Adjective: "Icelandic"},
	{Kind: PlaceKindCountry, Name: "India", Adjective: "Indian"},
	{Kind: PlaceKindCountry, Name: "Indonesia", Adjective: "Indonesian"},
	{Kind: PlaceKindCountry, Name: "Iran", Adjective: "Iranian"},
	{Kind: PlaceKindCountry, Name: "Iraq", Adjective: "Iraqi"},
	{Kind: PlaceKindCountry, Name: "Ireland", Adjective: "Irish"},
	{Kind: PlaceKindCountry, Name: "Isle of Man", Adjective: "Manx"},
	{Kind: PlaceKindCountry, Name: "Israel", Adjective: "Israeli"},
	{Kind: PlaceKindCountry, Name: "Italy", Adjective: "Italian"},
	{Kind: PlaceKindCountry, Name: "Jamaica", Adjective: "Jamaican"},
	{Kind: PlaceKindCountry, Name: "Japan", Adjective: "Japanese"},
	{Kind: PlaceKindCountry, Name: "Jersey", Adjective: "Jersey"},
	{Kind: PlaceKindCountry, Name: "Jordan", Adjective: "Jordanian"},
	{Kind: PlaceKindCountry, Name: "Kazakhstan", Adjective: "Kazakhstani"},
	{Kind: PlaceKindCountry, Name: "Kenya", Adjective: "Kenyan"},
	{Kind: PlaceKindCountry, Name: "Kuwait", Adjective: "Kuwaiti"},
	{Kind: PlaceKindCountry, Name: "Kyrgyzstan", Adjective: "Kyrgyzstani"},
	{Kind: PlaceKindCountry, Name: "Laos", Adjective: "Lao"},
	{Kind: PlaceKindCountry, Name: "Latvia", Adjective: "Latvians"},
	{Kind: PlaceKindCountry, Name: "Lebanon", Adjective: "Lebanese"},
	{Kind: PlaceKindCountry, Name: "Liberia", Adjective: "Liberian"},
	{Kind: PlaceKindCountry, Name: "Libya", Adjective: "Libyan"},
	{Kind: PlaceKindCountry, Name: "Liechtenstein", Adjective: "Liechtensteiner"},
	{Kind: PlaceKindCountry, Name: "Lithuania", Adjective: "Lithuanian"},
	{Kind: PlaceKindCountry, Name: "Luxembourg", Adjective: "Luxembourgish"},
	{Kind: PlaceKindCountry, Name: "Madagascar", Adjective: "Madagascan"},
	{Kind: PlaceKindCountry, Name: "Malaysia", Adjective: "Malaysian"},
	{Kind: PlaceKindCountry, Name: "Mali", Adjective: "Malian"},
	{Kind: PlaceKindCountry, Name: "Malta", Adjective: "Maltese"},
	{Kind: PlaceKindCountry, Name: "Marshall Islands", Adjective: "Marshallese"},
	{Kind: PlaceKindCountry, Name: "Martinique", Adjective: "Martiniquais"},
	{Kind: PlaceKindCountry, Name: "Mauritania", Adjective: "Mauritanian"},
	{Kind: PlaceKindCountry, Name: "Mauritius", Adjective: "Mauritian"},
	{Kind: PlaceKindCountry, Name: "Mayotte", Adjective: "Mahoran"},
	{Kind: PlaceKindCountry, Name: "Mexico", Adjective: "Mexican"},
	{Kind: PlaceKindCountry, Name: "Moldova", Adjective: "Moldovan"},
	{Kind: PlaceKindCountry, Name: "Monaco", Adjective: "Monégasque"},
	{Kind: PlaceKindCountry, Name: "Mongolia", Adjective: "Mongolian"},
	{Kind: PlaceKindCountry, Name: "Montenegro", Adjective: "Montenegrin"},
	{Kind: PlaceKindCountry, Name: "Montserrat", Adjective: "Montserratian"},
	{Kind: PlaceKindCountry, Name: "Morocco", Adjective: "Moroccan"},
	{Kind: PlaceKindCountry, Name: "Mozambique", Adjective: "Mozambican"},
	{Kind: PlaceKindCountry, Name: "Namibia", Adjective: "Namibian"},
	{Kind: PlaceKindCountry, Name: "Nauru", Adjective: "Nauruan"},
	{Kind: PlaceKindCountry, Name: "Nepal", Adjective: "Nepali"},
	{Kind: PlaceKindCountry, Name: "Netherlands", Adjective: "Dutch"},
	{Kind: PlaceKindCountry, Name: "New Caledonia", Adjective: "New Caledonian"},
	{Kind: PlaceKindCountry, Name: "New Zealand", Adjective: "New Zealand"},
	{Kind: PlaceKindCountry, Name: "Nicaragua", Adjective: "Nicaraguan"},
	{Kind: PlaceKindCountry, Name: "Nigeria", Adjective: "Nigerian"},
	{Kind: PlaceKindCountry, Name: "Norway", Adjective: "Norwegian"},
	{Kind: PlaceKindCountry, Name: "Oman", Adjective: "Omani"},
	{Kind: PlaceKindCountry, Name: "Pakistan", Adjective: "Pakistani"},
	{Kind: PlaceKindCountry, Name: "Palau", Adjective: "Palauan"},
	{Kind: PlaceKindCountry, Name: "Panama", Adjective: "Panamanian"},
	{Kind: PlaceKindCountry, Name: "Papua New Guinea", Adjective: "Papua New Guinean"},
	{Kind: PlaceKindCountry, Name: "Paraguay", Adjective: "Paraguayan"},
	{Kind: PlaceKindCountry, Name: "Peru", Adjective: "Peruvian"},
	{Kind: PlaceKindCountry, Name: "Philippines", Adjective: "Filipino"},
	{Kind: PlaceKindCountry, Name: "Poland", Adjective: "Polish"},
	{Kind: PlaceKindCountry, Name: "Portugal", Adjective: "Portuguese"},
	{Kind: PlaceKindCountry, Name: "Puerto Rico", Adjective: "Puerto Rican"},
	{Kind: PlaceKindCountry, Name: "Qatar", Adjective: "Qatari"},
	{Kind: PlaceKindCountry, Name: "Romania", Adjective: "Romanian"},
	{Kind: PlaceKindCountry, Name: "Rwanda", Adjective: "Rwandan"},
	{Kind: PlaceKindCountry, Name: "Samoa", Adjective: "Samoan"},
	{Kind: PlaceKindCountry, Name: "San Marino", Adjective: "Sammarinese"},
	{Kind: PlaceKindCountry, Name: "Saudi Arabia", Adjective: "Saudi"},
	{Kind: PlaceKindCountry, Name: "Senegal", Adjective: "Senegalese"},
	{Kind: PlaceKindCountry, Name: "Serbia", Adjective: "Serbian"},
	{Kind: PlaceKindCountry, Name: "Slovakia", Adjective: "Slovak"},
	{Kind: PlaceKindCountry, Name: "Slovenia", Adjective: "Slovenian"},
	{Kind: PlaceKindCountry, Name: "Somalia", Adjective: "Somalian>"},
	{Kind: PlaceKindCountry, Name: "Somaliland", Adjective: "Somalilander"},
	{Kind: PlaceKindCountry, Name: "South Africa", Adjective: "South African"},
	{Kind: PlaceKindCountry, Name: "Spain", Adjective: "Spanish"},
	{Kind: PlaceKindCountry, Name: "Sri Lanka", Adjective: "Sri Lankan"},
	{Kind: PlaceKindCountry, Name: "Palestine", Adjective: "Palestinian"},
	{Kind: PlaceKindCountry, Name: "Sudan", Adjective: "Sudanese"},
	{Kind: PlaceKindCountry, Name: "Suriname", Adjective: "Surinamese"},
	{Kind: PlaceKindCountry, Name: "Sweden", Adjective: "Swedish"},
	{Kind: PlaceKindCountry, Name: "Switzerland", Adjective: "Swiss"},
	{Kind: PlaceKindCountry, Name: "Syria", Adjective: "Syrian"},
	{Kind: PlaceKindCountry, Name: "Tanzania", Adjective: "Tanzanian"},
	{Kind: PlaceKindCountry, Name: "Thailand", Adjective: "Thai"},
	{Kind: PlaceKindCountry, Name: "Tonga", Adjective: "Tongan"},
	{Kind: PlaceKindCountry, Name: "Tunisia", Adjective: "Tunisian"},
	{Kind: PlaceKindCountry, Name: "Turkey", Adjective: "Turkish"},
	{Kind: PlaceKindCountry, Name: "Turkmenistan", Adjective: "Turkmen"},
	{Kind: PlaceKindCountry, Name: "Uganda", Adjective: "Ugandan"},
	{Kind: PlaceKindCountry, Name: "Ukraine", Adjective: "Ukrainian"},
	{Kind: PlaceKindCountry, Name: "United Arab Emirates", Adjective: "Emirati"},
	// {Kind: PlaceKindCountry, Name: "United Kingdom", Adjective: "British", Aliases: []string{"UK", "Britain"}, ChildHints: []Hint{MaybeUKNation{}}},
	{Kind: PlaceKindCountry, Name: "United States of America", Adjective: "American", Aliases: []string{"US", "USA", "United States"}},
	{Kind: PlaceKindCountry, Name: "Uruguay", Adjective: "Uruguayans"},
	{Kind: PlaceKindCountry, Name: "Uzbekistan", Adjective: "Uzbekistani"},
	{Kind: PlaceKindCountry, Name: "Vatican City", Adjective: "Vaticanian"},
	{Kind: PlaceKindCountry, Name: "Venezuela", Adjective: "Venezuelan"},
	{Kind: PlaceKindCountry, Name: "Vietnam", Adjective: "Vietnamese"},
	{Kind: PlaceKindCountry, Name: "Yemen", Adjective: "Yemeni"},
	{Kind: PlaceKindCountry, Name: "Zambia", Adjective: "Zambian"},
	{Kind: PlaceKindCountry, Name: "Zanzibar", Adjective: "Zanzibari"},
	{Kind: PlaceKindCountry, Name: "Zimbabwe", Adjective: "Zimbabwean"},
}

var ukNationNames = []*PlaceName{
	{Kind: PlaceKindUKNation, Name: "England", Adjective: "English", PartOf: []*PlaceName{countryUK}},
	{Kind: PlaceKindUKNation, Name: "Scotland", Adjective: "Scottish", PartOf: []*PlaceName{countryUK}},
	{Kind: PlaceKindUKNation, Name: "Northern Ireland", Adjective: "Northern Irish", PartOf: []*PlaceName{countryUK}},
	{Kind: PlaceKindUKNation, Name: "Wales", Adjective: "Welsh", PartOf: []*PlaceName{countryUK}},
}

func splitPlaceName(s string) []string {
	parts := []string{}
	var b strings.Builder
	b.Grow(len(s))

	var seenChar bool
	var prevWasSpace bool
	var prevWasSeparator bool
	for _, c := range s {
		if !unicode.IsGraphic(c) {
			continue
		}
		if unicode.IsSpace(c) {
			// collapse whitespace
			if prevWasSpace || !seenChar {
				continue
			}
			prevWasSpace = true
			continue
		}

		if c == ',' || c == ';' {
			if prevWasSeparator || !seenChar {
				continue
			}
			prevWasSeparator = true
			prevWasSpace = true
			continue
		}

		if (unicode.IsPunct(c) || unicode.IsSymbol(c)) && c != '-' {
			continue
		}

		if prevWasSeparator {
			parts = append(parts, b.String())
			b.Reset()
			prevWasSeparator = false
			prevWasSpace = false
		} else if prevWasSpace {
			b.WriteRune(' ')
			prevWasSpace = false
		}
		b.WriteRune(c)
		seenChar = true
	}

	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	return parts
}

func Clean(name string) string {
	return strings.Join(splitPlaceName(name), ", ")
}

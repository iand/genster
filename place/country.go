package place

import (
	"strings"
)

type Country struct {
	Name      string
	Adjective string
	Aliases   []string
	Unknown   bool
}

func (c *Country) IsUnknown() bool {
	if c == nil {
		return true
	}
	return c.Unknown
}

func (c *Country) SameAs(other *Country) bool {
	if c == nil || other == nil {
		return false
	}
	return c == other || (c.Name != "" && c.Name == other.Name)
}

func UnknownCountry() *Country {
	return &Country{
		Name:      "unknown",
		Adjective: "unknown",
		Unknown:   true,
	}
}

var countryLookup = map[string]Country{}

func init() {
	for _, c := range countries {
		countryLookup[strings.ToLower(c.Name)] = c
		for _, al := range c.Aliases {
			countryLookup[strings.ToLower(al)] = c
		}
	}
}

func LookupCountry(v string) (Country, bool) {
	c, ok := countryLookup[strings.ToLower(v)]
	return c, ok
}

// TODO: not all countries, or historic countries are here
var countries = []Country{
	{Name: "Abkhazia", Adjective: "Abkhaz"},
	{Name: "Afghanistan", Adjective: "Afghan"},
	{Name: "Albania", Adjective: "Albanian"},
	{Name: "Algeria", Adjective: "Algerian"},
	{Name: "American Samoa", Adjective: "American Samoan"},
	{Name: "Andorra", Adjective: "Andorran"},
	{Name: "Angola", Adjective: "Angolan"},
	{Name: "Anguilla", Adjective: "Anguillan"},
	{Name: "Antigua and Barbuda", Adjective: "Antiguan"},
	{Name: "Argentina", Adjective: "Argentine"},
	{Name: "Armenia", Adjective: "Armenian"},
	{Name: "Aruba", Adjective: "Aruban"},
	{Name: "Australia", Adjective: "Australian"},
	{Name: "Austria", Adjective: "Austrian"},
	{Name: "Azerbaijan", Adjective: "Azerbaijani"},
	{Name: "Bahrain", Adjective: "Bahraini"},
	{Name: "Bangladesh", Adjective: "Bangladeshi"},
	{Name: "Barbados", Adjective: "Barbadian"},
	{Name: "Belarus", Adjective: "Belarusian"},
	{Name: "Belgium", Adjective: "Belgian"},
	{Name: "Belize", Adjective: "Belizean"},
	{Name: "Benin", Adjective: "Beninese"},
	{Name: "Bermuda", Adjective: "Bermudian"},
	{Name: "Bhutan", Adjective: "Bhutanese"},
	{Name: "Bolivia", Adjective: "Bolivian"},
	{Name: "Cambodia", Adjective: "Cambodian"},
	{Name: "Chile", Adjective: "Chilean"},
	{Name: "China", Adjective: "Chinese"},
	{Name: "Colombia", Adjective: "Colombian"},
	{Name: "Comoros", Adjective: "Comoran"},
	{Name: "Cook Islands", Adjective: "Cook Island"},
	{Name: "Costa Rica", Adjective: "Costa Rican"},
	{Name: "Croatia", Adjective: "Croatian"},
	{Name: "Cuba", Adjective: "Cuban"},
	{Name: "Cyprus", Adjective: "Cypriot"},
	{Name: "Czech Republic", Adjective: "Czech"},
	{Name: "Denmark", Adjective: "Danish"},
	{Name: "Djibouti", Adjective: "Djiboutian"},
	{Name: "Dominica", Adjective: "Dominican"},
	{Name: "Dominican Republic", Adjective: "Dominican"},
	{Name: "Ecuador", Adjective: "Ecuadorian"},
	{Name: "Egypt", Adjective: "Egyptian"},
	{Name: "El Salvador", Adjective: "Salvadoran"},
	{Name: "England", Adjective: "English"},
	{Name: "Equatorial Guinea", Adjective: "Equatoguinean"},
	{Name: "Eritrea", Adjective: "Eritrean"},
	{Name: "Estonia", Adjective: "Estonian"},
	{Name: "Ethiopia", Adjective: "Ethiopian"},
	{Name: "Faroe Islands", Adjective: "Faroese"},
	{Name: "Fiji", Adjective: "Fijian"},
	{Name: "Finland", Adjective: "Finnish"},
	{Name: "France", Adjective: "French"},
	{Name: "French Guiana", Adjective: "French Guianese"},
	{Name: "French Polynesia", Adjective: "French Polynesian"},
	{Name: "Gabon", Adjective: "Gabonese"},
	{Name: "Georgia", Adjective: "Georgian"},
	{Name: "Germany", Adjective: "German"},
	{Name: "Ghana", Adjective: "Ghanaian"},
	{Name: "Gibraltar", Adjective: "Gibraltan"},
	{Name: "Greece", Adjective: "Greek"},
	{Name: "Guatemala", Adjective: "Guatemalan"},
	{Name: "Guernsey", Adjective: "Guernsey"},
	{Name: "Guinea-Bissau", Adjective: "Bissau-Guinean"},
	{Name: "Guinea", Adjective: "Guinean"},
	{Name: "Guyana", Adjective: "Guyanese"},
	{Name: "Haiti", Adjective: "Haitian"},
	{Name: "Honduras", Adjective: "Honduran"},
	{Name: "Hong Kong", Adjective: "Hong Kong"},
	{Name: "Hungary", Adjective: "Hungarian"},
	{Name: "Iceland", Adjective: "Icelandic"},
	{Name: "India", Adjective: "Indian"},
	{Name: "Indonesia", Adjective: "Indonesian"},
	{Name: "Iran", Adjective: "Iranian"},
	{Name: "Iraq", Adjective: "Iraqi"},
	{Name: "Isle of Man", Adjective: "Manx"},
	{Name: "Israel", Adjective: "Israeli"},
	{Name: "Italy", Adjective: "Italian"},
	{Name: "Jamaica", Adjective: "Jamaican"},
	{Name: "Japan", Adjective: "Japanese"},
	{Name: "Jersey", Adjective: "Jersey"},
	{Name: "Jordan", Adjective: "Jordanian"},
	{Name: "Kazakhstan", Adjective: "Kazakhstani"},
	{Name: "Kenya", Adjective: "Kenyan"},
	{Name: "Kuwait", Adjective: "Kuwaiti"},
	{Name: "Kyrgyzstan", Adjective: "Kyrgyzstani"},
	{Name: "Laos", Adjective: "Lao"},
	{Name: "Latvia", Adjective: "Latvians"},
	{Name: "Lebanon", Adjective: "Lebanese"},
	{Name: "Liberia", Adjective: "Liberian"},
	{Name: "Libya", Adjective: "Libyan"},
	{Name: "Liechtenstein", Adjective: "Liechtensteiner"},
	{Name: "Lithuania", Adjective: "Lithuanian"},
	{Name: "Luxembourg", Adjective: "Luxembourgish"},
	{Name: "Madagascar", Adjective: "Madagascan"},
	{Name: "Malaysia", Adjective: "Malaysian"},
	{Name: "Mali", Adjective: "Malian"},
	{Name: "Malta", Adjective: "Maltese"},
	{Name: "Marshall Islands", Adjective: "Marshallese"},
	{Name: "Martinique", Adjective: "Martiniquais"},
	{Name: "Mauritania", Adjective: "Mauritanian"},
	{Name: "Mauritius", Adjective: "Mauritian"},
	{Name: "Mayotte", Adjective: "Mahoran"},
	{Name: "Mexico", Adjective: "Mexican"},
	{Name: "Moldova", Adjective: "Moldovan"},
	{Name: "Monaco", Adjective: "Monégasque"},
	{Name: "Mongolia", Adjective: "Mongolian"},
	{Name: "Montenegro", Adjective: "Montenegrin"},
	{Name: "Montserrat", Adjective: "Montserratian"},
	{Name: "Morocco", Adjective: "Moroccan"},
	{Name: "Mozambique", Adjective: "Mozambican"},
	{Name: "Namibia", Adjective: "Namibian"},
	{Name: "Nauru", Adjective: "Nauruan"},
	{Name: "Nepal", Adjective: "Nepali"},
	{Name: "Netherlands", Adjective: "Dutch"},
	{Name: "New Caledonia", Adjective: "New Caledonian"},
	{Name: "New Zealand", Adjective: "New Zealand"},
	{Name: "Nicaragua", Adjective: "Nicaraguan"},
	{Name: "Nigeria", Adjective: "Nigerian"},
	{Name: "Northern Ireland", Adjective: "Northern Irish"},
	{Name: "Norway", Adjective: "Norwegian"},
	{Name: "Oman", Adjective: "Omani"},
	{Name: "Pakistan", Adjective: "Pakistani"},
	{Name: "Palau", Adjective: "Palauan"},
	{Name: "Panama", Adjective: "Panamanian"},
	{Name: "Papua New Guinea", Adjective: "Papua New Guinean"},
	{Name: "Paraguay", Adjective: "Paraguayan"},
	{Name: "Peru", Adjective: "Peruvian"},
	{Name: "Philippines", Adjective: "Filipino"},
	{Name: "Poland", Adjective: "Polish"},
	{Name: "Portugal", Adjective: "Portuguese"},
	{Name: "Puerto Rico", Adjective: "Puerto Rican"},
	{Name: "Qatar", Adjective: "Qatari"},
	{Name: "Romania", Adjective: "Romanian"},
	{Name: "Rwanda", Adjective: "Rwandan"},
	{Name: "Samoa", Adjective: "Samoan"},
	{Name: "San Marino", Adjective: "Sammarinese"},
	{Name: "Saudi Arabia", Adjective: "Saudi"},
	{Name: "Scotland", Adjective: "Scottish"},
	{Name: "Senegal", Adjective: "Senegalese"},
	{Name: "Serbia", Adjective: "Serbian"},
	{Name: "Slovakia", Adjective: "Slovak"},
	{Name: "Slovenia", Adjective: "Slovenian"},
	{Name: "Somalia", Adjective: "Somalian>"},
	{Name: "Somaliland", Adjective: "Somalilander"},
	{Name: "South Africa", Adjective: "South African"},
	{Name: "Spain", Adjective: "Spanish"},
	{Name: "Sri Lanka", Adjective: "Sri Lankan"},
	{Name: "Palestine", Adjective: "Palestinian"},
	{Name: "Sudan", Adjective: "Sudanese"},
	{Name: "Suriname", Adjective: "Surinamese"},
	{Name: "Sweden", Adjective: "Swedish"},
	{Name: "Switzerland", Adjective: "Swiss"},
	{Name: "Syria", Adjective: "Syrian"},
	{Name: "Tanzania", Adjective: "Tanzanian"},
	{Name: "Thailand", Adjective: "Thai"},
	{Name: "Tonga", Adjective: "Tongan"},
	{Name: "Tunisia", Adjective: "Tunisian"},
	{Name: "Turkey", Adjective: "Turkish"},
	{Name: "Turkmenistan", Adjective: "Turkmen"},
	{Name: "Uganda", Adjective: "Ugandan"},
	{Name: "Ukraine", Adjective: "Ukrainian"},
	{Name: "United Arab Emirates", Adjective: "Emirati"},
	{Name: "United Kingdom", Adjective: "British", Aliases: []string{"UK", "Britain"}},
	{Name: "United States of America", Adjective: "American", Aliases: []string{"US", "USA", "United States"}},
	{Name: "Uruguay", Adjective: "Uruguayans"},
	{Name: "Uzbekistan", Adjective: "Uzbekistani"},
	{Name: "Vatican City", Adjective: "Vaticanian"},
	{Name: "Venezuela", Adjective: "Venezuelan"},
	{Name: "Vietnam", Adjective: "Vietnamese"},
	{Name: "Wales", Adjective: "Welsh"},
	{Name: "Yemen", Adjective: "Yemeni"},
	{Name: "Zambia", Adjective: "Zambian"},
	{Name: "Zanzibar", Adjective: "Zanzibari"},
	{Name: "Zimbabwe", Adjective: "Zimbabwean"},
}

package place

// // TODO: Give places a Familiar Name, Sort Name, Full Name, Unique Name, just like Person
// // Remember when we mention it more than once and switch from Full Name to FamiliarName
// // Also remember the original unparsed text

type Place interface {
	// String returns the most detailed description of the place available
	String() string

	Country() string

	// Locality returns just the locality of the place such as the town or village
	Locality() string

	// FullLocality returns the locality of the place within it's national hierarchy such as town, county, country
	FullLocality() string
}

// func IsUnknown(p Place) bool {
// 	if p == nil {
// 		return true
// 	}
// 	_, ok := p.(*Unknown)
// 	return ok
// }

// var _ Place = (*Address)(nil)

// type Address struct {
// 	Full        string
// 	Line1       string
// 	Line2       string
// 	Line3       string
// 	City        string
// 	State       string
// 	PostalCode  string
// 	CountryName string
// }

// func (a *Address) String() string {
// 	return "an unknown place"
// }

// func (a *Address) Country() string {
// 	return a.CountryName
// }

// func (a *Address) Locality() string {
// 	if a.City != "" {
// 		return a.City
// 	}
// 	if a.State != "" {
// 		return a.State
// 	}
// 	if a.CountryName != "" {
// 		return a.CountryName
// 	}
// 	if a.Full != "" {
// 		return a.Full
// 	}
// 	return "unknown"
// }

// func (a *Address) FullLocality() string {
// 	if a.CountryName != "" {
// 		local := ""
// 		if a.City != "" {
// 			local = a.City
// 		} else if a.State != "" {
// 			local = a.State
// 		}

// 		if local != "" {
// 			return local + ", " + a.CountryName
// 		}
// 		return a.CountryName
// 	}

// 	if a.Full != "" {
// 		return a.Full
// 	}
// 	return "unknown"
// }

// var _ Place = (*Unknown)(nil)

// type Unknown struct {
// 	Text string
// }

// func (u *Unknown) String() string {
// 	return "an unknown place"
// }

// func (u *Unknown) Country() string {
// 	return "an unknown place"
// }

// func (u *Unknown) Locality() string {
// 	return "unknown"
// }

// func (u *Unknown) FullLocality() string {
// 	return "unknown"
// }

// var _ Place = (*UKRegistrationDistrict)(nil)

// type UKRegistrationDistrict struct {
// 	Name string
// }

// func (u *UKRegistrationDistrict) String() string {
// 	return u.Name + ", United Kingdom"
// }

// func (u *UKRegistrationDistrict) Country() string {
// 	return "United Kingdom"
// }

// func (u *UKRegistrationDistrict) Locality() string {
// 	return u.Name
// }

// func (u *UKRegistrationDistrict) FullLocality() string {
// 	return u.Name + ", United Kingdom"
// }

// var _ Place = (*UKRegistrationDistrict)(nil)

// type TownCountyCountry struct {
// 	Original    string
// 	Town        string
// 	County      string
// 	CountryName string
// }

// func (u *TownCountyCountry) String() string {
// 	return u.Original
// }

// func (u *TownCountyCountry) Country() string {
// 	return u.CountryName
// }

// func (u *TownCountyCountry) Locality() string {
// 	return u.Town
// }

// func (u *TownCountyCountry) FullLocality() string {
// 	return u.Town + ", " + u.County + ", " + u.CountryName
// }

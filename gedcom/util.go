package gedcom

import (
	"net/url"
	"strings"

	"github.com/iand/gdate"
	"github.com/iand/gedcom"
	"github.com/iand/genster/model"
)

func findUserDefinedTags(us []gedcom.UserDefinedTag, tag string, recurse bool) []gedcom.UserDefinedTag {
	res := make([]gedcom.UserDefinedTag, 0)
	for i := range us {
		if us[i].Tag == tag {
			res = append(res, us[i])
		}
		if recurse {
			more := findUserDefinedTags(us[i].UserDefined, tag, recurse)
			if len(more) > 0 {
				res = append(res, more...)
			}
		}
	}

	return res
}

func parseURL(u string) *model.Link {
	l := &model.Link{
		Title: u,
		URL:   u,
	}
	pu, err := url.Parse(u)
	if err == nil && pu != nil && pu.Host != "" {
		l.Title = refineHostName(pu.Host)
	}

	return l
}

func refineHostName(h string) string {
	h = strings.ToLower(h)

	h = strings.TrimPrefix(h, "www.")

	// Check against some well known hostnames
	// TODO: read well known hostnames from config
	switch h {
	case "familysearch.org":
		return "Family Search"
	case "findmypast.co.uk":
		return "FindMyPast"
	case "britishnewspaperarchive.co.uk":
		return "British Newspaper Archive"
	default:
		return h
	}
}

func stringOneOf(s string, alts ...string) bool {
	for _, a := range alts {
		if a == s {
			return true
		}
	}
	return false
}

func stripXref(s string) string {
	return strings.Trim(s, "@")
}

// reckoningForPlace attempts to find a ReckoningLocation based on the place
func reckoningForPlace(pl *model.Place) gdate.ReckoningLocation {
	if pl.IsUnknown() || pl.Country.IsUnknown() {
		return gdate.ReckoningLocationNone
	}
	switch strings.ToLower(pl.Country.Name) {
	case "england", "wales":
		return gdate.ReckoningLocationEnglandAndWales
	case "scotland":
		return gdate.ReckoningLocationScotland
	case "ireland":
		return gdate.ReckoningLocationIreland
	default:
		return gdate.ReckoningLocationNone
	}
}

package place

import (
	"strings"
	"unicode"

	"github.com/iand/genster/logging"
)

// ParseNameHierarchy attempts to parse a placename into a hierarchy
func ParseNameHierarchy(s string, hints ...Hint) (*PlaceNameHierarchy, bool) {
	parts := splitPlaceName(s)
	if len(parts) == 0 {
		return nil, false
	}

	ph := &PlaceNameHierarchy{}

	if len(parts) > 0 {
		ph.Name = PlaceName{Name: parts[0]}
		ph.NormalizedWithHierarchy = strings.Join(parts, ",")
	}

	p := ph
	for i := 1; i < len(parts); i++ {
		p.Parent = &PlaceNameHierarchy{
			Name:                    PlaceName{Name: parts[i]},
			NormalizedWithHierarchy: strings.Join(parts[i:], ","),
			Child:                   p,
		}

		p = p.Parent

	}

	broadest := ph
	for broadest.Parent != nil {
		broadest = broadest.Parent
	}

	_, ok := LookupPlaceName(broadest.Name.Name)
	logging.Debug("parse place hierarchy: checking if country", "name", broadest.Name.Name, "is_country", ok)
	if ok {
		broadest.Kind = PlaceKindCountry
	}

	return ph, true
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

func normalizeParts(parts []string) string {
	return strings.ToLower(strings.Join(parts, ", "))
}

func Normalize(name string) string {
	return normalizeParts(splitPlaceName(name))
}

func Clean(name string) string {
	return strings.Join(splitPlaceName(name), ", ")
}

func LastElement(name string) string {
	parts := splitPlaceName(name)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func CutLastElement(name string) (string, string) {
	parts := splitPlaceName(name)
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ", "), parts[len(parts)-1]
	}
	return "", name
}

package place

import (
	"strings"
	"unicode"
)

// Parse attempts to parse a placename. Additional context may be supplied by the use of hints.
// For example the user could hint that the place could be a UK registration district
// based on the source being the general register office.
func Parse(s string, hints ...Hint) (PlaceHierarchy, bool) {
	parts := splitPlaceName(s)
	if len(parts) == 0 {
		return PlaceHierarchy{}, false
	}

	ph := PlaceHierarchy{
		Hierarchy: []string{},
	}

	if len(parts) > 0 {
		ph.Name = parts[0]
	}
	if len(parts) > 1 {
		ph.Hierarchy = parts[1:]
	}

	// pl, conf := matchBest(pc, hints...)

	// if pl != nil && conf > 0 {
	// 	return pl, nil
	// }
	return ph, true
}

type PlaceHierarchy struct {
	Name      string
	Hierarchy []string
}

func (p *PlaceHierarchy) String() string {
	return strings.Join(append([]string{p.Name}, p.Hierarchy...), ", ")
}

func (p *PlaceHierarchy) Normalized() string {
	return normalizeParts(append([]string{p.Name}, p.Hierarchy...))
}

func (p *PlaceHierarchy) HasParent() bool {
	return len(p.Hierarchy) > 0
}

func (p *PlaceHierarchy) Parent() (PlaceHierarchy, bool) {
	if len(p.Hierarchy) == 0 {
		return PlaceHierarchy{}, false
	}
	return PlaceHierarchy{
		Name:      p.Hierarchy[0],
		Hierarchy: p.Hierarchy[1:],
	}, true
}

func (p *PlaceHierarchy) TrimHierarchy(n int) PlaceHierarchy {
	if n > len(p.Hierarchy) {
		n = len(p.Hierarchy)
	}

	return PlaceHierarchy{
		Name:      p.Name,
		Hierarchy: p.Hierarchy[:len(p.Hierarchy)-n],
	}
}

func matchBest(pc PlaceHierarchy, hints ...Hint) (Place, float64) {
	var bestPlace Place
	var bestConfidence float64

	for _, hint := range hints {
		p, conf := hint(pc)
		if conf > bestConfidence {
			bestConfidence = conf
			bestPlace = p
		}
	}

	return bestPlace, bestConfidence
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

package infer

import (
	"fmt"
	"strings"

	"github.com/iand/genster/model"
)

// maxDuplicateNameDistance is the largest number of letters by which the names of two partners
// born in the same year can differ for them to be taken as possibly the same person.
const maxDuplicateNameDistance = 2

// CheckDuplicatePartners records an anomaly for each pair of the person's partners who were
// born in the same year and whose names differ by no more than a couple of letters, since they
// are probably two records of one person.
func CheckDuplicatePartners(p *model.Person) {
	var partners []*model.Person
	seen := make(map[string]bool)
	for _, f := range uniqueFamilies(p) {
		o := f.OtherParent(p)
		if o.IsUnknown() || o.Unidentified || seen[o.ID] {
			continue
		}
		seen[o.ID] = true
		partners = append(partners, o)
	}

	for i := range partners {
		for j := i + 1; j < len(partners); j++ {
			a, b := partners[i], partners[j]
			year, ok := sameBirthYear(a, b)
			if !ok {
				continue
			}
			if editDistance(strings.ToLower(a.PreferredFullName), strings.ToLower(b.PreferredFullName)) > maxDuplicateNameDistance {
				continue
			}
			addAnomaly(p, "Possible duplicate partner", fmt.Sprintf("%s and %s, both born %d, may be the same person",
				a.PreferredFullName, b.PreferredFullName, year))
		}
	}
}

// sameBirthYear returns the birth year of a and b when it is known for both and the same.
func sameBirthYear(a, b *model.Person) (int, bool) {
	aev, bev := birthForAge(a), birthForAge(b)
	if aev == nil || bev == nil {
		return 0, false
	}
	ay, ok := aev.GetDate().Year()
	if !ok {
		return 0, false
	}
	by, ok := bev.GetDate().Year()
	if !ok || ay != by {
		return 0, false
	}
	return ay, true
}

// editDistance returns the number of single letter insertions, deletions or substitutions
// needed to change a into b.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}

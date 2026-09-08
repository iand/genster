package model

import "regexp"

var reGrampsID = regexp.MustCompile(`^([A-Z])0+([0-9]+)$`)

// NormalizeGrampsID normalizes a gramps id by removing any leading zeroes
// after the prefix character
func NormalizeGrampsID(id string) string {
	m := reGrampsID.FindStringSubmatch(id)
	if m == nil || len(m) != 3 {
		return id
	}
	return m[1] + m[2]
}

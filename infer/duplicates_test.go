package infer

import (
	"testing"

	"github.com/iand/genster/model"
)

// partnerOf returns a partner born in the given year, in a family with p.
func partnerOf(p *model.Person, id, name string, born int) *model.Person {
	o := person(id, name, model.GenderFemale)
	if born != 0 {
		withBirth(o, birthOf(o, model.Year(born)))
	}
	family("f"+id, p, o)
	return o
}

func TestCheckDuplicatePartners(t *testing.T) {
	t.Run("names a letter apart and born the same year", func(t *testing.T) {
		// Shaped like I2147.
		p := person("p", "Hans Larsen", model.GenderMale)
		partnerOf(p, "a", "Ellen Maria Sørensdatter", 1816)
		partnerOf(p, "b", "Elen Maria Sørensdatter", 1816)

		CheckDuplicatePartners(p)

		checkAnomalies(t, p, "Possible duplicate partner", "Ellen Maria Sørensdatter and Elen Maria Sørensdatter, both born 1816, may be the same person")
	})

	t.Run("names differing only in case", func(t *testing.T) {
		p := person("p", "John Smith", model.GenderMale)
		partnerOf(p, "a", "Mary Brown", 1800)
		partnerOf(p, "b", "MARY BROWN", 1800)

		CheckDuplicatePartners(p)

		checkAnomalies(t, p, "Possible duplicate partner", "Mary Brown and MARY BROWN, both born 1800, may be the same person")
	})

	t.Run("same name born in different years", func(t *testing.T) {
		p := person("p", "John Smith", model.GenderMale)
		partnerOf(p, "a", "Mary Brown", 1800)
		partnerOf(p, "b", "Mary Brown", 1810)

		CheckDuplicatePartners(p)

		checkAnomalies(t, p, "Possible duplicate partner")
	})

	t.Run("different names born the same year", func(t *testing.T) {
		p := person("p", "John Smith", model.GenderMale)
		partnerOf(p, "a", "Mary Brown", 1800)
		partnerOf(p, "b", "Mary Green", 1800)

		CheckDuplicatePartners(p)

		checkAnomalies(t, p, "Possible duplicate partner")
	})

	t.Run("birth year unknown", func(t *testing.T) {
		p := person("p", "John Smith", model.GenderMale)
		partnerOf(p, "a", "Mary Brown", 0)
		partnerOf(p, "b", "Mary Brown", 0)

		CheckDuplicatePartners(p)

		checkAnomalies(t, p, "Possible duplicate partner")
	})

	t.Run("same partner in two families", func(t *testing.T) {
		p := person("p", "John Smith", model.GenderMale)
		o := partnerOf(p, "a", "Mary Brown", 1800)
		family("fa2", p, o)

		CheckDuplicatePartners(p)

		checkAnomalies(t, p, "Possible duplicate partner")
	})
}

package tree

import (
	"slices"

	"github.com/iand/genster/model"
)

// WalkFamilyLines traverses the ancestor tree of keyPerson, following the
// paternal line at each generation and pushing maternal branches onto a
// stack. It returns one FamilyLine per direct line found.
func WalkFamilyLines(keyPerson *model.Person) []*model.FamilyLine {
	stack := make([]*model.Person, 0, 20)

	var familyLines []*model.FamilyLine
	familyLine := make([]*model.Family, 0, 20)
	lineage := make([]*model.Person, 0, 20)

	p := keyPerson
	for !p.IsUnknown() {
		if !p.ParentFamily.IsUnknown() {
			if !p.Father.IsUnknown() {
				if !p.Redacted {
					familyLine = append(familyLine, p.ParentFamily)
					lineage = append(lineage, p)
				}
				if !p.Mother.IsUnknown() {
					stack = append(stack, p.Mother)
				}
				p = p.Father
				continue
			}

			if !p.Mother.IsUnknown() {
				if !p.Redacted {
					familyLine = append(familyLine, p.ParentFamily)
					lineage = append(lineage, p)
				}
				p = p.Mother
				continue
			}
		}

		if len(familyLine) > 0 {
			fl := &model.FamilyLine{
				ID:       familyLine[len(familyLine)-1].ID,
				Name:     familyLine[len(familyLine)-1].PreferredUniqueName,
				Families: make([]*model.Family, 0, len(familyLine)),
				Lineage:  make([]*model.Person, 0, len(lineage)),
			}
			for _, f := range slices.Backward(familyLine) {
				fl.Families = append(fl.Families, f)
			}
			lineage = append(lineage, p)
			for _, f := range slices.Backward(lineage) {
				fl.Lineage = append(fl.Lineage, f)
			}
			familyLines = append(familyLines, fl)
			familyLine = make([]*model.Family, 0, 20)
		}
		if len(stack) > 0 {
			p = stack[0]
			stack = stack[1:]
			continue
		}

		break
	}

	return familyLines
}

package tree

import (
	"testing"

	"github.com/iand/genster/model"
)

const unknownFullName = model.UnknownNamePlaceholder + " " + model.UnknownNamePlaceholder

// addNamedPerson adds a person to the tree with the given full name and gender,
// born and died in the given years when they are not zero.
func addNamedPerson(tr *Tree, sid string, name string, gender model.Gender, born, died int) *model.Person {
	p := tr.FindPerson("test", sid)
	p.PreferredFullName = name
	p.PreferredUniqueName = name
	p.Gender = gender
	if born != 0 {
		p.Timeline = append(p.Timeline, &model.BirthEvent{
			GeneralEvent:           model.GeneralEvent{Date: model.Year(born)},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
		})
	}
	if died != 0 {
		p.Timeline = append(p.Timeline, &model.DeathEvent{
			GeneralEvent:           model.GeneralEvent{Date: model.Year(died)},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
		})
	}
	return p
}

// addFamily adds a family with the given parents and children, linking each
// person to it as the loaders do.
func addFamily(tr *Tree, sid string, father, mother *model.Person, children ...*model.Person) *model.Family {
	f := tr.FindFamily("test", sid)
	if father == nil {
		father = model.UnknownPerson()
	} else {
		father.Families = append(father.Families, f)
	}
	if mother == nil {
		mother = model.UnknownPerson()
	} else {
		mother.Families = append(mother.Families, f)
	}
	f.Father = father
	f.Mother = mother
	for _, c := range children {
		f.Children = append(f.Children, c)
		c.Father = father
		c.Mother = mother
	}
	return f
}

func TestUnidentifiedWifeNamedByHusband(t *testing.T) {
	tr := NewTree("test", nil, &SurnameGroups{})
	husband := addNamedPerson(tr, "husband", "William Russel", model.GenderMale, 1711, 1771)
	wife := addNamedPerson(tr, "wife", unknownFullName, model.GenderFemale, 0, 0)
	f := addFamily(tr, "f1", husband, wife)

	if err := tr.Generate(false); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if want := "Wife of William Russel (1711–1771)"; wife.PreferredUniqueName != want {
		t.Errorf("wife unique name: got %q, want %q", wife.PreferredUniqueName, want)
	}
	if want := "William Russel (1711–1771)"; f.PreferredUniqueName != want {
		t.Errorf("family unique name: got %q, want %q", f.PreferredUniqueName, want)
	}
}

func TestUnidentifiedHusbandNamedByWife(t *testing.T) {
	tr := NewTree("test", nil, &SurnameGroups{})
	husband := addNamedPerson(tr, "husband", unknownFullName, model.GenderMale, 0, 0)
	wife := addNamedPerson(tr, "wife", "Mary Smith", model.GenderFemale, 1750, 1800)
	f := addFamily(tr, "f1", husband, wife)

	if err := tr.Generate(false); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if want := "Husband of Mary Smith (1750–1800)"; husband.PreferredUniqueName != want {
		t.Errorf("husband unique name: got %q, want %q", husband.PreferredUniqueName, want)
	}
	if want := "Mary Smith (1750–1800)"; f.PreferredUniqueName != want {
		t.Errorf("family unique name: got %q, want %q", f.PreferredUniqueName, want)
	}
}

func TestUnidentifiedPersonNamedByEarliestSpouse(t *testing.T) {
	tr := NewTree("test", nil, &SurnameGroups{})
	wife := addNamedPerson(tr, "wife", unknownFullName, model.GenderFemale, 0, 0)
	second := addNamedPerson(tr, "second", "Thomas Brown", model.GenderMale, 1700, 1760)
	first := addNamedPerson(tr, "first", "John Green", model.GenderMale, 1690, 1725)
	f2 := addFamily(tr, "f2", second, wife)
	f2.BestStartDate = model.Year(1730)
	f1 := addFamily(tr, "f1", first, wife)
	f1.BestStartDate = model.Year(1715)

	if err := tr.Generate(false); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if want := "Wife of John Green (1690–1725)"; wife.PreferredUniqueName != want {
		t.Errorf("wife unique name: got %q, want %q", wife.PreferredUniqueName, want)
	}
}

func TestUnidentifiedSpouseOfUnknownGender(t *testing.T) {
	tr := NewTree("test", nil, &SurnameGroups{})
	husband := addNamedPerson(tr, "husband", "William Russel", model.GenderMale, 1711, 1771)
	other := addNamedPerson(tr, "other", unknownFullName, model.GenderUnknown, 0, 0)
	addFamily(tr, "f1", husband, other)

	if err := tr.Generate(false); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if want := "Spouse of William Russel (1711–1771)"; other.PreferredUniqueName != want {
		t.Errorf("unique name: got %q, want %q", other.PreferredUniqueName, want)
	}
}

func TestUnidentifiedParentNamedByEldestChild(t *testing.T) {
	tr := NewTree("test", nil, &SurnameGroups{})
	mother := addNamedPerson(tr, "mother", unknownFullName, model.GenderFemale, 0, 0)
	younger := addNamedPerson(tr, "younger", "Ann Russel", model.GenderFemale, 1748, 0)
	elder := addNamedPerson(tr, "elder", "Thomas Russel", model.GenderMale, 1745, 1829)
	addFamily(tr, "f1", nil, mother, younger, elder)

	if err := tr.Generate(false); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if want := "Mother of Thomas Russel (1745–1829)"; mother.PreferredUniqueName != want {
		t.Errorf("mother unique name: got %q, want %q", mother.PreferredUniqueName, want)
	}
}

func TestUnidentifiedChildNamedByParents(t *testing.T) {
	tr := NewTree("test", nil, &SurnameGroups{})
	father := addNamedPerson(tr, "father", "William Russel", model.GenderMale, 1711, 1771)
	mother := addNamedPerson(tr, "mother", "Mary Smith", model.GenderFemale, 1715, 1780)
	son := addNamedPerson(tr, "son", unknownFullName, model.GenderMale, 0, 0)
	addFamily(tr, "f1", father, mother, son)

	if err := tr.Generate(false); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if want := "Son of William Russel (1711–1771) and Mary Smith (1715–1780)"; son.PreferredUniqueName != want {
		t.Errorf("son unique name: got %q, want %q", son.PreferredUniqueName, want)
	}
}

func TestUnidentifiedPersonWithNoRelativesKeepsPlaceholder(t *testing.T) {
	tr := NewTree("test", nil, &SurnameGroups{})
	p := addNamedPerson(tr, "p", unknownFullName, model.GenderFemale, 0, 0)

	if err := tr.Generate(false); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if p.PreferredUniqueName != unknownFullName {
		t.Errorf("unique name: got %q, want %q", p.PreferredUniqueName, unknownFullName)
	}
}

func TestUnidentifiedPersonNotNamedByRedactedSpouse(t *testing.T) {
	tr := NewTree("test", nil, &SurnameGroups{})
	husband := addNamedPerson(tr, "husband", "William Russel", model.GenderMale, 1711, 1771)
	husband.Redacted = true
	wife := addNamedPerson(tr, "wife", unknownFullName, model.GenderFemale, 0, 0)
	addFamily(tr, "f1", husband, wife)

	if err := tr.Generate(false); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if wife.PreferredUniqueName != unknownFullName {
		t.Errorf("wife unique name: got %q, want %q", wife.PreferredUniqueName, unknownFullName)
	}
}

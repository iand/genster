package debug

import (
	"fmt"
	"strings"

	"github.com/iand/genster/model"
)

func DumpPerson(p *model.Person) error {
	fmt.Println("ID:", p.ID)
	fmt.Println("Redacted:", p.Redacted)
	fmt.Println("PreferredFullName:", p.PreferredFullName)
	fmt.Println("PreferredGivenName:", p.PreferredGivenName)
	fmt.Println("PreferredFamiliarName:", p.PreferredFamiliarName)
	fmt.Println("PreferredFamiliarFullName:", p.PreferredFamiliarFullName)
	fmt.Println("PreferredFamilyName:", p.PreferredFamilyName)
	fmt.Println("PreferredSortName:", p.PreferredSortName)
	fmt.Println("PreferredUniqueName:", p.PreferredUniqueName)
	fmt.Println("NickName:", p.NickName)
	fmt.Println("Olb:", p.Olb)
	fmt.Println("Gender:", p.Gender)
	fmt.Println("VitalYears:", p.VitalYears)
	fmt.Println("PossiblyAlive:", p.PossiblyAlive)
	fmt.Println("Unknown:", p.Unknown)
	fmt.Println("Unmarried:", p.Unmarried)
	fmt.Println("Childless:", p.Childless)

	fmt.Println("BestBirthlikeEvent:", ObjectTitle(p.BestBirthlikeEvent))
	fmt.Println("BestDeathlikeEvent:", ObjectTitle(p.BestDeathlikeEvent))

	fmt.Println("Tags:", strings.Join(p.Tags, ", "))

	fmt.Println()

	fmt.Println("Father:", ObjectTitle(p.Father))
	fmt.Println("Mother:", ObjectTitle(p.Mother))

	fmt.Println()

	if len(p.Spouses) == 0 {
		fmt.Println("Spouses: none")
	} else {
		fmt.Println("Spouses:")
		for _, s := range p.Spouses {
			fmt.Printf(" - %s (%s)\n", s.PreferredUniqueName, s.ID)
		}
	}

	fmt.Println()

	if len(p.Children) == 0 {
		fmt.Println("Children: none")
	} else {
		fmt.Println("Children:")
		for _, c := range p.Children {
			fmt.Printf(" - %s\n", ObjectTitle(c))
		}
	}

	fmt.Println()

	if len(p.Families) == 0 {
		fmt.Println("Families: none")
	} else {
		fmt.Println("Families:")
		for i, f := range p.Families {
			if i > 0 {
				fmt.Println()
			}
			fmt.Println("  ID:", f.ID)
			fmt.Println("  Father:", ObjectTitle(f.Father))
			fmt.Println("  Mother:", ObjectTitle(f.Mother))

			if len(f.Children) == 0 {
				fmt.Println("  Children: none")
			} else {
				fmt.Println("  Children:")
				for _, c := range f.Children {
					fmt.Printf("   - %s (%s)\n", c.PreferredUniqueName, c.ID)
				}
			}

			fmt.Println("  PreferredUniqueName:", f.PreferredUniqueName)
			fmt.Println("  Bond:", f.Bond)
			fmt.Println("  BestStartDate:", ObjectTitle(f.BestStartDate))
			fmt.Println("  BestEndDate:", ObjectTitle(f.BestEndDate))
			fmt.Println("  BestStartEvent:", ObjectTitle(f.BestStartEvent))
			fmt.Println("  BestEndEvent:", ObjectTitle(f.BestEndEvent))
			fmt.Println("  EndReason:", f.EndReason)
			fmt.Println("  EndDeathPerson:", ObjectTitle(f.EndDeathPerson))

			if len(f.Timeline) == 0 {
				fmt.Println("  Timeline events:", "none")
			} else {
				fmt.Println("  Timeline events:")
				for _, ev := range f.Timeline {
					fmt.Printf("   - %s\n", ObjectTitle(ev))
				}
			}
		}
	}

	fmt.Println()

	if len(p.Timeline) == 0 {
		fmt.Println("Timeline events:", "none")
	} else {
		fmt.Println("Timeline events:")
		for _, ev := range p.Timeline {
			fmt.Printf("   - %s\n", ObjectTitle(ev))
		}
	}

	fmt.Println()

	if len(p.Occupations) == 0 {
		fmt.Println("Occupations:", "none")
	} else {
		fmt.Println("Occupations:")
		for i, o := range p.Occupations {
			if i > 0 {
				fmt.Println()
			}
			fmt.Println("  Title:", o.Title)
			fmt.Println("  Detail:", o.Detail)
			fmt.Println("  Occurrences:", o.Occurrences)
			fmt.Println("  StartDate:", ObjectTitle(o.StartDate))
			fmt.Println("  EndDate:", ObjectTitle(o.EndDate))
			fmt.Println("  Place:", ObjectTitle(o.Place))
		}
	}

	fmt.Println()

	if len(p.MiscFacts) == 0 {
		fmt.Println("MiscFacts:", "none")
	} else {
		fmt.Println("MiscFacts:")
		for i, f := range p.MiscFacts {
			if i > 0 {
				fmt.Println()
			}
			fmt.Println("  Category:", f.Category)
			fmt.Println("  Detail:", f.Detail)
		}
	}

	fmt.Println()

	if len(p.Inferences) == 0 {
		fmt.Println("Inferences:", "none")
	} else {
		fmt.Println("Inferences:")
		for i, f := range p.Inferences {
			if i > 0 {
				fmt.Println()
			}
			fmt.Println("  Type:", f.Type)
			fmt.Println("  Value:", f.Value)
			fmt.Println("  Reason:", f.Reason)
		}
	}

	fmt.Println()

	if len(p.Anomalies) == 0 {
		fmt.Println("Anomalies:", "none")
	} else {
		fmt.Println("Anomalies:")
		for i, f := range p.Anomalies {
			if i > 0 {
				fmt.Println()
			}
			fmt.Println("  Category:", f.Category)
			fmt.Println("  Text:", f.Text)
			fmt.Println("  Context:", f.Context)
		}
	}

	return nil
}

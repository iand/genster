package debug

import (
	"fmt"
	"io"
	"strings"

	"github.com/iand/genster/model"
)

func DumpPerson(p *model.Person, w io.Writer) error {
	fmt.Fprintln(w, "ID:", p.ID)
	fmt.Fprintln(w, "Redacted:", p.Redacted)
	fmt.Fprintln(w, "PreferredFullName:", p.PreferredFullName)
	fmt.Fprintln(w, "PreferredGivenName:", p.PreferredGivenName)
	fmt.Fprintln(w, "PreferredFamiliarName:", p.PreferredFamiliarName)
	fmt.Fprintln(w, "PreferredFamiliarFullName:", p.PreferredFamiliarFullName)
	fmt.Fprintln(w, "PreferredFamilyName:", p.PreferredFamilyName)
	fmt.Fprintln(w, "PreferredSortName:", p.PreferredSortName)
	fmt.Fprintln(w, "PreferredUniqueName:", p.PreferredUniqueName)
	fmt.Fprintln(w, "NickName:", p.NickName)
	if len(p.KnownNames) == 0 {
		fmt.Fprintln(w, "Known Names: none")
	} else {
		fmt.Fprintln(w, "Known Names:")
		for _, n := range p.KnownNames {
			fmt.Fprintf(w, " - %s\n", n.Name)
		}
	}

	fmt.Fprintln(w, "Olb:", p.Olb)
	fmt.Fprintln(w, "Gender:", p.Gender)
	fmt.Fprintln(w, "VitalYears:", p.VitalYears)
	fmt.Fprintln(w, "PossiblyAlive:", p.PossiblyAlive)
	fmt.Fprintln(w, "Unknown:", p.Unknown)
	fmt.Fprintln(w, "Unmarried:", p.Unmarried)
	fmt.Fprintln(w, "Childless:", p.Childless)

	fmt.Fprintln(w, "BestBirthlikeEvent:", ObjectTitle(p.BestBirthlikeEvent))
	fmt.Fprintln(w, "BestDeathlikeEvent:", ObjectTitle(p.BestDeathlikeEvent))

	fmt.Fprintln(w, "Tags:", strings.Join(p.Tags, ", "))

	fmt.Fprintln(w)

	fmt.Fprintln(w, "Father:", ObjectTitle(p.Father))
	fmt.Fprintln(w, "Mother:", ObjectTitle(p.Mother))

	fmt.Fprintln(w)

	if len(p.Spouses) == 0 {
		fmt.Fprintln(w, "SPOUSES: none")
	} else {
		fmt.Fprintln(w, "SPOUSES:")
		for _, s := range p.Spouses {
			fmt.Fprintf(w, " - %s (%s)\n", s.PreferredUniqueName, s.ID)
		}
	}

	fmt.Fprintln(w)

	if len(p.Children) == 0 {
		fmt.Fprintln(w, "CHILDREN: none")
	} else {
		fmt.Fprintln(w, "CHILDREN:")
		for _, c := range p.Children {
			fmt.Fprintf(w, " - %s\n", ObjectTitle(c))
		}
	}

	fmt.Fprintln(w)

	if len(p.Families) == 0 {
		fmt.Fprintln(w, "FAMILIES: none")
	} else {
		fmt.Fprintln(w, "FAMILIES:")
		for i, f := range p.Families {
			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintln(w, "  ID:", f.ID)
			fmt.Fprintln(w, "  Father:", ObjectTitle(f.Father))
			fmt.Fprintln(w, "  Mother:", ObjectTitle(f.Mother))

			if len(f.Children) == 0 {
				fmt.Fprintln(w, "  Children: none")
			} else {
				fmt.Fprintln(w, "  Children:")
				for _, c := range f.Children {
					fmt.Fprintf(w, "   - %s (%s)\n", c.PreferredUniqueName, c.ID)
				}
			}

			fmt.Fprintln(w, "  PreferredUniqueName:", f.PreferredUniqueName)
			fmt.Fprintln(w, "  Bond:", f.Bond)
			fmt.Fprintln(w, "  BestStartDate:", ObjectTitle(f.BestStartDate))
			fmt.Fprintln(w, "  BestEndDate:", ObjectTitle(f.BestEndDate))
			fmt.Fprintln(w, "  BestStartEvent:", ObjectTitle(f.BestStartEvent))
			fmt.Fprintln(w, "  BestEndEvent:", ObjectTitle(f.BestEndEvent))
			fmt.Fprintln(w, "  EndReason:", f.EndReason)
			fmt.Fprintln(w, "  EndDeathPerson:", ObjectTitle(f.EndDeathPerson))

			if len(f.Timeline) == 0 {
				fmt.Fprintln(w, "  Timeline events:", "none")
			} else {
				fmt.Fprintln(w, "  Timeline events:")
				for _, ev := range f.Timeline {
					fmt.Fprintf(w, "   - %s\n", ObjectTitle(ev))
				}
			}
		}
	}

	fmt.Fprintln(w)

	if len(p.Timeline) == 0 {
		fmt.Fprintln(w, "TIMELINE EVENTS:", "none")
	} else {
		fmt.Fprintln(w, "TIMELINE EVENTS:")
		for _, ev := range p.Timeline {
			fmt.Fprintf(w, "   - %s\n", ObjectTitle(ev))
		}
	}

	fmt.Fprintln(w)

	if len(p.Occupations) == 0 {
		fmt.Fprintln(w, "OCCUPATIONS:", "none")
	} else {
		fmt.Fprintln(w, "OCCUPATIONS:")
		for i, o := range p.Occupations {
			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintln(w, "  Title:", o.Name)
			fmt.Fprintln(w, "  Detail:", o.Detail)
			fmt.Fprintln(w, "  Occurrences:", o.Occurrences)
			fmt.Fprintln(w, "  StartDate:", ObjectTitle(o.StartDate))
			fmt.Fprintln(w, "  EndDate:", ObjectTitle(o.EndDate))
			fmt.Fprintln(w, "  Place:", ObjectTitle(o.Place))
		}
	}

	fmt.Fprintln(w)

	if len(p.MiscFacts) == 0 {
		fmt.Fprintln(w, "MISC FACTS:", "none")
	} else {
		fmt.Fprintln(w, "MISC FACTS:")
		for i, f := range p.MiscFacts {
			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintln(w, "  Category:", f.Category)
			fmt.Fprintln(w, "  Detail:", f.Detail)
		}
	}

	fmt.Fprintln(w)

	if len(p.Inferences) == 0 {
		fmt.Fprintln(w, "INFERENCES:", "none")
	} else {
		fmt.Fprintln(w, "INFERENCES:")
		for i, f := range p.Inferences {
			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintln(w, "  Type:", f.Type)
			fmt.Fprintln(w, "  Value:", f.Value)
			fmt.Fprintln(w, "  Reason:", f.Reason)
		}
	}

	fmt.Fprintln(w)

	if len(p.Anomalies) == 0 {
		fmt.Fprintln(w, "ANOMALIES:", "none")
	} else {
		fmt.Fprintln(w, "ANOMALIES:")
		for i, f := range p.Anomalies {
			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintln(w, "  Category:", f.Category)
			fmt.Fprintln(w, "  Text:", f.Text)
			fmt.Fprintln(w, "  Context:", f.Context)
		}
	}

	return nil
}

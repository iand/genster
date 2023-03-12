package gedcom

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/iand/gdate"
	"github.com/iand/gedcom"
	"github.com/iand/genster/model"
	"golang.org/x/exp/slog"
)

var reUppercase = regexp.MustCompile(`^[A-Z \-]{3}[A-Z \-]+$`)

func (l *Loader) populatePersonFacts(m ModelFinder, in *gedcom.IndividualRecord) error {
	slog.Debug("populating from individual record", "xref", in.Xref)
	p := m.FindPerson(l.ScopeName, in.Xref)

	if len(in.Name) == 0 {
		p.PreferredFullName = "unknown"
		p.PreferredGivenName = "unknown"
		p.PreferredFamiliarName = "unknown"
		p.PreferredFamilyName = "unknown"
		p.PreferredSortName = "unknown"
		p.PreferredUniqueName = "unknown"
	} else {
		name := strings.ReplaceAll(in.Name[0].Name, "-?-", model.UnknownNamePlaceholder)

		prefName := gedcom.SplitPersonalName(name)

		if prefName.Surname == "" &&
			strings.Contains(prefName.Full, " ") &&
			!stringOneOf(prefName.Full, "Mary Ann") {
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: "Name",
				Text:     fmt.Sprintf("Person has no surname but full name %q contains more than one word.", prefName.Full),
				Context:  "Person's name",
			})
		}

		if prefName.Surname == "" {
			prefName.Surname = model.UnknownNamePlaceholder
			prefName.Full += model.UnknownNamePlaceholder
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: "Name",
				Text:     "Person has no surname, should replace with -?-.",
				Context:  "Person's name",
			})
		}

		if reUppercase.MatchString(prefName.Full) {
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: "Name",
				Text:     "Person's name is all uppercase, should change to proper case.",
				Context:  "Person's name",
			})
		} else if reUppercase.MatchString(prefName.Surname) {
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: "Name",
				Text:     "Person's surname is all uppercase, should change to proper case.",
				Context:  "Person's name",
			})
		}

		if prefName.Given == "" {
			prefName.Given = model.UnknownNamePlaceholder
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: "Name",
				Text:     "Person has no given name, should replace with -?-.",
				Context:  "Person's name",
			})
		}

		p.PreferredFullName = prefName.Full
		p.PreferredGivenName = prefName.Given

		p.PreferredFamilyName = prefName.Surname
		p.PreferredSortName = prefName.Surname + ", " + prefName.Given
		if prefName.Suffix != "" {
			p.PreferredSortName += " " + prefName.Suffix
		}
		p.PreferredUniqueName = prefName.Full
		p.NickName = prefName.Nickname

	}

	switch in.Sex {
	case "M", "m":
		p.Gender = model.GenderMale
	case "F", "f":
		p.Gender = model.GenderFemale
	default:
		p.Gender = model.GenderUnknown
	}

	events := append([]*gedcom.EventRecord{}, in.Event...)
	events = append(events, in.Attribute...)

	// collect occupation events and attempt to consolidate them later
	occupationEvents := make([]model.GeneralEvent, 0)

	for _, er := range events {
		pl, anoms := l.findPlaceForEvent(m, er)

		dt, err := gdate.Parse(er.Date)
		if err != nil {
			return fmt.Errorf("date: %w", err)
		}

		detail := er.Value
		if detail == "" {
			if len(er.Note) > 0 {
				detail = er.Note[0].Note
			}
		}

		gev := model.GeneralEvent{
			Date:   dt,
			Place:  pl,
			Detail: detail,
			Title:  fmt.Sprintf("%s", er.Tag),
		}

		giv := model.GeneralIndividualEvent{
			Principal: p,
		}

		gev.Citations = l.parseCitationRecords(m, er.Citation)

		var ev model.TimelineEvent

		switch er.Tag {
		case "BIRT":
			ev = &model.BirthEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}

		case "BAPM", "CHR":
			ev = &model.BaptismEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}

		case "DEAT":
			ev = &model.DeathEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}

		case "BURI":
			ev = &model.BurialEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}

		case "CREM":
			ev = &model.CremationEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}

		case "PROB":
			ev = &model.ProbateEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}

		case "RESI":
			// This might be an ancestry census event
			censusDate, isCensus := maybeFixCensusDate(er)
			if isCensus {
				gev.Date = censusDate
				ev = l.populateCensusRecord(er, gev, p)
			} else {
				ev = &model.ResidenceRecordedEvent{
					GeneralEvent:           gev,
					GeneralIndividualEvent: giv,
				}
			}
		case "CENS":
			censusDate, fixed := maybeFixCensusDate(er)
			if fixed {
				gev.Date = censusDate
			}
			ev = l.populateCensusRecord(er, gev, p)

		case "FACT":
			switch strings.ToUpper(er.Type) {
			case "OLB":
				p.Olb = er.Value
			default:
				category := factCategoryForType(er.Type)

				if er.Value != "" {
					p.MiscFacts = append(p.MiscFacts, model.Fact{
						Category:  category,
						Detail:    er.Value,
						Citations: gev.Citations,
					})
				}
				for _, n := range er.Note {
					p.MiscFacts = append(p.MiscFacts, model.Fact{
						Category:  category,
						Detail:    n.Note,
						Citations: l.parseCitationRecords(m, n.Citation),
					})
				}
			}
		case "EVEN":
			if strings.ToUpper(er.Type) == "OLB" && p.Olb == "" {
				p.Olb = er.Value
			} else {
				gev.Title = er.Type
				switch strings.ToLower(er.Type) {
				case "narrative":
					ev = &model.IndividualNarrativeEvent{
						GeneralEvent:           gev,
						GeneralIndividualEvent: giv,
					}
				case "arrival":
					ev = &model.ArrivalEvent{
						GeneralEvent:           gev,
						GeneralIndividualEvent: giv,
					}
				case "departure":
					ev = &model.DepartureEvent{
						GeneralEvent:           gev,
						GeneralIndividualEvent: giv,
					}
				default:
					// Outdoor Relief
					// Criminal Hearing
					// Settlement Examination
					// Court Hearing
					// Departure
					// Arrival
					// Sale of Farm Stock
					// Criminal Trial
					// Discharged from Workhouse
					// Marriage of mother to William Rouse
					// Notice
					// Indentured in Merchant Navy
					// Admitted to Workhouse
					// Absconded from Workhouse
					// Birth of daughter Mary while in workhouse
					// Discharged with her children from Shipmeadow Workhouse
					// Enlisted in Military
					// Military Rank
					// Witness
					// ev = &model.IndividualNarrativeEvent{
					// 	GeneralEvent:           gev,
					// 	GeneralIndividualEvent: giv,
					// }
					ev = &model.IndividualNarrativeEvent{
						GeneralEvent:           gev,
						GeneralIndividualEvent: giv,
					}

					// ev = &model.PlaceholderIndividualEvent{
					// 	GeneralEvent:           gev,
					// 	GeneralIndividualEvent: giv,
					// 	ExtraInfo:              fmt.Sprintf("Unknown individual event (Xref=%s, Tag=%s, Type=%s, Value=%s)", in.Xref, er.Tag, er.Type, er.Value),
					// }
				}
			}
		case "OCCU":
			occupationEvents = append(occupationEvents, gev)
		default:
			ev = &model.PlaceholderIndividualEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
				ExtraInfo:              fmt.Sprintf("Unknown individual event (Xref=%s, Tag=%s, Value=%s)", in.Xref, er.Tag, er.Value),
			}

		}

		if ev != nil {
			p.Timeline = append(p.Timeline, ev)
			if !pl.IsUnknown() {
				pl.Timeline = append(pl.Timeline, ev)
			}
			for _, anom := range anoms {
				anom.Context = "Place in " + ev.Type() + " event"
				if !gdate.IsUnknown(ev.GetDate()) {
					anom.Context += " " + ev.GetDate().Occurrence()
				}
				p.Anomalies = append(p.Anomalies, anom)
			}

		}
	}

	// Try and consolidate occupation events
	if len(occupationEvents) > 0 {
		sort.Slice(occupationEvents, func(i, j int) bool {
			return gdate.SortsBefore(occupationEvents[i].Date, occupationEvents[j].Date)
		})
		for i, gev := range occupationEvents {
			// slog.Debug("occupation event", "detail", gev.Detail)
			// HACK: these are common in my ancestry descriptions
			gev.Detail = strings.TrimPrefix(gev.Detail, "Occupation recorded as ")
			gev.Detail = strings.TrimPrefix(gev.Detail, "Recorded as ")
			gev.Detail = strings.TrimPrefix(gev.Detail, "Noted as ")
			gev.Detail = strings.TrimRight(gev.Detail, ".!, ")

			if i == 0 {
				oc := &model.Occupation{
					StartDate:   gev.Date,
					EndDate:     gev.Date,
					Place:       gev.Place,
					Title:       "Occupation",
					Detail:      gev.Detail,
					Citations:   gev.Citations,
					Occurrences: 1,
				}
				p.Occupations = append(p.Occupations, oc)
			} else {
				// See if we can merge this with the previous occupation
				previous := p.Occupations[len(p.Occupations)-1]
				oc := metrics.NewOverlapCoefficient()
				similarity := strutil.Similarity(gev.Detail, previous.Detail, oc)
				// fmt.Printf("Occupation similarity between %q and %q is %v\n", gev.Detail, previous.Detail, similarity)
				if similarity >= 0.7 {
					// consolidate
					previous.EndDate = gev.Date
					if len(gev.Detail) > len(previous.Detail) {
						previous.Detail = gev.Detail
					}
					previous.Citations = append(previous.Citations, gev.Citations...)
					previous.Occurrences++
				} else {
					oc := &model.Occupation{
						StartDate:   gev.Date,
						EndDate:     gev.Date,
						Place:       gev.Place,
						Title:       "Occupation",
						Detail:      gev.Detail,
						Citations:   gev.Citations,
						Occurrences: 1,
					}
					p.Occupations = append(p.Occupations, oc)
				}

			}
		}
	}

	// Add ancestry links
	if id, ok := l.Attrs["ANCESTRY_TREE_ID"]; ok {
		personID := strings.TrimPrefix(in.Xref, "I")
		// TODO: support other ancestry sites
		p.EditLink = &model.Link{
			Title: "Edit details at ancestry.co.uk",
			URL:   fmt.Sprintf("https://www.ancestry.co.uk/family-tree/person/tree/%s/person/%s/facts", id, personID),
		}
	}

	// Add ancestry tags
	for _, ud := range in.UserDefined {
		switch ud.Tag {
		case "_MTTAG":
			tag, ok := l.Tags[stripXref(ud.Value)]
			if ok {
				p.Tags = append(p.Tags, tag)
			}
		}
	}

	return nil
}

func maybeFixCensusDate(er *gedcom.EventRecord) (gdate.Date, bool) {
	if len(er.Citation) > 0 {
		for _, c := range er.Citation {
			if strings.Contains(c.Page, "Class: HO107") || strings.Contains(c.Page, "Class: HO 107") {
				// 1841 or 1851 census
				if er.Date == "1841" {
					return &gdate.Precise{Y: 1841, M: 6, D: 6}, true
				} else if er.Date == "1851" {
					return &gdate.Precise{Y: 1851, M: 3, D: 30}, true
				}
				return nil, false
			} else if strings.Contains(c.Page, "Class: RG9") || strings.Contains(c.Page, "Class: RG 9") {
				// 1861 census
				return &gdate.Precise{Y: 1861, M: 4, D: 7}, true
			} else if strings.Contains(c.Page, "Class: RG10") || strings.Contains(c.Page, "Class: RG 10") {
				// 1871 census
				return &gdate.Precise{Y: 1871, M: 4, D: 2}, true
			} else if strings.Contains(c.Page, "Class: RG11") || strings.Contains(c.Page, "Class: RG 11") {
				// 1881 census
				return &gdate.Precise{Y: 1881, M: 4, D: 3}, true
			} else if strings.Contains(c.Page, "Class: RG12") || strings.Contains(c.Page, "Class: RG 12") {
				// 1891 census
				return &gdate.Precise{Y: 1891, M: 4, D: 5}, true
			} else if strings.Contains(c.Page, "Class: RG13") || strings.Contains(c.Page, "Class: RG 13") {
				// 1901 census
				return &gdate.Precise{Y: 1901, M: 3, D: 31}, true
			} else if strings.Contains(c.Page, "Class: RG14") || strings.Contains(c.Page, "Class: RG 14") {
				// 1911 census
				return &gdate.Precise{Y: 1911, M: 4, D: 2}, true
			} else if strings.Contains(c.Page, "Class: RG15") || strings.Contains(c.Page, "Class: RG 15") {
				// 1921 census
				return &gdate.Precise{Y: 1921, M: 6, D: 19}, true
			}
		}
	}

	return nil, false
}

func findFirstUserDefinedTag(tag string, uds []gedcom.UserDefinedTag) (gedcom.UserDefinedTag, bool) {
	for _, ud := range uds {
		if ud.Tag == tag {
			return ud, true
		}
	}

	return gedcom.UserDefinedTag{}, false
}

func anyNonEmpty(ss ...string) bool {
	for _, s := range ss {
		if s != "" {
			return true
		}
	}
	return false
}

func factCategoryForType(ty string) string {
	switch strings.ToUpper(ty) {
	case "AKA":
		return model.FactCategoryAKA
	default:
		return ty
	}
}

func (l *Loader) populateCensusRecord(er *gedcom.EventRecord, gev model.GeneralEvent, p *model.Person) *model.CensusEvent {
	// TODO: lookup census event
	ev := &model.CensusEvent{GeneralEvent: gev}
	// CensusEntryRelation

	ce := &model.CensusEntry{
		Principal: p,
	}

	detail := er.Value
	if detail == "" {
		if len(er.Note) > 0 {
			detail = er.Note[0].Note
		}
	}

	fillCensusEntry(detail, ce)

	ev.Entries = append(ev.Entries, ce)
	ev.GeneralEvent.Detail = ce.Detail // TODO: change when census events are shared
	return ev
}

var censusEntryRelationLookup = map[string]model.CensusEntryRelation{
	"head":            model.CensusEntryRelationHead,
	"wife":            model.CensusEntryRelationWife,
	"husband":         model.CensusEntryRelationHusband,
	"son":             model.CensusEntryRelationSon,
	"daughter":        model.CensusEntryRelationDaughter,
	"child":           model.CensusEntryRelationChild,
	"lodger":          model.CensusEntryRelationLodger,
	"boarder":         model.CensusEntryRelationBoarder,
	"inmate":          model.CensusEntryRelationInmate,
	"patient":         model.CensusEntryRelationPatient,
	"servant":         model.CensusEntryRelationServant,
	"nephew":          model.CensusEntryRelationNephew,
	"niece":           model.CensusEntryRelationNiece,
	"brother":         model.CensusEntryRelationBrother,
	"sister":          model.CensusEntryRelationSister,
	"son-in-law":      model.CensusEntryRelationSonInLaw,
	"daughter-in-law": model.CensusEntryRelationDaughterInLaw,
	"father-in-law":   model.CensusEntryRelationFatherInLaw,
	"mother-in-law":   model.CensusEntryRelationMotherInLaw,
	"brother-in-law":  model.CensusEntryRelationBrotherInLaw,
	"sister-in-law":   model.CensusEntryRelationSisterInLaw,
	"grandson":        model.CensusEntryRelationGrandson,
	"granddaughter":   model.CensusEntryRelationGranddaughter,
	"visitor":         model.CensusEntryRelationVisitor,
	"soldier":         model.CensusEntryRelationSoldier,
	"father":          model.CensusEntryRelationFather,
	"mother":          model.CensusEntryRelationMother,
	"uncle":           model.CensusEntryRelationUncle,
	"aunt":            model.CensusEntryRelationAunt,
	"foster child":    model.CensusEntryRelationFosterChild,
}

var censusEntryMaritalStatusLookup = map[string]model.CensusEntryMaritalStatus{
	"married":   model.CensusEntryMaritalStatusMarried,
	"unmarried": model.CensusEntryMaritalStatusUnmarried,
	"single":    model.CensusEntryMaritalStatusUnmarried,
	"divorced":  model.CensusEntryMaritalStatusDivorced,
	"widow":     model.CensusEntryMaritalStatusWidowed,
	"widower":   model.CensusEntryMaritalStatusWidowed,
	"windower":  model.CensusEntryMaritalStatusWidowed,
	"widowed":   model.CensusEntryMaritalStatusWidowed,
}

func fillCensusEntry(v string, ce *model.CensusEntry) {
	v = strings.TrimSpace(v)
	if v == "" {
		return
	}
	reRelationshipToHead := regexp.MustCompile(`(?i)^(.*)\brelation(?:ship)?(?: to head(?: of house)?)?:\s*(.+?(?:-in-law)?)\b[\.,;]*(.*)$`)
	reMaritalStatus := regexp.MustCompile(`(?i)^(.*)\bmarital status:\s*(.+?)\b[\.,;]*(.*)$`)

	matches := reRelationshipToHead.FindStringSubmatch(v)
	if len(matches) > 3 {
		if rel, ok := censusEntryRelationLookup[strings.ToLower(matches[2])]; ok {
			ce.RelationToHead = rel
			v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
		}
	}

	matches = reMaritalStatus.FindStringSubmatch(v)
	if len(matches) > 3 {
		if status, ok := censusEntryMaritalStatusLookup[strings.ToLower(matches[2])]; ok {
			ce.MaritalStatus = status
			v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
		}
	}

	if status, ok := censusEntryMaritalStatusLookup[strings.ToLower(v)]; ok {
		ce.MaritalStatus = status
		v = ""
	}

	// TODO:
	// Relationship to Head: Son. Noted as "idiot"
	// Relation to Head of House: Son. Noted as "imbecile"
	// Relation to Head: Lodger  Occupation: Railway porter
	// Relation to Head: Captain
	// Relation to Head: Mate
	// Relation to Head: Seaman A B

	ce.Detail = v
}

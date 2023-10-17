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
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
)

var (
	reUppercase     = regexp.MustCompile(`^[A-Z \-]{3}[A-Z \-]+$`)
	reParanthesised = regexp.MustCompile(`^\((.+)\)$`)
	reGedcomTag     = regexp.MustCompile(`^[_A-Z][A-Z]+$`)
)

func (l *Loader) populatePersonFacts(m ModelFinder, in *gedcom.IndividualRecord) error {
	p := m.FindPerson(l.ScopeName, in.Xref)

	logger := logging.With("id", p.ID)
	logger.Debug("populating from individual record", "xref", in.Xref)

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
				Category: model.AnomalyCategoryName,
				Text:     fmt.Sprintf("Person has no surname but full name %q contains more than one word.", prefName.Full),
				Context:  "Person's name",
			})
		}

		if prefName.Surname == "" {
			prefName.Surname = model.UnknownNamePlaceholder
			prefName.Full += model.UnknownNamePlaceholder
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: model.AnomalyCategoryName,
				Text:     "Person has no surname, should replace with -?-.",
				Context:  "Person's name",
			})
		}

		if reUppercase.MatchString(prefName.Full) {
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: model.AnomalyCategoryName,
				Text:     "Person's name is all uppercase, should change to proper case.",
				Context:  "Person's name",
			})
		} else if reUppercase.MatchString(prefName.Surname) {
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: model.AnomalyCategoryName,
				Text:     "Person's surname is all uppercase, should change to proper case.",
				Context:  "Person's name",
			})
		}

		if prefName.Given == "" {
			prefName.Given = model.UnknownNamePlaceholder
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: model.AnomalyCategoryName,
				Text:     "Person has no given name, should replace with -?-.",
				Context:  "Person's name",
			})
		}

		prefName.Surname = strings.ReplaceAll(prefName.Surname, "\\", "/")
		prefName.Full = strings.ReplaceAll(prefName.Full, "\\", "/")

		p.PreferredFullName = prefName.Full
		p.PreferredGivenName = prefName.Given

		p.PreferredFamilyName = prefName.Surname
		p.PreferredSortName = prefName.Surname + ", " + prefName.Given
		if prefName.Suffix != "" {
			if matches := reParanthesised.FindStringSubmatch(prefName.Suffix); len(matches) == 2 {
				// suffix is paranthesised which is a convention for prominent tags
				tags := strings.Split(matches[1], ",")
				for _, tag := range tags {
					p.Tags = append(p.Tags, strings.TrimSpace(tag))
				}
				// remove the suffix
				p.PreferredFullName = strings.TrimSpace(p.PreferredFullName[:len(p.PreferredFullName)-len(prefName.Suffix)])
			} else {
				p.PreferredSortName += " " + prefName.Suffix
			}
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

	for _, ud := range in.UserDefined {
		switch ud.Tag {
		case "_MILT": // ancestry military event
			events = append(events, l.parseUserDefinedAsEvent("EVEN", ud))
		case "_EMPLOY": // ancestry employment event
			events = append(events, l.parseUserDefinedAsEvent("EVEN", ud))
		case "_DEST": // ancestry destination event
			events = append(events, l.parseUserDefinedAsEvent("EVEN", ud))
		case "_MILTID": // ancestry military ID fact
			events = append(events, l.parseUserDefinedAsEvent("FACT", ud))
		case "_MDCL": // ancestry medical event
			events = append(events, l.parseUserDefinedAsEvent("EVEN", ud))
		case "MARR": // ancestry individual marriage event
			events = append(events, l.parseUserDefinedAsEvent("EVEN", ud))
		case "ADDR": // ancestry address event
			events = append(events, l.parseUserDefinedAsEvent("EVEN", ud))
		case "_WLNK": // ancestry web link
			// not an event
		case "_MTTAG": // ancestry tag
			// not an event
		default:
			logger.Warn("found user defined tag that might be an event", "xref", in.Xref, "tag", ud.Tag, "value", ud.Value)
		}
	}

	// collect occupation events and attempt to consolidate them later
	occupationEvents := make([]model.GeneralEvent, 0)

	dp := &gdate.Parser{
		AssumeGROQuarter: true,
	}

	for _, er := range events {
		pl, planoms := l.findPlaceForEvent(m, er)

		logger.Debug("found gedcom event", "date", er.Date)

		dt, err := dp.Parse(er.Date)
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
			Date:   &model.Date{Date: dt},
			Place:  pl,
			Detail: detail,
			Title:  er.Tag,
		}

		giv := model.GeneralIndividualEvent{
			Principal: p,
		}

		var citanoms []*model.Anomaly
		if len(er.Citation) > 0 {
			gev.Citations, citanoms = l.parseCitationRecords(m, er.Citation)
		}

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

		case "WILL":
			ev = &model.WillEvent{
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
				logger.Debug("setting OLB from fact", "olb", gev.Detail)
				p.Olb = gev.Detail
			default:
				category, ok := factCategoryForType(er.Type)
				if ok {
					if er.Value != "" {
						p.MiscFacts = append(p.MiscFacts, model.Fact{
							Category:  category,
							Detail:    er.Value,
							Citations: gev.Citations,
						})
					}
					for _, n := range er.Note {
						cits, anoms := l.parseCitationRecords(m, n.Citation)
						p.MiscFacts = append(p.MiscFacts, model.Fact{
							Category:  category,
							Detail:    n.Note,
							Citations: cits,
						})
						p.Anomalies = append(p.Anomalies, anoms...)
					}
				} else {
					logger.Warn("unhandled fact", "xref", in.Xref, "tag", er.Tag, "type", er.Type, "value", er.Value)
				}
			}
		case "EVEN":
			if strings.ToUpper(er.Type) == "OLB" {
				logger.Debug("setting OLB from event", "olb", gev.Detail)
				p.Olb = gev.Detail
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
					if reGedcomTag.MatchString(er.Type) {
						switch er.Type {
						case "MARR": // for ancestry this is a marriage that is not linked to a second person
							p.Anomalies = append(p.Anomalies, &model.Anomaly{
								Category: model.AnomalyCategoryEvent,
								Text:     fmt.Sprintf("The marriage event dated %s is not linked to a second person. Update it to reference the spouse.", er.Date),
								Context:  "Marriage event",
							})
						case "_MILT": // ancestry generic military event
							p.Anomalies = append(p.Anomalies, &model.Anomaly{
								Category: model.AnomalyCategoryEvent,
								Text:     "A generic military event was found, remove it or replace with a descriptive custom event",
								Context:  "Generic military event",
							})
						default:
							logger.Warn("unhandled custom event", "xref", in.Xref, "tag", er.Tag, "type", er.Type, "date", er.Date, "value", er.Value)
						}
						// ev = &model.PlaceholderIndividualEvent{
						// 	GeneralEvent:           gev,
						// 	GeneralIndividualEvent: giv,
						// 	ExtraInfo:              fmt.Sprintf("Unknown custom event (Xref=%s, Tag=%s, Type=%s, Value=%s)", in.Xref, er.Tag, er.Type, er.Value),
						// }
					} else {
						ev = &model.IndividualNarrativeEvent{
							GeneralEvent:           gev,
							GeneralIndividualEvent: giv,
						}
					}
				}
			}
		case "OCCU":
			occupationEvents = append(occupationEvents, gev)
		default:
			logger.Warn("unhandled individual event", "xref", in.Xref, "tag", er.Tag, "type", er.Type, "value", er.Value)
			// ev = &model.PlaceholderIndividualEvent{
			// 	GeneralEvent:           gev,
			// 	GeneralIndividualEvent: giv,
			// 	ExtraInfo:              fmt.Sprintf("Unknown custom event (Xref=%s, Tag=%s, Type=%s, Value=%s)", in.Xref, er.Tag, er.Type, er.Value),
			// }

		}

		if ev != nil {
			logger.Debug("adding event to timeline", "what", ev.What(), "when", ev.When(), "where", ev.Where())

			p.Timeline = append(p.Timeline, ev)
			if !pl.IsUnknown() {
				pl.Timeline = append(pl.Timeline, ev)
			}
			for _, anom := range planoms {
				anom.Context = "Place in " + ev.Type() + " event"
				if !ev.GetDate().IsUnknown() {
					anom.Context += " " + ev.GetDate().When()
				}
				p.Anomalies = append(p.Anomalies, anom)
			}

			seenSource := make(map[*model.Source]bool)
			for _, c := range ev.GetCitations() {
				if c.Source != nil && !seenSource[c.Source] {
					c.Source.EventsCiting = append(c.Source.EventsCiting, ev)
					seenSource[c.Source] = true
				}
			}

			for _, anom := range citanoms {
				anom.Context = "Citation for " + ev.Type() + " event"
				p.Anomalies = append(p.Anomalies, anom)
			}

		}
	}

	// Try and consolidate occupation events
	if len(occupationEvents) > 0 {
		sort.Slice(occupationEvents, func(i, j int) bool {
			return occupationEvents[i].GetDate().SortsBefore(occupationEvents[j].GetDate())
		})
		for i, gev := range occupationEvents {
			// logging.Debug("occupation event", "detail", gev.Detail)
			// HACK: these are common in my ancestry descriptions
			gev.Detail = strings.TrimPrefix(gev.Detail, "Occupation recorded as ")
			gev.Detail = strings.TrimPrefix(gev.Detail, "Recorded as ")
			gev.Detail = strings.TrimPrefix(gev.Detail, "Noted as ")
			gev.Detail = strings.TrimRight(gev.Detail, ".!, ")

			if i == 0 {
				oc := &model.Occupation{
					StartDate:   gev.GetDate(),
					EndDate:     gev.GetDate(),
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
					previous.EndDate = gev.GetDate()
					if len(gev.Detail) > len(previous.Detail) {
						previous.Detail = gev.Detail
					}
					previous.Citations = append(previous.Citations, gev.Citations...)
					previous.Occurrences++
				} else {
					oc := &model.Occupation{
						StartDate:   gev.GetDate(),
						EndDate:     gev.GetDate(),
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

	// Add links to other ancestry trees
	// TODO: add these as a general citation
	// reAncestryTreeID := regexp.MustCompile(`^(\d+):1030:(\d+)$`)
	// for _, cr := range in.Citation {
	// 	if cr.Page == "Ancestry Family Tree" {
	// 		// 1 SOUR @S503070767@
	// 		// 2 PAGE Ancestry Family Tree
	// 		// 2 _APID 29347156557:1030:18090881
	// 		// 2 _APID 332290679880:1030:102833601
	// 		// 2 _APID 202436359813:1030:179850252
	// 		apIDs := findAllUserDefinedTagValues("_APID", cr.UserDefined)
	// 		for _, apID := range apIDs {
	// 			m := reAncestryTreeID.FindStringSubmatch(apID)
	// 			if len(m) > 2 {
	// 				treeID := m[2]
	// 				personID := m[1]
	// 				link := fmt.Sprintf("https://www.ancestry.co.uk/family-tree/person/tree/%s/person/%s/facts", treeID, personID)
	// 				p.Links = append(p.Links, model.Link{
	// 					Title: "Another tree at ancestry.co.uk",
	// 					URL:   link,
	// 				})
	// 			}
	// 			// logger.Debug("foun
	// 		}
	// 	}
	// }

	// Add ancestry tags
	for _, ud := range in.UserDefined {
		switch ud.Tag {
		case "_MTTAG":
			tag, ok := l.Tags[stripXref(ud.Value)]
			if ok {
				switch strings.ToLower(tag) {
				case "illegitimate":
					p.Illegitimate = true
					logger.Debug("found illegitimate tag, marking as illegitimate")
				case "never married":
					p.Unmarried = true
					logger.Debug("found never married tag, marking as unmarried")
				case "no children":
					p.Childless = true
					logger.Debug("found no children tag, marking as childless")
				case "twin":
					p.Twin = true
					logger.Debug("found twin tag, marking as twin")
				case "blind":
					p.Blind = true
					logger.Debug("found blind tag, marking as blind")
				case "deaf":
					p.Deaf = true
					logger.Debug("found deaf tag, marking as deaf")
				case "physically impaired":
					p.PhysicalImpairment = true
					logger.Debug("found physically impaired tag, marking as physically impaired")
				case "mentally impaired":
					p.MentalImpairment = true
					logger.Debug("found mentally impaired tag, marking as mentally impaired")
				case "died in childbirth":
					p.DiedInChildbirth = true
					logger.Debug("found died in childbirth impaired tag, marking as died in childbirth")
				case "transcribe death cert":
					// p.ToDos = append(p.ToDos, &model.ToDo{
					// 	Category: model.ToDoCategoryCitations,
					// 	Context:  "death event",
					// 	Goal:     "Transcribe the death certificate",
					// 	Reason:   "A copy of the certificate is available but it hasn't been transcribed to the source citation.",
					// })
				case "transcribe marriage cert":
					// p.ToDos = append(p.ToDos, &model.ToDo{
					// 	Category: model.ToDoCategoryCitations,
					// 	Context:  "marriage event",
					// 	Goal:     "Transcribe the marriage certificate",
					// 	Reason:   "A copy of the certificate is available but it hasn't been transcribed to the source citation.",
					// })
				case "transcribe birth cert":
					// p.ToDos = append(p.ToDos, &model.ToDo{
					// 	Category: model.ToDoCategoryCitations,
					// 	Context:  "birth event",
					// 	Goal:     "Transcribe the birth certificate",
					// 	Reason:   "A copy of the certificate is available but it hasn't been transcribed to the source citation.",
					// })
				case "transcribe army records":
					// p.ToDos = append(p.ToDos, &model.ToDo{
					// 	Category: model.ToDoCategoryCitations,
					// 	Context:  "army records",
					// 	Goal:     "Transcribe the army records",
					// 	Reason:   "A copy of the army records available but they haven't been transcribed to the source citation.",
					// })
				case "missing birth cert":
					// p.ToDos = append(p.ToDos, &model.ToDo{
					// 	Category: model.ToDoCategoryRecords,
					// 	Context:  "birth event",
					// 	Goal:     "Obtain a copy of the birth certificate",
					// 	Reason:   "The date and place of birth is known and it is within the period of Civil Registration, so a copy of the birth certificate can be requested.",
					// })
				case "missing death cert":
					// p.ToDos = append(p.ToDos, &model.ToDo{
					// 	Category: model.ToDoCategoryRecords,
					// 	Context:  "death event",
					// 	Goal:     "Obtain a copy of the death certificate",
					// 	Reason:   "The date and place of death is known and it is within the period of Civil Registration, so a copy of the death certificate can be requested.",
					// })
				case "missing marriage cert":
					// p.ToDos = append(p.ToDos, &model.ToDo{
					// 	Category: model.ToDoCategoryRecords,
					// 	Context:  "marriage event",
					// 	Goal:     "Obtain a copy of the marriage certificate",
					// 	Reason:   "The date and place of marriage is known and it is within the period of Civil Registration, so a copy of the marriage certificate can be requested.",
					// })
				case "find army records":
					p.ToDos = append(p.ToDos, &model.ToDo{
						Category: model.ToDoCategoryRecords,
						Context:  "military records",
						Goal:     "Obtain a copy of military records",
						Reason:   "This person is believed to have served in the military, so a copy of the records can be requested",
					})
				case "find death date":
				case "find birth date":
				case "find marriage":
				case "find in census":
				case "find surname":
					// handled generically elsewhere
				case "source baptism":
					// handled generically elsewhere
				case "source birth":
					// handled generically elsewhere
				case "source burial":
					// handled generically elsewhere
				case "source death":
					// handled generically elsewhere
				case "source marriage":
					// handled generically elsewhere
				case "transcription needed":
					p.ToDos = append(p.ToDos, &model.ToDo{
						Category: model.ToDoCategoryCitations,
						Context:  "transcribe records",
						Goal:     "transcribe records",
						Reason:   "records are available that have not been transcribed to the source citation",
					})
				case "find other children":
					p.ToDos = append(p.ToDos, &model.ToDo{
						Category: model.ToDoCategoryMissing,
						Context:  "children",
						Goal:     "find other children",
						Reason:   "one or more children are known but there are possibly others that have not been recorded",
					})
				case "actively researching":
					p.Puzzle = true
				case "brick wall":
					p.Puzzle = true
				case "featured":
					p.Featured = true
				case "dna match":
					// This person is on your DNA Match List.
				case "dna connection":
					// This person is a relative on the path between a DNA Match and a common ancestor.
				case "common dna ancestor":
					// This person is a common ancestor between yourself and at least one of your DNA Matches.
				default:
					p.Tags = append(p.Tags, tag)
				}
			}
		}
	}

	return nil
}

func maybeFixCensusDate(er *gedcom.EventRecord) (*model.Date, bool) {
	if len(er.Citation) > 0 {
		for _, c := range er.Citation {
			if strings.Contains(c.Page, "Class: HO107") || strings.Contains(c.Page, "Class: HO 107") {
				// 1841 or 1851 census
				if er.Date == "1841" {
					return model.PreciseDate(1841, 6, 6), true
				} else if er.Date == "1851" {
					return model.PreciseDate(1851, 3, 30), true
				}
				return nil, false
			} else if strings.Contains(c.Page, "Class: RG9") || strings.Contains(c.Page, "Class: RG 9") {
				// 1861 census
				return model.PreciseDate(1861, 4, 7), true
			} else if strings.Contains(c.Page, "Class: RG10") || strings.Contains(c.Page, "Class: RG 10") {
				// 1871 census
				return model.PreciseDate(1871, 4, 2), true
			} else if strings.Contains(c.Page, "Class: RG11") || strings.Contains(c.Page, "Class: RG 11") {
				// 1881 census
				return model.PreciseDate(1881, 4, 3), true
			} else if strings.Contains(c.Page, "Class: RG12") || strings.Contains(c.Page, "Class: RG 12") {
				// 1891 census
				return model.PreciseDate(1891, 4, 5), true
			} else if strings.Contains(c.Page, "Class: RG13") || strings.Contains(c.Page, "Class: RG 13") {
				// 1901 census
				return model.PreciseDate(1901, 3, 31), true
			} else if strings.Contains(c.Page, "Class: RG14") || strings.Contains(c.Page, "Class: RG 14") {
				// 1911 census
				return model.PreciseDate(1911, 4, 2), true
			} else if strings.Contains(c.Page, "Class: RG15") || strings.Contains(c.Page, "Class: RG 15") {
				// 1921 census
				return model.PreciseDate(1921, 6, 19), true
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

func factCategoryForType(ty string) (string, bool) {
	switch strings.ToUpper(ty) {
	case "AKA":
		return model.FactCategoryAKA, true
	case "_MILTID":
		return model.FactCategoryMilitaryServiceNumber, true
	default:
		return "", false
	}
}

var (
	reDetectMentalImpairment   = regexp.MustCompile(`(?i)\b(?:idiot|imbecile)\b`)
	reDetectPhysicalImpairment = regexp.MustCompile(`(?i)\b(?:crippled|disabled)\b`)
	reDetectPauper             = regexp.MustCompile(`(?i)\b(?:pauper)\b`)
)

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
	if ce.RelationToHead != "" {
		logging.Debug("noting relation to head as "+ce.RelationToHead.String()+" from information on census", "id", p.ID)
	}
	if ce.MaritalStatus != "" {
		logging.Debug("noting marital status as "+ce.MaritalStatus.String()+" from information on census", "id", p.ID)
	}
	if ce.Occupation != "" {
		logging.Debug("noting occupation as "+ce.Occupation+" from information on census", "id", p.ID)
	}

	if reDetectMentalImpairment.MatchString(ce.Detail) {
		p.MentalImpairment = true
		logging.Debug("marking as mentally impaired from information on census", "id", p.ID, "detail", ce.Detail)
	}

	if reDetectPhysicalImpairment.MatchString(ce.Detail) {
		p.PhysicalImpairment = true
		logging.Debug("marking as physically impaired from information on census", "id", p.ID, "detail", ce.Detail)
	}

	if reDetectPauper.MatchString(ce.Detail) {
		p.Pauper = true
		logging.Debug("marking as pauper from information on census", "id", p.ID, "detail", ce.Detail)
	}

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

var (
	reRelationshipToHead = regexp.MustCompile(`(?i)^(.*)\brelation(?:ship)?(?: to head(?: of house)?)?:\s*(.+?(?:-in-law)?)\b[\.,;]*(.*)$`)
	reMaritalStatus      = regexp.MustCompile(`(?i)^(.*)\bmarital status:\s*(.+?)\b[\.,;]*(.*)$`)
)

func fillCensusEntry(v string, ce *model.CensusEntry) {
	v = strings.TrimSpace(v)
	if v == "" {
		return
	}
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

// parseUserDefinedAsEvent converts a user defined tag to an event record
//
// Converts something like:
// 1 _MILT
// 2 DATE Aug 1799
// 2 SOUR @S506536161@
// 3 PAGE Piece 2197, 1798 - 1799
// 3 DATA
// 4 DATE 1798 - 1799
// 2 NOTE William was recorded as a sargeant in Capt Lamonte's Company
//
// Into
// 1 EVEN
// 2 TYPE Enlisted in Military
// 2 DATE 28 Jul 1799
// 2 SOUR @S506536161@
// 3 PAGE Piece 2197, 1798 - 1799
// 3 DATA
// 4 DATE 1798 - 1799
func (l *Loader) parseUserDefinedAsEvent(tag string, ud gedcom.UserDefinedTag) *gedcom.EventRecord {
	evr := &gedcom.EventRecord{
		Tag:   tag,
		Type:  ud.Tag,
		Value: ud.Value,
	}

	for _, ud2 := range ud.UserDefined {
		switch ud2.Tag {
		case "DATE":
			evr.Date = ud2.Value
		case "NOTE":
			evr.Note = append(evr.Note, &gedcom.NoteRecord{Note: ud2.Value})
		case "PLAC":
			evr.Place = gedcom.PlaceRecord{Name: ud2.Value}
		case "OBJE":
			mr, ok := l.MediaRecordsByXref[stripXref(ud2.Value)]
			if !ok {
				logging.Warn("unknown media record", "parent_tag", ud.Tag, "tag", ud2.Tag, "xref", ud2.Value)
				continue
			}
			evr.Media = append(evr.Media, mr)
		case "SOUR":
			sr, ok := l.SourceRecordsByXref[stripXref(ud2.Value)]
			if !ok {
				logging.Warn("unknown source record", "parent_tag", ud.Tag, "tag", ud2.Tag, "xref", ud2.Value)
				continue
			}

			cr := &gedcom.CitationRecord{
				Source: sr,
			}
			for _, ud3 := range ud2.UserDefined {
				switch ud3.Tag {
				case "PAGE":
					cr.Page = ud3.Value
				case "_APID":
					cr.UserDefined = append(cr.UserDefined, gedcom.UserDefinedTag{
						Tag:   ud3.Tag,
						Value: ud3.Value,
					})
				case "OBJE":
					mr, ok := l.MediaRecordsByXref[stripXref(ud3.Value)]
					if !ok {
						logging.Warn("unknown media record", "parent_tag", ud2.Tag, "tag", ud3.Tag, "xref", ud3.Value)
						continue
					}
					cr.Media = append(cr.Media, mr)
				case "DATA":
					for _, ud4 := range ud3.UserDefined {
						switch ud4.Tag {
						case "DATE":
							cr.Data.Date = ud4.Value
						case "TEXT":
							cr.Data.Text = append(cr.Data.Text, ud4.Value)
						case "WWW":
							cr.Data.UserDefined = append(cr.Data.UserDefined, gedcom.UserDefinedTag{
								Tag:   ud3.Tag,
								Value: ud3.Value,
							})
						default:
							logging.Warn("unhandled tag when converting user defined to event", "parent_tag", ud3.Tag, "tag", ud4.Tag, "value", ud4.Value)
						}
					}

				default:
					logging.Warn("unhandled tag when converting user defined to event", "parent_tag", ud2.Tag, "tag", ud3.Tag, "value", ud3.Value)
				}
			}

			evr.Citation = append(evr.Citation, cr)
		default:
			logging.Warn("unhandled tag when converting user defined to event", "parent_tag", ud.Tag, "tag", ud2.Tag, "value", ud2.Value)

		}
	}

	return evr
}

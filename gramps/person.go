package gramps

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/grampsxml"
)

var reUppercase = regexp.MustCompile(`^[A-Z \-]{3}[A-Z \-]+$`)

func (l *Loader) populatePersonFacts(m ModelFinder, gp *grampsxml.Person) error {
	id := pval(gp.ID, gp.Handle)
	p := m.FindPerson(l.ScopeName, id)

	if gp.ID != nil {
		p.GrampsID = *gp.ID
	}

	logger := logging.With("id", p.ID)
	logger.Debug("populating from person record", "handle", gp.Handle)

	if len(gp.Name) == 0 {
		p.PreferredFullName = "unknown"
		p.PreferredGivenName = "unknown"
		p.PreferredFamiliarName = "unknown"
		p.PreferredFamilyName = "unknown"
		p.PreferredSortName = "unknown"
		p.PreferredUniqueName = "unknown"
	} else {
		// TODO support multiple names with dates
		var name *grampsxml.Name
		for _, n := range gp.Name {
			if name == nil && !pval(n.Alt, false) {
				name = &n
			}
			oname := &model.Name{
				Name: formatName(n),
			}
			if len(n.Citationref) > 0 {
				oname.Citations, _ = l.parseCitationRecords(m, n.Citationref, logger)
			}

			p.KnownNames = append(p.KnownNames, oname)
		}
		// If none are marked as preferred then choose first
		if name == nil {
			name = &gp.Name[0]
		}

		var prefName struct {
			given   string
			surname string
			suffix  string
			nick    string
			call    string
		}

		if name.First != nil && *name.First != "" {
			prefName.given = *name.First
		} else {
			prefName.given = model.UnknownNamePlaceholder
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: model.AnomalyCategoryName,
				Text:     "Person has no given name, should replace with -?-.",
				Context:  "Person's name",
			})
		}

		switch len(name.Surname) {
		case 0:
			prefName.surname = model.UnknownNamePlaceholder
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: model.AnomalyCategoryName,
				Text:     "Person has no surname, should replace with -?-.",
				Context:  "Person's name",
			})
		case 1:
			prefName.surname = name.Surname[0].Surname
		default:
			return fmt.Errorf("multiple surnames not supported yet (person id: %s)", pval(gp.ID, gp.Handle))
		}

		prefName.given = strings.ReplaceAll(prefName.given, "-?-", model.UnknownNamePlaceholder)
		prefName.surname = strings.ReplaceAll(prefName.surname, "-?-", model.UnknownNamePlaceholder)

		prefName.call = strings.TrimSpace(prefName.given)
		prefName.suffix = strings.TrimSpace(pval(name.Suffix, ""))
		prefName.nick = strings.TrimSpace(pval(name.Nick, ""))
		prefName.call = strings.TrimSpace(pval(name.Call, ""))

		p.PreferredGivenName = prefName.given
		p.PreferredFamilyName = prefName.surname
		p.PreferredSortName = prefName.surname + ", " + prefName.given
		p.PreferredFullName = prefName.given + " " + prefName.surname
		p.PreferredFamiliarName = prefName.call
		p.PreferredFamiliarFullName = prefName.call + " " + prefName.surname
		p.NickName = prefName.nick

		if prefName.suffix != "" {
			p.PreferredFullName += " " + prefName.suffix
			p.PreferredFamiliarFullName += " " + prefName.suffix
		}

		if prefName.suffix != "" {
			p.PreferredSortName += " " + prefName.suffix
		}

		p.PreferredUniqueName = p.PreferredFullName

		if group, ok := l.familyNameGroups[p.PreferredFamilyName]; ok {
			p.FamilyNameGrouping = group
		}

		if reUppercase.MatchString(p.PreferredFullName) {
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: model.AnomalyCategoryName,
				Text:     "Person's name is all uppercase, should change to proper case.",
				Context:  "Person's name",
			})
		} else if reUppercase.MatchString(prefName.surname) {
			p.Anomalies = append(p.Anomalies, &model.Anomaly{
				Category: model.AnomalyCategoryName,
				Text:     "Person's surname is all uppercase, should change to proper case.",
				Context:  "Person's name",
			})
		}

	}

	switch gp.Gender {
	case "M", "m":
		p.Gender = model.GenderMale
	case "F", "f":
		p.Gender = model.GenderFemale
	default:
		p.Gender = model.GenderUnknown
	}

	pgcs, _ := l.parseCitationRecords(m, gp.Citationref, logger)

	for _, pgc := range pgcs {
		if pgc.Source == nil {
			continue
		}
		if strings.HasPrefix(pgc.Detail, "https://www.ancestry.co.uk/family-tree/person/") {
			p.Links = append(p.Links, model.Link{
				Title: "Ancestry",
				URL:   pgc.Detail,
			})
		} else if pgc.Source.Title == "WikiTree" {
			p.Links = append(p.Links, model.Link{
				Title: "WikiTree",
				URL:   "https://www.wikitree.com/wiki/" + pgc.Detail,
			})
		}
	}

	// Add attributes
	for _, att := range gp.Attribute {
		if pval(att.Priv, false) {
			logger.Debug("skipping attribute marked as private", "type", att.Type)
			continue
		}
		switch strings.ToLower(att.Type) {
		case "ancestry url":
			anom := &model.Anomaly{
				Category: model.AnomalyCategoryAttribute,
				Text:     "Person has 'ancestry url' attribute",
				Context:  "Attribute",
			}
			p.Anomalies = append(p.Anomalies, anom)

		case "wikitree id":
			anom := &model.Anomaly{
				Category: model.AnomalyCategoryAttribute,
				Text:     "Person has 'wikitree id' attribute",
				Context:  "Attribute",
			}
			p.Anomalies = append(p.Anomalies, anom)
			p.WikiTreeID = att.Value
			p.Links = append(p.Links, model.Link{
				Title: "WikiTree",
				URL:   "https://www.wikitree.com/wiki/" + att.Value,
			})
		case "illegitimate":
			p.Illegitimate = true
		case "unmarried", "never married":
			p.Unmarried = true
		case "childless", "died without issue":
			p.Childless = true
		case "cause of death":
			codcits, _ := l.parseCitationRecords(m, att.Citationref, logger)
			p.CauseOfDeath = model.ParseCauseOfDeathFact(att.Value, codcits)
		case "mode of death":
			switch strings.ToLower(att.Value) {
			case "suicide":
				p.ModeOfDeath = model.ModeOfDeathSuicide
			case "lost at sea":
				p.ModeOfDeath = model.ModeOfDeathLostAtSea
			case "killed in action":
				p.ModeOfDeath = model.ModeOfDeathKilledInAction
			case "drowned", "drowning":
				p.ModeOfDeath = model.ModeOfDeathDrowned
			case "executed", "execution":
				p.ModeOfDeath = model.ModeOfDeathExecuted
			default:
				logger.Warn("unhandled mode of death attribute", "type", att.Type, "value", att.Value)
			}
		case "military number":
			cits, _ := l.parseCitationRecords(m, att.Citationref, logger)
			p.MiscFacts = append(p.MiscFacts, model.Fact{
				Category:  model.FactCategoryMilitaryServiceNumber,
				Detail:    att.Value,
				Citations: cits,
			})
		case "seamans ticket":
			cits, _ := l.parseCitationRecords(m, att.Citationref, logger)
			p.MiscFacts = append(p.MiscFacts, model.Fact{
				Category:  model.FactCategorySeamansTicket,
				Detail:    att.Value,
				Citations: cits,
			})
		case "slug":
			p.Slug = att.Value
		case "olb":
			p.Olb = att.Value
		default:
			logger.Warn("unhandled person attribute", "type", att.Type, "value", att.Value)
		}
	}

	// Add tags
	for _, tref := range gp.Tagref {
		tag, ok := l.TagsByHandle[tref.Hlink]
		if !ok {
			logger.Warn("could not find tag", "hlink", tref.Hlink)
			continue
		}
		switch strings.ToLower(tag.Name) {
		case "puzzle":
			p.Puzzle = true
		case "featured":
			p.Featured = true
		case "publish":
			p.Publish = true
		}
	}

	// // collect occupation events and attempt to consolidate them later
	// occupationEvents := make([]model.GeneralEvent, 0)

	for _, grer := range gp.Eventref {
		role := pval(grer.Role, "unknown")
		if role != "Primary" {
			continue
		}

		grev, ok := l.EventsByHandle[grer.Hlink]
		if !ok {
			logger.Warn("could not find event", "hlink", grer.Hlink)
			continue
		}

		gev, eventAnomalies, err := l.parseEvent(m, grev, &grer, logger)
		if err != nil {
			logger.Warn("could not parse event", "error", err.Error(), "hlink", grer.Hlink)
			anom := &model.Anomaly{
				Category: model.AnomalyCategoryEvent,
				Text:     err.Error(),
				Context:  "Parsing event data",
			}
			p.Anomalies = append(p.Anomalies, anom)
			continue
		}

		giv := model.GeneralIndividualEvent{
			Principal: p,
		}

		var ev model.TimelineEvent

		switch strings.ToLower(pval(grev.Type, "unknown")) {
		case "birth":
			ev = &model.BirthEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "baptism":
			ev = &model.BaptismEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "death":
			ev = &model.DeathEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "burial":
			ev = &model.BurialEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "cremation":
			ev = &model.CremationEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "memorial":
			ev = &model.MemorialEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "census":
			censusDate, fixed := maybeFixCensusDate(grev)
			if fixed {
				gev.Date = censusDate
			}
			ev = l.populateCensusRecord(grev, &grer, gev, p, m)
		case "residence":
			ev = l.getResidenceEvent(grev, &grer, gev, p, m)
		case "probate":
			ev = &model.ProbateEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "will":
			ev = &model.WillEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "apprentice":
			ev = &model.ApprenticeEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		// case "Property":
		// 	ev = &model.PropertyEvent{
		// 		GeneralEvent:           gev,
		// 		GeneralIndividualEvent: giv,
		// 	}
		case "sale of property":
			ev = &model.SaleOfPropertyEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "economic status":
			if desc := pval(grev.Description, ""); desc != "" {
				gev.Title = "Economic status recorded as " + desc
				ev = &model.EconomicStatusEvent{
					GeneralEvent:           gev,
					GeneralIndividualEvent: giv,
				}
			}

		case "occupation":
			if desc := pval(grev.Description, ""); desc != "" {
				name, status, group := parseOccupation(desc)
				if strings.ToLower(name) == "pauper" {
					p.Anomalies = append(p.Anomalies, &model.Anomaly{
						Category: model.AnomalyCategoryEvent,
						Text:     "Occupation event looks like an economic status: " + name,
						Context:  "Detail",
					})
				}
				oc := &model.Occupation{
					Date:        gev.GetDate(),
					StartDate:   gev.GetDate(),
					EndDate:     gev.GetDate(),
					Place:       gev.Place,
					Name:        name,
					Status:      status,
					Group:       group,
					Detail:      desc,
					Citations:   gev.Citations,
					Occurrences: 1,
				}
				p.Occupations = append(p.Occupations, oc)
				if oc.Name != "" {
					str := ""
					if oc.Status != model.OccupationStatusUnknown {
						str = oc.Status.String() + " "
					}
					str += oc.Name

					gev.Title = "Occupation recorded as " + str
					ev = &model.OccupationEvent{
						GeneralEvent:           gev,
						GeneralIndividualEvent: giv,
					}
				}
			}

			// logger.Debug("found occupation", "what", gev.Detail, "when", gev.When(), "where", gev.Where())
			// occupationEvents = append(occupationEvents, gev)
			// ev = &model.OccupationEvent{
			// 	GeneralEvent:           gev,
			// 	GeneralIndividualEvent: giv,
			// }
		case "narrative":
			ev = &model.IndividualNarrativeEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}

		case "enlistment":
			if _, ok := gev.Attributes[model.EventAttributeRegiment]; !ok {
				if _, ok := gev.Attributes[model.EventAttributeService]; !ok {
					p.Anomalies = append(p.Anomalies, &model.Anomaly{
						Category: model.AnomalyCategoryEvent,
						Text:     "Enlistment event is missing either a regiment or service attribute",
						Context:  "Attributes",
					})
				}
			}

			if reg, ok := gev.Attributes[model.EventAttributeRegiment]; ok {
				gev.Title = "enlisted in the " + reg
			} else if svc, ok := gev.Attributes[model.EventAttributeService]; ok {
				gev.Title = "enlisted in the " + svc
			} else {
				gev.Title = "enlisted"
			}
			ev = &model.EnlistmentEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "promotion":
			if desc := pval(grev.Description, ""); desc != "" {
				gev.Title = "promoted to " + desc
			} else {
				gev.Title = "promoted"
			}
			ev = &model.PromotionEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "demotion":
			if desc := pval(grev.Description, ""); desc != "" {
				gev.Title = "demoted to " + desc
			} else {
				gev.Title = "demoted"
			}
			ev = &model.DemotionEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "muster":
			ev = l.getMusterEvent(grev, &grer, gev, p, m)
			if gev.Title == "" {
				if _, ok := gev.Attributes[model.EventAttributeRegiment]; !ok {
					if _, ok := gev.Attributes[model.EventAttributeService]; !ok {
						p.Anomalies = append(p.Anomalies, &model.Anomaly{
							Category: model.AnomalyCategoryEvent,
							Text:     "Muster event is missing either a regiment or service attribute",
							Context:  "Attributes",
						})
					}
				}
			}
		case "battle":
			ev = l.getBattleEvent(grev, &grer, gev, p, m)
			if gev.Title == "" {
				if _, ok := gev.Attributes[model.EventAttributeRegiment]; !ok {
					if _, ok := gev.Attributes[model.EventAttributeService]; !ok {
						p.Anomalies = append(p.Anomalies, &model.Anomaly{
							Category: model.AnomalyCategoryEvent,
							Text:     "Battle event is missing either a regiment or service attribute",
							Context:  "Attributes",
						})
					}
				}
			}
		case "institution entry":
			ev = &model.InstitutionEntryEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "institution departure":
			ev = &model.InstitutionDepartureEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		default:
			// TODO:
			// Economic Status
			// Settlement Examination
			// Inquest
			// Medical Information
			// Property
			// Arrival
			// Departure
			// Adopted

			logger.Warn("unhandled person event type", "hlink", grer.Hlink, "type", pval(grev.Type, "unknown"))

		}

		if ev != nil {
			logger.Debug("adding event to timeline", "what", ev.What(), "when", ev.When(), "where", ev.Where())

			p.Timeline = append(p.Timeline, ev)
			pl := ev.GetPlace()
			if !pl.IsUnknown() {
				pl.Timeline = append(pl.Timeline, ev)
			}

			seenSource := make(map[*model.Source]bool)
			for _, c := range ev.GetCitations() {
				if c.Source != nil && !seenSource[c.Source] {
					c.Source.EventsCiting = append(c.Source.EventsCiting, ev)
					seenSource[c.Source] = true
				}
			}
			p.Anomalies = append(p.Anomalies, eventAnomalies...)
		}

	}

	// Try and consolidate occupation events
	if len(p.Occupations) > 0 {
		// sort.Slice(occupationEvents, func(i, j int) bool {
		// 	return occupationEvents[i].GetDate().SortsBefore(occupationEvents[j].GetDate())
		// })
		groupCounts := map[model.OccupationGroup]int{}
		for _, oc := range p.Occupations {
			groupCounts[oc.Group]++
		}

		// for _, gev := range occupationEvents {
		// 	title, status, group := parseOccupation(gev.Detail)
		// 	groupCounts[group]++
		// 	oc := &model.Occupation{
		// 		Date:        gev.GetDate(),
		// 		StartDate:   gev.GetDate(),
		// 		EndDate:     gev.GetDate(),
		// 		Place:       gev.Place,
		// 		Name:        title,
		// 		Status:      status,
		// 		Group:       group,
		// 		Detail:      gev.Detail,
		// 		Citations:   gev.Citations,
		// 		Occurrences: 1,
		// 	}
		// 	p.Occupations = append(p.Occupations, oc)
		// }

		bestGroup := model.OccupationGroupUnknown
		bestGroupCount := 0
		for grp, count := range groupCounts {
			if count > bestGroupCount {
				bestGroup = grp
			}
		}
		p.OccupationGroup = bestGroup

	}

	// Add notes
	for _, nr := range gp.Noteref {
		gn, ok := l.NotesByHandle[nr.Hlink]
		if !ok {
			continue
		}
		if pval(gn.Priv, false) {
			logger.Debug("skipping person note marked as private", "handle", gn.Handle)
			continue
		}

		switch strings.ToLower(gn.Type) {
		case "person note":
			p.Comments = append(p.Comments, l.parseNote(gn, m))
		case "research":
			// research notes are always assumed to be markdown
			t := l.parseNote(gn, m)
			t.Markdown = true
			p.ResearchNotes = append(p.ResearchNotes, t)
		default:
			// ignore note
		}
	}

	// Add to families
	for _, co := range gp.Childof {
		gfam, ok := l.FamiliesByHandle[co.Hlink]
		if !ok {
			return fmt.Errorf("person child of unknown family (person id: %s, childof hlink:%s)", pval(gp.ID, gp.Handle), co.Hlink)
		}
		fam := m.FindFamily(l.ScopeName, pval(gfam.ID, gfam.Handle))
		fam.Children = append(fam.Children, p)
	}

	// Add associations
	for _, pr := range gp.Personref {
		ap, ok := l.PeopleByHandle[pr.Hlink]
		if !ok {
			continue
		}
		id := pval(ap.ID, ap.Handle)
		other := m.FindPerson(l.ScopeName, id)

		switch strings.ToLower(pr.Rel) {
		case "twin":
			assoc := model.Association{
				Kind:  model.AssociationKindTwin,
				Other: other,
			}
			if len(pr.Citationref) > 0 {
				assoc.Citations, _ = l.parseCitationRecords(m, pr.Citationref, logger)
			}

			p.Associations = append(p.Associations, assoc)
		default:
			logger.Warn("unhandled person reference relation", "handle", gp.Handle, "rel", pr.Rel)
		}
	}

	// add media objects
	for _, gor := range gp.Objref {
		if pval(gor.Priv, false) {
			logger.Debug("skipping media object marked as private", "handle", gor.Hlink)
			continue
		}
		gob, ok := l.ObjectsByHandle[gor.Hlink]
		if ok {
			mo := m.FindMediaObject(gob.File.Src)

			cmo := &model.CitedMediaObject{
				Object: mo,
			}
			if gor.Region != nil && gor.Region.Corner1x != nil && gor.Region.Corner1y != nil && gor.Region.Corner2x != nil && gor.Region.Corner2y != nil {
				cmo.Highlight = &model.Region{
					Left:   *gor.Region.Corner1x,
					Bottom: 100 - *gor.Region.Corner2y,
					Width:  *gor.Region.Corner2x - *gor.Region.Corner1x,
					Height: *gor.Region.Corner2y - *gor.Region.Corner1y,
				}
			}

			p.Gallery = append(p.Gallery, cmo)
			if p.FeatureImage == nil {
				p.FeatureImage = cmo
			}
		}
	}

	return nil
}

func formatName(n grampsxml.Name) string {
	if n.Display != nil && *n.Display != "" {
		return *n.Display
	}

	name := ""

	if n.First != nil && *n.First != "" {
		name = *n.First
	} else {
		name = model.UnknownNamePlaceholder
	}

	switch len(n.Surname) {
	case 0:
		name += " " + model.UnknownNamePlaceholder
	case 1:
		name += " " + n.Surname[0].Surname
	default:
		name += " " + n.Surname[0].Surname
		// TODO: multiple surnames
	}
	name = strings.ReplaceAll(name, "-?-", model.UnknownNamePlaceholder)

	return name
}

var (
	reApprenticeParan = regexp.MustCompile(`(?i)^(.*)\(apprentice\)(.*)$`)
	reApprentice      = regexp.MustCompile(`(?i)^(.*)\bapprentice\b(.*)$`)
	reJourneymanParan = regexp.MustCompile(`(?i)^(.*)\(journeyman\)(.*)$`)
	reJourneyman      = regexp.MustCompile(`(?i)^(.*)\bjourneyman\b(.*)$`)
	reMasterParan     = regexp.MustCompile(`(?i)^(.*)\(master\)(.*)$`)
	reMaster          = regexp.MustCompile(`(?i)^(.*)\bmaster\b(.*)$`)
	reRetiredParan    = regexp.MustCompile(`(?i)^(.*)\(retired\)(.*)$`)
	reRetired         = regexp.MustCompile(`(?i)^(.*)\bretired\b(.*)$`)
	reDeceasedParan   = regexp.MustCompile(`(?i)^(.*)\(deceased\)(.*)$`)
	reDeceased        = regexp.MustCompile(`(?i)^(.*)\bdeceased\b(.*)$`)
	reUnemployedParan = regexp.MustCompile(`(?i)^(.*)\(unemployed\)(.*)$`)
	reUnemployed      = regexp.MustCompile(`(?i)^(.*)\bunemployed\b(.*)$`)
)

var (
	reGroupLabourer   = regexp.MustCompile(`(?i)\b(labourer|farmer|plate layer|bricklayer|husbandman|hind|gardener|dairyman|shepherd|porter|carter|waggoner|cartman|miller|boatman|excavator|maltster)\b`)
	reGroupIndustrial = regexp.MustCompile(`(?i)\b(miner|pitman|collier|stoker|shipwright|plater|calker|rivetter|blacksmith|engineer|glassman|glassmaker|glass maker|bottle maker)\b`)
	reGroupClerical   = regexp.MustCompile(`(?i)\bclerk|printer|teacher\b`)
	reGroupMilitary   = regexp.MustCompile(`(?i)\b(soldier|sergeant|private|corporal|quartermaster)\b`)
	reGroupPolice     = regexp.MustCompile(`(?i)\b(policeman|police|prison warder)\b`)
	reGroupMaritime   = regexp.MustCompile(`(?i)\b(seaman|sailor|mariner)\b`)
	reGroupCrafts     = regexp.MustCompile(`(?i)\b(baker|shoemaker|carpenter|mason|cobbler|tailor|dressmaker|seamstress|dress maker|lacemaker|lace maker|lace runner|shoe maker|bootmaker|machinist|cordwainer|joiner|glover|butcher)\b`)
	reGroupCommercial = regexp.MustCompile(`(?i)\b(victualer|grocer|publican|dealer|hairdresser)\b`)
	reGroupService    = regexp.MustCompile(`(?i)\b(nurse|servant|valet|housekeeper|charwoman|washerwoman|washer woman|cook|housemaid|maid|milkmaid)\b`)
)

func parseOccupation(s string) (string, model.OccupationStatus, model.OccupationGroup) {
	status := model.OccupationStatusUnknown
	group := model.OccupationGroupUnknown
	s = strings.TrimSpace(s)

	// values enclosed by quotes preserve their case
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		s = strings.Trim(s, `"`)
	} else {
		s = strings.ToLower(s)
	}

	tokenStatuses := []struct {
		re     *regexp.Regexp
		status model.OccupationStatus
	}{
		{re: reJourneymanParan, status: model.OccupationStatusJourneyman},
		{re: reJourneyman, status: model.OccupationStatusJourneyman},
		{re: reMasterParan, status: model.OccupationStatusMaster},
		{re: reMaster, status: model.OccupationStatusMaster},
		{re: reApprenticeParan, status: model.OccupationStatusApprentice},
		{re: reApprentice, status: model.OccupationStatusApprentice},
		{re: reRetiredParan, status: model.OccupationStatusRetired},
		{re: reRetired, status: model.OccupationStatusRetired},
		{re: reDeceasedParan, status: model.OccupationStatusFormer},
		{re: reDeceased, status: model.OccupationStatusFormer},
		{re: reUnemployedParan, status: model.OccupationStatusUnemployed},
		{re: reUnemployed, status: model.OccupationStatusUnemployed},
	}

	for _, st := range tokenStatuses {
		matches := st.re.FindStringSubmatch(s)
		if len(matches) > 2 {
			s = matches[1] + matches[2]
			status = st.status
			break
		}
	}

	kindMatchers := []struct {
		re    *regexp.Regexp
		group model.OccupationGroup
	}{
		{re: reGroupLabourer, group: model.OccupationGroupLabouring},
		{re: reGroupIndustrial, group: model.OccupationGroupIndustrial},
		{re: reGroupMaritime, group: model.OccupationGroupMaritime},
		{re: reGroupCrafts, group: model.OccupationGroupCrafts},
		{re: reGroupClerical, group: model.OccupationGroupClerical},
		{re: reGroupCommercial, group: model.OccupationGroupCommercial},
		{re: reGroupMilitary, group: model.OccupationGroupMilitary},
		{re: reGroupPolice, group: model.OccupationGroupPolice},
		{re: reGroupService, group: model.OccupationGroupService},
	}

	for _, km := range kindMatchers {
		if km.re.MatchString(s) {
			group = km.group
			break
		}
	}

	if group == model.OccupationGroupUnknown {
		logging.Debug("did not match occupation group", "occupation", s)
	}

	s = text.RemoveRedundantWhitespace(s)
	return s, status, group
}

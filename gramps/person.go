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
	p.NativeID = id

	changeTime, err := changeToTime(gp.Change)
	if err == nil {
		p.UpdateTime = &changeTime
	}

	createdTime, err := createdTimeFromHandle(gp.Handle)
	if err == nil {
		p.CreateTime = &createdTime
	}

	if gp.ID != nil {
		p.GrampsID = *gp.ID
	}

	logger := logging.With("source", "person", "id", p.ID, "native_id", id)
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
				oname.Citations = l.parseCitationRecords(m, n.Citationref, logger)
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
			var sb strings.Builder
			for i, s := range name.Surname {
				if i > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(s.Surname)
			}
			prefName.surname = sb.String()
			// return fmt.Errorf("multiple surnames not supported yet (person id: %s)", pval(gp.ID, gp.Handle))
		}

		prefName.given = strings.ReplaceAll(prefName.given, "-?-", model.UnknownNamePlaceholder)
		prefName.surname = strings.ReplaceAll(prefName.surname, "-?-", model.UnknownNamePlaceholder)

		prefName.suffix = strings.TrimSpace(pval(name.Suffix, ""))
		prefName.nick = strings.TrimSpace(pval(name.Nick, ""))
		prefName.call = strings.TrimSpace(pval(name.Call, prefName.given))

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

	pgcs := l.parseCitationRecords(m, gp.Citationref, logger)

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
			anom := &model.Anomaly{
				Category: model.AnomalyCategoryCitation,
				Text:     "Person has 'wikitree' citation",
				Context:  "Citation",
			}
			p.Anomalies = append(p.Anomalies, anom)
			// p.Links = append(p.Links, model.Link{
			// 	Title: "WikiTree",
			// 	URL:   pgc.Detail,
			// })
		} else if strings.HasPrefix(pgc.Detail, "https://www.familysearch.org/tree/person/") {
			anom := &model.Anomaly{
				Category: model.AnomalyCategoryCitation,
				Text:     "Person has 'familysearch' citation",
				Context:  "Citation",
			}
			p.Anomalies = append(p.Anomalies, anom)
			// p.Links = append(p.Links, model.Link{
			// 	Title: "FamilySearch",
			// 	URL:   pgc.Detail,
			// })
		}

	}

	var literacyFact *model.Fact

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
			p.WikiTreeID = att.Value
			p.Links = append(p.Links, model.Link{
				Title: "WikiTree",
				URL:   "https://www.wikitree.com/wiki/" + att.Value,
			})
		case "familysearch id":
			p.FamilySearchID = att.Value
			p.Links = append(p.Links, model.Link{
				Title: "FamilySearch",
				URL:   "https://www.familysearch.org/tree/person/details/" + att.Value,
			})
		case "wikitree category":
			p.WikiTreeCategories = append(p.WikiTreeCategories, att.Value)
		case "illegitimate":
			p.Illegitimate = true
		case "unmarried", "never married":
			p.Unmarried = true
		case "childless", "died without issue":
			p.Childless = true
		case "cause of death":
			codcits := l.parseCitationRecords(m, att.Citationref, logger)
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
			case "childbirth":
				p.ModeOfDeath = model.ModeOfDeathChildbirth
			default:
				logger.Warn("unhandled mode of death attribute", "type", att.Type, "value", att.Value)
			}
		case "military number":
			cits := l.parseCitationRecords(m, att.Citationref, logger)
			p.MiscFacts = append(p.MiscFacts, model.Fact{
				Category:  model.FactCategoryMilitaryServiceNumber,
				Detail:    att.Value,
				Citations: cits,
			})
		case "seamans ticket":
			cits := l.parseCitationRecords(m, att.Citationref, logger)
			p.MiscFacts = append(p.MiscFacts, model.Fact{
				Category:  model.FactCategorySeamansTicket,
				Detail:    att.Value,
				Citations: cits,
			})
		case "slug":
			p.Slug = att.Value
		case "olb":
			p.Olb = att.Value
		case "merged gramps id":
			// ignore
		case "died in childbirth":
			p.ModeOfDeath = model.ModeOfDeathChildbirth
		case "could sign name":
			cits := l.parseCitationRecords(m, att.Citationref, logger)
			const detailCouldSign = "could sign their name"
			const detailCouldNotSign = "could not sign their name"
			const detailBoth = "could sign their name at times"
			switch strings.ToLower(att.Value) {
			case "yes":
				if literacyFact == nil {
					literacyFact = &model.Fact{
						Category:  model.FactCategoryLiteracy,
						Detail:    detailCouldSign,
						Citations: cits,
					}
				} else {
					literacyFact.Citations = append(literacyFact.Citations, cits...)
					if literacyFact.Detail == detailCouldNotSign {
						literacyFact.Detail = detailBoth
					}
				}
			case "no":
				if literacyFact == nil {
					literacyFact = &model.Fact{
						Category:  model.FactCategoryLiteracy,
						Detail:    detailCouldNotSign,
						Citations: cits,
					}
				} else {
					literacyFact.Citations = append(literacyFact.Citations, cits...)
					if literacyFact.Detail == detailCouldSign {
						literacyFact.Detail = detailBoth
					}
				}
			default:
				logger.Error("unsupported value for attribute", "type", att.Type, "value", att.Value)
			}

		default:
			logger.Warn("unhandled person attribute", "type", att.Type, "value", att.Value)
		}
	}

	if literacyFact != nil {
		p.MiscFacts = append(p.MiscFacts, *literacyFact)
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

	for _, grer := range gp.Eventref {
		logger := logger.With("event_handle", grer.Hlink)

		grampsRole := strings.ToLower(pval(grer.Role, "unknown"))
		if grampsRole == "unknown" {
			logger.Warn("anomaly: event has unknown role")
		}

		var roleAttrs map[string]string
		if len(grer.Attribute) > 0 {
			roleAttrs = make(map[string]string, len(grer.Attribute))
			for _, att := range grer.Attribute {
				if pval(att.Priv, false) {
					logger.Debug("skipping event reference attribute marked as private", "type", att.Type)
					continue
				}
				roleAttrs[strings.ToLower(att.Type)] = att.Value
			}
		}

		var ev model.TimelineEvent
		var ok bool
		ev, ok = l.lookupEvent(&grer)

		if ok {
			role := model.EventRoleUnknown
			if grampsRole == "primary" {
				role = model.EventRolePrincipal
			}

			switch tev := ev.(type) {
			case *model.BirthEvent:
				switch grampsRole {
				case "informant":
					role = model.EventRoleInformant
				}
			case *model.BaptismEvent:
				switch grampsRole {
				case "godparent":
					role = model.EventRoleGodparent
				}
			case *model.MarriageEvent:
				switch grampsRole {
				case "witness":
					role = model.EventRoleWitness
				}
			case *model.DeathEvent:
				switch grampsRole {
				case "informant":
					role = model.EventRoleInformant
				}

			case *model.OccupationEvent:
				switch grampsRole {
				case "primary":
					p.Occupations = append(p.Occupations, &tev.Occupation)
				}

			case *model.WillEvent:
				switch grampsRole {
				case "witness":
					role = model.EventRoleWitness
				case "beneficiary":
					role = model.EventRoleBeneficiary
				case "executor":
					role = model.EventRoleExecutor
				}
			case *model.ProbateEvent:
				switch grampsRole {
				case "witness":
					role = model.EventRoleWitness
				case "beneficiary":
					role = model.EventRoleBeneficiary
				case "executor":
					role = model.EventRoleExecutor
				}

			case *model.CensusEvent:
				// prevent event participant being added
				role = model.EventRoleUnknown

				if grampsRole != "primary" {
					logger.Warn("anomaly: census event has an unexpected role, should be primary", "role", grampsRole)
				}
				ce := &model.CensusEntry{
					Principal: p,
				}

				var transcript string
				for _, gnr := range grer.Noteref {
					gn, ok := l.NotesByHandle[gnr.Hlink]
					if !ok {
						continue
					}
					if pval(gn.Priv, false) {
						logging.Debug("skipping census entry note marked as private", "id", p.ID, "handle", gn.Handle)
						continue
					}
					if gn.Type == "Transcript" {
						transcript = gn.Text
					} else if gn.Type == "Narrative" {
						ce.Narrative = gn.Text
					}
				}

				fillCensusEntry(transcript, ce)
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

				tev.Entries = append(tev.Entries, ce)

			}

			if role != model.EventRoleUnknown {
				ev.AddParticipant(&model.EventParticipant{
					Person:     p,
					Role:       role,
					Attributes: roleAttrs,
				})
			} else if grampsRole != "primary" {
				logger.Warn("unhandled person event role", "role", grampsRole, "type", ev.Type())
				// ignore the event
				ev = nil
			}

		} else {
			grev, ok := l.EventsByHandle[grer.Hlink]
			if !ok {
				panic("could not find event with hlink " + grer.Hlink)
			}
			logger.Debug("missing person event", "type", pval(grev.Type, "unknown"))
			continue
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
		case "intro":
			intro := l.parseNote(gn, m)
			p.Intro = &intro
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
				assoc.Citations = l.parseCitationRecords(m, pr.Citationref, logger)
			}

			p.Associations = append(p.Associations, assoc)
		case "dna":
			assoc := model.Association{
				Kind:  model.AssociationKindDNA,
				Other: other,
			}
			if len(pr.Citationref) > 0 {
				assoc.Citations = l.parseCitationRecords(m, pr.Citationref, logger)
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
		for _, s := range n.Surname {
			name += " "
			name += s.Surname
		}
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

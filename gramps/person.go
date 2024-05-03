package gramps

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

var (
	reUppercase     = regexp.MustCompile(`^[A-Z \-]{3}[A-Z \-]+$`)
	reParanthesised = regexp.MustCompile(`^\((.+)\)$`)
)

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

		prefName.call = prefName.given
		prefName.suffix = pval(name.Suffix, "")
		prefName.nick = pval(name.Nick, "")
		prefName.call = pval(name.Call, "")

		p.PreferredGivenName = prefName.given
		p.PreferredFamilyName = prefName.surname
		p.PreferredSortName = prefName.surname + ", " + prefName.given
		p.PreferredFullName = prefName.given + " " + prefName.surname
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
			if p.EditLink == nil {
				p.EditLink = &model.Link{
					Title: "Edit details at ancestry.co.uk",
					URL:   pgc.Detail,
				}
				continue
			}
			p.Links = append(p.Links, model.Link{
				Title: "Family tree at ancestry.co.uk",
				URL:   pgc.Detail,
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

			if p.EditLink == nil {
				p.EditLink = &model.Link{
					Title: "Edit details at ancestry.co.uk",
					URL:   att.Value,
				}
			}
		case "wikitree id":
			anom := &model.Anomaly{
				Category: model.AnomalyCategoryAttribute,
				Text:     "Person has 'wikitree id' attribute",
				Context:  "Attribute",
			}
			p.Anomalies = append(p.Anomalies, anom)
			p.WikiTreeID = att.Value
		case "illegitimate":
			p.Illegitimate = true
		case "unmarried", "never married":
			p.Unmarried = true
		case "childless":
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
			case "drowned":
				p.ModeOfDeath = model.ModeOfDeathDrowned
			case "executed", "execution":
				p.ModeOfDeath = model.ModeOfDeathExecuted
			default:
				logger.Warn("unhandled mode of death attribute", "type", att.Type, "value", att.Value)
			}

		default:
			logger.Warn("unhandled person attribute", "type", att.Type, "value", att.Value)
		}
	}

	// collect occupation events and attempt to consolidate them later
	occupationEvents := make([]model.GeneralEvent, 0)

	for _, er := range gp.Eventref {
		role := pval(er.Role, "unknown")
		if role != "Primary" {
			continue
		}

		grev, ok := l.EventsByHandle[er.Hlink]
		if !ok {
			logger.Warn("could not find event", "hlink", er.Hlink)
			continue
		}

		pl, planoms := l.findPlaceForEvent(m, grev)

		dt, err := EventDate(grev)
		if err != nil {
			logger.Warn("could not parse event date", "error", err.Error(), "hlink", er.Hlink)
			anom := &model.Anomaly{
				Category: model.AnomalyCategoryEvent,
				Text:     err.Error(),
				Context:  "Event date",
			}
			p.Anomalies = append(p.Anomalies, anom)

			continue
		}

		gev := model.GeneralEvent{
			Date:   dt,
			Place:  pl,
			Detail: "", // TODO: notes
			Title:  pval(grev.Description, ""),
		}

		giv := model.GeneralIndividualEvent{
			Principal: p,
		}

		var citanoms []*model.Anomaly
		if len(grev.Citationref) > 0 {
			gev.Citations, citanoms = l.parseCitationRecords(m, grev.Citationref, logger)
		}

		var ev model.TimelineEvent

		switch pval(grev.Type, "unknown") {
		case "Birth":
			ev = &model.BirthEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "Baptism":
			ev = &model.BaptismEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "Death":
			ev = &model.DeathEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "Burial":
			ev = &model.BurialEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "Census":
			censusDate, fixed := maybeFixCensusDate(grev)
			if fixed {
				gev.Date = censusDate
			}
			ev = l.populateCensusRecord(grev, &er, gev, p)
		case "Residence":
			ev = &model.ResidenceRecordedEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "Probate":
			ev = &model.ProbateEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		case "Will":
			ev = &model.WillEvent{
				GeneralEvent:           gev,
				GeneralIndividualEvent: giv,
			}
		// case "Property":
		// 	ev = &model.PropertyEvent{
		// 		GeneralEvent:           gev,
		// 		GeneralIndividualEvent: giv,
		// 	}
		case "Occupation":
			occupationEvents = append(occupationEvents, gev)

		default:
			logger.Warn("unhandled person event type", "hlink", er.Hlink, "type", pval(grev.Type, "unknown"))

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

	// Add notes
	for _, nr := range gp.Noteref {
		n, ok := l.NotesByHandle[nr.Hlink]
		if !ok {
			continue
		}
		if pval(n.Priv, false) {
			logger.Debug("skipping person note marked as private", "handle", n.Handle)
			continue
		}

		switch strings.ToLower(n.Type) {
		case "person note", "research":
			p.ResearchNotes = append(p.ResearchNotes, &model.Note{
				Title:         "",
				Author:        "",
				Date:          "",
				Markdown:      n.Text,
				PrimaryPerson: p,
			})
		default:
			// ignore note
		}

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

	return nil
}

func (l *Loader) findPlaceForEvent(m ModelFinder, er *grampsxml.Event) (*model.Place, []*model.Anomaly) {
	if er.Place == nil {
		return model.UnknownPlace(), nil
	}

	po, ok := l.PlacesByHandle[er.Place.Hlink]
	if !ok {
		return model.UnknownPlace(), nil
	}

	id := pval(po.ID, po.Handle)
	pl := m.FindPlace(l.ScopeName, id)
	return pl, nil
}

func maybeFixCensusDate(grev *grampsxml.Event) (*model.Date, bool) {
	return nil, false
}

func (l *Loader) populateCensusRecord(grev *grampsxml.Event, er *grampsxml.Eventref, gev model.GeneralEvent, p *model.Person) *model.CensusEvent {
	id := pval(grev.ID, grev.Handle)

	ev, ok := l.censusEvents[id]
	if !ok {
		ev = &model.CensusEvent{GeneralEvent: gev}
		l.censusEvents[id] = ev
	}

	// CensusEntryRelation

	ce := &model.CensusEntry{
		Principal: p,
	}

	var transcript string
	for _, gnr := range er.Noteref {
		gn, ok := l.NotesByHandle[gnr.Hlink]
		if !ok {
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

	ev.Entries = append(ev.Entries, ce)
	// ev.GeneralEvent.Detail = ce.Detail // TODO: change when census events are shared
	return ev
}

var censusEntryRelationLookup = map[string]model.CensusEntryRelation{
	"head":            model.CensusEntryRelationHead,
	"wife":            model.CensusEntryRelationWife,
	"husband":         model.CensusEntryRelationHusband,
	"son":             model.CensusEntryRelationSon,
	"dau":             model.CensusEntryRelationDaughter,
	"daur":            model.CensusEntryRelationDaughter,
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
	"m":         model.CensusEntryMaritalStatusMarried,
	"mar":       model.CensusEntryMaritalStatusMarried,
	"married":   model.CensusEntryMaritalStatusMarried,
	"unmarried": model.CensusEntryMaritalStatusUnmarried,
	"u":         model.CensusEntryMaritalStatusUnmarried,
	"unmar":     model.CensusEntryMaritalStatusUnmarried,
	"unm":       model.CensusEntryMaritalStatusUnmarried,
	"single":    model.CensusEntryMaritalStatusUnmarried,
	"divorced":  model.CensusEntryMaritalStatusDivorced,
	"w":         model.CensusEntryMaritalStatusWidowed,
	"widow":     model.CensusEntryMaritalStatusWidowed,
	"wid":       model.CensusEntryMaritalStatusWidowed,
	"widower":   model.CensusEntryMaritalStatusWidowed,
	"windower":  model.CensusEntryMaritalStatusWidowed,
	"widowed":   model.CensusEntryMaritalStatusWidowed,
	"widwr":     model.CensusEntryMaritalStatusWidowed,
}

var (
	reRelationshipToHead       = regexp.MustCompile(`(?i)^(.*)\brelation(?:ship)?(?: to head(?: of house)?)?:\s*(.+?(?:-in-law)?)\b[\.,;]*(.*)$`)
	reMaritalStatus            = regexp.MustCompile(`(?i)^(.*)\b(?:marital status|condition):\s*(.+?)\b[\.,;]*(.*)$`)
	rePlaceOfBirth             = regexp.MustCompile(`(?i)^(.*)\b(?:born|birth|place of birth):\s*(.+?)\b[\.,;]*(.*)$`)
	rePlaceOfBirth2            = regexp.MustCompile(`(?i)^(.*)\b(?:born|birth|place of birth):\s*(.+?)$`)
	reName                     = regexp.MustCompile(`(?i)^(.*)\b(?:name):\s*(.+?)((?:\b[a-zA-Z]+:).*)$`)
	reAge                      = regexp.MustCompile(`(?i)^(.*)\b(?:age):\s*(.+?)\b[\.,;]*(.*)$`)
	reSex                      = regexp.MustCompile(`(?i)^(.*)\b(?:sex|gender):\s*(.+?)\b[\.,;]*(.*)$`)
	reImpairment               = regexp.MustCompile(`(?i)^(.*)\b(?:impairment|disability):\s*(.+?)\b[\.,;]*(.*)$`)
	reOccupation               = regexp.MustCompile(`(?i)^(.*)\b(?:occupation|occ|occ\.):\s*(.+?)\b[\.,;]*(.*)$`)
	reDetectMentalImpairment   = regexp.MustCompile(`(?i)\b(?:idiot|imbecile)\b`)
	reDetectPhysicalImpairment = regexp.MustCompile(`(?i)\b(?:crippled|disabled)\b`)
	reDetectPauper             = regexp.MustCompile(`(?i)\b(?:pauper)\b`)
)

func fillCensusEntry(v string, ce *model.CensusEntry) {
	v = strings.TrimSpace(v)
	if v == "" {
		return
	}

	// Check if this is a multi-line transcription, which we assume to be lines of "key: value"
	if strings.Contains(v, "\n") {
		lines := strings.Split(v, "\n")
		v = ""
		for _, line := range lines {
			field, value, ok := strings.Cut(line, ":")
			if !ok {
				v += line + "\n"
				continue
			}

			switch strings.ToLower(strings.TrimSpace(field)) {
			case "name":
				ce.Name = strings.TrimSpace(value)
			case "relation", "relationship", "relation to head", "relationship to head":
				if rel, ok := censusEntryRelationLookup[strings.ToLower(strings.TrimSpace(value))]; ok {
					ce.RelationToHead = rel
				} else {
					v += line + "\n"
				}

			case "condition", "marital status":
				if status, ok := censusEntryMaritalStatusLookup[strings.ToLower(strings.TrimSpace(value))]; ok {
					ce.MaritalStatus = status
				} else {
					v += line + "\n"
				}

			case "age":
				ce.Age = strings.TrimSpace(value)
			case "sex", "gender":
				ce.Sex = strings.TrimSpace(value)
			case "born", "birth", "place of birth":
				ce.PlaceOfBirth = strings.TrimSpace(value)
			case "impairment", "disability":
				ce.Impairment = strings.TrimSpace(value)
			case "occupation", "occ", "occ.", "occup", "occup.":
				ce.Occupation = strings.TrimSpace(value)
			default:
				v += line + "\n"
			}

		}

	} else {
		// Parse single line

		matches := reRelationshipToHead.FindStringSubmatch(v)
		if len(matches) > 3 {
			if rel, ok := censusEntryRelationLookup[strings.ToLower(matches[2])]; ok {
				ce.RelationToHead = rel
				v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
			}
		}

		matches = reMaritalStatus.FindStringSubmatch(v)
		if len(matches) > 3 {
			logging.Warn("XXXXXXXXXXX found marital status", "m1", matches[1], "m2", matches[2], "m3", matches[3])
			if status, ok := censusEntryMaritalStatusLookup[strings.ToLower(matches[2])]; ok {
				ce.MaritalStatus = status
				v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
			}
		}

		if status, ok := censusEntryMaritalStatusLookup[strings.ToLower(v)]; ok {
			ce.MaritalStatus = status
			v = ""
		}

		matches = reAge.FindStringSubmatch(v)
		if len(matches) > 3 {
			ce.Age = strings.TrimSpace(matches[2])
			v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
		}

		matches = reName.FindStringSubmatch(v)
		if len(matches) > 3 {
			ce.Name = strings.TrimSpace(matches[2])
			v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
		}

		matches = reSex.FindStringSubmatch(v)
		if len(matches) > 3 {
			ce.Sex = strings.TrimSpace(matches[2])
			v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
		}

		matches = reOccupation.FindStringSubmatch(v)
		if len(matches) > 3 {
			ce.Occupation = strings.TrimSpace(matches[2])
			v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
		}

		matches = reImpairment.FindStringSubmatch(v)
		if len(matches) > 3 {
			ce.Impairment = strings.TrimSpace(matches[2])
			v = strings.TrimRight(strings.TrimSpace(matches[1]+matches[3]), ",;")
		}

		matches = rePlaceOfBirth2.FindStringSubmatch(v)
		if len(matches) > 2 {
			ce.PlaceOfBirth = strings.TrimSpace(matches[2])
			v = strings.TrimRight(strings.TrimSpace(matches[1]), ",;")
		}
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

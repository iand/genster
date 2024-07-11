package gramps

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/iand/gdate"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/grampsxml"
)

func (l *Loader) parseEvent(m ModelFinder, grev *grampsxml.Event, grer *grampsxml.Eventref, logger *slog.Logger) (model.GeneralEvent, []*model.Anomaly, error) {
	pl := l.findPlaceForEvent(m, grev)

	dt, err := EventDate(grev)
	if err != nil {
		return model.GeneralEvent{}, nil, err
	}

	gev := model.GeneralEvent{
		Date:       dt,
		Place:      pl,
		Detail:     "",
		Title:      pval(grev.Type, ""),
		Attributes: make(map[string]string),
	}

	// add shared attributes
	for _, att := range grev.Attribute {
		if pval(att.Priv, false) {
			logger.Debug("skipping event attribute marked as private", "type", att.Type)
			continue
		}
		gev.Attributes[strings.ToLower(att.Type)] = att.Value
	}

	// add attributes for this reference
	for _, att := range grer.Attribute {
		if pval(att.Priv, false) {
			logger.Debug("skipping event reference attribute marked as private", "type", att.Type)
			continue
		}
		gev.Attributes[strings.ToLower(att.Type)] = att.Value
	}

	var anomalies []*model.Anomaly
	if len(grev.Citationref) > 0 {
		var citanoms []*model.Anomaly
		gev.Citations, citanoms = l.parseCitationRecords(m, grev.Citationref, logger)
		for _, anom := range citanoms {
			anom.Context = "Citation for " + pval(grev.Type, "unknown") + " event"
			anomalies = append(anomalies, anom)
		}
	}

	for _, gnr := range grev.Noteref {
		gn, ok := l.NotesByHandle[gnr.Hlink]
		if !ok {
			continue
		}
		if pval(gn.Priv, false) {
			logger.Debug("skipping event note marked as private", "handle", gn.Handle)
			continue
		}
		switch strings.ToLower(gn.Type) {
		case "narrative":
			if gev.Narrative.Text != "" {
				logger.Warn("overwriting narrative with Narrative note", "hlink", gnr.Hlink)
			}
			gev.Narrative = l.parseNote(gn, m)
		case "event note":
			cit := &model.GeneralCitation{
				ID:     gnr.Hlink,
				Detail: gn.Text,
			}
			gev.Citations = append(gev.Citations, cit)
		}
	}
	return gev, anomalies, nil
}

func (l *Loader) findPlaceForEvent(m ModelFinder, grev *grampsxml.Event) *model.Place {
	if grev.Place == nil {
		return model.UnknownPlace()
	}

	po, ok := l.PlacesByHandle[grev.Place.Hlink]
	if !ok {
		return model.UnknownPlace()
	}

	id := pval(po.ID, po.Handle)
	pl := m.FindPlace(l.ScopeName, id)
	return pl
}

func maybeFixCensusDate(grev *grampsxml.Event) (*model.Date, bool) {
	return nil, false
}

func (l *Loader) populateCensusRecord(grev *grampsxml.Event, er *grampsxml.Eventref, gev model.GeneralEvent, p *model.Person, m ModelFinder) *model.CensusEvent {
	id := pval(grev.ID, grev.Handle)

	ev, ok := l.censusEvents[id]
	if !ok {
		ev = &model.CensusEvent{GeneralEvent: gev}
		l.censusEvents[id] = ev

		for _, gnr := range grev.Noteref {
			gn, ok := l.NotesByHandle[gnr.Hlink]
			if !ok {
				continue
			}
			if pval(gn.Priv, false) {
				logging.Debug("skipping census note marked as private", "id", p.ID, "handle", gn.Handle)
				continue
			}
			if gn.Type == "Narrative" {
				ev.Narrative = l.parseNote(gn, m)
			}
		}

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
	rePlaceOfBirth             = regexp.MustCompile(`(?i)^(.*)\b(?:born|birth|place of birth):\s*(.+?)$`)
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

		matches = rePlaceOfBirth.FindStringSubmatch(v)
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

func EventDate(grev *grampsxml.Event) (*model.Date, error) {
	dp := &gdate.Parser{
		AssumeGROQuarter: false,
	}

	if grev.Dateval != nil {
		dt, err := ParseDateval(*grev.Dateval)
		if err != nil {
			return nil, fmt.Errorf("parse date value: %w", err)
		}
		return dt, nil
	} else if grev.Daterange != nil {
		dt, err := ParseDaterange(*grev.Daterange)
		if err != nil {
			return nil, fmt.Errorf("parse date range: %w", err)
		}
		return dt, nil
	} else if grev.Datespan != nil {
		dt, err := ParseDatespan(*grev.Datespan)
		if err != nil {
			return nil, fmt.Errorf("parse date span: %w", err)
		}
		return dt, nil
	} else if grev.Datestr != nil {
		dt, err := dp.Parse(grev.Datestr.Val)
		if err != nil {
			return nil, fmt.Errorf("parse date value: %w", err)
		}
		return &model.Date{Date: dt}, nil
	}
	return model.UnknownDate(), nil
}

func ParseDateval(dv grampsxml.Dateval) (*model.Date, error) {
	dp := &gdate.Parser{
		ReckoningLocation: gdate.EnglandAndWales,
	}

	if dv.Cformat != nil {
		switch strings.ToLower(*dv.Cformat) {
		case "gregorian":
			dp.ReckoningLocation = gdate.ReckoningLocationNone
			dp.Calendar = gdate.Gregorian
		case "julian":
			dp.ReckoningLocation = gdate.ReckoningLocationNone
			dp.Calendar = gdate.Julian25Mar
		default:
			return nil, fmt.Errorf("date Cformat not supported")
		}
	}
	if dv.Dualdated != nil {
		return nil, fmt.Errorf("date Dualdated not supported")
	}
	if dv.Newyear != nil {
		return nil, fmt.Errorf("date Newyear not supported")
	}

	// Quality:
	// - Regular
	// - Estimated
	// - Calculated

	var deriv model.DateDerivation
	if dv.Quality != nil {
		switch strings.ToLower(*dv.Quality) {
		case "regular":
			deriv = model.DateDerivationStandard
		case "estimated":
			deriv = model.DateDerivationEstimated
		case "calculated":
			deriv = model.DateDerivationCalculated
		default:
			return nil, fmt.Errorf("quality value %q not supported", *dv.Quality)

		}
	}

	dt, err := dp.Parse(dv.Val)
	if err != nil {
		return nil, fmt.Errorf("parse date value: %w", err)
	}

	dateType := pval(dv.Type, "Regular")
	switch strings.ToLower(dateType) {
	case "before":
		dyear, ok := dt.(*gdate.Year)
		if !ok {
			return nil, fmt.Errorf("'before' type not supported for dates other than years")
		}
		dt = &gdate.BeforeYear{
			C: dyear.C,
			Y: dyear.Y,
		}
	case "after":
		dyear, ok := dt.(*gdate.Year)
		if !ok {
			return nil, fmt.Errorf("'after' type not supported for dates other than years")
		}
		dt = &gdate.AfterYear{
			C: dyear.C,
			Y: dyear.Y,
		}
	case "about":
		dyear, ok := dt.(*gdate.Year)
		if !ok {
			return nil, fmt.Errorf("'about' type not supported for dates other than years")
		}
		dt = &gdate.AboutYear{
			C: dyear.C,
			Y: dyear.Y,
		}
	case "regular":
		break
	default:
		return nil, fmt.Errorf("unexpected date type: %s", dateType)
	}

	return &model.Date{
		Date:       dt,
		Derivation: deriv,
	}, nil
}

func ParseDaterange(dr grampsxml.Daterange) (*model.Date, error) {
	if dr.Cformat != nil {
		return nil, fmt.Errorf("date Cformat not supported")
	}
	if dr.Dualdated != nil {
		return nil, fmt.Errorf("date Dualdated not supported")
	}
	if dr.Newyear != nil {
		return nil, fmt.Errorf("date Newyear not supported")
	}

	// Quality:
	// - Regular
	// - Estimated
	// - Calculated
	var deriv model.DateDerivation
	if dr.Quality != nil {
		switch strings.ToLower(*dr.Quality) {
		case "regular":
			deriv = model.DateDerivationStandard
		case "estimated":
			deriv = model.DateDerivationEstimated
		case "calculated":
			deriv = model.DateDerivationCalculated
		default:
			return nil, fmt.Errorf("quality value %q not supported", *dr.Quality)

		}
	}

	// Currently only support quarter ranges
	dp := &gdate.Parser{}

	dstart, err := dp.Parse(dr.Start)
	if err != nil {
		return nil, fmt.Errorf("parse start value %q: %w", dr.Start, err)
	}

	var mystart *gdate.MonthYear
	switch tstart := dstart.(type) {
	case *gdate.MonthYear:
		mystart = tstart
	case *gdate.Precise:
		if tstart.D == 1 {
			mystart = &gdate.MonthYear{
				C: tstart.C,
				M: tstart.M,
				Y: tstart.Y,
			}
		}
	}

	if mystart == nil {
		return nil, fmt.Errorf("parse start value %q: unsupported start date type", dr.Start)
	}

	dstop, err := dp.Parse(dr.Stop)
	if err != nil {
		return nil, fmt.Errorf("parse stop value %q: %w", dr.Stop, err)
	}
	var mystop *gdate.MonthYear
	switch tstop := dstop.(type) {
	case *gdate.MonthYear:
		mystop = tstop
	case *gdate.Precise:
		switch tstop.M {
		case 1, 3, 5, 7, 8, 10, 12:
			if tstop.D == 31 {
				mystop = &gdate.MonthYear{
					C: tstop.C,
					M: tstop.M,
					Y: tstop.Y,
				}
			}
		case 4, 6, 9, 11:
			if tstop.D == 30 {
				mystop = &gdate.MonthYear{
					C: tstop.C,
					M: tstop.M,
					Y: tstop.Y,
				}
			}
		case 2:
			if tstop.Y%4 == 0 && (tstop.Y%100 != 0 || tstop.Y%400 == 0) {
				if tstop.D == 29 {
					mystop = &gdate.MonthYear{
						C: tstop.C,
						M: tstop.M,
						Y: tstop.Y,
					}
				}
			} else {
				if tstop.D == 28 {
					mystop = &gdate.MonthYear{
						C: tstop.C,
						M: tstop.M,
						Y: tstop.Y,
					}
				}
			}
		}
	}

	if mystop == nil {
		return nil, fmt.Errorf("parse stop value %q: unsupported stop date type", dr.Stop)
	}

	if mystart.Y == mystop.Y {
		if mystart.M == 1 && mystop.M == 3 {
			return &model.Date{
				Date: &gdate.YearQuarter{
					C: mystart.C,
					Y: mystart.Y,
					Q: 1,
				},
				Derivation: deriv,
			}, nil
		} else if mystart.M == 4 && mystop.M == 6 {
			return &model.Date{
				Date: &gdate.YearQuarter{
					C: mystart.C,
					Y: mystart.Y,
					Q: 2,
				},
				Derivation: deriv,
			}, nil
		} else if mystart.M == 7 && mystop.M == 9 {
			return &model.Date{
				Date: &gdate.YearQuarter{
					C: mystart.C,
					Y: mystart.Y,
					Q: 3,
				},
				Derivation: deriv,
			}, nil
		} else if mystart.M == 10 && mystop.M == 12 {
			return &model.Date{
				Date: &gdate.YearQuarter{
					C: mystart.C,
					Y: mystart.Y,
					Q: 4,
				},
				Derivation: deriv,
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported range")
}

func ParseDatespan(ds grampsxml.Datespan) (*model.Date, error) {
	return nil, fmt.Errorf("unsupported span")
}

func (l *Loader) getResidenceEvent(grev *grampsxml.Event, er *grampsxml.Eventref, gev model.GeneralEvent, p *model.Person, m ModelFinder) *model.ResidenceRecordedEvent {
	id := pval(grev.ID, grev.Handle)

	var ev *model.ResidenceRecordedEvent

	mev, ok := l.multipartyEvents[id]
	if ok {
		ev, ok = mev.(*model.ResidenceRecordedEvent)
		if !ok {
			panic(fmt.Sprintf("expected multiparty event with id %q to be a ResidenceRecordedEvent but it was a %T", id, mev))
		}
	} else {
		ev = &model.ResidenceRecordedEvent{GeneralEvent: gev}
		l.multipartyEvents[id] = ev

		for _, gnr := range grev.Noteref {
			gn, ok := l.NotesByHandle[gnr.Hlink]
			if !ok {
				continue
			}
			if pval(gn.Priv, false) {
				logging.Debug("skipping residence note marked as private", "id", p.ID, "handle", gn.Handle)
				continue
			}
			if gn.Type == "Narrative" {
				ev.Narrative = l.parseNote(gn, m)
			}
		}

	}

	ev.Participants = append(ev.Participants, &model.EventParticipant{
		Person: p,
		Role:   model.EventRolePrincipal,
	})

	return ev
}

func (l *Loader) getMusterEvent(grev *grampsxml.Event, er *grampsxml.Eventref, gev model.GeneralEvent, p *model.Person, m ModelFinder) *model.MusterEvent {
	id := pval(grev.ID, grev.Handle)

	var ev *model.MusterEvent

	mev, ok := l.multipartyEvents[id]
	if ok {
		ev, ok = mev.(*model.MusterEvent)
		if !ok {
			panic(fmt.Sprintf("expected multiparty event with id %q to be a MusterEvent but it was a %T", id, mev))
		}
	} else {
		ev = &model.MusterEvent{GeneralEvent: gev}
		l.multipartyEvents[id] = ev

		for _, gnr := range grev.Noteref {
			gn, ok := l.NotesByHandle[gnr.Hlink]
			if !ok {
				continue
			}
			if pval(gn.Priv, false) {
				logging.Debug("skipping muster note marked as private", "id", p.ID, "handle", gn.Handle)
				continue
			}
			if gn.Type == "Narrative" {
				ev.Narrative = l.parseNote(gn, m)
			}
		}

		title := text.JoinSentenceParts("recorded at muster")
		if regiment, ok := ev.GetAttribute(model.EventAttributeRegiment); ok {
			if battalion, ok := ev.GetAttribute(model.EventAttributeBattalion); ok {
				title = text.JoinSentenceParts(title, "for the", battalion, "battalion,", regiment)
			} else {
				title = text.JoinSentenceParts(title, "for the", regiment, "regiment")
			}
		}

		ev.Title = title

	}

	ev.Participants = append(ev.Participants, &model.EventParticipant{
		Person: p,
		Role:   model.EventRolePrincipal,
	})

	return ev
}

func (l *Loader) getBattleEvent(grev *grampsxml.Event, er *grampsxml.Eventref, gev model.GeneralEvent, p *model.Person, m ModelFinder) *model.BattleEvent {
	id := pval(grev.ID, grev.Handle)

	var ev *model.BattleEvent

	mev, ok := l.multipartyEvents[id]
	if ok {
		ev, ok = mev.(*model.BattleEvent)
		if !ok {
			panic(fmt.Sprintf("expected multiparty event with id %q to be a BattleEvent but it was a %T", id, mev))
		}
	} else {
		ev = &model.BattleEvent{GeneralEvent: gev}
		l.multipartyEvents[id] = ev

		if desc := pval(grev.Description, ""); desc != "" {
			ev.Title = "participated in the " + desc
		} else {
			ev.Title = "participated in battle"
		}

		for _, gnr := range grev.Noteref {
			gn, ok := l.NotesByHandle[gnr.Hlink]
			if !ok {
				continue
			}
			if pval(gn.Priv, false) {
				logging.Debug("skipping battle note marked as private", "id", p.ID, "handle", gn.Handle)
				continue
			}
			if gn.Type == "Narrative" {
				ev.Narrative = l.parseNote(gn, m)
			}
		}

		// title := text.JoinSentenceParts("participated in battle")
		// if regiment, ok := ev.GetAttribute(model.EventAttributeRegiment); ok {
		// 	if battalion, ok := ev.GetAttribute(model.EventAttributeBattalion); ok {
		// 		title = text.JoinSentenceParts(title, "in the", battalion, "battalion,", regiment)
		// 	} else {
		// 		title = text.JoinSentenceParts(title, "in the", regiment, "regiment")
		// 	}
		// }

		// ev.Title = title

	}

	ev.Participants = append(ev.Participants, &model.EventParticipant{
		Person: p,
		Role:   model.EventRolePrincipal,
	})

	return ev
}

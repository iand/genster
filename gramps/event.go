package gramps

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/iand/gdate"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/grampsxml"
)

func (l *Loader) populateEventFacts(m ModelFinder, grev *grampsxml.Event) error {
	id := pval(grev.ID, grev.Handle)
	logger := logging.With("id", id)
	logger.Debug("populating from event record", "handle", grev.Handle)

	pl := l.findPlaceForEvent(m, grev)
	dp := gdate.Parser{
		ReckoningLocation: reckoningForPlace(pl),
		AssumeGROQuarter:  false,
	}

	dt, err := EventDate(grev, dp)
	if err != nil {
		logger.Warn("unable to parse event date", "handle", grev.Handle, "error", err)
		dt = model.UnknownDate()
	}

	gev := model.GeneralEvent{
		Date:       dt,
		Place:      pl,
		Detail:     pval(grev.Description, ""),
		Title:      pval(grev.Type, ""),
		Attributes: make(map[string]string),
	}

	changeTime, err := changeToTime(grev.Change)
	if err == nil {
		gev.UpdateTime = &changeTime
	}

	createdTime, err := createdTimeFromHandle(grev.Handle)
	if err == nil {
		gev.CreateTime = &createdTime
	}

	// add shared attributes
	for _, att := range grev.Attribute {
		if pval(att.Priv, false) {
			logger.Debug("skipping event attribute marked as private", "type", att.Type)
			continue
		}
		gev.Attributes[strings.ToLower(att.Type)] = att.Value
	}

	if len(grev.Citationref) > 0 {
		gev.Citations = l.parseCitationRecords(m, grev.Citationref, logger)
	}

	for _, gor := range grev.Objref {
		if pval(gor.Priv, false) {
			logger.Debug("skipping citation object marked as private", "handle", gor.Hlink)
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

			gev.MediaObjects = append(gev.MediaObjects, cmo)
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

	var ev model.TimelineEvent
	evtype := strings.ToLower(pval(grev.Type, "unknown"))
	switch evtype {
	case "birth":
		ev = &model.BirthEvent{
			GeneralEvent: gev,
		}
	case "baptism":
		bev := &model.BaptismEvent{
			GeneralEvent: gev,
		}
		if _, ok := gev.Attributes[model.EventAttributePrivateBaptism]; ok {
			bev.Private = true
		}
		ev = bev
	case "received":
		ev = &model.ReceivedIntoCommunionEvent{
			GeneralEvent: gev,
		}
	case "naming":
		ev = &model.NamingEvent{
			GeneralEvent: gev,
		}
	case "death":
		ev = &model.DeathEvent{
			GeneralEvent: gev,
		}
	case "burial":
		ev = &model.BurialEvent{
			GeneralEvent: gev,
		}
	case "cremation":
		ev = &model.CremationEvent{
			GeneralEvent: gev,
		}
	case "memorial":
		ev = &model.MemorialEvent{
			GeneralEvent: gev,
		}
	case "inquest":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.InquestEvent{
			GeneralEvent: gev,
		}
	case "will":
		ev = &model.WillEvent{
			GeneralEvent: gev,
		}
	case "probate":
		ev = &model.ProbateEvent{
			GeneralEvent: gev,
		}
	case "apprentice":
		ev = &model.ApprenticeEvent{
			GeneralEvent: gev,
		}
	case "physical description":
		if desc := pval(grev.Description, ""); desc != "" {
			ev = &model.PhysicalDescriptionEvent{
				GeneralEvent: gev,
				Description:  desc,
			}
		}
	case "sale of property":
		ev = &model.SaleOfPropertyEvent{
			GeneralEvent: gev,
		}
	case "economic status":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = "Economic status recorded as " + desc
			ev = &model.EconomicStatusEvent{
				GeneralEvent: gev,
				Status:       desc,
			}
		}
	case "narrative":
		ev = &model.IndividualNarrativeEvent{
			GeneralEvent: gev,
		}
	case "institution entry":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.InstitutionEntryEvent{
			GeneralEvent: gev,
		}
	case "institution departure":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.InstitutionDepartureEvent{
			GeneralEvent: gev,
		}
	case "court":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.CourtEvent{
			GeneralEvent: gev,
		}
	case "conviction":
		if desc := pval(grev.Description, ""); desc != "" {
			ev = &model.ConvictionEvent{
				GeneralEvent: gev,
				Crime:        desc,
			}
		}

	case "immigration":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.ImmigrationEvent{
			GeneralEvent: gev,
		}
	case "departure":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.DepartureEvent{
			GeneralEvent: gev,
		}
	case "arrival":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.ArrivalEvent{
			GeneralEvent: gev,
		}
	case "possible birth":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.PossibleBirthEvent{
			GeneralEvent: gev,
		}
	case "possible death":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = desc
		}
		ev = &model.PossibleDeathEvent{
			GeneralEvent: gev,
		}

	case "enlistment":
		if _, ok := gev.Attributes[model.EventAttributeRegiment]; !ok {
			if _, ok := gev.Attributes[model.EventAttributeService]; !ok {
				logger.Warn("anomaly: enlistment event is missing either a regiment or service attribute")
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
			GeneralEvent: gev,
		}
	case "promotion":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = "promoted to " + desc
		} else {
			gev.Title = "promoted"
		}
		ev = &model.PromotionEvent{
			GeneralEvent: gev,
		}
	case "demotion":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = "demoted to " + desc
		} else {
			gev.Title = "demoted"
		}
		ev = &model.DemotionEvent{
			GeneralEvent: gev,
		}
	case "marriage":
		ev = &model.MarriageEvent{
			GeneralEvent: gev,
		}
	case "marriage license":
		ev = &model.MarriageLicenseEvent{
			GeneralEvent: gev,
		}
	case "marriage banns":
		ev = &model.MarriageBannsEvent{
			GeneralEvent: gev,
		}
	case "divorce":
		ev = &model.DivorceEvent{
			GeneralEvent: gev,
		}
	case "separation":
		ev = &model.SeparationEvent{
			GeneralEvent: gev,
		}
	case "residence":
		ev = &model.ResidenceRecordedEvent{
			GeneralEvent: gev,
		}
	case "census":
		censusDate, fixed := maybeFixCensusDate(grev)
		if fixed {
			gev.Date = censusDate
		}
		ev = &model.CensusEvent{
			GeneralEvent: gev,
		}

	case "occupation":

		if desc := pval(grev.Description, ""); desc != "" {
			name, status, group := parseOccupation(desc)
			if strings.ToLower(name) == "pauper" {
				logger.Warn("anomaly: occupation event looks like an economic status: " + name)
			}
			oc := model.Occupation{
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
			if oc.Name != "" {
				str := ""
				if oc.Status != model.OccupationStatusUnknown {
					str = oc.Status.String() + " "
				}
				str += oc.Name

				gev.Title = "Occupation recorded as " + str
				ev = &model.OccupationEvent{
					GeneralEvent: gev,
					Occupation:   oc,
				}
			}
		}

	case "muster":
		title := text.JoinSentenceParts("recorded at muster")
		if regiment, ok := gev.Attributes[model.EventAttributeRegiment]; ok {
			if battalion, ok := gev.Attributes[model.EventAttributeBattalion]; ok {
				title = text.JoinSentenceParts(title, "for the", battalion, "battalion,", regiment)
			} else {
				logger.Warn("anomaly: muster missing battalion attribute")
				title = text.JoinSentenceParts(title, "for the", regiment, "regiment")
			}
		} else {
			logger.Warn("anomaly: muster missing regiment attribute")
		}
		gev.Title = title
		ev = &model.MusterEvent{
			GeneralEvent: gev,
		}

	case "battle":
		if desc := pval(grev.Description, ""); desc != "" {
			gev.Title = "participated in the " + desc
		} else {
			gev.Title = "participated in battle"
		}
		ev = &model.BattleEvent{
			GeneralEvent: gev,
		}

	default:
		logger.Warn("unhandled general event type", "type", pval(grev.Type, "unknown"))
	}

	if ev != nil {
		l.timelineEvents[id] = ev
	}

	return nil
}

func (l *Loader) lookupEvent(grer *grampsxml.Eventref) (model.TimelineEvent, bool) {
	// TODO: consider priv attribute, maybe don't return private references
	grev, ok := l.EventsByHandle[grer.Hlink]
	if !ok {
		return nil, false
	}
	id := pval(grev.ID, grev.Handle)
	ev, ok := l.timelineEvents[id]

	return ev, ok
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

// isLastDayOfMonth reports whether d is the last day of month m in year y.
func isLastDayOfMonth(y, m, d int) bool {
	// First day of the next month minus one day gives the last day of this month.
	t := time.Date(y, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC)
	return d == t.Day()
}

// asPrecise converts a gdate.Date to a *gdate.Precise. If start is true, a
// MonthYear is promoted to the first of the month; otherwise to the last day.
func asPrecise(d gdate.Date, start bool) *gdate.Precise {
	switch dt := d.(type) {
	case *gdate.Precise:
		return dt
	case *gdate.MonthYear:
		day := 1
		if !start {
			t := time.Date(dt.Y, time.Month(dt.M)+1, 0, 0, 0, 0, 0, time.UTC)
			day = t.Day()
		}
		return &gdate.Precise{C: dt.C, Y: dt.Y, M: dt.M, D: day}
	}
	return nil
}

func EventDate(grev *grampsxml.Event, dp gdate.Parser) (*model.Date, error) {
	if grev.Dateval != nil {
		dt, err := ParseDateval(*grev.Dateval, dp)
		if err != nil {
			return nil, fmt.Errorf("parse date value: %w", err)
		}
		return dt, nil
	} else if grev.Daterange != nil {
		dt, err := ParseDaterange(*grev.Daterange, dp)
		if err != nil {
			return nil, fmt.Errorf("parse date range: %w", err)
		}
		return dt, nil
	} else if grev.Datespan != nil {
		dt, err := ParseDatespan(*grev.Datespan, dp)
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

func ParseDateval(dv grampsxml.Dateval, dp gdate.Parser) (*model.Date, error) {
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
	if dv.Dualdated != nil && *dv.Dualdated && dp.Calendar == gdate.Julian25Mar {
		// Need to subtract a year if before 25 March
		switch tdt := dt.(type) {
		case *gdate.Precise:
			if tdt.M == 1 || tdt.M == 2 || (tdt.M == 3 && tdt.D < 25) {
				tdt.Y--
			}
		}
	}

	dateType := pval(dv.Type, "Regular")
	switch strings.ToLower(dateType) {
	case "before":
		switch d := dt.(type) {
		case *gdate.Year:
			dt = &gdate.BeforeYear{C: d.C, Y: d.Y}
		case *gdate.Precise:
			dt = &gdate.BeforePrecise{C: d.C, Y: d.Y, M: d.M, D: d.D}
		case *gdate.MonthYear:
			dt = &gdate.BeforePrecise{C: d.C, Y: d.Y, M: d.M, D: 1}
		default:
			return nil, fmt.Errorf("'before' type not supported for date %T", dt)
		}
	case "after":
		switch d := dt.(type) {
		case *gdate.Year:
			dt = &gdate.AfterYear{C: d.C, Y: d.Y}
		case *gdate.Precise:
			dt = &gdate.AfterPrecise{C: d.C, Y: d.Y, M: d.M, D: d.D}
		case *gdate.MonthYear:
			dt = &gdate.AfterPrecise{C: d.C, Y: d.Y, M: d.M, D: 1}
		default:
			return nil, fmt.Errorf("'after' type not supported for date %T", dt)
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

func ParseDaterange(dr grampsxml.Daterange, dp gdate.Parser) (*model.Date, error) {
	// copy the parser so we functions can override if need be
	if dr.Cformat != nil {
		return nil, fmt.Errorf("date range Cformat not supported")
	}
	if dr.Dualdated != nil {
		return nil, fmt.Errorf("date range Dualdated not supported")
	}
	if dr.Newyear != nil {
		return nil, fmt.Errorf("date range Newyear not supported")
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

	dstart, err := dp.Parse(dr.Start)
	if err != nil {
		return nil, fmt.Errorf("parse start value %q: %w", dr.Start, err)
	}

	dstop, err := dp.Parse(dr.Stop)
	if err != nil {
		return nil, fmt.Errorf("parse stop value %q: %w", dr.Stop, err)
	}

	// Between two bare years → YearRange.
	if startY, ok := dstart.(*gdate.Year); ok {
		if stopY, ok := dstop.(*gdate.Year); ok {
			return &model.Date{
				Date:       &gdate.YearRange{C: startY.C, Lower: startY.Y, Upper: stopY.Y},
				Derivation: deriv,
			}, nil
		}
	}

	// Extract MonthYear for both endpoints (for quarter detection).
	// A Precise date with D==1 maps cleanly to its MonthYear for the start;
	// a Precise date on the last day of the month maps cleanly for the stop.
	var mystart *gdate.MonthYear
	switch tstart := dstart.(type) {
	case *gdate.MonthYear:
		mystart = tstart
	case *gdate.Precise:
		if tstart.D == 1 {
			mystart = &gdate.MonthYear{C: tstart.C, M: tstart.M, Y: tstart.Y}
		}
	}

	var mystop *gdate.MonthYear
	switch tstop := dstop.(type) {
	case *gdate.MonthYear:
		mystop = tstop
	case *gdate.Precise:
		if isLastDayOfMonth(tstop.Y, tstop.M, tstop.D) {
			mystop = &gdate.MonthYear{C: tstop.C, M: tstop.M, Y: tstop.Y}
		}
	}

	// Try to collapse to a GRO quarter.
	if mystart != nil && mystop != nil && mystart.Y == mystop.Y {
		switch {
		case mystart.M == 1 && mystop.M == 3:
			return &model.Date{Date: &gdate.YearQuarter{C: mystart.C, Y: mystart.Y, Q: 1}, Derivation: deriv}, nil
		case mystart.M == 4 && mystop.M == 6:
			return &model.Date{Date: &gdate.YearQuarter{C: mystart.C, Y: mystart.Y, Q: 2}, Derivation: deriv}, nil
		case mystart.M == 7 && mystop.M == 9:
			return &model.Date{Date: &gdate.YearQuarter{C: mystart.C, Y: mystart.Y, Q: 3}, Derivation: deriv}, nil
		case mystart.M == 10 && mystop.M == 12:
			return &model.Date{Date: &gdate.YearQuarter{C: mystart.C, Y: mystart.Y, Q: 4}, Derivation: deriv}, nil
		}
	}

	// Fall back to a BetweenPrecise using the actual dates, promoting MonthYear
	// endpoints to first/last day of month as needed.
	startP, stopP := asPrecise(dstart, true), asPrecise(dstop, false)
	if startP == nil {
		return nil, fmt.Errorf("parse start value %q: unsupported date type for range", dr.Start)
	}
	if stopP == nil {
		return nil, fmt.Errorf("parse stop value %q: unsupported date type for range", dr.Stop)
	}
	return &model.Date{
		Date: &gdate.BetweenPrecise{
			C:          startP.C,
			StartYear:  startP.Y,
			StartMonth: startP.M,
			StartDay:   startP.D,
			EndYear:    stopP.Y,
			EndMonth:   stopP.M,
			EndDay:     stopP.D,
		},
		Derivation: deriv,
	}, nil
}

func ParseDatespan(ds grampsxml.Datespan, dp gdate.Parser) (*model.Date, error) {
	if ds.Cformat != nil {
		return nil, fmt.Errorf("date span Cformat not supported")
	}
	if ds.Dualdated != nil {
		return nil, fmt.Errorf("date span Dualdated not supported")
	}
	if ds.Newyear != nil {
		return nil, fmt.Errorf("date span Newyear not supported")
	}

	// Quality:
	// - Regular
	// - Estimated
	// - Calculated
	var deriv model.DateDerivation
	if ds.Quality != nil {
		switch strings.ToLower(*ds.Quality) {
		case "regular":
			deriv = model.DateDerivationStandard
		case "estimated":
			deriv = model.DateDerivationEstimated
		case "calculated":
			deriv = model.DateDerivationCalculated
		default:
			return nil, fmt.Errorf("quality value %q not supported", *ds.Quality)

		}
	}

	dstart, err := dp.Parse(ds.Start)
	if err != nil {
		return nil, fmt.Errorf("parse start value %q: %w", ds.Start, err)
	}

	var mystartY *gdate.Year
	var mystartMY *gdate.MonthYear
	switch tstart := dstart.(type) {
	case *gdate.Year:
		mystartY = tstart
	case *gdate.MonthYear:
		mystartMY = tstart
	}

	if mystartY == nil && mystartMY == nil {
		return nil, fmt.Errorf("parse start value %q: unsupported start date type", ds.Start)
	}

	dstop, err := dp.Parse(ds.Stop)
	if err != nil {
		return nil, fmt.Errorf("parse stop value %q: %w", ds.Stop, err)
	}
	var mystopY *gdate.Year
	var mystopMY *gdate.MonthYear
	switch tstop := dstop.(type) {
	case *gdate.Year:
		mystopY = tstop
	case *gdate.MonthYear:
		mystopMY = tstop
	}

	if mystopY == nil && mystopMY == nil {
		return nil, fmt.Errorf("parse stop value %q: unsupported stop date type", ds.Stop)
	}

	// If either endpoint has month precision, use a MonthYearRange, promoting
	// bare years to Jan (start) or Dec (stop) as appropriate.
	if mystartMY != nil || mystopMY != nil {
		if mystartMY == nil {
			mystartMY = &gdate.MonthYear{C: mystartY.C, Y: mystartY.Y, M: 1}
		}
		if mystopMY == nil {
			mystopMY = &gdate.MonthYear{C: mystopY.C, Y: mystopY.Y, M: 12}
		}
		return &model.Date{
			Date: &gdate.MonthYearRange{
				C:          mystartMY.C,
				LowerYear:  mystartMY.Y,
				LowerMonth: mystartMY.M,
				UpperYear:  mystopMY.Y,
				UpperMonth: mystopMY.M,
			},
			Derivation: deriv,
			Span:       true,
		}, nil
	}

	return &model.Date{
		Date: &gdate.YearRange{
			C:     mystartY.C,
			Lower: mystartY.Y,
			Upper: mystopY.Y,
		},
		Derivation: deriv,
		Span:       true,
	}, nil
}

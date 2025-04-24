package gedcom

import (
	"fmt"

	"github.com/iand/gdate"
	"github.com/iand/gedcom"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

func (l *Loader) populateFamilyFacts(m ModelFinder, fr *gedcom.FamilyRecord) error {
	logger := logging.With("source", "family", "xref", fr.Xref)
	logger.Debug("populating from family record")

	// Ignore step parent families
	frels := findUserDefinedTags(fr.UserDefined, "_FREL", false)
	if len(frels) != 0 {
		return nil
	}

	fatherPresent := fr.Husband != nil
	motherPresent := fr.Wife != nil

	var father, mother *model.Person

	if fatherPresent {
		father = m.FindPerson(l.ScopeName, fr.Husband.Xref)
	} else {
		father = model.UnknownPerson()
	}

	if motherPresent {
		mother = m.FindPerson(l.ScopeName, fr.Wife.Xref)
	} else {
		mother = model.UnknownPerson()
	}

	for _, ch := range fr.Child {
		child := m.FindPerson(l.ScopeName, ch.Xref)

		if fatherPresent {
			if child.Father.IsUnknown() {
				child.Father = father
			} else {
				child.Anomalies = append(child.Anomalies, &model.Anomaly{
					Category: "GEDCOM",
					Text:     "Person appeared as a child in two GEDCOM family records with different husband records",
					Context:  "Family ref " + fr.Xref + ", Husband ref " + fr.Husband.Xref + ", Child ref " + ch.Xref,
				})
				father.Anomalies = append(father.Anomalies, &model.Anomaly{
					Category: "GEDCOM",
					Text:     "Person appeared as a husband in two GEDCOM family records with the same child",
					Context:  "Family ref " + fr.Xref + ", Husband ref " + fr.Husband.Xref + ", Child ref " + ch.Xref,
				})
			}
		}

		if motherPresent {
			if child.Mother.IsUnknown() {
				child.Mother = mother
			} else {
				child.Anomalies = append(child.Anomalies, &model.Anomaly{
					Category: "GEDCOM",
					Text:     "Person appeared as a child in two GEDCOM family records with different wife records",
					Context:  "Family ref " + fr.Xref + ", Wife ref " + fr.Wife.Xref + ", Child ref " + ch.Xref,
				})
				mother.Anomalies = append(mother.Anomalies, &model.Anomaly{
					Category: "GEDCOM",
					Text:     "Person appeared in wife record in two GEDCOM family records with the same child",
					Context:  "Family ref " + fr.Xref + ", Wife ref " + fr.Wife.Xref + ", Child ref " + ch.Xref,
				})
			}
		}

	}

	events := append([]*gedcom.EventRecord{}, fr.Event...)

	for _, er := range events {
		pl, _ := l.findPlaceForEvent(m, er)
		dp := &gdate.Parser{
			AssumeGROQuarter: true,
		}

		dt, err := dp.Parse(er.Date)
		if err != nil {
			return fmt.Errorf("date: %w", err)
		}

		gev := model.GeneralEvent{
			Date:       &model.Date{Date: dt},
			Place:      pl,
			Detail:     er.Value,
			Title:      fmt.Sprintf("%s event %s", er.Tag, dt.Occurrence()),
			Attributes: make(map[string]string),
		}
		var anoms []*model.Anomaly
		gev.Citations, anoms = l.parseCitationRecords(m, er.Citation, logger)
		for _, anom := range anoms {
			if fatherPresent {
				father.Anomalies = append(father.Anomalies, anom)
			}
			if motherPresent {
				mother.Anomalies = append(mother.Anomalies, anom)
			}
		}

		gpe := model.GeneralUnionEvent{
			Husband: father,
			Wife:    mother,
		}

		var ev model.TimelineEvent
		switch er.Tag {
		case "MARR":
			ev = &model.MarriageEvent{
				GeneralEvent:      gev,
				GeneralUnionEvent: gpe,
			}

		case "MARB":
			ev = &model.MarriageBannsEvent{
				GeneralEvent:      gev,
				GeneralUnionEvent: gpe,
			}
		case "MARL":
			ev = &model.MarriageLicenseEvent{
				GeneralEvent:      gev,
				GeneralUnionEvent: gpe,
			}
		case "DIV":
			ev = &model.DivorceEvent{
				GeneralEvent:      gev,
				GeneralUnionEvent: gpe,
			}
		case "ANUL":
			ev = &model.AnnulmentEvent{
				GeneralEvent:      gev,
				GeneralUnionEvent: gpe,
			}
		default:
			logging.Warn("unknown family event", "xref", fr.Xref, "tag", er.Tag, "value", er.Value)
		}

		if !mother.IsUnknown() {
			mother.Timeline = append(mother.Timeline, ev)
		}

		if !father.IsUnknown() {
			father.Timeline = append(father.Timeline, ev)
		}
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

	return nil
}

func (l *Loader) buildFamilies(t *tree.Tree, p *model.Person) error {
	var parentFamily *model.Family
	if p.Father.IsUnknown() {
		if p.Mother.IsUnknown() {
			// no known mother or father
			return nil
		} else {
			parentFamily = t.FindFamilyOneParent(p.Mother, p)
		}
	} else {
		if p.Mother.IsUnknown() {
			parentFamily = t.FindFamilyOneParent(p.Father, p)
		} else {
			parentFamily = t.FindFamilyByParents(p.Father, p.Mother)
		}
	}
	parentFamily.Children = append(parentFamily.Children, p)

	sortMaleFemale := func(p1 *model.Person, p2 *model.Person) (*model.Person, *model.Person, bool) {
		if p1.Gender == model.GenderMale && p2.Gender == model.GenderFemale {
			return p1, p2, true
		}
		if p1.Gender == model.GenderFemale && p2.Gender == model.GenderMale {
			return p2, p1, true
		}

		return p1, p2, false
	}

	addMarriage := func(t *tree.Tree, ev model.UnionTimelineEvent) {
		p1 := ev.GetHusband()
		p2 := ev.GetWife()
		if p1.IsUnknown() || p2.IsUnknown() {
			return
		}
		father, mother, ok := sortMaleFemale(p1, p2)
		if !ok {
			return
		}

		marriageFamily := t.FindFamilyByParents(father, mother)
		marriageFamily.Bond = model.FamilyBondMarried
		marriageFamily.Timeline = append(marriageFamily.Timeline, ev)
		marriageFamily.BestStartEvent = ev
		marriageFamily.BestStartDate = ev.GetDate()
	}

	for _, ev := range p.Timeline {
		switch tev := ev.(type) {
		case *model.MarriageEvent:
			addMarriage(t, tev)
		case *model.MarriageLicenseEvent:
			addMarriage(t, tev)
		case *model.MarriageBannsEvent:
			addMarriage(t, tev)
		case *model.DivorceEvent:
		case *model.AnnulmentEvent:
		}
	}

	return nil
}

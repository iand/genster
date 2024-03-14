package gramps

import (
	"fmt"

	"github.com/iand/gdate"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

func (l *Loader) populateFamilyFacts(m ModelFinder, fr *grampsxml.Family) error {
	id := pval(fr.ID, fr.Handle)

	logger := logging.With("source", "family", "id", id)
	logger.Debug("populating from family record")

	var fatherPresent, motherPresent bool
	var father, mother *model.Person

	if fr.Father != nil {
		fp, ok := l.PeopleByHandle[fr.Father.Hlink]
		if ok {
			father = m.FindPerson(l.ScopeName, pval(fp.ID, fp.Handle))
			fatherPresent = true
		}
	}

	if !fatherPresent {
		father = model.UnknownPerson()
	}

	if fr.Mother != nil {
		mp, ok := l.PeopleByHandle[fr.Mother.Hlink]
		if ok {
			mother = m.FindPerson(l.ScopeName, pval(mp.ID, mp.Handle))
			motherPresent = true
		}
	}

	if !motherPresent {
		mother = model.UnknownPerson()
	}

	if fr.Rel != nil {
		fam := m.FindFamilyByParents(father, mother)
		switch fr.Rel.Type {
		case "Married":
			fam.Bond = model.FamilyBondMarried
		case "Unmarried":
			fam.Bond = model.FamilyBondUnmarried
		}
	}

	for _, cr := range fr.Childref {
		cp, ok := l.PeopleByHandle[cr.Hlink]
		if !ok {
			logger.Warn("could not find child with handle", "handle", cr.Hlink)
			continue
		}
		child := m.FindPerson(l.ScopeName, pval(cp.ID, cp.Handle))

		if fatherPresent {
			if child.Father.IsUnknown() {
				child.Father = father
			} else {
				child.Anomalies = append(child.Anomalies, &model.Anomaly{
					Category: "Gramps",
					Text:     "Person appeared as a child in two Gramps family records with different father records",
					Context:  "Family handle " + fr.Handle + ", Father handle " + fr.Father.Hlink + ", Child handle " + cp.Handle,
				})
				father.Anomalies = append(father.Anomalies, &model.Anomaly{
					Category: "Gramps",
					Text:     "Person appeared as a father in two Gramps family records with the same child",
					Context:  "Family handle " + fr.Handle + ", Father handle " + fr.Father.Hlink + ", Child handle " + cp.Handle,
				})
			}
		}

		if motherPresent {
			if child.Mother.IsUnknown() {
				child.Mother = mother
			} else {
				child.Anomalies = append(child.Anomalies, &model.Anomaly{
					Category: "Gramps",
					Text:     "Person appeared as a child in two Gramps family records with different mother records",
					Context:  "Family handle " + fr.Handle + ", Mother handle " + fr.Mother.Hlink + ", Child handle " + cp.Handle,
				})
				mother.Anomalies = append(mother.Anomalies, &model.Anomaly{
					Category: "Gramps",
					Text:     "Person appeared as a mother in two Gramps family records with the same child",
					Context:  "Family handle " + fr.Handle + ", Mother handle " + fr.Mother.Hlink + ", Child handle " + cp.Handle,
				})
			}
		}
	}

	dp := &gdate.Parser{
		AssumeGROQuarter: true,
	}

	for _, er := range fr.Eventref {
		grev, ok := l.EventsByHandle[er.Hlink]
		if !ok {
			logger.Warn("could not find event", "hlink", er.Hlink)
			continue
		}
		pl, _ := l.findPlaceForEvent(m, grev)

		var dt gdate.Date
		var err error
		if grev.Dateval != nil {
			dt, err = dp.Parse(grev.Dateval.Val)
			if err != nil {
				return fmt.Errorf("date: %w", err)
			}
		} else {
			logger.Warn("could not parse event date", "hlink", er.Hlink)
			continue

		}

		gev := model.GeneralEvent{
			Date:   &model.Date{Date: dt},
			Place:  pl,
			Detail: "", // TODO: notes
			Title:  pval(grev.Description, ""),
		}

		// var anoms []*model.Anomaly
		// gev.Citations, anoms = l.parseCitationRecords(m, er.Citation, logger)
		// for _, anom := range anoms {
		// 	if fatherPresent {
		// 		father.Anomalies = append(father.Anomalies, anom)
		// 	}
		// 	if motherPresent {
		// 		mother.Anomalies = append(mother.Anomalies, anom)
		// 	}
		// }

		gpe := model.GeneralPartyEvent{
			Party1: father,
			Party2: mother,
		}

		var ev model.TimelineEvent

		switch pval(grev.Type, "unknown") {
		case "Marriage":
			ev = &model.MarriageEvent{
				GeneralEvent:      gev,
				GeneralPartyEvent: gpe,
			}
		case "Divorce":
			ev = &model.DivorceEvent{
				GeneralEvent:      gev,
				GeneralPartyEvent: gpe,
			}
		default:
			logger.Warn("unhandled family event type", "hlink", er.Hlink, "type", pval(grev.Type, "unknown"))

		}

		if ev != nil {

			if !mother.IsUnknown() {
				mother.Timeline = append(mother.Timeline, ev)
			}

			if !father.IsUnknown() {
				father.Timeline = append(father.Timeline, ev)
			}
			if !pl.IsUnknown() {
				pl.Timeline = append(pl.Timeline, ev)
			}

			// seenSource := make(map[*model.Source]bool)
			// for _, c := range ev.GetCitations() {
			// 	if c.Source != nil && !seenSource[c.Source] {
			// 		c.Source.EventsCiting = append(c.Source.EventsCiting, ev)
			// 		seenSource[c.Source] = true
			// 	}
			// }
		}
	}
	return nil
}

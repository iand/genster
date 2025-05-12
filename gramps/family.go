package gramps

import (
	"fmt"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

func (l *Loader) populateFamilyFacts(m ModelFinder, fr *grampsxml.Family) error {
	id := pval(fr.ID, fr.Handle)
	fam := m.FindFamily(l.ScopeName, id)

	logger := logging.With("source", "family", "id", fam.ID, "native_id", id)
	logger.Debug("populating from family record")

	if fr.Rel != nil {
		switch fr.Rel.Type {
		case "Married":
			fam.Bond = model.FamilyBondMarried
		case "Unmarried":
			fam.Bond = model.FamilyBondUnmarried
		}
	}

	var fatherPresent, motherPresent bool
	var father, mother *model.Person

	if fr.Father != nil {
		fp, ok := l.PeopleByHandle[fr.Father.Hlink]
		if ok {
			father = m.FindPerson(l.ScopeName, pval(fp.ID, fp.Handle))
			fatherPresent = true
			fam.Father = father
			father.Families = append(father.Families, fam)
		}
	}

	if !fatherPresent {
		father = model.UnknownPerson()
		fam.Father = father
	}

	if fr.Mother != nil {
		mp, ok := l.PeopleByHandle[fr.Mother.Hlink]
		if ok {
			mother = m.FindPerson(l.ScopeName, pval(mp.ID, mp.Handle))
			motherPresent = true
			fam.Mother = mother
			mother.Families = append(mother.Families, fam)
		}
	}

	if !motherPresent {
		mother = model.UnknownPerson()
		fam.Mother = mother
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

	for _, grer := range fr.Eventref {
		grev, ok := l.EventsByHandle[grer.Hlink]
		if !ok {
			logger.Warn("could not find event", "hlink", grer.Hlink)
			continue
		}

		role := strings.ToLower(pval(grer.Role, "unknown"))
		evtype := strings.ToLower(pval(grev.Type, "unknown"))

		var ev model.TimelineEvent
		ev, ok = l.lookupEvent(&grer)
		if !ok {
			panic(fmt.Sprintf("family event not found (evtype=%s, hlink=%s)", evtype, grer.Hlink))
		}
		if role == "family" {
			ev.AddParticipant(&model.EventParticipant{
				Person: father,
				Role:   model.EventRoleHusband,
			})
			ev.AddParticipant(&model.EventParticipant{
				Person: mother,
				Role:   model.EventRoleWife,
			})
			switch evtype {
			case "marriage":
				fam.Bond = model.FamilyBondMarried
				fam.Timeline = append(fam.Timeline, ev)
				fam.BestStartEvent = ev
				fam.BestStartDate = ev.GetDate()
			case "marriage license":
				fam.Timeline = append(fam.Timeline, ev)
				if fam.BestStartEvent == nil {
					fam.BestStartEvent = ev
					fam.BestStartDate = ev.GetDate()
				}
			case "marriage banns":
				fam.Timeline = append(fam.Timeline, ev)
				if fam.BestStartEvent == nil {
					fam.BestStartEvent = ev
					fam.BestStartDate = ev.GetDate()
				}
			case "divorce":
				fam.BestEndEvent = ev
				fam.BestEndDate = ev.GetDate()
			default:
				logger.Warn("unhandled family event type", "hlink", grer.Hlink, "type", evtype, "role", role)
			}
		} else {
			logger.Warn("unhandled family event role", "hlink", grer.Hlink, "type", evtype, "role", role)
		}

		if ev != nil {
			if !mother.IsUnknown() {
				mother.Timeline = append(mother.Timeline, ev)
			}

			if !father.IsUnknown() {
				father.Timeline = append(father.Timeline, ev)
			}
			pl := ev.GetPlace()
			if !pl.IsUnknown() {
				pl.Timeline = append(pl.Timeline, ev)
			}
		}
	}

	// Add attributes
	for _, att := range fr.Attribute {
		if pval(att.Priv, false) {
			logger.Debug("skipping attribute marked as private", "type", att.Type)
			continue
		}
		switch strings.ToLower(att.Type) {
		case "number of children":
			n, err := model.ParseNumberOfChildren(att.Value)
			if err != nil {
				logger.Warn("could not parse number of children", "error", err.Error(), "handle", fr.Handle)
				break
			}
			fam.NumberOfChildren = n
		case "all children known":
			fam.AllChildrenKnown = true
		}
	}

	for _, tref := range fr.Tagref {
		tag, ok := l.TagsByHandle[tref.Hlink]
		if !ok {
			logger.Warn("could not find tag", "hlink", tref.Hlink)
			continue
		}
		switch strings.ToLower(tag.Name) {
		case "publish":
			fam.PublishChildren = true
		}
	}

	return nil
}

package gedcom

import (
	"fmt"

	"github.com/iand/gdate"
	"github.com/iand/gedcom"
	"github.com/iand/genster/model"
	"golang.org/x/exp/slog"
)

func (l *Loader) populateFamilyFacts(m ModelFinder, fr *gedcom.FamilyRecord) error {
	slog.Debug("populating from family record", "xref", fr.Xref)

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
	dp := &gdate.Parser{
		AssumeGROQuarter: true,
	}

	for _, er := range events {
		pl, _ := l.findPlaceForEvent(m, er)

		dt, err := dp.Parse(er.Date)
		if err != nil {
			return fmt.Errorf("date: %w", err)
		}

		gev := model.GeneralEvent{
			Date:   &model.Date{Date: dt},
			Place:  pl,
			Detail: er.Value,
			Title:  fmt.Sprintf("%s event %s", er.Tag, dt.Occurrence()),
		}
		var anoms []*model.Anomaly
		gev.Citations, anoms = l.parseCitationRecords(m, er.Citation)
		for _, anom := range anoms {
			if fatherPresent {
				father.Anomalies = append(father.Anomalies, anom)
			}
			if motherPresent {
				mother.Anomalies = append(mother.Anomalies, anom)
			}
		}

		gpe := model.GeneralPartyEvent{
			Party1: father,
			Party2: mother,
		}

		var ev model.TimelineEvent
		switch er.Tag {
		case "MARR":
			ev = &model.MarriageEvent{
				GeneralEvent:      gev,
				GeneralPartyEvent: gpe,
			}

		case "MARB":
			ev = &model.MarriageBannsEvent{
				GeneralEvent:      gev,
				GeneralPartyEvent: gpe,
			}
		case "MARL":
			ev = &model.MarriageLicenseEvent{
				GeneralEvent:      gev,
				GeneralPartyEvent: gpe,
			}
		case "DIV":
			ev = &model.DivorceEvent{
				GeneralEvent:      gev,
				GeneralPartyEvent: gpe,
			}
		case "ANUL":
			ev = &model.AnnulmentEvent{
				GeneralEvent:      gev,
				GeneralPartyEvent: gpe,
			}
		default:
			ev = &model.PlaceholderPartyEvent{
				GeneralEvent:      gev,
				GeneralPartyEvent: gpe,
				ExtraInfo:         fmt.Sprintf("Unknown family event (Xref=%s, Tag=%s, Value=%s)", fr.Xref, er.Tag, er.Value),
			}
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

		if len(ev.GetCitations()) == 0 && ev.GetDate().IsFirm() {
			if !mother.IsUnknown() {
				mother.ToDos = append(mother.ToDos, &model.ToDo{
					Category: "Citation",
					Text:     fmt.Sprintf("This event appears to have a firm date %q but no source citation", ev.GetDate().String()),
					Context:  "No citation for " + ev.Type() + " event",
				})
			}
			if !father.IsUnknown() {
				father.ToDos = append(father.ToDos, &model.ToDo{
					Category: "Citation",
					Text:     fmt.Sprintf("This event appears to have a firm date %q but no source citation", ev.GetDate().String()),
					Context:  "No citation for " + ev.Type() + " event",
				})
			}
		}

	}

	return nil
}

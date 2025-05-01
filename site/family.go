package site

import (
	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
)

func RenderFamilyPage(s *Site, f *model.Family) (render.Document[md.Text], error) {
	doc := s.NewDocument()
	doc.Layout(PageLayoutFamily.String())
	doc.Category(PageCategoryFamily)
	doc.SetSitemapDisable()
	doc.ID(f.ID)
	doc.Title(f.PreferredUniqueName)
	// if p.Redacted {
	// }
	n := BuildFamilyNarrative(f, false)

	nc := &narrative.DefaultNameChooser{}

	n.Render(doc, nc)

	return doc, nil
}

func BuildFamilyNarrative(f *model.Family, inlineMedia bool) *narrative.FamilyNarrative[md.Text] {
	n := &narrative.FamilyNarrative[md.Text]{
		Family: f,
	}

	var timeline []model.TimelineEvent

	fintro := &narrative.IntroStatement[md.Text]{
		Principal:           f.Father,
		SuppressRelation:    true,
		IncludeMedia:        inlineMedia,
		CropMediaHighlights: inlineMedia,
	}
	if !f.Father.IsUnknown() {
		n.FatherStatements = append(n.FatherStatements, fintro)
		timeline = append(timeline, f.Father.Timeline...)

		// If death is known, add it
		if f.Father.BestDeathlikeEvent != nil {
			n.FatherStatements = append(n.FatherStatements, &narrative.DeathStatement[md.Text]{
				Principal:               f.Father,
				ExcludeSurvivingPartner: true,
				IncludeMedia:            inlineMedia,
				CropMediaHighlights:     inlineMedia,
			})
		}
		for _, pf := range f.Father.Families {
			if f.SameAs(pf) {
				continue
			}
			n.FatherStatements = append(n.FatherStatements, &narrative.FamilyStatement[md.Text]{
				Principal: f.Father,
				Family:    pf,
			})
			if !f.BestEndDate.IsUnknown() && f.BestEndEvent != nil && !f.BestEndEvent.IsInferred() {
				n.FatherStatements = append(n.FatherStatements, &narrative.FamilyEndStatement[md.Text]{
					Principal: f.Father,
					Family:    pf,
				})
			}
		}

		if f.Mother.IsUnknown() {
			n.FatherStatements = append(n.FatherStatements, &narrative.FamilyStatement[md.Text]{
				Principal: f.Father,
				Family:    f,
			})
		}
	}

	mintro := &narrative.IntroStatement[md.Text]{
		Principal:           f.Mother,
		SuppressRelation:    true,
		IncludeMedia:        inlineMedia,
		CropMediaHighlights: inlineMedia,
	}
	if !f.Mother.IsUnknown() {
		n.MotherStatements = append(n.MotherStatements, mintro)
		timeline = append(timeline, f.Mother.Timeline...)
		// If death is known, add it
		if f.Mother.BestDeathlikeEvent != nil {
			n.MotherStatements = append(n.MotherStatements, &narrative.DeathStatement[md.Text]{
				Principal:               f.Mother,
				ExcludeSurvivingPartner: true,
				IncludeMedia:            inlineMedia,
				CropMediaHighlights:     inlineMedia,
			})
		}
		for _, pf := range f.Mother.Families {
			if f.SameAs(pf) {
				continue
			}
			n.MotherStatements = append(n.MotherStatements, &narrative.FamilyStatement[md.Text]{
				Principal: f.Mother,
				Family:    pf,
			})
			if !f.BestEndDate.IsUnknown() && f.BestEndEvent != nil && !f.BestEndEvent.IsInferred() {
				n.MotherStatements = append(n.MotherStatements, &narrative.FamilyEndStatement[md.Text]{
					Principal: f.Mother,
					Family:    pf,
				})
			}
		}

		if f.Father.IsUnknown() {
			n.MotherStatements = append(n.MotherStatements, &narrative.FamilyStatement[md.Text]{
				Principal: f.Mother,
				Family:    f,
			})
		}
	}

	if !f.Father.IsUnknown() && !f.Mother.IsUnknown() {
		n.FamilyStatements = append(n.FamilyStatements, &narrative.FamilyStatement[md.Text]{
			Principal: f.Father,
			Family:    f,
		})
	}

	isFamilyEvent := func(ev model.TimelineEvent) bool {
		if !ev.DirectlyInvolves(f.Father) && !ev.DirectlyInvolves(f.Mother) {
			return false
		}
		if ev.GetDate().SortsBefore(f.BestStartDate) {
			return false
		}
		if f.BestEndDate.SortsBefore(ev.GetDate()) {
			return false
		}
		if f.Bond == model.FamilyBondUnmarried || f.Bond == model.FamilyBondLikelyUnmarried {
			// events are only family events for unmarried couples if they are both directly involved
			if !ev.DirectlyInvolves(f.Father) || !ev.DirectlyInvolves(f.Mother) {
				return false
			}
		}
		return true
	}

	seenSharedEvents := map[model.TimelineEvent]bool{}

	for _, ev := range timeline {
		switch tev := ev.(type) {
		case *model.BaptismEvent:
			if tev.DirectlyInvolves(f.Father) {
				fintro.Baptisms = append(fintro.Baptisms, tev)
			} else if tev.DirectlyInvolves(f.Mother) {
				mintro.Baptisms = append(mintro.Baptisms, tev)
			}
		case *model.CensusEvent:
			s := &narrative.CensusStatement[md.Text]{
				Principal:           f.Father, // TODO: change
				Event:               tev,
				IncludeMedia:        inlineMedia,
				CropMediaHighlights: inlineMedia,
			}
			if isFamilyEvent(ev) {
				if !seenSharedEvents[ev] {
					n.FamilyStatements = append(n.FamilyStatements, s)
					seenSharedEvents[ev] = true
				}
			} else if tev.DirectlyInvolves(f.Father) {
				n.FatherStatements = append(n.FatherStatements, s)
			} else if tev.DirectlyInvolves(f.Mother) {
				n.MotherStatements = append(n.MotherStatements, s)
			}
		case *model.IndividualNarrativeEvent:
			s := &narrative.GeneralEventStatement[md.Text]{
				Principal: f.Father, // TODO: change
				Event:     tev,
			}
			if isFamilyEvent(ev) {
				if !seenSharedEvents[ev] {
					n.FamilyStatements = append(n.FamilyStatements, s)
					seenSharedEvents[ev] = true
				}
			} else if tev.DirectlyInvolves(f.Father) {
				n.FatherStatements = append(n.FatherStatements, s)
			} else if tev.DirectlyInvolves(f.Mother) {
				n.MotherStatements = append(n.MotherStatements, s)
			}
		case *model.BirthEvent:
		case *model.DeathEvent:
		case *model.BurialEvent:
		case *model.CremationEvent:
		default:
			if tev.GetNarrative().Text != "" {
				s := &narrative.GeneralEventStatement[md.Text]{
					Principal: f.Father, // TODO: change
					Event:     tev,
				}
				if isFamilyEvent(ev) {
					if !seenSharedEvents[ev] {
						n.FamilyStatements = append(n.FamilyStatements, s)
						seenSharedEvents[ev] = true
					}
				} else if tev.DirectlyInvolves(f.Father) {
					n.FatherStatements = append(n.FatherStatements, s)
				} else if tev.DirectlyInvolves(f.Mother) {
					n.MotherStatements = append(n.MotherStatements, s)
				}
			}

		}
	}

	return n
}

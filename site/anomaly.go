package site

import (
	"fmt"
	"time"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
)

func (s *Site) ScanPersonForAnomalies(p *model.Person) {
	for _, ev := range p.Timeline {
		anoms := ScanTimelineEventForAnomalies(ev)
		if len(anoms) > 0 {
			for _, anom := range anoms {
				p.Anomalies = append(p.Anomalies, anom)
			}
		}
	}
}

func ScanTimelineEventForAnomalies(ev model.TimelineEvent) []*model.Anomaly {
	var anomalies []*model.Anomaly

	for _, cit := range ev.GetCitations() {
		anoms := ScanGeneralCitationForAnomalies(cit)
		if len(anoms) > 0 {
			for _, anom := range anoms {
				// TODO: add context
				anom.Context += " for " + ev.Type() + " " + ev.GetDate().Occurrence()
				anomalies = append(anomalies, anom)
			}
		}
	}

	return anomalies
}

func ScanGeneralCitationForAnomalies(cit *model.GeneralCitation) []*model.Anomaly {
	var anomalies []*model.Anomaly

	var name string
	if cit.Source != nil && cit.Source.Title != "" {
		name = cit.Source.Title
	} else {
		name = cit.Detail
	}

	recent := &gdate.Year{Y: time.Now().Year() - 25}
	if len(cit.TranscriptionText) != 0 && !gdate.IsUnknown(cit.TranscriptionDate) && gdate.SortsBefore(cit.TranscriptionDate, recent) {
		// Transcription date might be the date of the original record

		anomalies = append(anomalies, &model.Anomaly{
			Category: "Citation",
			Text:     fmt.Sprintf("%q might be the date of the original record, it should be the date the transcription was made.", cit.TranscriptionDate.String()),
			Context:  "Transcription date for citation of " + name,
		})
	}

	return anomalies
}

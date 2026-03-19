package site

import (
	"fmt"
	"strconv"

	"github.com/iand/gdate"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
)

func RenderTimeline[T render.EncodedText](t *model.Timeline, pov *model.POV, enc render.ContentBuilder[T], fmtr TimelineEntryFormatter[T]) error {
	enc.EmptyPara()
	if len(t.Events) == 0 {
		return nil
	}
	model.SortTimelineEvents(t.Events)

	logger := logging.Default()
	if !pov.Person.IsUnknown() {
		logger = logger.With("id", pov.Person.ID, "native_id", pov.Person.NativeID)
	}

	monthNames := []string{
		1:  "Jan",
		2:  "Feb",
		3:  "Mar",
		4:  "Apr",
		5:  "May",
		6:  "Jun",
		7:  "Jul",
		8:  "Aug",
		9:  "Sep",
		10: "Oct",
		11: "Nov",
		12: "Dec",
	}

	var row render.TimelineRow[T]
	events := make([]render.TimelineRow[T], 0, len(t.Events))
	for i, ev := range t.Events {
		if !IncludeInTimeline(ev) {
			continue
		}
		title := fmtr.Title(i, ev)
		if title == "" {
			continue
		}

		dt := ev.GetDate()
		if dt.IsUnknown() {
			continue
		}

		var sy, sd string
		if y, m, d, ok := dt.YMD(); ok {
			sy = strconv.Itoa(y)
			sd = fmt.Sprintf("%d %s", d, monthNames[m])
		} else if dt.Span {
			switch d := dt.Date.(type) {
			case *gdate.YearRange:
				sy = d.String()
				sd = ""
			default:
				logger.Warn("timeline: unsupported date span", "type", fmt.Sprintf("%T", d), "value", d.String(), "event", fmt.Sprintf("%T", ev))
			}
		} else {
			switch d := dt.Date.(type) {
			case *gdate.BeforeYear:
				sy = d.Occurrence()
				sd = ""
			case *gdate.AfterYear:
				sy = d.Occurrence()
				sd = ""
			case *gdate.AboutYear:
				sy = d.Occurrence()
				sd = ""
			case *gdate.Year:
				sy = d.String()
				sd = ""
			case *gdate.YearQuarter:
				sy = strconv.Itoa(d.Year())
				sd = d.MonthRange()
			case *gdate.MonthYear:
				sy = strconv.Itoa(d.Year())
				sd = monthNames[d.M]
			case *gdate.BeforePrecise:
				sy = d.Occurrence()
				sd = ""
			case *gdate.AfterPrecise:
				sy = d.Occurrence()
				sd = ""
			case *gdate.BetweenPrecise:
				sy = d.Occurrence()
				sd = ""
			case *gdate.YearRange:
				sy = d.String()
				sd = ""
			default:
				logger.Warn("timeline: unsupported date type", "type", fmt.Sprintf("%T", d), "value", d.String(), "event", fmt.Sprintf("%T", ev))
			}
		}

		if row.Year == sy && row.Date == sd {
			row.Details = append(row.Details, enc.EncodeText(title))
			continue
		} else {
			if row.Year != "" {
				events = append(events, row)
			}
			row = render.TimelineRow[T]{
				Year:    sy,
				Date:    sd,
				Details: []T{enc.EncodeText(title)},
			}
		}
	}
	if row.Year != "" {
		events = append(events, row)
	}
	enc.Timeline(events)
	return nil
}

func IncludeInTimeline(ev model.TimelineEvent) bool {
	switch ev.(type) {
	case *model.IndividualNarrativeEvent:
		return false
	default:
		return true
	}
}

type TimelineEntryFormatter[T render.EncodedText] interface {
	Title(seq int, ev model.TimelineEvent) string
	Detail(seq int, ev model.TimelineEvent) string
}

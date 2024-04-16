package gramps

import (
	"fmt"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

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
	if dv.Cformat != nil {
		return nil, fmt.Errorf("Cformat not supported")
	}
	if dv.Dualdated != nil {
		return nil, fmt.Errorf("Dualdated not supported")
	}
	if dv.Newyear != nil {
		return nil, fmt.Errorf("Newyear not supported")
	}

	// Quality:
	// - Regular
	// - Estimated
	// - Calculated
	if dv.Quality != nil && *dv.Quality != "Regular" {
		return nil, fmt.Errorf("Quality not supported")
	}

	dp := &gdate.Parser{}
	dt, err := dp.Parse(dv.Val)
	if err != nil {
		return nil, fmt.Errorf("parse date value: %w", err)
	}

	dateType := pval(dv.Type, "Regular")
	switch dateType {
	case "Before":
		dyear, ok := dt.(*gdate.Year)
		if !ok {
			return nil, fmt.Errorf("'before' type not supported for dates other than years")
		}
		dt = &gdate.BeforeYear{
			C: dyear.C,
			Y: dyear.Y,
		}
	case "After":
		dyear, ok := dt.(*gdate.Year)
		if !ok {
			return nil, fmt.Errorf("'after' type not supported for dates other than years")
		}
		dt = &gdate.AfterYear{
			C: dyear.C,
			Y: dyear.Y,
		}
	case "About":
		dyear, ok := dt.(*gdate.Year)
		if !ok {
			return nil, fmt.Errorf("'about' type not supported for dates other than years")
		}
		dt = &gdate.AboutYear{
			C: dyear.C,
			Y: dyear.Y,
		}
	case "Regular":
		break
	}

	return &model.Date{Date: dt}, nil
}

func ParseDaterange(dr grampsxml.Daterange) (*model.Date, error) {
	if dr.Cformat != nil {
		return nil, fmt.Errorf("Cformat not supported")
	}
	if dr.Dualdated != nil {
		return nil, fmt.Errorf("Dualdated not supported")
	}
	if dr.Newyear != nil {
		return nil, fmt.Errorf("Newyear not supported")
	}

	// Quality:
	// - Regular
	// - Estimated
	// - Calculated
	if dr.Quality != nil && *dr.Quality != "Regular" {
		return nil, fmt.Errorf("Quality not supported")
	}

	// Currently only support quarter ranges
	dp := &gdate.Parser{}

	dstart, err := dp.Parse(dr.Start)
	if err != nil {
		return nil, fmt.Errorf("start value: %w", err)
	}

	mystart, ok := dstart.(*gdate.MonthYear)
	if !ok {
		return nil, fmt.Errorf("unsupported range")
	}

	dstop, err := dp.Parse(dr.Stop)
	if err != nil {
		return nil, fmt.Errorf("stop value: %w", err)
	}
	mystop, ok := dstop.(*gdate.MonthYear)
	if !ok {
		return nil, fmt.Errorf("unsupported range")
	}

	if mystart.C != mystop.C {
		return nil, fmt.Errorf("unsupported range: mismatched calendars")
	}

	if mystart.Y == mystop.Y {
		if mystart.M == 1 && mystop.M == 3 {
			return &model.Date{
				Date: &gdate.YearQuarter{
					C: mystart.C,
					Y: mystart.Y,
					Q: 1,
				},
			}, nil
		} else if mystart.M == 4 && mystop.M == 6 {
			return &model.Date{
				Date: &gdate.YearQuarter{
					C: mystart.C,
					Y: mystart.Y,
					Q: 2,
				},
			}, nil
		} else if mystart.M == 7 && mystop.M == 9 {
			return &model.Date{
				Date: &gdate.YearQuarter{
					C: mystart.C,
					Y: mystart.Y,
					Q: 3,
				},
			}, nil
		} else if mystart.M == 10 && mystop.M == 12 {
			return &model.Date{
				Date: &gdate.YearQuarter{
					C: mystart.C,
					Y: mystart.Y,
					Q: 4,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported range")
}

func ParseDatespan(ds grampsxml.Datespan) (*model.Date, error) {
	return nil, fmt.Errorf("unsupported span")
}

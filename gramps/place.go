package gramps

import (
	"fmt"
	"strconv"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

func (l *Loader) populatePlaceFacts(m ModelFinder, gp *grampsxml.Placeobj) error {
	id := pval(gp.ID, gp.Handle)
	pl := m.FindPlace(l.ScopeName, id)

	logger := logging.With("id", pl.ID)
	logger.Debug("populating from place record", "handle", gp.Handle)

	switch gp.Type {
	case "Country":
		pl.PlaceType = model.PlaceTypeCountry
	case "Building":
		pl.PlaceType = model.PlaceTypeBuilding
	case "Street":
		pl.PlaceType = model.PlaceTypeStreet
	default:
		pl.PlaceType = model.PlaceTypeUnknown
	}

	if len(gp.Pname) == 0 {
		pl.PreferredName = "unknown"
		pl.PreferredFullName = "an unknown place"
		pl.PreferredUniqueName = "an unknown place"
		pl.PreferredSortName = "unknown place"
	} else {
		// TODO: support multiple place names
		pl.PreferredName = gp.Pname[0].Value
		pl.PreferredFullName = gp.Pname[0].Value
		pl.PreferredUniqueName = gp.Pname[0].Value
		pl.PreferredSortName = gp.Pname[0].Value
	}

	if gp.Coord != nil {
		var err error
		pl.Longitude, err = strconv.ParseFloat(gp.Coord.Long, 64)
		if err != nil {
			logger.Warn("could not parse longitude of place", "long", gp.Coord.Long)
		}

		pl.Latitude, err = strconv.ParseFloat(gp.Coord.Lat, 64)
		if err != nil {
			logger.Warn("could not parse latitude of place", "long", gp.Coord.Lat)
		}
	}

	if len(gp.Placeref) > 0 {
		for _, pr := range gp.Placeref {
			paro, ok := l.PlacesByHandle[pr.Hlink]
			if !ok {
				continue
			}
			if !l.populatedPlaces[paro.Handle] {
				if err := l.populatePlaceFacts(m, paro); err != nil {
					return fmt.Errorf("populate parent place: %w", err)
				}
			}
			po := m.FindPlace(l.ScopeName, pval(paro.ID, paro.Handle))
			if !po.IsUnknown() {
				pl.Parent = po
				if po.PreferredFullName != "" {
					pl.PreferredFullName += ", " + po.PreferredFullName
				}
				if po.PreferredUniqueName != "" {
					if po.PlaceType == model.PlaceTypeStreet || po.PlaceType == model.PlaceTypeBuilding {
						pl.PreferredUniqueName += ", " + po.PreferredUniqueName
					} else {
						pl.PreferredUniqueName += ", " + po.PreferredName
					}
				}
				if po.PreferredSortName != "" {
					pl.PreferredSortName = po.PreferredSortName + ", " + pl.PreferredSortName
				}
			}
			break
		}
	}

	l.populatedPlaces[gp.Handle] = true

	return nil
}

func (l *Loader) populatePlaceHierarchy(m ModelFinder, gp *grampsxml.Placeobj) error {
	return nil
}

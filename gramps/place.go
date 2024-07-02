package gramps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/grampsxml"
)

func (l *Loader) populatePlaceFacts(m ModelFinder, gp *grampsxml.Placeobj) error {
	id := pval(gp.ID, gp.Handle)
	pl := m.FindPlace(l.ScopeName, id)

	logger := logging.With("id", pl.ID)
	logger.Debug("populating from place record", "handle", gp.Handle)

	switch strings.ToLower(gp.Type) {
	case "country":
		pl.PlaceType = model.PlaceTypeCountry
	case "building":
		pl.PlaceType = model.PlaceTypeBuilding
	case "street":
		pl.PlaceType = model.PlaceTypeStreet
	case "ship":
		pl.PlaceType = model.PlaceTypeShip
	case "category":
		pl.PlaceType = model.PlaceTypeCategory
	case "register office":
		pl.PlaceType = model.PlaceTypeBuilding
		pl.BuildingKind = model.BuildingKindRegisterOffice
		pl.Singular = true
	case "workhouse":
		pl.PlaceType = model.PlaceTypeBuilding
		pl.BuildingKind = model.BuildingKindWorkhouse
		pl.Singular = true
	case "church":
		pl.PlaceType = model.PlaceTypeBuilding
		pl.BuildingKind = model.BuildingKindChurch
	case "farm":
		pl.PlaceType = model.PlaceTypeBuilding
		pl.BuildingKind = model.BuildingKindFarm
	case "cemetery":
		pl.PlaceType = model.PlaceTypeBurialGround
	case "hospital":
		pl.PlaceType = model.PlaceTypeBuilding
		pl.BuildingKind = model.BuildingKindHospital
	case "parish":
		pl.PlaceType = model.PlaceTypeParish
	case "ancient county", "county":
		pl.PlaceType = model.PlaceTypeCounty
	case "city":
		pl.PlaceType = model.PlaceTypeCity
	case "town":
		pl.PlaceType = model.PlaceTypeTown
	case "village":
		pl.PlaceType = model.PlaceTypeVillage
	case "hamlet":
		pl.PlaceType = model.PlaceTypeHamlet
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
		name := strings.TrimSpace(gp.Pname[0].Value)

		if pl.PlaceType == model.PlaceTypeBuilding {
			if text.StartsWithNumeral(name) {
				pl.Numbered = true
			} else {
				if strings.HasPrefix(strings.ToLower(name), "the ") {
					pl.Singular = true
				} else if pl.Singular {
					name = "The " + name
				}
			}
		}

		pl.PreferredName = name
		pl.PreferredFullName = name
		pl.PreferredUniqueName = name
		pl.PreferredSortName = name
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

	// Add notes
	for _, nr := range gp.Noteref {
		gn, ok := l.NotesByHandle[nr.Hlink]
		if !ok {
			continue
		}
		if pval(gn.Priv, false) {
			logger.Debug("skipping place note marked as private", "handle", gn.Handle)
			continue
		}

		switch strings.ToLower(gn.Type) {
		case "place note":
			pl.Comments = append(pl.Comments, l.parseNote(gn, m))
		case "research":
			// research notes are always assumed to be markdown
			t := l.parseNote(gn, m)
			t.Markdown = true
			pl.ResearchNotes = append(pl.ResearchNotes, t)
		default:
			// ignore note
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
			if !po.IsUnknown() && pl.Numbered && pl.PlaceType == model.PlaceTypeBuilding && (po.PlaceType == model.PlaceTypeStreet || po.PlaceType == model.PlaceTypeBuilding) {
				// combine into a single place
				if po.PreferredName != "" {
					pl.PreferredName += " " + po.PreferredName
				}
				if po.PreferredFullName != "" {
					pl.PreferredFullName += " " + po.PreferredFullName
				}
				if po.PreferredUniqueName != "" {
					if po.PlaceType == model.PlaceTypeStreet || po.PlaceType == model.PlaceTypeBuilding {
						pl.PreferredUniqueName += " " + po.PreferredUniqueName
					} else {
						pl.PreferredUniqueName += " " + po.PreferredName
					}
				}
				if po.PreferredSortName != "" {
					pl.PreferredSortName = po.PreferredSortName + " " + pl.PreferredSortName
				}
				po = po.Parent
			}

			if !po.IsUnknown() && po.PlaceType != model.PlaceTypeCategory {
				connector := ", "
				if pl.Singular {
					connector = " " + po.InAt() + " "
				}

				pl.Parent = po
				if po.PreferredFullName != "" {
					pl.PreferredFullName += connector + po.PreferredFullName
				}
				if po.PreferredUniqueName != "" {
					if po.PlaceType == model.PlaceTypeStreet || po.PlaceType == model.PlaceTypeBuilding {
						pl.PreferredUniqueName += connector + po.PreferredUniqueName
					} else {
						pl.PreferredUniqueName += connector + po.PreferredName
					}
				}
				if po.PreferredSortName != "" {
					pl.PreferredSortName = po.PreferredSortName + ", " + pl.PreferredSortName
				}
			}
			break
		}
	}

	// add media objects
	for _, gor := range gp.Objref {
		if pval(gor.Priv, false) {
			logger.Debug("skipping media object marked as private", "handle", gor.Hlink)
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

			pl.Gallery = append(pl.Gallery, cmo)
		}
	}

	l.populatedPlaces[gp.Handle] = true

	return nil
}

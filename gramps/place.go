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

	changeTime, err := changeToTime(gp.Change)
	if err == nil {
		pl.LastUpdated = &changeTime
	}

	logger := logging.With("id", pl.ID)
	logger.Debug("populating from place record", "handle", gp.Handle)
	pl.Name += "@@" + gp.Type + "@@"

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
	case "registration district":
		pl.PlaceType = model.PlaceTypeRegistrationDistrict
	case "prestegjeld":
		// Norwegian parish equivalent to a modern municipality
		pl.PlaceType = model.PlaceTypeParish
	case "kirkesokn":
		// Norwegian sub-parish
		pl.PlaceType = model.PlaceTypeParish
	default:
		pl.PlaceType = model.PlaceTypeUnknown
	}

	if len(gp.Pname) == 0 {
		pl.Name = model.UnknownPlaceName
		pl.FullName = model.UnknownPlaceName
		// pl.PreferredName = "unknown"
		// pl.PreferredFullName = "an unknown place"
		// pl.PreferredUniqueName = "an unknown place"
		pl.PreferredSortName = "unknown place"
		// pl.PreferredLocalityName = "unknown place"
		// pl.PreferredVerboseName = "an unknown place"
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

		pl.Name = name
		// pl.PreferredName = name
		// pl.PreferredFullName = name
		// pl.PreferredUniqueName = name
		pl.PreferredSortName = name
		// pl.PreferredLocalityName = name

		if strings.Contains(strings.ToLower(pl.Name), "church ") || strings.Contains(strings.ToLower(pl.Name), " church") {
			pl.BuildingKind = model.BuildingKindChurch
		}

		if pl.PlaceType == model.PlaceTypeRegistrationDistrict {
			pl.Name = strings.TrimSuffix(pl.Name, " Registration District")
			// pl.PreferredName = strings.TrimSuffix(pl.PreferredName, " Registration District")
		}

	}

	if gp.Coord != nil {
		var err error
		long, err := strconv.ParseFloat(gp.Coord.Long, 64)
		if err != nil {
			logger.Warn("could not parse longitude of place", "long", gp.Coord.Long)
		} else {
			lat, err := strconv.ParseFloat(gp.Coord.Lat, 64)
			if err != nil {
				logger.Warn("could not parse latitude of place", "long", gp.Coord.Lat)
			} else {
				pl.GeoLocation = &model.GeoLocation{
					Latitude:  lat,
					Longitude: long,
				}
			}
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

	// Enhance names with parent context
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
			parent := m.FindPlace(l.ScopeName, pval(paro.ID, paro.Handle))

			// handle buildings or streets
			if !parent.IsUnknown() && pl.Numbered && pl.PlaceType == model.PlaceTypeBuilding && (parent.PlaceType == model.PlaceTypeStreet || parent.PlaceType == model.PlaceTypeBuilding) {
				// combine into a single place
				if parent.Name != "" {
					pl.Name += " " + parent.Name
				}
				// if parent.PreferredFullName != "" {
				// 	pl.PreferredFullName += " " + parent.PreferredName
				// }
				// if parent.PreferredUniqueName != "" {
				// 	if parent.PlaceType == model.PlaceTypeStreet || parent.PlaceType == model.PlaceTypeBuilding {
				// 		pl.PreferredUniqueName += " " + parent.PreferredName
				// 	} else {
				// 		pl.PreferredUniqueName += " " + parent.PreferredName
				// 	}
				// }
				if parent.PreferredSortName != "" {
					pl.PreferredSortName = parent.PreferredSortName + " " + pl.PreferredSortName
				}
				parent = parent.Parent
			}

			if !parent.IsUnknown() && parent.PlaceType != model.PlaceTypeCategory {
				connector := ", "
				if pl.Singular {
					connector = " " + parent.InAt() + " "
				}

				pl.Parent = parent
				if parent.FullName != "" {
					pl.FullName += connector + parent.FullName
				}
				// if parent.PreferredUniqueName != "" {
				// 	if parent.PlaceType == model.PlaceTypeStreet || parent.PlaceType == model.PlaceTypeBuilding {
				// 		pl.PreferredUniqueName += connector + parent.PreferredUniqueName
				// 	} else {
				// 		pl.PreferredUniqueName += connector + parent.PreferredName
				// 	}
				// }
				if parent.PreferredSortName != "" {
					pl.PreferredSortName = parent.PreferredSortName + ", " + pl.PreferredSortName
				}

				// if pl.PlaceType == model.PlaceTypeStreet || pl.PlaceType == model.PlaceTypeBuilding {
				// 	pl.PreferredLocalityName = parent.PreferredLocalityName
				// } else {
				// 	pl.PreferredLocalityName += connector + parent.Name
				// }
			}

			if pl.GeoLocation == nil && parent.GeoLocation != nil {
				pl.GeoLocation = parent.GeoLocation
			}

			break
		}
	}

	// // Build verbose name
	// // TODO: expand to cover more cases
	// pl.PreferredVerboseName = pl.Name

	// parent := pl.Parent
	// for parent != nil {
	// 	switch parent.PlaceType {
	// 	case model.PlaceTypeParish:
	// 		pl.PreferredVerboseName += " in the parish of " + parent.Name
	// 	case model.PlaceTypeRegistrationDistrict:
	// 		pl.PreferredVerboseName += " in the registration district of " + parent.Name
	// 	case model.PlaceTypeCounty:
	// 		pl.PreferredVerboseName += " in " + parent.Name
	// 	default:
	// 		pl.PreferredVerboseName += ", " + parent.Name
	// 	}

	// 	parent = parent.Parent
	// }

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

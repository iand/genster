package gedcom

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/iand/gdate"
	"github.com/iand/gedcom"
	"github.com/iand/genster/identifier"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/place"
	"github.com/iand/genster/tree"
)

var _ = logging.Debug

var startsWithNumber = regexp.MustCompile(`^[1-9]`)

type ModelFinder interface {
	FindPerson(scope string, id string) *model.Person
	FindSource(scope string, id string) *model.Source
	FindPlaceUnstructured(name string, hints ...place.Hint) *model.Place
	AddAlias(alias string, canonical string)
}

type Loader struct {
	ScopeName           string
	Gedcom              *gedcom.Gedcom
	Attrs               map[string]string
	Citations           map[string]*model.GeneralCitation
	Tags                map[string]string
	SourcesByAPID       map[string]*model.Source
	SourceRecordsByXref map[string]*gedcom.SourceRecord
	MediaRecordsByXref  map[string]*gedcom.MediaRecord
}

func NewLoader(filename string) (*Loader, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("open gedcom file: %w", err)
	}

	d := gedcom.NewDecoder(bytes.NewReader(data))

	g, err := d.Decode()
	if err != nil {
		return nil, fmt.Errorf("decode gedcom: %w", err)
	}
	sort.SliceStable(g.Source, func(a, b int) bool { return g.Source[a].Xref < g.Source[b].Xref })
	sort.SliceStable(g.Repository, func(a, b int) bool { return g.Repository[a].Xref < g.Repository[b].Xref })
	sort.SliceStable(g.Individual, func(a, b int) bool { return g.Individual[a].Xref < g.Individual[b].Xref })
	sort.SliceStable(g.Family, func(a, b int) bool { return g.Family[a].Xref < g.Family[b].Xref })
	sort.SliceStable(g.Media, func(a, b int) bool { return g.Media[a].Xref < g.Media[b].Xref })

	l := &Loader{
		Gedcom:              g,
		Attrs:               make(map[string]string),
		ScopeName:           filename,
		Citations:           make(map[string]*model.GeneralCitation),
		Tags:                make(map[string]string),
		SourcesByAPID:       make(map[string]*model.Source),
		SourceRecordsByXref: make(map[string]*gedcom.SourceRecord),
		MediaRecordsByXref:  make(map[string]*gedcom.MediaRecord),
	}
	l.readAttrs()
	l.readTags()

	if id, ok := l.Attrs["ANCESTRY_TREE_ID"]; ok {
		l.ScopeName = fmt.Sprintf("ANCESTRY_TREE_%s", id)
	}

	return l, nil
}

func (l *Loader) Scope() string {
	return l.ScopeName
}

func (l *Loader) readAttrs() error {
	// Look for an ancestry tree identifier
	if l.Gedcom.Header.SourceSystem.BusinessName == "Ancestry.com" {
		for _, hud := range l.Gedcom.Header.SourceSystem.UserDefined {
			if hud.Tag == "_TREE" {
				if hud.Value != "" {
					l.Attrs["ANCESTRY_TREE_NAME"] = hud.Value
				}
				for _, tud := range hud.UserDefined {
					if tud.Value != "" {
						switch tud.Tag {
						case "RIN":
							l.Attrs["ANCESTRY_TREE_ID"] = tud.Value
						case "NOTE":
							l.Attrs["ANCESTRY_TREE_NOTE"] = tud.Value
						}
					}
				}
			}
		}
	}

	return nil
}

func (l *Loader) readTags() error {
	// Look for ancestry style tags using _MTTAG
	for _, ud := range l.Gedcom.UserDefined {
		if ud.Tag != "_MTTAG" {
			continue
		}
		for _, uds := range ud.UserDefined {
			if uds.Tag == "NAME" {
				l.Tags[ud.Xref] = uds.Value
				break
			}
		}
	}

	return nil
}

func (l *Loader) Load(t *tree.Tree) error {
	if name, ok := l.Attrs["ANCESTRY_TREE_NAME"]; ok {
		t.Name = name
	}
	if desc, ok := l.Attrs["ANCESTRY_TREE_NOTE"]; ok {
		t.Description = desc
	}

	for _, mr := range l.Gedcom.Media {
		if err := l.populateMediaFacts(t, mr); err != nil {
			return fmt.Errorf("media: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d media records", len(l.Gedcom.Media)))

	for _, sr := range l.Gedcom.Source {
		if err := l.populateSourceFacts(t, sr); err != nil {
			return fmt.Errorf("source: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d source records", len(l.Gedcom.Source)))

	for _, in := range l.Gedcom.Individual {
		if err := l.populatePersonFacts(t, in); err != nil {
			return fmt.Errorf("person: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d individual records", len(l.Gedcom.Individual)))

	for _, fr := range l.Gedcom.Family {
		if err := l.populateFamilyFacts(t, fr); err != nil {
			return fmt.Errorf("family: %w", err)
		}
	}

	logging.Info(fmt.Sprintf("loaded %d family records", len(l.Gedcom.Family)))

	for _, p := range t.People {
		l.buildFamilies(t, p)
	}

	return nil
}

func (l *Loader) findPlaceForEvent(m ModelFinder, er *gedcom.EventRecord) (*model.Place, []*model.Anomaly) {
	var name string
	var anomalies []*model.Anomaly
	if len(er.Address.Address) > 0 && (er.Address.Address[0].Full != "" || er.Address.Address[0].Line1 != "") {
		// just use first address for now
		// TODO: handle multiple addresses
		full := er.Address.Address[0].Full

		// TODO: use address structure as hint for finding place

		if full == "" {
			a := er.Address.Address[0]
			comma := func(s string) string {
				if s == "" {
					return s
				}
				return s + ", "
			}
			name = comma(a.Line1) + comma(a.Line2) + comma(a.Line3) + comma(a.City) + comma(a.State) + comma(a.PostalCode) + comma(a.Country)
		} else {
			name = strings.ReplaceAll(full, "\n", ", ")
		}
	} else {
		name = er.Place.Name
	}

	if name == "" {
		return model.UnknownPlace(), nil
	} else {
		// if _, country := place.LookupPlaceName(name); !country {
		// 	if !strings.Contains(name, ",") {
		// 		anomalies = append(anomalies, &model.Anomaly{
		// 			Category: "Name",
		// 			Text:     fmt.Sprintf("Place name does not appear to be structured: %q", name),
		// 			Context:  "Place in event",
		// 		})
		// 	}
		// }

		if reUppercase.MatchString(name) {
			anomalies = append(anomalies, &model.Anomaly{
				Category: "Name",
				Text:     fmt.Sprintf("Place name is all uppercase, should change to proper case: %q", name),
				Context:  "Place in event",
			})
		}
		// if na me == "Newport Market, Glamorgan, Gwent, Monmouthshire, United Kingdom" {
		// 	anomalies = append(anomalies, &model.Anomaly{
		// 		Category: "Name",
		// 		Text:     fmt.Sprintf("Place name should be Newport, Monmouthshire, England (ancestry database incorrectly links Newport M with wrong place): %q", name),
		// 		Context:  "Place in event",
		// 	})
		// }

		pl := m.FindPlaceUnstructured(name)

		if startsWithNumber.MatchString(name) {
			pl.PlaceType = model.PlaceTypeAddress
		}

		c := pl.CountryName
		if c.IsUnknown() {
			anomalies = append(anomalies, &model.Anomaly{
				Category: "Name",
				Text:     fmt.Sprintf("Place name does not include a country: %q", name),
				Context:  "Place in event",
			})
		} else if c.Name == "United Kingdom" && pl.UKNationName == nil {
			// This is just my personal preference
			anomalies = append(anomalies, &model.Anomaly{
				Category: "Name",
				Text:     fmt.Sprintf("Place name has United Kingdom as country, change to use England, Scotland or Wales: %q", name),
				Context:  "Place in event",
			})
		}

		return pl, anomalies
	}
}

func (l *Loader) parseCitationRecords(m ModelFinder, crs []*gedcom.CitationRecord, logger *slog.Logger) ([]*model.GeneralCitation, []*model.Anomaly) {
	cits := make([]*model.GeneralCitation, 0)
	anomalies := make([]*model.Anomaly, 0)
	for _, cr := range crs {
		pc, err := l.parseCitation(m, cr, logger)
		if err != nil {
			anomalies = append(anomalies, &model.Anomaly{
				Category: "GEDCOM",
				Text:     err.Error(),
				Context:  "Citation",
			})
			logger.Warn("skipping citation with no source", "error", err.Error())
			continue
		}
		cits = append(cits, pc)
	}
	return cits, anomalies
}

// 2024-02-26 Ancestry have replaced _APID in family marriage ciations into separate
// idenfifiers for the husband's citation and the wife's citation
//
// An example of the new IDs
//
//	0 @F4@ FAM
//	1 HUSB @I192525701960@
//	1 WIFE @I192525702042@
//	1 CHIL @I192525702043@
//	1 CHIL @I192525702044@
//	1 MARR
//	2 DATE 23 Oct 1836
//	2 PLAC Huddersfield, Yorkshire, England
//	2 SOUR @S556044042@
//	3 PAGE West Yorkshire Archive Service; Leeds, Yorkshire, England; Reference Number: WDP32/31
//	3 _HPID 1,2253::5644507
//	3 _WPID 1,2253::16776616
//
// The source is the marriage register:
//
// The citation view for the husband (Charles Spencer):
// https://www.ancestry.co.uk/discoveryui-content/view/5644507:2253
//
// The citation view for the wife (Mary Ann Hayley):
// https://www.ancestry.co.uk/discoveryui-content/view/16776616:2253
//
// Previously this was
//
//	0 @F4@ FAM
//	1 HUSB @I192525701960@
//	1 WIFE @I192525702042@
//	1 CHIL @I192525702043@
//	1 CHIL @I192525702044@
//	1 MARR
//	2 DATE 23 Oct 1836
//	2 PLAC Huddersfield, Yorkshire, England
//	2 SOUR @S556044042@
//	3 PAGE West Yorkshire Archive Service; Leeds, Yorkshire, England; Reference Number: WDP32/31
//	3 _APID 1,2253::5644507
func (l *Loader) parseCitation(m ModelFinder, cr *gedcom.CitationRecord, logger *slog.Logger) (*model.GeneralCitation, error) {
	type scope struct {
		name string
		id   string
	}

	var sourceScope scope
	var citationScope scope

	// Find the most specific identifying informaion that can be used to lookup or generate the citation id
	if cr.Source != nil && cr.Source.Xref != "" {
		sourceScope.name = l.ScopeName
		sourceScope.id = cr.Source.Xref

		citationScope.name = cr.Source.Xref
		citationScope.id = cr.Page
	} else if cr.Page != "" {
		citationScope.name = "Page"
		citationScope.id = cr.Page
	}

	// Look for an id that indicates a shared citation
	ud, found := findFirstUserDefinedTag("_APID", cr.UserDefined)
	if found && ud.Value != "" {
		citationScope.name = "_APID"
		citationScope.id = ud.Value
	} else {
		udh, found := findFirstUserDefinedTag("_HPID", cr.UserDefined)
		if found && udh.Value != "" {
			citationScope.name = "_HPID"
			citationScope.id = udh.Value
		} else {
			udw, found := findFirstUserDefinedTag("_WPID", cr.UserDefined)
			if found && udw.Value != "" {
				citationScope.name = "_WPID"
				citationScope.id = udw.Value
			}
		}
	}

	// If no source ID given in gedcom but an ancestry ID was found then use that as source
	// identifier
	if sourceScope.id == "" && citationScope.id != "" {
		matches := reApid.FindStringSubmatch(citationScope.id)
		if len(matches) > 1 {
			sourceScope.name = "_APID"
			sourceScope.id = fmt.Sprintf("1,%s::0", matches[1])
		}

	}

	if citationScope.name == "" || citationScope.id == "" {
		return nil, fmt.Errorf("no citation information found to generate id")
	}
	id := identifier.New(citationScope.name, citationScope.id)

	cit, ok := l.Citations[id]
	if ok {
		return cit, nil
	}

	cit = &model.GeneralCitation{
		ID:     id,
		Detail: cr.Page,
	}

	cit.Detail = cleanCitationDetail(cit.Detail)

	if sourceScope.name != "" && sourceScope.id != "" {
		cit.Source = m.FindSource(sourceScope.name, sourceScope.id)
	} else {
		// ensure we have some citation text
		if cit.Detail == "" {
			cit.Detail = "unknown source"
		}
		logger.Warn("no source found for citation", "source_name", sourceScope.name, "source_id", sourceScope.id, "cit_id", cit.ID, "detail", cit.Detail)
	}

	if cr.Data.Date != "" {
		dt, err := gdate.Parse(cr.Data.Date)
		if err == nil {
			cit.TranscriptionDate = &model.Date{Date: dt}
		}
	}

	for _, ct := range cr.Data.Text {
		cit.TranscriptionText = append(cit.TranscriptionText, model.Text{Text: ct})
	}

	wwws := findUserDefinedTags(cr.Data.UserDefined, "WWW", false)
	if len(wwws) > 0 {
		cit.URL = parseURL(wwws[0].Value)
	}

	// for _, mr := range cr.Media {
	// }

	// 1 OBJE @O109@
	// 2 _PRIM Y
	// 2 _CROP
	// 3 _LEFT 50
	// 3 _TOP 62
	// 3 _WDTH 302
	// 3 _HGHT 302
	// 3 _TYPE primary

	// 0 @O90@ OBJE
	// 1 FILE
	// 2 FORM jpg
	// 3 TYPE image
	// 3 _MTYPE document
	// 3 _STYPE png
	// 3 _SIZE 594731
	// 3 _WDTH 2666
	// 3 _HGHT 834
	// 2 TITL MarriageOfMatthewHallAndMaryMiller1844
	// 1 RIN 4cb4264f-b059-4f66-8828-920e44280d75
	// 1 DATE 9 Nov 1844
	// 1 _META <metadataxml><transcription></transcription></metadataxml>
	// 1 _CREA 2021-06-02 00:14:49.000
	// 1 _USER Mqz7bRWFFpfgkHH9MUIp4WnvYJqZK602KqxfoyNVPvdfi/2brGe/0qZpGfG2hI8OaEURgDqW4KV0lOSrZ0uZNw==
	// 2 _ENCR 1
	// 1 _ORIG u
	// 1 _ATL N

	// Note PNG iin following:

	// 0 @O258@ OBJE
	// 1 FILE
	// 2 FORM jpg
	// 3 TYPE image
	// 3 _MTYPE document
	// 3 _STYPE png
	// 3 _SIZE 282616
	// 3 _WDTH 1947
	// 3 _HGHT 1339
	// 2 TITL MarriageOfJamesBrightenAndRebeccaPritty1794
	// 1 RIN e3f657e8-cf55-4cf0-bb7f-662d8a33cd55
	// 1 DATE 17 Sep 1794
	// 1 _META <metadataxml><transcription>Marriage by banns at Brockdish, Norfolk on 17th Sep 1794 of James Brighten and Rebecca Pritty, single woman. He could not sign his name but she could. The witnesses were John King and Thomas Pritty, both of whom could si
	// 2 CONC gn their names</transcription></metadataxml>
	// 1 _CREA 2021-04-08 19:45:24.000
	// 1 _USER zm9sFpl23GrxS8/rol5sKLkEqP7R/cGzVSv5CtVaXqD4QtVxrco5kt7Xi+3H63CY/Or98z5MJG8/dSXtE65qRw==
	// 2 _ENCR 1
	// 1 _ORIG u
	// 1 _ATL N

	l.Citations[id] = cit
	// Source      *SourceRecord
	// Page        string
	// Data        DataRecord
	// Quay        string
	// Media       []*MediaRecord
	// Note        []*NoteRecord
	// UserDefined []UserDefinedTag

	return cit, nil
}

// cleanCitationDetail removes some redundant information that isn't necessary when a source is included
func cleanCitationDetail(page string) string {
	page = strings.TrimPrefix(page, "The National Archives of the UK (TNA); Kew, Surrey, England; Census Returns of England and Wales, 1891;")
	page = strings.TrimPrefix(page, "The National Archives; Kew, London, England; 1871 England Census; ")
	return page
}

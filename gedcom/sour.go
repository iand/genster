package gedcom

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iand/gedcom"
	"github.com/iand/genster/identifier"
	"github.com/iand/genster/logging"
)

var reApid = regexp.MustCompile(`^\d+,(\d+)::`)

func (l *Loader) populateSourceFacts(m ModelFinder, sr *gedcom.SourceRecord) error {
	scope := l.ScopeName
	scopeID := sr.Xref

	ud, found := findFirstUserDefinedTag("_APID", sr.UserDefined)
	if found && ud.Value != "" {
		scope = "_APID"
		scopeID = ud.Value
		logging.Debug("populating from source record using _APID", "apid", ud.Value)
	} else {
		logging.Debug("populating from source record using Xref", "xref", sr.Xref)
	}
	so := m.FindSource(scope, scopeID)
	so.Title = strings.TrimSpace(sr.Title)

	logger := logging.With("id", so.ID)
	if scope == "_APID" {
		matches := reApid.FindStringSubmatch(ud.Value)
		if len(matches) > 1 {
			so.SearchLink = fmt.Sprintf("https://www.ancestry.com/search/collections/%s/", matches[1])
		}

		alias := identifier.New(l.ScopeName, sr.Xref)
		m.AddAlias(alias, so.ID)
		logger.Debug("adding source alias", "alias", alias)
	}

	if so.Title == "" {
		logger.Warn("source has empty title", "xref", sr.Xref)
		so.Title = "Unknown Source"
	}

	if sr.Repository != nil && sr.Repository.Repository != nil {
		rr := sr.Repository.Repository
		so.RepositoryName = rr.Name

		if len(rr.Address.WWW) > 0 {
			so.RepositoryLink = rr.Address.WWW[0]
		}

		if so.RepositoryLink == "" && len(rr.Note) > 0 {
			for _, n := range rr.Note {
				if n == nil {
					continue
				}
				if strings.HasPrefix(n.Note, "https://") || strings.HasPrefix(n.Note, "http://") {
					so.RepositoryLink = n.Note
					break
				}
			}
		}

	}

	return nil
}

package place

type PlaceHierarchy struct {
	parts []string
}

func (p *PlaceHierarchy) Last() (string, bool) {
	if len(p.parts) == 0 {
		return "", false
	}
	return p.parts[len(p.parts)-1], true
}

func (p *PlaceHierarchy) CutLast() (*PlaceHierarchy, string) {
	if len(p.parts) == 0 {
		return nil, ""
	}
	if len(p.parts) == 1 {
		return nil, p.parts[len(p.parts)-1]
	}
	return &PlaceHierarchy{parts: p.parts[:len(p.parts)-1]}, p.parts[len(p.parts)-1]
}

func ClassifyName(name string, hints ...Hint) *PlaceName {
	pn, ok := LookupPlaceName(name)
	if ok {
		return pn
	}

	parts := splitPlaceName(name)
	if len(parts) == 0 {
		return UnknownPlaceName()
	}

	pn = &PlaceName{
		Name: Clean(name),
		Kind: PlaceKindUnknown,
	}

	for i := len(parts) - 1; i >= 0; i-- {
		po, ok := LookupPlaceName(parts[i])
		if ok {
			pn.PartOf = append(pn.PartOf, po)
		}
	}

	return pn
}

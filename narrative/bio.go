package narrative

import (
	"sort"
	"strconv"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

// Bio returns a plain-text short biography of p.
func Bio(p *model.Person) string {
	return bio(p)
}

// bio assembles the biography as a sequence of terse fragment sentences in the
// style of a biographical dictionary entry: the most distinctive fact, then
// birth, occupation and residence, marriages and children in order, the notable
// events of the life, and finally death. The subject's name is the heading and
// never appears in the body, so the prose carries no pronouns.
func bio(p *model.Person) string {
	b := &bioBuilder{p: p}
	b.addNotable()
	b.addBirth()
	b.addParentage()
	b.addOrphaned()
	b.addAltSurname()
	b.addNameChange()
	b.addOccupation()
	b.addMilitary()
	b.addResidence()
	b.addFamily()
	b.addCrime()
	b.addTravel()
	b.addHardship()
	b.addDeath()
	b.addCauseOfDeath()
	return strings.Join(b.frags, " ")
}

// bioBuilder accumulates the fragment sentences of a single biography.
type bioBuilder struct {
	p           *model.Person
	frags       []string
	widowCount  int      // number of widowhoods reported, so the second reads "widowed again"
	firstPlace  []string // components of the first full place shown, used to elide repeated trailing parts
	lastTown    string   // town of the most recent place mentioned, used to drop a redundant marriage place
	lastCountry string   // country of the most recent place mentioned, used to decide how much of a marriage place to give
}

// placeWhere renders a full place (birth or death) for prose. The first full
// place is shown complete; each later place has the trailing components it
// shares with the first removed (always keeping its locality), so a repeated
// country is stated once and a place unchanged since birth collapses to its
// locality.
func (b *bioBuilder) placeWhere(pl *model.Place) string {
	b.noteLocation(pl)
	parts := placeParts(pl.FullName)
	if len(b.firstPlace) == 0 {
		b.firstPlace = parts
		return pl.Where()
	}
	return pl.InAt() + " " + strings.Join(elideCommonSuffix(parts, b.firstPlace), ", ")
}

// marriagePlace returns the place clause for a marriage, or the empty string
// when the marriage was in the town already mentioned. A marriage elsewhere is
// given by its town alone, unless the country has changed, in which case the
// town, region and country are given so a move abroad reads in full.
func (b *bioBuilder) marriagePlace(pl *model.Place) string {
	if pl == nil || pl.IsUnknown() {
		return ""
	}
	town := placeTown(pl)
	if town == "" {
		return ""
	}
	if strings.EqualFold(town, b.lastTown) {
		return ""
	}

	country := placeCountryName(pl)
	changed := b.lastCountry != "" && country != "" && !strings.EqualFold(country, b.lastCountry)
	b.lastTown = town
	if country != "" {
		b.lastCountry = country
	}

	if !changed {
		return "in " + town
	}
	locale := []string{town}
	if region := placeRegionName(pl); region != "" {
		locale = append(locale, region)
	}
	locale = append(locale, country)
	return "in " + strings.Join(locale, ", ")
}

// noteLocation records the town and country of a place as the running location,
// against which a later marriage place is judged.
func (b *bioBuilder) noteLocation(pl *model.Place) {
	if town := placeTown(pl); town != "" {
		b.lastTown = town
	}
	if country := placeCountryName(pl); country != "" {
		b.lastCountry = country
	}
}

// placeTown returns the town or parish of a place, the locality a marriage is
// reported at, falling back to the bare place name.
func placeTown(pl *model.Place) string {
	if pl.District != nil && pl.District.Name != "" {
		return pl.District.Name
	}
	return pl.Name
}

// placeRegionName returns the region (county or state) of a place, or the empty
// string when it is not known.
func placeRegionName(pl *model.Place) string {
	if pl.Region != nil && pl.Region.Name != "" {
		return pl.Region.Name
	}
	return ""
}

// placeCountryName returns the country of a place, or the empty string when it
// is not known.
func placeCountryName(pl *model.Place) string {
	if pl.Country != nil && pl.Country.Name != "" {
		return pl.Country.Name
	}
	return ""
}

// placeParts splits a full place name into its comma-separated components,
// trimming surrounding space and dropping empty components.
func placeParts(fullName string) []string {
	raw := strings.Split(fullName, ",")
	parts := make([]string, 0, len(raw))
	for _, p := range raw {
		if p = strings.TrimSpace(p); p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// elideCommonSuffix returns parts with the longest run of trailing components it
// shares with ref removed, always leaving at least the leading component.
func elideCommonSuffix(parts, ref []string) []string {
	n := 0
	for n < len(parts)-1 && n < len(ref) {
		if parts[len(parts)-1-n] != ref[len(ref)-1-n] {
			break
		}
		n++
	}
	return parts[:len(parts)-n]
}

// add appends fragment as a sentence, capitalising its first letter and giving
// it a full stop. Empty fragments are ignored.
func (b *bioBuilder) add(fragment string) {
	fragment = text.UpperFirst(fragment)
	if fragment == "" {
		return
	}
	if !strings.HasSuffix(fragment, ".") {
		fragment += "."
	}
	b.frags = append(b.frags, fragment)
}

// addNotable leads the biography with the curated notable fact, stated as a
// fragment without embellishment.
func (b *bioBuilder) addNotable() {
	n := strings.TrimSpace(b.p.Notable)
	if n == "" {
		return
	}
	if _, ok := cutPrefixFold(n, "changed name from "); ok {
		return // rendered as a name change instead
	}
	b.add(n)
}

// addNameChange notes a recorded change of name, drawn from the notable fact of
// the form "Changed name from X to Y".
func (b *bioBuilder) addNameChange() {
	rest, ok := cutPrefixFold(b.p.Notable, "changed name from ")
	if !ok {
		return
	}
	from, to, ok := strings.Cut(rest, " to ")
	if !ok || from == "" || to == "" {
		return
	}
	b.add("took the name " + to + " in place of " + from)
}

// cutPrefixFold returns the remainder of s after a case-insensitive prefix, and
// whether the prefix was present.
func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return "", false
	}
	return s[len(prefix):], true
}

// addAltSurname notes any further surnames the subject was recorded under, drawn
// from a slash-joined dual surname.
func (b *bioBuilder) addAltSurname() {
	_, rest, ok := strings.Cut(b.p.PreferredFamilyName, "/")
	if !ok || rest == "" {
		return
	}
	alts := strings.Split(rest, "/")
	noun := "surname"
	if len(alts) > 1 {
		noun = "surnames"
	}
	b.add("also recorded under the " + noun + " " + text.JoinList(alts))
}

func (b *bioBuilder) addBirth() {
	twin := b.isTwin()
	ev := b.p.BestBirthlikeEvent
	var date *model.Date
	var place *model.Place
	if ev != nil {
		date = ev.GetDate()
		place = ev.GetPlace()
	}
	hasDate := date != nil && !date.IsUnknown()
	hasPlace := place != nil && !place.IsUnknown()

	if hasDate || hasPlace {
		verb := "born"
		if _, ok := ev.(*model.BaptismEvent); ok {
			verb = "baptised"
		}
		parts := []string{verb}
		if hasDate {
			parts = append(parts, dateBare(date))
		}
		if hasPlace {
			parts = append(parts, b.placeWhere(place))
		}
		frag := strings.Join(parts, " ")
		if twin {
			frag += ", a twin"
		}
		b.add(frag)
	} else if twin {
		b.add("born a twin")
	}

	if b.p.Illegitimate && b.fatherUnknown() {
		b.add("father unknown")
	}
}

// isTwin reports whether the subject is recorded as a twin, by the flag or an
// association.
func (b *bioBuilder) isTwin() bool {
	if b.p.Twin {
		return true
	}
	for _, a := range b.p.Associations {
		if a.Kind == model.AssociationKindTwin {
			return true
		}
	}
	return false
}

// fatherUnknown reports whether no father is recorded for the subject.
func (b *bioBuilder) fatherUnknown() bool {
	return b.p.Father == nil || b.p.Father.IsUnknown()
}

// addParentage names the subject's parents, led by their position among their
// siblings when it is known, such as the eldest son or an only child. Nothing is
// stated when neither parent is named.
func (b *bioBuilder) addParentage() {
	father := parentName(b.p.Father)
	mother := parentName(b.p.Mother)
	if father == "" && mother == "" {
		return
	}

	rel := PositionInFamily(b.p)
	if rel == "" {
		rel = text.LowerFirst(b.p.Gender.RelationToParentNoun())
	}

	var parents string
	switch {
	case father != "" && mother != "":
		parents = father + " and " + mother
	case father != "":
		parents = father
	default:
		parents = mother
	}
	b.add(rel + " of " + parents)
}

// parentName returns a parent's full name for prose, or the empty string when no
// usable name is recorded.
func parentName(p *model.Person) string {
	if p == nil || p.IsUnknown() {
		return ""
	}
	if name := p.PreferredFullName; name != "" && !strings.Contains(name, model.UnknownNamePlaceholder) {
		return name
	}
	return ""
}

// orphanedBelowAge is the age below which the death of both parents is taken to
// have orphaned the subject in childhood.
const orphanedBelowAge = 16

// addOrphaned notes a subject both of whose recorded parents died during their
// childhood, stating the age reached when the second parent died.
func (b *bioBuilder) addOrphaned() {
	fatherAge, ok := b.ageAtParentDeath(b.p.Father)
	if !ok || fatherAge >= orphanedBelowAge {
		return
	}
	motherAge, ok := b.ageAtParentDeath(b.p.Mother)
	if !ok || motherAge >= orphanedBelowAge {
		return
	}
	b.add("orphaned " + AgeQualifier(max(fatherAge, motherAge)))
}

// ageAtParentDeath returns the subject's age when a parent died, and whether it
// could be determined, needing a known birth for the subject and a known death
// for the parent.
func (b *bioBuilder) ageAtParentDeath(parent *model.Person) (int, bool) {
	if parent == nil || parent.IsUnknown() {
		return 0, false
	}
	ev := parent.BestDeathlikeEvent
	if ev == nil {
		return 0, false
	}
	d := ev.GetDate()
	if d == nil || d.IsUnknown() {
		return 0, false
	}
	return b.p.AgeInYearsAt(d)
}

// subjectDeathKnown reports whether the subject has a known date of death, the
// evidence that they outlived a partner whose death ended a marriage.
func (b *bioBuilder) subjectDeathKnown() bool {
	ev := b.p.BestDeathlikeEvent
	if ev == nil {
		return false
	}
	d := ev.GetDate()
	return d != nil && !d.IsUnknown()
}

func (b *bioBuilder) addFamily() {
	unions := mergeUnions(b.unions())

	deathKnown := b.subjectDeathKnown()
	married := false
	emitted := false
	for i, u := range unions {
		phrase := b.unionPhrase(u)
		if phrase == "" {
			continue
		}
		b.add(phrase)
		emitted = true
		if u.married {
			married = true
		}
		widowConfirmed := deathKnown || i < len(unions)-1
		if ending := b.unionEnding(u, widowConfirmed); ending != "" {
			b.add(ending)
		}
	}

	hasChildren := len(b.p.Children) > 0
	if !emitted && hasChildren {
		b.add(text.CardinalWithUnit(len(b.p.Children), "child", "children"))
	}

	b.addMaritalStatus(married, hasChildren)
}

// addMaritalStatus notes a life certainly without marriage or children, drawn
// from the Unmarried and Childless flags and combined so they read as one
// statement. A flag is not repeated when the family itself already shows a
// marriage or children.
func (b *bioBuilder) addMaritalStatus(married, hasChildren bool) {
	unmarried := b.p.Unmarried && !married
	childless := b.p.Childless && !hasChildren
	switch {
	case unmarried && childless:
		b.add("never married and had no children")
	case unmarried:
		b.add("never married")
	case childless:
		b.add("had no children")
	}
}

// unionEnding describes how a marriage ended, as a fragment, or the empty string
// when there is nothing to report. Widowhood is reported only when widowConfirmed
// is set, meaning the subject is known to have outlived the partner.
func (b *bioBuilder) unionEnding(u union, widowConfirmed bool) string {
	if !u.married {
		return ""
	}
	switch u.endReason {
	case model.FamilyEndReasonDeath:
		if u.widowed && widowConfirmed {
			b.widowCount++
			if b.widowCount > 1 {
				return "widowed again"
			}
			return "widowed"
		}
	case model.FamilyEndReasonDivorce:
		return "divorced"
	case model.FamilyEndReasonAnulment:
		return "the marriage was annulled"
	}
	return ""
}

// unionPhrase renders one union as a fragment without the subject. An empty
// string means the union carries nothing worth stating.
func (b *bioBuilder) unionPhrase(u union) string {
	if u.married {
		p := "married"
		if u.partner != "" {
			p = text.JoinSentenceParts(p, u.partner)
		}
		if u.when != "" {
			p = text.JoinSentenceParts(p, u.when)
		}
		if where := b.marriagePlace(u.place); where != "" {
			p = text.JoinSentenceParts(p, where)
		}
		if u.children > 0 {
			p += "; " + text.CardinalWithUnit(u.children, "child", "children")
		}
		return p
	}

	if u.children == 0 {
		return ""
	}
	cc := text.CardinalWithUnit(u.children, "child", "children")
	if u.partner != "" {
		return cc + " with " + u.partner
	}
	if u.unknownParents > 1 {
		return cc + " by unknown " + b.otherParentNoun() + "s"
	}
	return cc + " by an unknown " + b.otherParentNoun()
}

// otherParentNoun returns the word for the parent of the subject's children who
// is not the subject, chosen from the subject's gender.
func (b *bioBuilder) otherParentNoun() string {
	switch {
	case b.p.Gender.IsFemale():
		return "father"
	case b.p.Gender.IsMale():
		return "mother"
	default:
		return "parent"
	}
}

func (b *bioBuilder) addOccupation() {
	desc := strings.TrimSpace(b.p.Epithet)
	if desc == "" {
		if o := b.primaryOccupation(); o != nil {
			desc = o.Name
		}
	}
	if desc == "" {
		return
	}
	b.add(desc)
}

// addResidence notes where the subject spent much of their life: a lifelong
// residence when every placed event falls in one place, otherwise a dominant
// district or county across their residence and census records.
func (b *bioBuilder) addResidence() {
	if place, ok := b.lifelongPlace(); ok {
		b.add("lifelong resident of " + place)
		b.lastTown = place
		return
	}

	districts := map[string]int{}
	regions := map[string]int{}
	total := 0
	for _, ev := range b.p.Timeline {
		switch ev.(type) {
		case *model.ResidenceRecordedEvent, *model.CensusEvent:
			pl := ev.GetPlace()
			if pl == nil || pl.IsUnknown() {
				continue
			}
			total++
			if pl.District != nil && pl.District.Name != "" {
				districts[pl.District.Name]++
			}
			if pl.Region != nil && pl.Region.Name != "" {
				regions[pl.Region.Name]++
			}
		}
	}
	if total < 3 {
		return
	}

	place := dominantPlace(districts, regions, total)
	if place == "" {
		return
	}
	b.add("resident of " + place + " for many years")
	b.lastTown = place
}

// lifelongPlace returns the district, or failing that the region, that a subject
// never left: every placed event falls within it and a placed birth and death
// bracket the life. The empty string and false mean no such place is evident.
func (b *bioBuilder) lifelongPlace() (string, bool) {
	birthPl := eventPlace(b.p.BestBirthlikeEvent)
	deathPl := eventPlace(b.p.BestDeathlikeEvent)
	if birthPl == nil || deathPl == nil {
		return "", false
	}

	districts := map[string]bool{}
	regions := map[string]bool{}
	placed := 0
	districtEverywhere := true
	regionEverywhere := true
	note := func(pl *model.Place) {
		if pl == nil || pl.IsUnknown() {
			return
		}
		placed++
		if d := placeDistrictName(pl); d != "" {
			districts[d] = true
		} else {
			districtEverywhere = false
		}
		if r := placeRegionName(pl); r != "" {
			regions[r] = true
		} else {
			regionEverywhere = false
		}
	}

	note(birthPl)
	note(deathPl)
	for _, ev := range b.p.Timeline {
		if ev == b.p.BestBirthlikeEvent || ev == b.p.BestDeathlikeEvent {
			continue
		}
		note(ev.GetPlace())
	}

	if placed < 4 {
		return "", false
	}
	if districtEverywhere && len(districts) == 1 {
		return soleKey(districts), true
	}
	if regionEverywhere && len(regions) == 1 {
		return soleKey(regions), true
	}
	return "", false
}

// eventPlace returns the known place of an event, or nil when the event or its
// place is absent or unknown.
func eventPlace(ev model.TimelineEvent) *model.Place {
	if ev == nil {
		return nil
	}
	pl := ev.GetPlace()
	if pl == nil || pl.IsUnknown() {
		return nil
	}
	return pl
}

// placeDistrictName returns the district (town or parish) of a place, or the
// empty string when it is not known.
func placeDistrictName(pl *model.Place) string {
	if pl.District != nil && pl.District.Name != "" {
		return pl.District.Name
	}
	return ""
}

// soleKey returns the single key of a one-element set.
func soleKey(m map[string]bool) string {
	for k := range m {
		return k
	}
	return ""
}

// dominantPlace returns the district, or failing that the county, where a clear
// majority of a person's residence records fall, or the empty string when none
// dominates.
func dominantPlace(districts, regions map[string]int, total int) string {
	if name, count := maxEntry(districts); count >= 4 && count*100 >= total*60 {
		return name
	}
	if name, count := maxEntry(regions); count >= 4 && count*100 >= total*60 {
		return name
	}
	return ""
}

// maxEntry returns the highest-counted key, breaking ties by name for stability.
func maxEntry(m map[string]int) (string, int) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	bestName, bestCount := "", 0
	for _, k := range keys {
		if m[k] > bestCount {
			bestName, bestCount = k, m[k]
		}
	}
	return bestName, bestCount
}

// addMilitary notes the named battles the subject took part in, the vivid part
// of a military life; enlistment and rank are left to the occupation segment.
func (b *bioBuilder) addMilitary() {
	var names []string
	seen := map[string]bool{}
	for _, ev := range b.p.Timeline {
		be, ok := ev.(*model.BattleEvent)
		if !ok {
			continue
		}
		name := strings.TrimPrefix(be.What(), "participated in ")
		name = strings.TrimPrefix(name, "the ")
		if name == "" || name == "battle" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	if len(names) == 0 {
		return
	}

	capped := false
	if len(names) > 3 {
		names = names[:3]
		capped = true
	}
	list := battleList(names)
	if capped {
		list += ", among others"
	}
	b.add("saw action at " + list)
}

// battleList renders a set of battle names, collapsing the shared "Battle of"
// prefix into "the Battles of X, Y and Z" when every name carries it.
func battleList(names []string) string {
	places := make([]string, 0, len(names))
	for _, n := range names {
		p, ok := strings.CutPrefix(n, "Battle of ")
		if !ok {
			places = places[:0]
			break
		}
		places = append(places, p)
	}
	if len(places) == len(names) {
		if len(places) == 1 {
			return "the Battle of " + places[0]
		}
		return "the Battles of " + text.JoinList(places)
	}

	full := make([]string, len(names))
	for i, n := range names {
		full[i] = "the " + n
	}
	return text.JoinList(full)
}

// addHardship notes that the subject was, at some point, recorded as a pauper.
func (b *bioBuilder) addHardship() {
	pauper := b.p.Pauper
	if !pauper {
		for _, ev := range b.p.Timeline {
			if e, ok := ev.(*model.EconomicStatusEvent); ok && strings.Contains(strings.ToLower(e.GetDetail()), "pauper") {
				pauper = true
				break
			}
		}
	}
	if !pauper {
		return
	}
	b.add("reduced to pauperism")
}

// addTravel notes a move to another country, reading it as transportation when
// a conviction preceded the arrival and as emigration otherwise.
func (b *bioBuilder) addTravel() {
	var dest *model.Place
	var arrivalDate *model.Date
	for _, ev := range b.p.Timeline {
		switch ev.(type) {
		case *model.ArrivalEvent, *model.ImmigrationEvent:
			if pl := ev.GetPlace(); pl != nil && !pl.IsUnknown() {
				dest = pl
				arrivalDate = ev.GetDate()
			}
		}
	}
	if dest == nil {
		return
	}
	destName := placeCountryOrRegion(dest)
	if destName == "" {
		return
	}

	if b.p.BestBirthlikeEvent != nil {
		bp := b.p.BestBirthlikeEvent.GetPlace()
		if bp != nil && bp.Country != nil && dest.Country != nil && bp.Country.SameAs(dest.Country) {
			return
		}
	}

	transported := false
	for _, ev := range b.p.Timeline {
		if _, ok := ev.(*model.ConvictionEvent); ok && arrivalDate != nil && ev.GetDate().SortsBefore(arrivalDate) {
			transported = true
			break
		}
	}

	if transported {
		b.add("transported to " + destName)
	} else {
		b.add("emigrated to " + destName)
	}
}

// placeCountryOrRegion returns the region name of a place, or its country name,
// as a destination suitable for describing a move.
func placeCountryOrRegion(pl *model.Place) string {
	if pl.Region != nil && !pl.Region.IsUnknown() && pl.Region.Name != "" {
		return pl.Region.Name
	}
	if pl.Country != nil && !pl.Country.IsUnknown() && pl.Country.Name != "" {
		return pl.Country.Name
	}
	return pl.Name
}

// addCrime notes convictions and court appearances. Named offences are listed;
// court appearances are summarised by a representative, most serious one.
func (b *bioBuilder) addCrime() {
	var crimes []string
	seen := map[string]bool{}
	var court []string
	for _, ev := range b.p.Timeline {
		switch e := ev.(type) {
		case *model.ConvictionEvent:
			c := strings.ToLower(strings.TrimSpace(e.Crime))
			if c != "" && !seen[c] {
				seen[c] = true
				crimes = append(crimes, c)
			}
		case *model.CourtEvent:
			if d := strings.TrimSpace(e.What()); d != "" {
				court = append(court, d)
			}
		}
	}

	if len(crimes) > 0 {
		b.add("convicted of " + text.JoinList(crimes))
	}

	if len(court) == 0 {
		return
	}
	rep := representativeCourt(court)
	if len(court) >= 3 {
		b.add("repeated brushes with the law")
	}
	b.add(rep)
}

// representativeCourt returns the most serious court appearance, preferring the
// most recent one that mentions a sentence, otherwise the most recent.
func representativeCourt(descs []string) string {
	rep := descs[len(descs)-1]
	for _, d := range descs {
		if courtNotable(d) {
			rep = d
		}
	}
	return rep
}

// courtNotable reports whether a court appearance resulted in a sentence.
func courtNotable(d string) bool {
	l := strings.ToLower(d)
	for _, kw := range []string{"sentenced", "hard labour", "imprison", "penal", "transport", "gaol", "jail"} {
		if strings.Contains(l, kw) {
			return true
		}
	}
	return false
}

// primaryOccupation returns the occupation recorded most often, which best
// represents how the subject made a living, or nil when none is known.
func (b *bioBuilder) primaryOccupation() *model.Occupation {
	var best *model.Occupation
	for _, o := range b.p.Occupations {
		if o == nil || o.Unknown || o.Name == "" {
			continue
		}
		if best == nil || o.Occurrences > best.Occurrences {
			best = o
		}
	}
	return best
}

func (b *bioBuilder) addDeath() {
	ev := b.p.BestDeathlikeEvent
	if ev == nil {
		return
	}
	date := ev.GetDate()
	place := ev.GetPlace()
	hasDate := date != nil && !date.IsUnknown()
	hasPlace := place != nil && !place.IsUnknown()
	if !hasDate && !hasPlace {
		return
	}

	verb := "died"
	switch ev.(type) {
	case *model.BurialEvent:
		verb = "buried"
	case *model.CremationEvent:
		verb = "cremated"
	default:
		if b.p.ModeOfDeath != model.ModeOfDeathNatural {
			verb = b.p.ModeOfDeath.What()
		}
	}
	parts := []string{verb}
	if hasDate {
		parts = append(parts, dateBare(date))
	}
	if hasPlace {
		parts = append(parts, b.placeWhere(place))
	}
	frag := strings.Join(parts, " ")
	if hasDate {
		if age, ok := b.p.AgeInYearsAt(date); ok && age >= 0 {
			frag += ", aged " + strconv.Itoa(age)
		}
	}
	b.add(frag)
}

// unremarkableCause lists causes of death not worth noting, being the ordinary
// end of a long life rather than a distinguishing fact.
var unremarkableCause = map[string]bool{
	"old age":          true,
	"natural decay":    true,
	"senile decay":     true,
	"natural causes":   true,
	"natural death":    true,
	"decay":            true,
	"decline":          true,
	"senility":         true,
	"debility":         true,
	"general debility": true,
	"senile debility":  true,
	"gradual decay":    true,
}

// maxGlossWords is the longest parenthetical gloss kept alongside a cause of
// death; a longer one is a full definition better left out of a terse line.
const maxGlossWords = 4

// addCauseOfDeath notes a recorded cause of death, attributed rather than stated
// outright because the underlying records vary in wording and reliability.
// Unremarkable causes such as old age are left out.
func (b *bioBuilder) addCauseOfDeath() {
	if b.p.CauseOfDeath == nil {
		return
	}
	if term := causeTerm(b.p.CauseOfDeath.Detail); term == "" || unremarkableCause[term] {
		return
	}
	b.add("death attributed to " + causeDisplay(b.p.CauseOfDeath.Detail))
}

// causeTerm reduces a cause-of-death fact detail to its bare term, dropping the
// quotes and any parenthetical gloss the loader adds.
func causeTerm(detail string) string {
	term, _, _ := strings.Cut(strings.TrimSpace(detail), " (")
	return strings.ToLower(strings.Trim(term, `"`))
}

// causeDisplay renders a cause-of-death fact detail for the biography: the bare
// term with the loader's quotes removed, keeping a short parenthetical gloss but
// dropping a longer one.
func causeDisplay(detail string) string {
	term, gloss, _ := strings.Cut(strings.TrimSpace(detail), " (")
	term = strings.Trim(term, `"`)
	gloss = strings.TrimSuffix(gloss, ")")
	if gloss != "" && len(strings.Fields(gloss)) <= maxGlossWords {
		return term + " (" + gloss + ")"
	}
	return term
}

// dateBare returns a date for leading a "Born" or "Died" fragment: the plain
// date without the "on" preposition or the comma the model inserts into precise
// dates, but keeping an "in", "before", "about" or "after" qualifier.
func dateBare(d *model.Date) string {
	return strings.TrimPrefix(dateWhen(d), "on ")
}

// dateWhen returns a date as a clause with its natural preposition, suitable for
// a marriage, without the comma the model inserts into precise dates.
func dateWhen(d *model.Date) string {
	return strings.Replace(d.When(), ", ", " ", 1)
}

// union holds the phrasing components of one family the subject was a parent in.
type union struct {
	partner        string       // cleaned partner name, empty when unknown
	married        bool         // true for a married or likely-married bond
	when           string       // date phrase for the start of the union, empty when unknown
	date           *model.Date  // start date, used only for ordering
	place          *model.Place // place the union began, used for a marriage, nil when unknown
	children       int          // number of children in this union
	unknownParents int          // number of distinct unknown partners, set only on merged unknown liaisons
	endReason      string       // model.FamilyEndReason* describing how the union ended
	widowed        bool         // true when the union ended with the death of the partner, not the subject
}

// unions returns the subject's families in chronological order of their start.
func (b *bioBuilder) unions() []union {
	var out []union
	for _, fam := range b.p.Families {
		u := union{
			married:  fam.Bond == model.FamilyBondMarried || fam.Bond == model.FamilyBondLikelyMarried,
			children: len(fam.Children),
			partner:  partnerName(fam.OtherParent(b.p)),
		}
		if fam.BestStartDate != nil && !fam.BestStartDate.IsUnknown() {
			u.date = fam.BestStartDate
			u.when = dateWhen(fam.BestStartDate)
		}
		if fam.BestStartEvent != nil {
			u.place = fam.BestStartEvent.GetPlace()
		}
		u.endReason = fam.EndReason
		if fam.EndReason == model.FamilyEndReasonDeath && fam.EndDeathPerson != nil && !fam.EndDeathPerson.SameAs(b.p) {
			u.widowed = true
		}
		out = append(out, u)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].date == nil || out[j].date == nil {
			return out[i].date != nil && out[j].date == nil
		}
		return out[i].date.SortsBefore(out[j].date)
	})
	return out
}

// mergeUnions collapses runs of adjacent unions that produced children with an
// unknown partner into a single union, so that "one child by an unknown father"
// repeated reads as "two children by unknown fathers". Other unions pass through
// unchanged and the chronological order is preserved.
func mergeUnions(in []union) []union {
	var out []union
	for _, u := range in {
		if isUnknownLiaison(u) {
			if len(out) > 0 && isUnknownLiaison(out[len(out)-1]) {
				out[len(out)-1].children += u.children
				out[len(out)-1].unknownParents++
				continue
			}
			u.unknownParents = 1
		}
		out = append(out, u)
	}
	return out
}

// isUnknownLiaison reports whether u is an unmarried union that produced
// children with a partner who is not named.
func isUnknownLiaison(u union) bool {
	return !u.married && u.partner == "" && u.children > 0
}

// partnerName returns a name for a partner suitable for prose, falling back to
// the given name when the surname is unknown and to the empty string when no
// usable name is recorded.
func partnerName(p *model.Person) string {
	if p == nil || p.IsUnknown() || p.Unidentified {
		return ""
	}
	if name := p.PreferredFamiliarFullName; !strings.Contains(name, model.UnknownNamePlaceholder) {
		return name
	}
	if given := p.PreferredFamiliarName; given != "" && !strings.Contains(given, model.UnknownNamePlaceholder) {
		return given
	}
	return ""
}

package genbio

import (
	"strconv"
	"strings"
	"testing"

	"github.com/iand/genster/model"
)

func place(fullName string) *model.Place {
	return &model.Place{FullName: fullName}
}

// placeFull builds a place with the structured town, region and country fields
// populated, as the loader would supply them.
func placeFull(town, region, country string) *model.Place {
	pl := &model.Place{Name: town}
	var parts []string
	if town != "" {
		parts = append(parts, town)
		pl.District = &model.Place{Name: town}
	}
	if region != "" {
		parts = append(parts, region)
		pl.Region = &model.Place{Name: region}
	}
	if country != "" {
		parts = append(parts, country)
		pl.Country = &model.Place{Name: country}
	}
	pl.FullName = strings.Join(parts, ", ")
	return pl
}

func battle(title string) *model.BattleEvent {
	return &model.BattleEvent{GeneralEvent: model.GeneralEvent{Date: model.UnknownDate(), Title: title}}
}

func conviction(crime string) *model.ConvictionEvent {
	return &model.ConvictionEvent{GeneralEvent: model.GeneralEvent{Date: model.UnknownDate()}, Crime: crime}
}

func court(title string) *model.CourtEvent {
	return &model.CourtEvent{GeneralEvent: model.GeneralEvent{Date: model.UnknownDate(), Title: title}}
}

func convictionOn(crime string, d *model.Date) *model.ConvictionEvent {
	return &model.ConvictionEvent{GeneralEvent: model.GeneralEvent{Date: d}, Crime: crime}
}

func arrival(d *model.Date, pl *model.Place) *model.ArrivalEvent {
	return &model.ArrivalEvent{GeneralEvent: model.GeneralEvent{Date: d, Place: pl}}
}

func placeIn(name, country string) *model.Place {
	return &model.Place{Name: name, FullName: name, Country: &model.Place{Name: country}}
}

func econ(detail string) *model.EconomicStatusEvent {
	return &model.EconomicStatusEvent{GeneralEvent: model.GeneralEvent{Date: model.UnknownDate(), Detail: detail}}
}

func placeDistrict(district, region string) *model.Place {
	return &model.Place{District: &model.Place{Name: district}, Region: &model.Place{Name: region}}
}

func residence(district, region string) *model.ResidenceRecordedEvent {
	return &model.ResidenceRecordedEvent{GeneralEvent: model.GeneralEvent{Date: model.UnknownDate(), Place: placeDistrict(district, region)}}
}

func census(district, region string) *model.CensusEvent {
	return &model.CensusEvent{GeneralEvent: model.GeneralEvent{Date: model.UnknownDate(), Place: placeDistrict(district, region)}}
}

func birth(p *model.Person, d *model.Date, pl *model.Place) *model.BirthEvent {
	return &model.BirthEvent{
		GeneralEvent:           model.GeneralEvent{Date: d, Place: pl},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
}

func death(p *model.Person, d *model.Date, pl *model.Place) *model.DeathEvent {
	return &model.DeathEvent{
		GeneralEvent:           model.GeneralEvent{Date: d, Place: pl},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
}

func marriage(d *model.Date, pl *model.Place) *model.MarriageEvent {
	return &model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: d, Place: pl}}
}

func baptism(p *model.Person, d *model.Date, pl *model.Place) *model.BaptismEvent {
	return &model.BaptismEvent{
		GeneralEvent:           model.GeneralEvent{Date: d, Place: pl},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
}

func burial(p *model.Person, d *model.Date, pl *model.Place) *model.BurialEvent {
	return &model.BurialEvent{
		GeneralEvent:           model.GeneralEvent{Date: d, Place: pl},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
}

func cremation(p *model.Person, d *model.Date, pl *model.Place) *model.CremationEvent {
	return &model.CremationEvent{
		GeneralEvent:           model.GeneralEvent{Date: d, Place: pl},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
}

// fullMale returns a person with birth, marriage, children and death all known.
func fullMale() *model.Person {
	p := &model.Person{
		ID:                "p1",
		PreferredFullName: "John Smith",
		Gender:            model.GenderMale,
	}
	p.BestBirthlikeEvent = birth(p, model.PreciseDate(1820, 3, 15), place("Manchester, Lancashire"))
	p.BestDeathlikeEvent = death(p, model.PreciseDate(1878, 6, 20), place("Salford, Lancashire"))
	spouse := &model.Person{ID: "p1w", PreferredFamiliarFullName: "Mary Jones", Gender: model.GenderFemale}
	kids := []*model.Person{{ID: "c1"}, {ID: "c2"}, {ID: "c3"}}
	p.Children = kids
	p.Families = []*model.Family{
		{ID: "f1", Bond: model.FamilyBondMarried, Father: p, Mother: spouse, BestStartDate: model.Year(1845), Children: kids},
	}
	return p
}

// kids returns n placeholder children for populating a family.
func kids(n int) []*model.Person {
	out := make([]*model.Person, n)
	for i := range out {
		out[i] = &model.Person{ID: "k" + strconv.Itoa(i)}
	}
	return out
}

func TestBioGolden(t *testing.T) {
	// A female known only by birth year and a count of children.
	female := &model.Person{ID: "p2", PreferredFullName: "Mary Ann Jones", Gender: model.GenderFemale}
	female.BestBirthlikeEvent = birth(female, model.Year(1900), nil)
	female.Children = []*model.Person{{ID: "d1"}, {ID: "d2"}}

	// A person of unknown gender with a birthplace and a year of death.
	they := &model.Person{ID: "p3", PreferredFullName: "Alex Doe", Gender: model.GenderUnknown}
	they.BestBirthlikeEvent = birth(they, nil, place("Yorkshire"))
	they.BestDeathlikeEvent = death(they, model.Year(1950), nil)

	// Two marriages, children in each.
	twice := &model.Person{ID: "p4", PreferredFullName: "Thomas Frost", Gender: model.GenderMale}
	wifeA := &model.Person{ID: "p4a", PreferredFamiliarFullName: "Anne Adams", Gender: model.GenderFemale}
	wifeB := &model.Person{ID: "p4b", PreferredFamiliarFullName: "Betty Brown", Gender: model.GenderFemale}
	twice.Families = []*model.Family{
		{ID: "f4a", Bond: model.FamilyBondMarried, Father: twice, Mother: wifeA, BestStartDate: model.Year(1840), Children: kids(2)},
		{ID: "f4b", Bond: model.FamilyBondMarried, Father: twice, Mother: wifeB, BestStartDate: model.Year(1850), Children: kids(1)},
	}

	// Illegitimate children by an unknown father, then a marriage.
	illeg := &model.Person{ID: "p5", PreferredFullName: "Jane Poole", Gender: model.GenderFemale}
	husband := &model.Person{ID: "p5h", PreferredFamiliarFullName: "John Rouse", Gender: model.GenderMale}
	illeg.Families = []*model.Family{
		{ID: "f5a", Bond: model.FamilyBondUnmarried, Mother: illeg, BestStartDate: model.Year(1830), Children: kids(2)},
		{ID: "f5b", Bond: model.FamilyBondMarried, Father: husband, Mother: illeg, BestStartDate: model.Year(1840), Children: kids(1)},
	}

	// Consecutive children by unknown fathers, then a marriage.
	merged := &model.Person{ID: "p8", PreferredFullName: "Ada West", Gender: model.GenderFemale}
	husband2 := &model.Person{ID: "p8h", PreferredFamiliarFullName: "Cyril Cole", Gender: model.GenderMale}
	merged.Families = []*model.Family{
		{ID: "f8a", Bond: model.FamilyBondUnmarried, Mother: merged, BestStartDate: model.Year(1855), Children: kids(1)},
		{ID: "f8b", Bond: model.FamilyBondUnmarried, Mother: merged, BestStartDate: model.Year(1858), Children: kids(1)},
		{ID: "f8c", Bond: model.FamilyBondMarried, Father: husband2, Mother: merged, BestStartDate: model.Year(1860), Children: kids(2)},
	}

	// An unmarried union with a known partner.
	liaison := &model.Person{ID: "p6", PreferredFullName: "Mary Read", Gender: model.GenderFemale}
	partner := &model.Person{ID: "p6p", PreferredFamiliarFullName: "Daniel Palmer", Gender: model.GenderMale}
	liaison.Families = []*model.Family{
		{ID: "f6", Bond: model.FamilyBondUnmarried, Father: partner, Mother: liaison, BestStartDate: model.Year(1860), Children: kids(3)},
	}

	// Three marriages: the first ended in widowhood, the second in divorce.
	thrice := &model.Person{ID: "p9", PreferredFullName: "Bridget Nash", Gender: model.GenderFemale}
	ash := &model.Person{ID: "p9a", PreferredFamiliarFullName: "Alan Ash", Gender: model.GenderMale}
	beech := &model.Person{ID: "p9b", PreferredFamiliarFullName: "Brian Beech", Gender: model.GenderMale}
	cook := &model.Person{ID: "p9c", PreferredFamiliarFullName: "Colin Cook", Gender: model.GenderMale}
	thrice.Families = []*model.Family{
		{ID: "f9a", Bond: model.FamilyBondMarried, Father: ash, Mother: thrice, BestStartDate: model.Year(1860), Children: kids(1), EndReason: model.FamilyEndReasonDeath, EndDeathPerson: ash},
		{ID: "f9b", Bond: model.FamilyBondMarried, Father: beech, Mother: thrice, BestStartDate: model.Year(1865), EndReason: model.FamilyEndReasonDivorce},
		{ID: "f9c", Bond: model.FamilyBondMarried, Father: cook, Mother: thrice, BestStartDate: model.Year(1870), Children: kids(2)},
	}

	// A single marriage that ended in the death of the spouse, the subject's own
	// death known so the widowhood is confirmed.
	widow := &model.Person{ID: "p10", PreferredFullName: "Sarah Vale", Gender: model.GenderFemale}
	widow.BestDeathlikeEvent = death(widow, model.Year(1890), nil)
	vale := &model.Person{ID: "p10h", PreferredFamiliarFullName: "Peter Vale", Gender: model.GenderMale}
	widow.Families = []*model.Family{
		{ID: "f10", Bond: model.FamilyBondMarried, Father: vale, Mother: widow, BestStartDate: model.Year(1880), Children: kids(1), EndReason: model.FamilyEndReasonDeath, EndDeathPerson: vale},
	}

	// The marriage ended in the spouse's death, but the subject's own death is
	// unknown, so the subject may have predeceased; widowhood is not claimed.
	maybeWidow := &model.Person{ID: "p44", PreferredFullName: "Owen Gray", Gender: model.GenderMale}
	gwen := &model.Person{ID: "p44w", PreferredFamiliarFullName: "Gwen Gray", Gender: model.GenderFemale}
	maybeWidow.Families = []*model.Family{
		{ID: "f44", Bond: model.FamilyBondMarried, Father: maybeWidow, Mother: gwen, BestStartDate: model.Year(1860), Children: kids(1), EndReason: model.FamilyEndReasonDeath, EndDeathPerson: gwen},
	}

	// A tradesman known only by occupation.
	tradesman := &model.Person{ID: "p11", PreferredFullName: "Henry Cole", Gender: model.GenderMale}
	tradesman.Occupations = []*model.Occupation{
		{Name: "labourer", Occurrences: 1},
		{Name: "blacksmith", Occurrences: 3},
	}

	// A soldier, whose occupation is in the military group.
	soldier := &model.Person{ID: "p12", PreferredFullName: "George Ash", Gender: model.GenderMale}
	soldier.Occupations = []*model.Occupation{
		{Name: "soldier", Group: model.OccupationGroupMilitary, Occurrences: 2},
	}

	// A soldier who fought in two named battles.
	fighter := &model.Person{ID: "p13", PreferredFullName: "Tom Gunn", Gender: model.GenderMale}
	fighter.Timeline = []model.TimelineEvent{
		battle("participated in the Battle of Waterloo"),
		battle("participated in the Battle of Trafalgar"),
	}

	// A soldier who fought in more battles than are listed.
	veteran := &model.Person{ID: "p14", PreferredFullName: "Max Steel", Gender: model.GenderMale}
	veteran.Timeline = []model.TimelineEvent{
		battle("participated in the Battle of Alma"),
		battle("participated in the Battle of Balaclava"),
		battle("participated in the Battle of Inkerman"),
		battle("participated in the Battle of Sevastopol"),
	}

	// Convicted of two named offences.
	convict := &model.Person{ID: "p15", PreferredFullName: "Ned Kelly", Gender: model.GenderMale}
	convict.Timeline = []model.TimelineEvent{
		conviction("burglary"),
		conviction("Cattle stealing"),
	}

	// Several court appearances, one of them serious.
	repeatOffender := &model.Person{ID: "p16", PreferredFullName: "Moll Cutpurse", Gender: model.GenderFemale}
	repeatOffender.Timeline = []model.TimelineEvent{
		court("Fined for creating a disturbance"),
		court("Indicted for keeping a disorderly house and sentenced to nine months with hard labour"),
		court("Fined for keeping a disorderly house"),
	}

	// A single court appearance.
	oneCharge := &model.Person{ID: "p17", PreferredFullName: "Sam Vane", Gender: model.GenderMale}
	oneCharge.Timeline = []model.TimelineEvent{
		court("Tried on a charge of piratical revolt"),
	}

	// A sailor lost at sea.
	sailor := &model.Person{ID: "p18", PreferredFullName: "Jack Tar", Gender: model.GenderMale, ModeOfDeath: model.ModeOfDeathLostAtSea}
	sailor.BestDeathlikeEvent = death(sailor, model.Year(1800), nil)

	// A woman who died in childbirth.
	mother := &model.Person{ID: "p19", PreferredFullName: "Mary Reed", Gender: model.GenderFemale, ModeOfDeath: model.ModeOfDeathChildbirth}
	mother.BestDeathlikeEvent = death(mother, model.Year(1850), nil)

	// A convict transported overseas.
	transportee := &model.Person{ID: "p20", PreferredFullName: "Jem Nab", Gender: model.GenderMale}
	transportee.BestBirthlikeEvent = birth(transportee, model.Year(1790), placeIn("London, England", "England"))
	transportee.Timeline = []model.TimelineEvent{
		convictionOn("theft", model.Year(1815)),
		arrival(model.Year(1816), placeIn("Sydney, New South Wales, Australia", "Australia")),
	}

	// A free emigrant to another country.
	emigrant := &model.Person{ID: "p21", PreferredFullName: "Ada Ford", Gender: model.GenderFemale}
	emigrant.BestBirthlikeEvent = birth(emigrant, model.Year(1850), placeIn("Norwich, England", "England"))
	emigrant.Timeline = []model.TimelineEvent{
		arrival(model.Year(1875), placeIn("Toronto, Canada", "Canada")),
	}

	// Recorded as a pauper via an economic-status event.
	pauper := &model.Person{ID: "p22", PreferredFullName: "Tom Poor", Gender: model.GenderMale}
	pauper.Timeline = []model.TimelineEvent{econ("Pauper")}

	// Described by a curated epithet rather than a raw occupation.
	epithetPerson := &model.Person{ID: "p24", PreferredFullName: "Eli Mason", Gender: model.GenderMale, Epithet: "master mason"}

	// A recorded change of name.
	renamed := &model.Person{ID: "p25", PreferredFullName: "Will Green", Gender: model.GenderMale, Notable: "Changed name from Brown to Green"}

	// A notable fact that becomes a headline (and suppresses the closer).
	notableHero := &model.Person{ID: "p26", PreferredFullName: "Kit Vale", Gender: model.GenderMale, Notable: "Transported to Australia for theft"}
	notableHero.BestBirthlikeEvent = birth(notableHero, model.Year(1800), nil)

	// A notable fact that already carries its own verb.
	notableActive := &model.Person{ID: "p27", PreferredFullName: "Meg Roe", Gender: model.GenderFemale, Notable: "Died in a fire"}
	notableActive.BestBirthlikeEvent = birth(notableActive, model.Year(1850), nil)

	// Settled in one district for most of life.
	settled := &model.Person{ID: "p28", PreferredFullName: "Sam Kent", Gender: model.GenderMale}
	settled.Timeline = []model.TimelineEvent{
		census("Ashford", "Kent"), residence("Ashford", "Kent"), census("Ashford", "Kent"),
		residence("Ashford", "Kent"), census("Ashford", "Kent"),
	}

	// Moved between districts but stayed in one county.
	countyDweller := &model.Person{ID: "p29", PreferredFullName: "Ann Vale", Gender: model.GenderFemale}
	countyDweller.Timeline = []model.TimelineEvent{
		census("Exeter", "Devon"), residence("Tiverton", "Devon"), census("Barnstaple", "Devon"),
		residence("Totnes", "Devon"), census("Honiton", "Devon"),
	}

	// A dual surname joined with a slash.
	dualName := &model.Person{ID: "p23", PreferredFullName: "Benjamin Fetters/Fletcher", PreferredFamilyName: "Fetters/Fletcher", Gender: model.GenderMale}
	dualName.BestBirthlikeEvent = birth(dualName, model.Year(1767), nil)

	// A spouse whose surname is unknown, carrying the name placeholder.
	unknownSpouse := &model.Person{ID: "p7", PreferredFullName: "Richard Woods", Gender: model.GenderMale}
	hannah := &model.Person{ID: "p7w", PreferredFamiliarFullName: "Hannah " + model.UnknownNamePlaceholder, PreferredFamiliarName: "Hannah", Gender: model.GenderFemale}
	unknownSpouse.Families = []*model.Family{
		{ID: "f7", Bond: model.FamilyBondMarried, Father: unknownSpouse, Mother: hannah, BestStartDate: model.BeforeYear(1685), Children: kids(4)},
	}

	// Born and died in the same place, so the death place collapses to its locality.
	homebody := &model.Person{ID: "p30", PreferredFullName: "Noah Dale", Gender: model.GenderMale}
	homebody.BestBirthlikeEvent = birth(homebody, model.PreciseDate(1830, 2, 1), place("Oakham, Rutland, England"))
	homebody.BestDeathlikeEvent = death(homebody, model.PreciseDate(1900, 4, 2), place("Oakham, Rutland, England"))

	// Born and died in the same country, so the death place drops the country.
	traveller := &model.Person{ID: "p31", PreferredFullName: "Ruth Dale", Gender: model.GenderFemale}
	traveller.BestBirthlikeEvent = birth(traveller, model.PreciseDate(1830, 2, 1), place("Oakham, Rutland, England"))
	traveller.BestDeathlikeEvent = death(traveller, model.PreciseDate(1900, 4, 2), place("Derby, Derbyshire, England"))

	// Married where born, then married again in another town of the same
	// country: the first place is redundant and dropped, the second is given by
	// its town alone.
	remarried := &model.Person{ID: "p32", PreferredFullName: "Abel Stone", Gender: model.GenderMale}
	remarried.BestBirthlikeEvent = birth(remarried, model.PreciseDate(1800, 1, 1), placeFull("Shalstone", "Buckinghamshire", "England"))
	wife1 := &model.Person{ID: "p32a", PreferredFamiliarFullName: "Jane Fisher", Gender: model.GenderFemale}
	wife2 := &model.Person{ID: "p32b", PreferredFamiliarFullName: "Mary Gold", Gender: model.GenderFemale}
	remarried.Families = []*model.Family{
		{
			ID: "f32a", Bond: model.FamilyBondMarried, Father: remarried, Mother: wife1,
			BestStartDate:  model.PreciseDate(1822, 3, 4),
			BestStartEvent: marriage(model.PreciseDate(1822, 3, 4), placeFull("Shalstone", "Buckinghamshire", "England")),
			Children:       kids(2),
		},
		{
			ID: "f32b", Bond: model.FamilyBondMarried, Father: remarried, Mother: wife2,
			BestStartDate:  model.PreciseDate(1835, 6, 7),
			BestStartEvent: marriage(model.PreciseDate(1835, 6, 7), placeFull("Westbury", "Wiltshire", "England")),
			Children:       kids(1),
		},
	}

	// Born in one country and married in another: the marriage abroad is given
	// with its region and country for context.
	abroad := &model.Person{ID: "p33", PreferredFullName: "Cora Vane", Gender: model.GenderFemale}
	abroad.BestBirthlikeEvent = birth(abroad, model.PreciseDate(1850, 5, 5), placeFull("Oakham", "Rutland", "England"))
	groom := &model.Person{ID: "p33h", PreferredFamiliarFullName: "Louis Roy", Gender: model.GenderMale}
	abroad.Families = []*model.Family{
		{
			ID: "f33", Bond: model.FamilyBondMarried, Father: groom, Mother: abroad,
			BestStartDate:  model.PreciseDate(1875, 7, 8),
			BestStartEvent: marriage(model.PreciseDate(1875, 7, 8), placeFull("Paris", "Île-de-France", "France")),
			Children:       kids(1),
		},
	}

	// Known only by baptism and burial rather than birth and death.
	churchRecords := &model.Person{ID: "p34", PreferredFullName: "Silas Wood", Gender: model.GenderMale}
	churchRecords.BestBirthlikeEvent = baptism(churchRecords, model.PreciseDate(1700, 3, 2), place("Ely, Cambridgeshire, England"))
	churchRecords.BestDeathlikeEvent = burial(churchRecords, model.PreciseDate(1760, 8, 9), place("Ely, Cambridgeshire, England"))

	// Cremated rather than buried.
	crematedPerson := &model.Person{ID: "p35", PreferredFullName: "Iris Frost", Gender: model.GenderFemale}
	crematedPerson.BestDeathlikeEvent = cremation(crematedPerson, model.Year(1980), place("Golders Green, Middlesex, England"))

	// A twin, noted in the birth line without naming the twin.
	twinFlagged := &model.Person{ID: "p39", PreferredFullName: "Silas Poole", Gender: model.GenderMale, Twin: true}
	twinFlagged.BestBirthlikeEvent = birth(twinFlagged, model.PreciseDate(1850, 3, 3), place("Ipswich, Suffolk, England"))

	// Twin status carried by an association rather than the flag.
	twinAssoc := &model.Person{ID: "p40", PreferredFullName: "Iris Poole", Gender: model.GenderFemale}
	twinAssoc.Associations = []model.Association{{Kind: model.AssociationKindTwin}}
	twinAssoc.BestBirthlikeEvent = birth(twinAssoc, model.Year(1852), nil)

	// Certainly unmarried and childless, combined into one statement.
	spinster := &model.Person{ID: "p41", PreferredFullName: "Edith Vane", Gender: model.GenderFemale, Unmarried: true, Childless: true}
	spinster.BestBirthlikeEvent = birth(spinster, model.Year(1800), nil)

	// Illegitimate with no recorded father, noted after the birth.
	baseborn := &model.Person{ID: "p42", PreferredFullName: "Tom Reed", Gender: model.GenderMale, Illegitimate: true}
	baseborn.BestBirthlikeEvent = birth(baseborn, model.Year(1830), place("Norwich, England"))

	// A childless marriage: the marriage shows, then the childlessness.
	childlessCouple := &model.Person{ID: "p43", PreferredFullName: "Hugh Frost", Gender: model.GenderMale, Childless: true}
	wifeC := &model.Person{ID: "p43w", PreferredFamiliarFullName: "Ada Frost", Gender: model.GenderFemale}
	childlessCouple.Families = []*model.Family{
		{ID: "f43", Bond: model.FamilyBondMarried, Father: childlessCouple, Mother: wifeC, BestStartDate: model.Year(1850)},
	}

	// A remarkable cause of death is noted, attributed rather than asserted.
	causePerson := &model.Person{ID: "p37", PreferredFullName: "Eli Ward", Gender: model.GenderMale}
	causePerson.BestDeathlikeEvent = death(causePerson, model.Year(1875), nil)
	causePerson.CauseOfDeath = model.ParseCauseOfDeathFact("lockjaw", nil)

	// An unremarkable cause of death is left out.
	oldAge := &model.Person{ID: "p38", PreferredFullName: "Ada Hale", Gender: model.GenderFemale}
	oldAge.BestDeathlikeEvent = death(oldAge, model.Year(1900), nil)
	oldAge.CauseOfDeath = model.ParseCauseOfDeathFact("old age", nil)

	// Every placed event falls in one county, bracketed by a birth and death there.
	neverLeft := &model.Person{ID: "p36", PreferredFullName: "Rebecca Field", Gender: model.GenderFemale}
	neverLeft.BestBirthlikeEvent = birth(neverLeft, model.PreciseDate(1841, 5, 8), placeFull("Withersdale", "Suffolk", "England"))
	neverLeft.BestDeathlikeEvent = death(neverLeft, model.PreciseDate(1881, 8, 24), placeFull("Wilby", "Suffolk", "England"))
	neverLeft.Timeline = []model.TimelineEvent{
		census("Framlingham", "Suffolk"), residence("Hoxne", "Suffolk"), census("Eye", "Suffolk"),
	}

	testCases := []struct {
		name string
		p    *model.Person
		want string
	}{
		{
			name: "full_male",
			p:    fullMale(),
			want: "Born 15 Mar 1820 in Manchester, Lancashire. Married Mary Jones in 1845; three children. Died 20 Jun 1878 in Salford, aged 58.",
		},
		{
			name: "female_partial",
			p:    female,
			want: "Born in 1900. Two children.",
		},
		{
			name: "unknown_gender",
			p:    they,
			want: "Born in Yorkshire. Died in 1950.",
		},
		{
			name: "two_marriages",
			p:    twice,
			want: "Married Anne Adams in 1840; two children. Married Betty Brown in 1850; one child.",
		},
		{
			name: "illegitimate_then_marriage",
			p:    illeg,
			want: "Two children by an unknown father. Married John Rouse in 1840; one child.",
		},
		{
			name: "merged_unknown_fathers",
			p:    merged,
			want: "Two children by unknown fathers. Married Cyril Cole in 1860; two children.",
		},
		{
			name: "unmarried_known_partner",
			p:    liaison,
			want: "Three children with Daniel Palmer.",
		},
		{
			name: "unknown_spouse_surname",
			p:    unknownSpouse,
			want: "Married Hannah before 1685; four children.",
		},
		{
			name: "three_marriages_widowed_and_divorced",
			p:    thrice,
			want: "Married Alan Ash in 1860; one child. Widowed. Married Brian Beech in 1865. Divorced. Married Colin Cook in 1870; two children.",
		},
		{
			name: "single_marriage_widowed",
			p:    widow,
			want: "Married Peter Vale in 1880; one child. Widowed. Died in 1890.",
		},
		{
			name: "widowhood_unconfirmed_when_death_unknown",
			p:    maybeWidow,
			want: "Married Gwen Gray in 1860; one child.",
		},
		{
			name: "occupation_most_recorded",
			p:    tradesman,
			want: "Blacksmith.",
		},
		{
			name: "occupation_military",
			p:    soldier,
			want: "Soldier.",
		},
		{
			name: "battles_named",
			p:    fighter,
			want: "Saw action at the Battles of Waterloo and Trafalgar.",
		},
		{
			name: "battles_capped",
			p:    veteran,
			want: "Saw action at the Battles of Alma, Balaclava and Inkerman, among others.",
		},
		{
			name: "crime_convictions",
			p:    convict,
			want: "Convicted of burglary and cattle stealing.",
		},
		{
			name: "crime_court_repeated",
			p:    repeatOffender,
			want: "Repeated brushes with the law. Indicted for keeping a disorderly house and sentenced to nine months with hard labour.",
		},
		{
			name: "crime_court_single",
			p:    oneCharge,
			want: "Tried on a charge of piratical revolt.",
		},
		{
			name: "mode_of_death_lost_at_sea",
			p:    sailor,
			want: "Lost at sea in 1800.",
		},
		{
			name: "mode_of_death_childbirth",
			p:    mother,
			want: "Died in childbirth in 1850.",
		},
		{
			name: "transported",
			p:    transportee,
			want: "Born in 1790 in London, England. Convicted of theft. Transported to Australia.",
		},
		{
			name: "emigrated",
			p:    emigrant,
			want: "Born in 1850 in Norwich, England. Emigrated to Canada.",
		},
		{
			name: "pauper",
			p:    pauper,
			want: "Reduced to pauperism.",
		},
		{
			name: "notable_headline",
			p:    notableHero,
			want: "Transported to Australia for theft. Born in 1800.",
		},
		{
			name: "notable_headline_active_verb",
			p:    notableActive,
			want: "Died in a fire. Born in 1850.",
		},
		{
			name: "epithet_occupation",
			p:    epithetPerson,
			want: "Master mason.",
		},
		{
			name: "name_change",
			p:    renamed,
			want: "Took the name Green in place of Brown.",
		},
		{
			name: "residence_dominant_district",
			p:    settled,
			want: "Resident of Ashford for many years.",
		},
		{
			name: "residence_dominant_county",
			p:    countyDweller,
			want: "Resident of Devon for many years.",
		},
		{
			name: "dual_surname",
			p:    dualName,
			want: "Born in 1767. Also recorded under the surname Fletcher.",
		},
		{
			name: "same_place_collapses_to_locality",
			p:    homebody,
			want: "Born 1 Feb 1830 in Oakham, Rutland, England. Died 2 Apr 1900 in Oakham, aged 70.",
		},
		{
			name: "same_country_drops_country",
			p:    traveller,
			want: "Born 1 Feb 1830 in Oakham, Rutland, England. Died 2 Apr 1900 in Derby, Derbyshire, aged 70.",
		},
		{
			name: "marriage_place_town_only",
			p:    remarried,
			want: "Born 1 Jan 1800 in Shalstone, Buckinghamshire, England. Married Jane Fisher on 4 Mar 1822; two children. Married Mary Gold on 7 Jun 1835 in Westbury; one child.",
		},
		{
			name: "marriage_place_abroad_gives_country",
			p:    abroad,
			want: "Born 5 May 1850 in Oakham, Rutland, England. Married Louis Roy on 8 Jul 1875 in Paris, Île-de-France, France; one child.",
		},
		{
			name: "baptism_and_burial",
			p:    churchRecords,
			want: "Baptised 2 Mar 1700 in Ely, Cambridgeshire, England. Buried 9 Aug 1760 in Ely, aged 60.",
		},
		{
			name: "cremation",
			p:    crematedPerson,
			want: "Cremated in 1980 in Golders Green, Middlesex, England.",
		},
		{
			name: "lifelong_resident",
			p:    neverLeft,
			want: "Born 8 May 1841 in Withersdale, Suffolk, England. Lifelong resident of Suffolk. Died 24 Aug 1881 in Wilby, aged 40.",
		},
		{
			name: "twin_flagged",
			p:    twinFlagged,
			want: "Born 3 Mar 1850 in Ipswich, Suffolk, England, a twin.",
		},
		{
			name: "unmarried_and_childless",
			p:    spinster,
			want: "Born in 1800. Never married and had no children.",
		},
		{
			name: "illegitimate_father_unknown",
			p:    baseborn,
			want: "Born in 1830 in Norwich, England. Father unknown.",
		},
		{
			name: "childless_marriage",
			p:    childlessCouple,
			want: "Married Ada Frost in 1850. Had no children.",
		},
		{
			name: "twin_via_association",
			p:    twinAssoc,
			want: "Born in 1852, a twin.",
		},
		{
			name: "cause_of_death_noted",
			p:    causePerson,
			want: "Died in 1875. Death attributed to lockjaw (tetanus).",
		},
		{
			name: "cause_of_death_unremarkable_omitted",
			p:    oldAge,
			want: "Died in 1900.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := bio(FromPerson(tc.p))
			if got != tc.want {
				t.Errorf("got:\n  %q\nwant:\n  %q", got, tc.want)
			}
		})
	}
}

func TestBioDeterministic(t *testing.T) {
	p := fullMale()
	first := BioFromPerson(p)
	for range 5 {
		if got := BioFromPerson(p); got != first {
			t.Fatalf("Bio is not deterministic: %q then %q", first, got)
		}
	}
	if first == "" {
		t.Error("Bio returned empty string")
	}
}

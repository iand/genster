package site

import (
	"regexp"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

// A person born at country level must not have their birth described as "here".
func TestRenderPersonPageCountryLevelBirthIsNotHere(t *testing.T) {
	tr := tree.NewTree("test", nil, &tree.SurnameGroups{})

	scotland := tr.FindPlace("test", "scotland")
	scotland.Name = "Scotland"
	scotland.FullName = "Scotland"
	scotland.NameWithDistrict = "Scotland"
	scotland.NameWithRegion = "Scotland"
	scotland.NameWithCountry = "Scotland"
	scotland.PlaceType = model.PlaceTypeCountry
	scotland.Country = scotland

	p := tr.FindPerson("test", "ann")
	p.PreferredGivenName = "Ann"
	p.PreferredFamiliarName = "Ann"
	p.PreferredFullName = "Ann Smith"
	p.PreferredUniqueName = "Ann Smith"
	p.Gender = model.GenderFemale

	birth := &model.BirthEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.Year(1802), Place: scotland},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
	p.BestBirthlikeEvent = birth
	p.Timeline = []model.TimelineEvent{birth}

	s := NewSite("/trees/test/", tr)
	ps, err := NewPublishSet(tr, func(*model.Person) bool { return true })
	if err != nil {
		t.Fatalf("NewPublishSet: %v", err)
	}
	s.PublishSet = ps

	doc, err := RenderPersonPage(s, p)
	if err != nil {
		t.Fatalf("RenderPersonPage: %v", err)
	}

	got := doc.String()
	if regexp.MustCompile(`\bhere\b`).MatchString(got) {
		t.Errorf("page describes the birth place as \"here\":\n%s", got)
	}
}

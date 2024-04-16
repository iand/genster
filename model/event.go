package model

import (
	"sort"
)

type TimelineEvent interface {
	GetDate() *Date
	GetDateType() EventDateType
	GetPlace() *Place
	GetTitle() string
	GetDetail() string
	GetCitations() []*GeneralCitation
	Type() string
	ShortDescription() string        // returns the abbreviated name of the event and its date, e.g. "b. 4 Jul 1928"
	What() string                    // married, born, divorced
	When() string                    // text description of date
	Where() string                   // text description of place
	IsInferred() bool                // whether or not the event was inferred to exist, i.e. has no supporting evidence
	DirectlyInvolves(p *Person) bool // whether or not the event directly involves a person as a principal or party
	Participants() []*Person
	SortsBefore(other TimelineEvent) bool
}

// GeneralPartyEvent is a timeline event involving one individual.
type IndividualTimelineEvent interface {
	TimelineEvent
	GetPrincipal() *Person
}

// PartyTimelineEvent is a timeline event involving two parties.
type PartyTimelineEvent interface {
	TimelineEvent
	GetParty1() *Person
	GetParty2() *Person
	GetOther(p *Person) *Person // returns the first of party1 or party2 that is not p
}

func SortTimelineEvents(evs []TimelineEvent) {
	sort.Slice(evs, func(i, j int) bool {
		return evs[i].SortsBefore(evs[j])
	})
}

type GeneralEvent struct {
	Date      *Date
	Place     *Place
	Title     string
	Detail    string
	Citations []*GeneralCitation
	Inferred  bool
	Narrative string // hand written narrative, if any
}

func (e *GeneralEvent) GetDate() *Date {
	return e.Date
}

func (e *GeneralEvent) When() string {
	return e.Date.When()
}

func (e *GeneralEvent) GetDateType() EventDateType {
	return EventDateTypeOccurred
}

func (e *GeneralEvent) GetPlace() *Place {
	return e.Place
}

func (e *GeneralEvent) Where() string {
	return e.Place.Where()
}

func (e *GeneralEvent) GetTitle() string {
	return e.Title
}

func (e *GeneralEvent) GetDetail() string {
	return e.Detail
}

func (e *GeneralEvent) GetNarrative() string {
	return e.Narrative
}

func (e *GeneralEvent) GetCitations() []*GeneralCitation {
	return e.Citations
}

func (e *GeneralEvent) EventDate() *Date {
	return e.Date
}

func (e *GeneralEvent) Type() string {
	return "general event"
}

func (e *GeneralEvent) ShortDescription() string {
	return e.abbrev("")
}

func (e *GeneralEvent) IsInferred() bool {
	return e.Inferred
}

func (e *GeneralEvent) abbrev(prefix string) string {
	if prefix == "" {
		return e.Date.String()
	}
	return prefix + ". " + e.Date.String()
}

func (e *GeneralEvent) What() string { return "had an event" }

func (e *GeneralEvent) SortsBefore(other TimelineEvent) bool {
	if e == nil || e.Date == nil {
		return false
	}
	if other == nil || other.GetDate() == nil {
		return true
	}

	return e.Date.SortsBefore(other.GetDate())
}

// GeneralPartyEvent is a general event involving one individual.
type GeneralIndividualEvent struct {
	Principal *Person
}

func (e *GeneralIndividualEvent) GetPrincipal() *Person {
	return e.Principal
}

func (e *GeneralIndividualEvent) DirectlyInvolves(p *Person) bool {
	return e.Principal.SameAs(p)
}

func (e *GeneralIndividualEvent) Participants() []*Person {
	return []*Person{e.Principal}
}

// GeneralPartyEvent is a general event involving two principal parties.
type GeneralPartyEvent struct {
	Party1 *Person
	Party2 *Person
}

func (e *GeneralPartyEvent) GetParty1() *Person {
	return e.Party1
}

func (e *GeneralPartyEvent) GetParty2() *Person {
	return e.Party2
}

func (e *GeneralPartyEvent) DirectlyInvolves(p *Person) bool {
	return e.Party1.SameAs(p) || e.Party2.SameAs(p)
}

func (e *GeneralPartyEvent) GetOther(p *Person) *Person {
	if !e.Party1.SameAs(p) {
		return e.Party1
	}
	if !e.Party2.SameAs(p) {
		return e.Party2
	}
	return p
}

func (e *GeneralPartyEvent) Participants() []*Person {
	return []*Person{e.Party1, e.Party2}
}

// POV represents a point of view. It is used to provide contect when constructing a description of an event.
type POV struct {
	Person *Person // the person observing or experiencing the event
	Place  *Place  // the place in which the observing is taking place
}

// PlaceholderIndividualEvent represents an event involving one individual that has not been interpreted and is a placeholder until the necessary
// processing has been written.
type PlaceholderIndividualEvent struct {
	GeneralEvent
	GeneralIndividualEvent
	ExtraInfo string
}

func (e *PlaceholderIndividualEvent) GetTitle() string {
	return e.ExtraInfo
}

// PlaceholderPartyEvent represents an event involving two parties that has not been interpreted and is a placeholder until the necessary
// processing has been written.
type PlaceholderPartyEvent struct {
	GeneralEvent
	GeneralPartyEvent
	ExtraInfo string
}

func (e *PlaceholderPartyEvent) GetTitle() string {
	return e.ExtraInfo
}

var _ TimelineEvent = (*PlaceholderIndividualEvent)(nil)

// BirthEvent represents the birth of a person in their timeline
type BirthEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *BirthEvent) Type() string             { return "birth" }
func (e *BirthEvent) ShortDescription() string { return e.abbrev("b") }
func (e *BirthEvent) What() string             { return "born" }

var (
	_ TimelineEvent           = (*BirthEvent)(nil)
	_ IndividualTimelineEvent = (*BirthEvent)(nil)
)

// BaptismEvent represents the baptism of a person in their timeline
type BaptismEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *BaptismEvent) Type() string             { return "baptism" }
func (e *BaptismEvent) ShortDescription() string { return e.abbrev("bap") }
func (e *BaptismEvent) What() string             { return "baptised" }

var (
	_ TimelineEvent           = (*BaptismEvent)(nil)
	_ IndividualTimelineEvent = (*BaptismEvent)(nil)
)

// DeathEvent represents the death of a person in their timeline
type DeathEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *DeathEvent) Type() string             { return "death" }
func (e *DeathEvent) ShortDescription() string { return e.abbrev("d") }
func (e *DeathEvent) What() string             { return "died" }

var (
	_ TimelineEvent           = (*DeathEvent)(nil)
	_ IndividualTimelineEvent = (*DeathEvent)(nil)
)

// BurialEvent represents the burial of a person in their timeline
type BurialEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *BurialEvent) Type() string             { return "burial" }
func (e *BurialEvent) ShortDescription() string { return e.abbrev("bur") }
func (e *BurialEvent) What() string             { return "buried" }

var (
	_ TimelineEvent           = (*BurialEvent)(nil)
	_ IndividualTimelineEvent = (*BurialEvent)(nil)
)

// CremationEvent represents the cremation of a person in their timeline
type CremationEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *CremationEvent) Type() string             { return "cremation" }
func (e *CremationEvent) ShortDescription() string { return e.abbrev("crem") }
func (e *CremationEvent) What() string             { return "cremated" }

var (
	_ TimelineEvent           = (*CremationEvent)(nil)
	_ IndividualTimelineEvent = (*CremationEvent)(nil)
)

// DepartureEvent represents the departure of a person from a place
type DepartureEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *DepartureEvent) Type() string             { return "departure" }
func (e *DepartureEvent) ShortDescription() string { return e.abbrev("dep") }
func (e *DepartureEvent) What() string             { return "departed" }

var (
	_ TimelineEvent           = (*DepartureEvent)(nil)
	_ IndividualTimelineEvent = (*DepartureEvent)(nil)
)

// ArrivalEvent represents the arrival of a person at a place
type ArrivalEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *ArrivalEvent) Type() string             { return "arrival" }
func (e *ArrivalEvent) ShortDescription() string { return e.abbrev("arr") }
func (e *ArrivalEvent) What() string             { return "arrived" }

var (
	_ TimelineEvent           = (*ArrivalEvent)(nil)
	_ IndividualTimelineEvent = (*ArrivalEvent)(nil)
)

// IndividualNarrativeEvent represents some narrative that can be used as-is
type IndividualNarrativeEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *IndividualNarrativeEvent) ShortDescription() string { return e.abbrev("narr") }
func (e *IndividualNarrativeEvent) What() string             { return e.GetTitle() }

var (
	_ TimelineEvent           = (*IndividualNarrativeEvent)(nil)
	_ IndividualTimelineEvent = (*IndividualNarrativeEvent)(nil)
)

// CensusEvent represents the taking of a census in a timeline
// It is a shared event
type CensusEvent struct {
	GeneralEvent
	Entries []*CensusEntry
}

type CensusEntry struct {
	Principal      *Person
	RelationToHead CensusEntryRelation // cleaned
	Name           string              // as recorded
	Sex            string              // as recorded
	MaritalStatus  CensusEntryMaritalStatus
	Age            string // as recorded
	Occupation     string // as recorded
	PlaceOfBirth   string // as recorded
	Impairment     string // as recorded (deaf and dumb, blind, imbecile etc)
	Detail         string // any remaining unparsed detail
	Narrative      string // hand written narrative, if any
}

func (e *CensusEvent) Type() string             { return "census" }
func (e *CensusEvent) ShortDescription() string { return e.abbrev("cens") }
func (e *CensusEvent) What() string             { return "appeared in census" }

func (e *CensusEvent) DirectlyInvolves(p *Person) bool {
	if p.IsUnknown() {
		return false
	}
	for _, en := range e.Entries {
		if p.SameAs(en.Principal) {
			return true
		}
	}
	return false
}

func (e *CensusEvent) Entry(p *Person) (*CensusEntry, bool) {
	for _, en := range e.Entries {
		if en.Principal.SameAs(p) {
			return en, true
		}
	}
	return nil, false
}

func (e *CensusEvent) Head() *Person {
	for _, en := range e.Entries {
		if en.RelationToHead == CensusEntryRelationHead {
			return en.Principal
		}
	}
	return nil
}

func (e *CensusEvent) Participants() []*Person {
	// TODO implement CensusEvent.Participants
	return []*Person{}
}

var _ TimelineEvent = (*CensusEvent)(nil)

// ProbateEvent represents the granting of probate for a person in their timeline
type ProbateEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*ProbateEvent)(nil)
	_ IndividualTimelineEvent = (*ProbateEvent)(nil)
)

func (e *ProbateEvent) Type() string             { return "probate" }
func (e *ProbateEvent) ShortDescription() string { return e.abbrev("prob") }
func (e *ProbateEvent) What() string             { return "had probate granted" }

// WillEvent represents the writing of a will by a person in their timeline
type WillEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*WillEvent)(nil)
	_ IndividualTimelineEvent = (*WillEvent)(nil)
)

func (e *WillEvent) Type() string             { return "will" }
func (e *WillEvent) ShortDescription() string { return e.abbrev("will") }
func (e *WillEvent) What() string             { return "wrote a will" }

// ResidenceRecordedEvent represents the event of a person's occupation being recorded / noted
type ResidenceRecordedEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *ResidenceRecordedEvent) Type() string               { return "residence" }
func (e *ResidenceRecordedEvent) ShortDescription() string   { return e.abbrev("res") }
func (e *ResidenceRecordedEvent) What() string               { return "resided" }
func (e *ResidenceRecordedEvent) GetDateType() EventDateType { return EventDateTypeRecorded }
func (e *ResidenceRecordedEvent) GetTitle() string           { return "residence" }

var (
	_ TimelineEvent           = (*ResidenceRecordedEvent)(nil)
	_ IndividualTimelineEvent = (*ResidenceRecordedEvent)(nil)
)

// MarriageEvent represents the joining of two people in marriage in a timeline
type MarriageEvent struct {
	GeneralEvent
	GeneralPartyEvent
}

var (
	_ TimelineEvent      = (*MarriageEvent)(nil)
	_ PartyTimelineEvent = (*MarriageEvent)(nil)
)

func (e *MarriageEvent) Type() string             { return "marriage" }
func (e *MarriageEvent) ShortDescription() string { return e.abbrev("m") }
func (e *MarriageEvent) What() string             { return "married" }

// MarriageLicenseEvent represents the event where two people obtain a license to marry
type MarriageLicenseEvent struct {
	GeneralEvent
	GeneralPartyEvent
}

var (
	_ TimelineEvent      = (*MarriageLicenseEvent)(nil)
	_ PartyTimelineEvent = (*MarriageLicenseEvent)(nil)
)

func (e *MarriageLicenseEvent) Type() string             { return "marriage license" }
func (e *MarriageLicenseEvent) ShortDescription() string { return e.abbrev("lic.") }
func (e *MarriageLicenseEvent) What() string             { return "obtained licensed to marry" }

func (e *MarriageLicenseEvent) GetParty1() *Person {
	return e.Party1
}

func (e *MarriageLicenseEvent) GetParty2() *Person {
	return e.Party2
}

// MarriageBannsEvent represents the event that public notice is given that two people intend to marry
type MarriageBannsEvent struct {
	GeneralEvent
	GeneralPartyEvent
}

var (
	_ TimelineEvent      = (*MarriageBannsEvent)(nil)
	_ PartyTimelineEvent = (*MarriageBannsEvent)(nil)
)

func (e *MarriageBannsEvent) Type() string             { return "marriage banns" }
func (e *MarriageBannsEvent) ShortDescription() string { return e.abbrev("ban") }
func (e *MarriageBannsEvent) What() string             { return "had marriage banns read" }

func (e *MarriageBannsEvent) GetParty1() *Person {
	return e.Party1
}

func (e *MarriageBannsEvent) GetParty2() *Person {
	return e.Party2
}

// DivorceEvent represents the ending of a marriage by divorce in a timeline
type DivorceEvent struct {
	GeneralEvent
	GeneralPartyEvent
}

var (
	_ TimelineEvent      = (*DivorceEvent)(nil)
	_ PartyTimelineEvent = (*DivorceEvent)(nil)
)

func (e *DivorceEvent) Type() string             { return "divorce" }
func (e *DivorceEvent) ShortDescription() string { return e.abbrev("div") }
func (e *DivorceEvent) What() string             { return "divorced" }

func (e *DivorceEvent) GetParty1() *Person {
	return e.Party1
}

func (e *DivorceEvent) GetParty2() *Person {
	return e.Party2
}

// AnnulmentEvent represents the ending of a marriage by anulment in a timeline
type AnnulmentEvent struct {
	GeneralEvent
	GeneralPartyEvent
}

var (
	_ TimelineEvent      = (*AnnulmentEvent)(nil)
	_ PartyTimelineEvent = (*AnnulmentEvent)(nil)
)

func (e *AnnulmentEvent) Type() string             { return "annulment" }
func (e *AnnulmentEvent) ShortDescription() string { return e.abbrev("anul") }
func (e *AnnulmentEvent) What() string             { return "had marriage anulled" }

func (e *AnnulmentEvent) GetParty1() *Person {
	return e.Party1
}

func (e *AnnulmentEvent) GetParty2() *Person {
	return e.Party2
}

// FamilyStartEvent represents the start of a family grouping, if no other more specific event is available
type FamilyStartEvent struct {
	GeneralEvent
	GeneralPartyEvent
}

var (
	_ TimelineEvent      = (*FamilyStartEvent)(nil)
	_ PartyTimelineEvent = (*FamilyStartEvent)(nil)
)

func (e *FamilyStartEvent) ShortDescription() string { return e.abbrev("start") }
func (e *FamilyStartEvent) What() string             { return "started a family" }

func (e *FamilyStartEvent) GetParty1() *Person {
	return e.Party1
}

func (e *FamilyStartEvent) GetParty2() *Person {
	return e.Party2
}

// FamilyEndEvent represents the end of a family grouping, if no other more specific event is available
type FamilyEndEvent struct {
	GeneralEvent
	GeneralPartyEvent
}

var (
	_ TimelineEvent      = (*FamilyEndEvent)(nil)
	_ PartyTimelineEvent = (*FamilyEndEvent)(nil)
)

func (e *FamilyEndEvent) ShortDescription() string { return e.abbrev("end") }
func (e *FamilyEndEvent) What() string             { return "ended a family" }

func (e *FamilyEndEvent) GetParty1() *Person {
	return e.Party1
}

func (e *FamilyEndEvent) GetParty2() *Person {
	return e.Party2
}

// PartyNarrativeEvent represents some narrative that can be used as-is
type PartyNarrativeEvent struct {
	GeneralEvent
	GeneralPartyEvent
}

func (e *PartyNarrativeEvent) ShortDescription() string { return e.abbrev("narr") }
func (e *PartyNarrativeEvent) What() string             { return "narrative" }

var (
	_ TimelineEvent      = (*PartyNarrativeEvent)(nil)
	_ PartyTimelineEvent = (*PartyNarrativeEvent)(nil)
)

type EventMatcher func(TimelineEvent) bool

// FilterEventList returns a new slice that includes only the events that match the
// supplied EventMatcher
func FilterEventList(evs []TimelineEvent, include EventMatcher) []TimelineEvent {
	switch len(evs) {
	case 0:
		return []TimelineEvent{}
	case 1:
		if include(evs[0]) {
			return []TimelineEvent{evs[0]}
		}
		return []TimelineEvent{}
	default:
		l := make([]TimelineEvent, 0, len(evs))
		for _, ev := range evs {
			if include(ev) {
				l = append(l, ev)
			}
		}
		return l
	}
}

// CollapseEventList returns a new slice that includes only unique events
func CollapseEventList(evs []TimelineEvent) []TimelineEvent {
	switch len(evs) {
	case 0:
		return []TimelineEvent{}
	case 1:
		return []TimelineEvent{evs[0]}
	case 2:
		if evs[0] == evs[1] {
			return []TimelineEvent{evs[0]}
		}
		return []TimelineEvent{evs[0], evs[1]}
	default:
		seen := make(map[TimelineEvent]bool)
		l := make([]TimelineEvent, 0, len(evs))
		for _, ev := range evs {
			if seen[ev] {
				continue
			}
			seen[ev] = true
			l = append(l, ev)
		}
		return l
	}
}

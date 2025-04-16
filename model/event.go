package model

import (
	"sort"
	"time"
)

type TimelineEvent interface {
	GetDate() *Date
	GetPlace() *Place
	GetDetail() string
	GetNarrative() Text
	GetCitations() []*GeneralCitation
	GetAttribute(name string) (string, bool)
	Type() string                    // name of the type of event, usually a single word
	What() string                    // text description of what happened, an active verb in the past tense, such as married, born, died
	When() string                    // text description of date
	Where() string                   // text description of place
	IsInferred() bool                // whether or not the event was inferred to exist, i.e. has no supporting evidence
	DirectlyInvolves(p *Person) bool // whether or not the event directly involves a person as a principal or party
	GetParticipants() []*EventParticipant
	GetParticipantsByRole(EventRole) []*EventParticipant
	GetMediaObjects() []*CitedMediaObject
}

// IndividualTimelineEvent is a timeline event involving one individual.
type IndividualTimelineEvent interface {
	TimelineEvent
	GetPrincipal() *Person
}

// UnionTimelineEvent is a timeline event involving the union of two parties.
type UnionTimelineEvent interface {
	TimelineEvent
	GetHusband() *Person
	GetWife() *Person
	GetOther(p *Person) *Person // returns the first of party1 or party2 that is not p
}

// MultipartyTimelineEvent is a timeline event involving multiple principal parties.
type MultipartyTimelineEvent interface {
	TimelineEvent
	GetPrincipals() []*Person
}

// EventSortsBefore reports whether ev sorts before other
func EventSortsBefore(ev, other TimelineEvent) bool {
	if ev == nil || ev.GetDate() == nil {
		return false
	}
	if other == nil || other.GetDate() == nil {
		return true
	}

	if ev.GetDate().SortsBefore(other.GetDate()) {
		return true
	}

	if ev.GetDate().SameDate(other.GetDate()) {
		if sb, ok := ev.(interface {
			SortsBefore(other TimelineEvent) bool
		}); ok {
			return sb.SortsBefore(other)
		}
	}

	return false
}

type EventRole string

const (
	EventRoleUnknown     EventRole = "unknown"
	EventRolePrincipal   EventRole = "principal"
	EventRoleHusband     EventRole = "husband"
	EventRoleWife        EventRole = "wife"
	EventRoleWitness     EventRole = "witness" // witness to a marriage, will or other legal document
	EventRoleGodparent   EventRole = "godparent"
	EventRoleBeneficiary EventRole = "beneficiary" // beneficiary of a will
	EventRoleExecutor    EventRole = "executor"    // executor of a will
)

type EventParticipant struct {
	Person *Person
	Role   EventRole
}

func (e *EventParticipant) IsUnknown() bool {
	return e == nil || e.Person.IsUnknown()
}

func SortTimelineEvents(evs []TimelineEvent) {
	sort.Slice(evs, func(i, j int) bool {
		return EventSortsBefore(evs[i], evs[j])
	})
}

const (
	EventAttributeEmployer  = "employer"
	EventAttributeService   = "service" // military service: merchant navy, army, royal artillery etc.
	EventAttributeRegiment  = "regiment"
	EventAttributeBattalion = "battalion"
	EventAttributeCompany   = "company"
	EventAttributeRank      = "rank"
)

type GeneralEvent struct {
	Date         *Date
	Place        *Place
	Title        string // used for the return value of "What()"
	Detail       string
	Citations    []*GeneralCitation
	Inferred     bool
	Narrative    Text // hand written narrative, if any
	Attributes   map[string]string
	MediaObjects []*CitedMediaObject
	UpdateTime   *time.Time // time of last update, if known
	CreateTime   *time.Time // time of creation if known
}

func (e *GeneralEvent) GetDate() *Date {
	return e.Date
}

func (e *GeneralEvent) When() string {
	return e.Date.When()
}

func (e *GeneralEvent) GetPlace() *Place {
	return e.Place
}

func (e *GeneralEvent) Where() string {
	return e.Place.Where()
}

func (e *GeneralEvent) GetDetail() string {
	return e.Detail
}

func (e *GeneralEvent) GetNarrative() Text {
	return e.Narrative
}

func (e *GeneralEvent) GetAttribute(name string) (string, bool) {
	v, ok := e.Attributes[name]
	return v, ok
}

func (e *GeneralEvent) GetCitations() []*GeneralCitation {
	return e.Citations
}

func (e *GeneralEvent) GetMediaObjects() []*CitedMediaObject {
	return e.MediaObjects
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

func (e *GeneralEvent) What() string {
	if e.Title != "" {
		return e.Title
	}
	return e.Type()
}

func (e *GeneralEvent) Created() (time.Time, bool) {
	if e.CreateTime == nil {
		return time.Time{}, false
	}
	return *e.CreateTime, true
}

func (e *GeneralEvent) Updated() (time.Time, bool) {
	if e.UpdateTime == nil {
		return time.Time{}, false
	}
	return *e.UpdateTime, true
}

// GeneralPartyEvent is a general event involving one individual.
type GeneralIndividualEvent struct {
	Principal         *Person
	OtherParticipants []*EventParticipant
}

func (e *GeneralIndividualEvent) GetPrincipal() *Person {
	return e.Principal
}

func (e *GeneralIndividualEvent) DirectlyInvolves(p *Person) bool {
	return e.Principal.SameAs(p)
}

func (e *GeneralIndividualEvent) GetParticipants() []*EventParticipant {
	return []*EventParticipant{{
		Person: e.Principal,
		Role:   EventRolePrincipal,
	}}
}

func (e *GeneralIndividualEvent) GetParticipantsByRole(r EventRole) []*EventParticipant {
	if r == EventRolePrincipal {
		return []*EventParticipant{{
			Person: e.Principal,
			Role:   EventRolePrincipal,
		}}
	}

	var eps []*EventParticipant
	for _, ep := range e.OtherParticipants {
		if ep.Role == r {
			eps = append(eps, ep)
		}
	}
	return eps
}

// GeneralUnionEvent is a general event involving the union of two parties.
type GeneralUnionEvent struct {
	Husband           *Person
	Wife              *Person
	OtherParticipants []*EventParticipant
}

func (e *GeneralUnionEvent) GetHusband() *Person {
	return e.Husband
}

func (e *GeneralUnionEvent) GetWife() *Person {
	return e.Wife
}

func (e *GeneralUnionEvent) GetHusband1() *Person {
	return e.Husband
}

func (e *GeneralUnionEvent) GetWife1() *Person {
	return e.Wife
}

func (e *GeneralUnionEvent) DirectlyInvolves(p *Person) bool {
	return e.Husband.SameAs(p) || e.Wife.SameAs(p)
}

func (e *GeneralUnionEvent) GetOther(p *Person) *Person {
	if !e.Husband.SameAs(p) {
		return e.Husband
	}
	if !e.Wife.SameAs(p) {
		return e.Wife
	}
	return p
}

func (e *GeneralUnionEvent) GetParticipants() []*EventParticipant {
	return []*EventParticipant{
		{
			Person: e.Husband,
			Role:   EventRoleHusband,
		},
		{
			Person: e.Wife,
			Role:   EventRoleWife,
		},
	}
}

func (e *GeneralUnionEvent) GetParticipantsByRole(r EventRole) []*EventParticipant {
	switch r {
	case EventRoleHusband:
		return []*EventParticipant{
			{
				Person: e.Husband,
				Role:   EventRoleHusband,
			},
		}
	case EventRoleWife:
		return []*EventParticipant{
			{
				Person: e.Wife,
				Role:   EventRoleWife,
			},
		}
	default:
		var eps []*EventParticipant
		for _, ep := range e.OtherParticipants {
			if ep.Role == r {
				eps = append(eps, ep)
			}
		}
		return eps
	}
}

// GeneralMultipartyEvent is a general event involving multiple parties.
type GeneralMultipartyEvent struct {
	Participants []*EventParticipant
}

func (e *GeneralMultipartyEvent) DirectlyInvolves(p *Person) bool {
	for _, ep := range e.Participants {
		if ep.Person.SameAs(p) && ep.Role == EventRolePrincipal {
			return true
		}
	}
	return false
}

func (e *GeneralMultipartyEvent) GetPrincipals() []*Person {
	var ps []*Person
	for _, p := range e.Participants {
		if p.Role == EventRolePrincipal {
			ps = append(ps, p.Person)
		}
	}
	return ps
}

func (e *GeneralMultipartyEvent) GetParticipants() []*EventParticipant {
	return e.Participants
}

func (e *GeneralMultipartyEvent) GetParticipantsByRole(r EventRole) []*EventParticipant {
	var eps []*EventParticipant
	for _, ep := range e.Participants {
		if ep.Role == r {
			eps = append(eps, ep)
		}
	}
	return eps
}

// POV represents a point of view. It is used to provide contect when constructing a description of an event.
type POV struct {
	Person *Person // the person observing or experiencing the event
	Place  *Place  // the place in which the observing is taking place
	Family *Family // the family observing or experiencing the event
}

// BirthEvent represents the birth of a person in their timeline
type BirthEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *BirthEvent) Type() string                         { return "birth" }
func (e *BirthEvent) ShortDescription() string             { return e.abbrev("b") }
func (e *BirthEvent) What() string                         { return "born" }
func (e *BirthEvent) SortsBefore(other TimelineEvent) bool { return true }

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
func (e *BaptismEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *BirthEvent:
		return false
	default:
		return true
	}
}

var (
	_ TimelineEvent           = (*BaptismEvent)(nil)
	_ IndividualTimelineEvent = (*BaptismEvent)(nil)
)

// NamingEvent represents the naming of a person in their timeline
type NamingEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *NamingEvent) Type() string             { return "naming" }
func (e *NamingEvent) ShortDescription() string { return e.abbrev("name") }
func (e *NamingEvent) What() string             { return "named" }
func (e *NamingEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *BirthEvent:
		return false
	default:
		return true
	}
}

var (
	_ TimelineEvent           = (*NamingEvent)(nil)
	_ IndividualTimelineEvent = (*NamingEvent)(nil)
)

// DeathEvent represents the death of a person in their timeline
type DeathEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *DeathEvent) Type() string                                { return "death" }
func (e *DeathEvent) ShortDescription() string                    { return e.abbrev("d") }
func (e *DeathEvent) What() string                                { return "died" }
func (e *DeathEvent) PassiveWhat() string                         { return "died" }
func (e *DeathEvent) ConditionalWhat(adverb string) string        { return adverb + " died" }
func (e *DeathEvent) PassiveConditionalWhat(adverb string) string { return adverb + " died" }
func (e *DeathEvent) PresentPerfectWhat() string                  { return "have died" }
func (e *DeathEvent) PastPerfectWhat() string                     { return "had died" }

func (e *DeathEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *BirthEvent, *BaptismEvent:
		return false
	default:
		return true
	}
}

var (
	_ TimelineEvent           = (*DeathEvent)(nil)
	_ IndividualTimelineEvent = (*DeathEvent)(nil)
	_ IrregularWhater         = (*DeathEvent)(nil)
)

// BurialEvent represents the burial of a person in their timeline
type BurialEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *BurialEvent) Type() string             { return "burial" }
func (e *BurialEvent) ShortDescription() string { return e.abbrev("bur") }
func (e *BurialEvent) What() string             { return "buried" }
func (e *BurialEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *BirthEvent, *BaptismEvent, *DeathEvent:
		return false
	default:
		return true
	}
}

var (
	_ TimelineEvent           = (*BurialEvent)(nil)
	_ IndividualTimelineEvent = (*BurialEvent)(nil)
)

// CremationEvent represents the cremation of a person in their timeline
type CremationEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*CremationEvent)(nil)
	_ IndividualTimelineEvent = (*CremationEvent)(nil)
)

func (e *CremationEvent) Type() string { return "cremation" }
func (e *CremationEvent) What() string { return "cremated" }
func (e *CremationEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *BirthEvent, *BaptismEvent, *DeathEvent:
		return false
	default:
		return true
	}
}

// MemorialEvent represents the memorial of a person in their timeline
type MemorialEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*MemorialEvent)(nil)
	_ IndividualTimelineEvent = (*MemorialEvent)(nil)
)

func (e *MemorialEvent) Type() string { return "memorial" }
func (e *MemorialEvent) What() string { return "memorial" }
func (e *MemorialEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *BirthEvent, *BaptismEvent, *DeathEvent, *BurialEvent, *CremationEvent:
		return false
	default:
		return true
	}
}

// EnlistmentEvent represents the enlisting of a person to a military service
type EnlistmentEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *EnlistmentEvent) Type() string             { return "enlistment" }
func (e *EnlistmentEvent) ShortDescription() string { return e.abbrev("enl") }

var (
	_ TimelineEvent           = (*EnlistmentEvent)(nil)
	_ IndividualTimelineEvent = (*EnlistmentEvent)(nil)
)

// MusterEvent represents the recording of a person in a muster call
type MusterEvent struct {
	GeneralEvent
	GeneralMultipartyEvent
}

func (e *MusterEvent) Type() string             { return "muster" }
func (e *MusterEvent) ShortDescription() string { return e.abbrev("must") }

var (
	_ TimelineEvent           = (*MusterEvent)(nil)
	_ MultipartyTimelineEvent = (*MusterEvent)(nil)
)

// BattleEvent represents the recording of a person's participation in a battle
type BattleEvent struct {
	GeneralEvent
	GeneralMultipartyEvent
}

func (e *BattleEvent) Type() string             { return "battle" }
func (e *BattleEvent) ShortDescription() string { return e.abbrev("bat") }

var (
	_ TimelineEvent           = (*BattleEvent)(nil)
	_ MultipartyTimelineEvent = (*BattleEvent)(nil)
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

// ImmigrationEvent represents the immigration of a person to a new country of residence
type ImmigrationEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *ImmigrationEvent) Type() string             { return "immigration" }
func (e *ImmigrationEvent) ShortDescription() string { return e.abbrev("imm") }
func (e *ImmigrationEvent) What() string             { return "immigrated" }

var (
	_ TimelineEvent           = (*ImmigrationEvent)(nil)
	_ IndividualTimelineEvent = (*ImmigrationEvent)(nil)
)

// OccupationEvent represents the the recording of a person's occupation
type OccupationEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *OccupationEvent) Type() string             { return "occupation" }
func (e *OccupationEvent) ShortDescription() string { return e.abbrev("occ") }

var (
	_ TimelineEvent           = (*OccupationEvent)(nil)
	_ IndividualTimelineEvent = (*OccupationEvent)(nil)
)

// ApprenticeEvent represents the commencement of an apprenticeship of a person
type ApprenticeEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *ApprenticeEvent) Type() string             { return "apprentice" }
func (e *ApprenticeEvent) ShortDescription() string { return e.abbrev("app") }
func (e *ApprenticeEvent) What() string             { return "apprenticed" }

var (
	_ TimelineEvent           = (*ApprenticeEvent)(nil)
	_ IndividualTimelineEvent = (*ApprenticeEvent)(nil)
)

// Promotion represents the promotion of a person in their employment
type PromotionEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *PromotionEvent) Type() string             { return "promotion" }
func (e *PromotionEvent) ShortDescription() string { return e.abbrev("prom") }

var (
	_ TimelineEvent           = (*PromotionEvent)(nil)
	_ IndividualTimelineEvent = (*PromotionEvent)(nil)
)

// Demotion represents the demotion of a person in their employment
type DemotionEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *DemotionEvent) Type() string             { return "demotion" }
func (e *DemotionEvent) ShortDescription() string { return e.abbrev("dem") }

var (
	_ TimelineEvent           = (*DemotionEvent)(nil)
	_ IndividualTimelineEvent = (*DemotionEvent)(nil)
)

// IndividualNarrativeEvent represents some narrative that can be used as-is
type IndividualNarrativeEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *IndividualNarrativeEvent) ShortDescription() string { return e.abbrev("narr") }

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
func (e *CensusEvent) What() string             { return "recorded in the census" }

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

func (e *CensusEvent) GetParticipants() []*EventParticipant {
	var ps []*EventParticipant

	// TODO deduplicate
	for _, ce := range e.Entries {
		if ce.Principal != nil {
			ps = append(ps, &EventParticipant{Person: ce.Principal})
		}
	}
	return ps
}

func (e *CensusEvent) GetParticipantsByRole(r EventRole) []*EventParticipant {
	panic("GeneralPartyEvent.GetParticipantsByRole not implemented")
}

func (e *CensusEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *OccupationEvent, *ResidenceRecordedEvent, *EconomicStatusEvent:
		return true
	default:
		return false
	}
}

var _ TimelineEvent = (*CensusEvent)(nil)

// ProbateEvent represents the granting of probate for a person who has died
type ProbateEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*ProbateEvent)(nil)
	_ IndividualTimelineEvent = (*ProbateEvent)(nil)
	_ IrregularWhater         = (*ProbateEvent)(nil)
)

func (e *ProbateEvent) Type() string             { return "probate" }
func (e *ProbateEvent) ShortDescription() string { return e.abbrev("prob") }
func (e *ProbateEvent) What() string             { return "probate granted" }
func (e *ProbateEvent) PassiveWhat() string      { return "probate was read" }
func (e *ProbateEvent) ConditionalWhat(adverb string) string {
	return "probate " + adverb + " granted"
}

func (e *ProbateEvent) PassiveConditionalWhat(adverb string) string {
	return "probate was " + adverb + " granted"
}
func (e *ProbateEvent) PresentPerfectWhat() string { return "have had probate granted" }
func (e *ProbateEvent) PastPerfectWhat() string    { return "had probate granted" }

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

// ResidenceRecordedEvent represents the event of a person's residence being recorded / noted
type ResidenceRecordedEvent struct {
	GeneralEvent
	GeneralMultipartyEvent
}

func (e *ResidenceRecordedEvent) Type() string             { return "residence" }
func (e *ResidenceRecordedEvent) ShortDescription() string { return e.abbrev("res") }
func (e *ResidenceRecordedEvent) What() string             { return "resided" }
func (e *ResidenceRecordedEvent) GetTitle() string         { return "residence" }

func (e *ResidenceRecordedEvent) PassiveWhat() string { return "was resident" }
func (e *ResidenceRecordedEvent) ConditionalWhat(adverb string) string {
	return adverb + " resident"
}

func (e *ResidenceRecordedEvent) PassiveConditionalWhat(adverb string) string {
	return adverb + " resident"
}

func (e *ResidenceRecordedEvent) PresentPerfectWhat() string {
	return "have resided"
}
func (e *ResidenceRecordedEvent) PastPerfectWhat() string { return "had resided" }

var (
	_ TimelineEvent           = (*ResidenceRecordedEvent)(nil)
	_ MultipartyTimelineEvent = (*ResidenceRecordedEvent)(nil)
	_ IrregularWhater         = (*ResidenceRecordedEvent)(nil)
)

// SaleOfPropertyEvent represents the sale of some property person in their timeline
type SaleOfPropertyEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*SaleOfPropertyEvent)(nil)
	_ IndividualTimelineEvent = (*SaleOfPropertyEvent)(nil)
)

func (e *SaleOfPropertyEvent) Type() string             { return "sale of property" }
func (e *SaleOfPropertyEvent) ShortDescription() string { return e.abbrev("sale") }
func (e *SaleOfPropertyEvent) What() string             { return "sold some property" }

// InstitutionEntryEvent represents the admission of a person to an institution
type InstitutionEntryEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*InstitutionEntryEvent)(nil)
	_ IndividualTimelineEvent = (*InstitutionEntryEvent)(nil)
)

func (e *InstitutionEntryEvent) Type() string             { return "institution entry" }
func (e *InstitutionEntryEvent) ShortDescription() string { return e.abbrev("adm") }
func (e *InstitutionEntryEvent) What() string             { return "entered" }

// InstitutionDepartureEvent represents the discharge of a person from an institution
type InstitutionDepartureEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*InstitutionDepartureEvent)(nil)
	_ IndividualTimelineEvent = (*InstitutionDepartureEvent)(nil)
)

func (e *InstitutionDepartureEvent) Type() string             { return "institution departure" }
func (e *InstitutionDepartureEvent) ShortDescription() string { return e.abbrev("adm") }
func (e *InstitutionDepartureEvent) What() string             { return "left" }

// EconomicStatusEvent represents the economic status of a person
type EconomicStatusEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*EconomicStatusEvent)(nil)
	_ IndividualTimelineEvent = (*EconomicStatusEvent)(nil)
)

func (e *EconomicStatusEvent) Type() string             { return "economic status" }
func (e *EconomicStatusEvent) ShortDescription() string { return e.abbrev("anul") }

// CourtEvent represents the appearance of a person in a court or a court hearing they are involved in
type CourtEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

var (
	_ TimelineEvent           = (*CourtEvent)(nil)
	_ IndividualTimelineEvent = (*CourtEvent)(nil)
)

func (e *CourtEvent) Type() string             { return "court" }
func (e *CourtEvent) ShortDescription() string { return e.abbrev("crt") }

// MarriageEvent represents the joining of two people in marriage in a timeline
type MarriageEvent struct {
	GeneralEvent
	GeneralUnionEvent
}

var (
	_ TimelineEvent      = (*MarriageEvent)(nil)
	_ UnionTimelineEvent = (*MarriageEvent)(nil)
)

func (e *MarriageEvent) Type() string             { return "marriage" }
func (e *MarriageEvent) ShortDescription() string { return e.abbrev("m") }
func (e *MarriageEvent) What() string             { return "married" }

func (e *MarriageEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *OccupationEvent, *ResidenceRecordedEvent:
		return true
	default:
		return false
	}
}

// MarriageLicenseEvent represents the event where two people obtain a license to marry
type MarriageLicenseEvent struct {
	GeneralEvent
	GeneralUnionEvent
}

var (
	_ TimelineEvent      = (*MarriageLicenseEvent)(nil)
	_ UnionTimelineEvent = (*MarriageLicenseEvent)(nil)
	_ IrregularWhater    = (*MarriageLicenseEvent)(nil)
)

func (e *MarriageLicenseEvent) Type() string             { return "marriage license" }
func (e *MarriageLicenseEvent) ShortDescription() string { return e.abbrev("lic.") }
func (e *MarriageLicenseEvent) What() string             { return "marriage license obtained" }

func (e *MarriageLicenseEvent) PassiveWhat() string { return "marriage license was obtained" }
func (e *MarriageLicenseEvent) ConditionalWhat(adverb string) string {
	return "marriage license " + adverb + " obtained"
}

func (e *MarriageLicenseEvent) PassiveConditionalWhat(adverb string) string {
	return "marriage license was " + adverb + " obtained"
}
func (e *MarriageLicenseEvent) PresentPerfectWhat() string { return "have obtained a marriage license" }
func (e *MarriageLicenseEvent) PastPerfectWhat() string    { return "had obtained a marriage license" }

func (e *MarriageLicenseEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *OccupationEvent, *ResidenceRecordedEvent, *MarriageEvent:
		return true
	default:
		return false
	}
}

// MarriageBannsEvent represents the event that public notice is given that two people intend to marry
type MarriageBannsEvent struct {
	GeneralEvent
	GeneralUnionEvent
}

var (
	_ TimelineEvent      = (*MarriageBannsEvent)(nil)
	_ UnionTimelineEvent = (*MarriageBannsEvent)(nil)
	_ IrregularWhater    = (*MarriageBannsEvent)(nil)
)

func (e *MarriageBannsEvent) Type() string             { return "marriage banns" }
func (e *MarriageBannsEvent) ShortDescription() string { return e.abbrev("ban") }
func (e *MarriageBannsEvent) What() string             { return "marriage banns read" }
func (e *MarriageBannsEvent) PassiveWhat() string      { return "marriage banns were read" }
func (e *MarriageBannsEvent) ConditionalWhat(adverb string) string {
	return "marriage banns " + adverb + " read"
}

func (e *MarriageBannsEvent) PassiveConditionalWhat(adverb string) string {
	return "marriage banns were " + adverb + " read"
}
func (e *MarriageBannsEvent) PresentPerfectWhat() string { return "have had marriage banns read" }
func (e *MarriageBannsEvent) PastPerfectWhat() string    { return "had marriage banns read" }

func (e *MarriageBannsEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *OccupationEvent, *ResidenceRecordedEvent, *MarriageEvent:
		return true
	default:
		return false
	}
}

// DivorceEvent represents the ending of a marriage by divorce in a timeline
type DivorceEvent struct {
	GeneralEvent
	GeneralUnionEvent
}

var (
	_ TimelineEvent      = (*DivorceEvent)(nil)
	_ UnionTimelineEvent = (*DivorceEvent)(nil)
)

func (e *DivorceEvent) Type() string             { return "divorce" }
func (e *DivorceEvent) ShortDescription() string { return e.abbrev("div") }
func (e *DivorceEvent) What() string             { return "divorced" }
func (e *DivorceEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *OccupationEvent, *ResidenceRecordedEvent:
		return true
	default:
		return false
	}
}

// AnnulmentEvent represents the ending of a marriage by anulment in a timeline
type AnnulmentEvent struct {
	GeneralEvent
	GeneralUnionEvent
}

var (
	_ TimelineEvent      = (*AnnulmentEvent)(nil)
	_ UnionTimelineEvent = (*AnnulmentEvent)(nil)
)

func (e *AnnulmentEvent) Type() string             { return "annulment" }
func (e *AnnulmentEvent) ShortDescription() string { return e.abbrev("anul") }
func (e *AnnulmentEvent) What() string             { return "had marriage anulled" }
func (e *AnnulmentEvent) SortsBefore(other TimelineEvent) bool {
	switch other.(type) {
	case *OccupationEvent, *ResidenceRecordedEvent:
		return true
	default:
		return false
	}
}

// PossibleBirthEvent represents an indirect and unproven record of the birth of a person in their timeline
type PossibleBirthEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *PossibleBirthEvent) Type() string                         { return "possible birth" }
func (e *PossibleBirthEvent) ShortDescription() string             { return e.abbrev("poss. b") }
func (e *PossibleBirthEvent) What() string                         { return "possibly born" }
func (e *PossibleBirthEvent) SortsBefore(other TimelineEvent) bool { return true }

var (
	_ TimelineEvent           = (*PossibleBirthEvent)(nil)
	_ IndividualTimelineEvent = (*PossibleBirthEvent)(nil)
)

// PossibleDeathEvent represents an indirect and unproven record of the death of a person in their timeline
type PossibleDeathEvent struct {
	GeneralEvent
	GeneralIndividualEvent
}

func (e *PossibleDeathEvent) Type() string                         { return "possible death" }
func (e *PossibleDeathEvent) ShortDescription() string             { return e.abbrev("poss. d") }
func (e *PossibleDeathEvent) What() string                         { return "possibly died" }
func (e *PossibleDeathEvent) SortsBefore(other TimelineEvent) bool { return true }

var (
	_ TimelineEvent           = (*PossibleDeathEvent)(nil)
	_ IndividualTimelineEvent = (*PossibleDeathEvent)(nil)
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

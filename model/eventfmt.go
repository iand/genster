package model

func AbbrevWhatWhen(ev TimelineEvent) string {
	prefix := "ev"
	switch ev.(type) {
	case *BirthEvent:
		prefix = "b"
	case *BaptismEvent:
		prefix = "bap"
	case *DeathEvent:
		prefix = "d"
	case *BurialEvent:
		prefix = "bur"
	case *CremationEvent:
		prefix = "crem"
	case *MarriageEvent:
		prefix = "m"
	case *MarriageLicenseEvent:
		prefix = "lic"
	case *MarriageBannsEvent:
		prefix = "ban"
	case *DivorceEvent:
		prefix = "div"
	case *AnnulmentEvent:
		prefix = "anul"
	}

	dt := ev.GetDate()
	if dt.IsUnknown() {
		return prefix + ". ?"
	}

	qual := dt.Derivation.Abbrev()
	if qual != "" {
		qual += " "
	}

	return prefix + ". " + qual + dt.Date.String()
}

func AbbrevWhatWhenWhere(ev TimelineEvent) string {
	if ev.GetPlace().IsUnknown() {
		return AbbrevWhatWhen(ev)
	}
	return AbbrevWhatWhen(ev) + ", " + ev.GetPlace().PreferredLocalityName
}

type Whater interface {
	What() string
}

// optional interface for event grammar
type IrregularWhater interface {
	PassiveWhat() string                         // text description of what happened, a passive verb in the past tense, such as "was married", "was born", "died"
	ConditionalWhat(adverb string) string        // text description of what happened, an active verb in the past tense with a conditonal, such as "probably married", "probate probably granted"
	PassiveConditionalWhat(adverb string) string // text description of what happened, a passive verb in the past tense with a conditonal, such as was "was probably married", "probate was probably granted"
	PresentPerfectWhat() string                  // text description of what happened, a passive verb in the present perfect tense, usually prefixed by "inferred to ", such as "[inferred to ]have been married", "[inferred to ]have died"
	PastPerfectWhat() string                     // text description of what happened, a passive verb in the past perfect tense, such as "had been married", "had died"
}

// What returns an active verb phrase in the past tense describing what happened.
func What(w Whater) string {
	return w.What()
}

// PassiveWhat returns a passive verb phrase in the past tense describing what happened.
func PassiveWhat(w Whater) string {
	if iw, ok := w.(IrregularWhater); ok {
		return iw.PassiveWhat()
	}
	return "was " + w.What()
}

// ConditionalWhat returns an active verb phrase with a conditional such as "probably" in the past tense describing what happened.
func ConditionalWhat(w Whater, adverb string) string {
	if iw, ok := w.(IrregularWhater); ok {
		return iw.ConditionalWhat(adverb)
	}
	return adverb + " " + w.What()
}

// PassiveConditionalWhat returns a passive verb phrase with a conditional such as "probably" in the past tense describing what happened.
func PassiveConditionalWhat(w Whater, adverb string) string {
	if iw, ok := w.(IrregularWhater); ok {
		return iw.PassiveConditionalWhat(adverb)
	}
	return "was " + adverb + " " + w.What()
}

func PresentPerfectWhat(w Whater) string {
	if gev, ok := w.(IrregularWhater); ok {
		return gev.PresentPerfectWhat()
	}
	return "have been " + w.What()
}

func PastPerfectWhat(w Whater) string {
	if iw, ok := w.(IrregularWhater); ok {
		return iw.PastPerfectWhat()
	}
	return "had been " + w.What()
}

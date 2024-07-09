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

package debug

import (
	"fmt"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
)

func ObjectTitle(obj any) string {
	if obj == nil {
		return "none"
	}
	switch tobj := obj.(type) {
	case model.IndividualTimelineEvent:
		return fmt.Sprintf("%s [d=%s; pl=%s; p=%s]", tobj.Type(), ObjectTitle(tobj.GetDate()), ObjectTitle(tobj.GetPlace()), ObjectTitle(tobj.GetPrincipal()))
	case model.PartyTimelineEvent:
		return fmt.Sprintf("%s [d=%s; pl=%s; p1=%s; p2=%s]", tobj.Type(), ObjectTitle(tobj.GetDate()), ObjectTitle(tobj.GetPlace()), ObjectTitle(tobj.GetParty1()), ObjectTitle(tobj.GetParty2()))
	case model.TimelineEvent:
		return fmt.Sprintf("%s [d=%s; pl=%s]", tobj.GetTitle(), ObjectTitle(tobj.GetDate()), ObjectTitle(tobj.GetPlace()))
	case *model.Person:
		return fmt.Sprintf("%s (%s)", tobj.PreferredUniqueName, tobj.ID)
	case *model.Place:
		if tobj == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%s (%s)", tobj.PreferredUniqueName, tobj.ID)
	case *model.Date:
		return tobj.String()
	case gdate.Date:
		return tobj.String()
	}

	return "unknown type"
}

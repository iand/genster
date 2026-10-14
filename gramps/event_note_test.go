package gramps

import (
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/iand/grampsxml"
)

// TestEventNoteSharedByEventsIsOneCitation checks that an event note attached
// to several events becomes a single citation registered with the tree, so
// every event refers to the same citation object.
func TestEventNoteSharedByEventsIsOneCitation(t *testing.T) {
	noteID := "N0001"
	note := &grampsxml.Note{
		ID:     &noteID,
		Handle: "_note",
		Type:   "Event Note",
		Text:   "The baptism register is missing entries between 1755 and 1768",
	}

	l := &Loader{
		ScopeName:       "test",
		NotesByHandle:   map[string]*grampsxml.Note{note.Handle: note},
		EventsByHandle:  map[string]*grampsxml.Event{},
		PlacesByHandle:  map[string]*grampsxml.Placeobj{},
		populatedPlaces: map[string]bool{},
		timelineEvents:  map[string]model.TimelineEvent{},
	}
	tr := tree.NewTree("test", nil, &tree.SurnameGroups{})

	birth := "Birth"
	ids := []string{"E0001", "E0002"}
	for _, id := range ids {
		grev := &grampsxml.Event{
			ID:      &id,
			Handle:  "_" + id,
			Type:    &birth,
			Noteref: []grampsxml.Noteref{{Hlink: note.Handle}},
		}
		l.EventsByHandle[grev.Handle] = grev
		if err := l.populateEventFacts(tr, grev); err != nil {
			t.Fatalf("populateEventFacts(%s): %v", id, err)
		}
	}

	var cits []*model.GeneralCitation
	for _, id := range ids {
		ev, ok := l.timelineEvents[persistentID(&id, "_"+id)]
		if !ok {
			t.Fatalf("event %s not created", id)
		}
		if len(ev.GetCitations()) != 1 {
			t.Fatalf("event %s has %d citations, want 1", id, len(ev.GetCitations()))
		}
		cits = append(cits, ev.GetCitations()[0])
	}

	if cits[0] != cits[1] {
		t.Errorf("events refer to different citation objects for the same note")
	}
	if _, ok := tr.Citations[cits[0].ID]; !ok {
		t.Errorf("note citation %q is not registered with the tree", cits[0].ID)
	}
	if cits[0].Detail != note.Text {
		t.Errorf("citation detail is %q, want the note text", cits[0].Detail)
	}
}

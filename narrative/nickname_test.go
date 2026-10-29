package narrative

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// A person with known parents and no birth-like event has their nickname once, directly after
// their given name.
func TestIntroNicknameWithoutBirth(t *testing.T) {
	father := &model.Person{ID: "f", PreferredUniqueName: "James Hall", PreferredFamiliarName: "James", Gender: model.GenderMale}
	p := &model.Person{ID: "p", PreferredGivenName: "James", PreferredFamiliarName: "James", NickName: "Jimmy", Gender: model.GenderMale, Father: father, Mother: model.UnknownPerson()}

	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
	enc := &md.Content{}
	(&IntroStatement[md.Text]{Principal: p}).RenderDetail(0, intro, enc, &DefaultNameChooser{})
	got := enc.String()

	if n := strings.Count(got, "(known as Jimmy)"); n != 1 {
		t.Errorf("nickname appears %d times, want 1:\n%s", n, got)
	}
	checkContains(t, got, "James (known as Jimmy) was the son of")
}

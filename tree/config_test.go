package tree

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestReadConfig(t *testing.T) {
	const kdlContent = `
tree id="cg" {
    name "Chambers and Guiver Family Tree"
    description r#"
        The Chambers family originated from Suffolk, England.
        The Guivers are on Ian's paternal side.
        "#
}

surname-groups {
    Dockrell "Dockaril" "Dockarell" "Dockarill"
    Martin "Martyn"
}
`
	f, err := os.CreateTemp("", "treeconfig-*.kdl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(kdlContent); err != nil {
		t.Fatal(err)
	}
	f.Close()

	got, err := ReadConfig(f.Name())
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}

	want := &Config{
		ID:          "cg",
		Name:        "Chambers and Guiver Family Tree",
		Description: "The Chambers family originated from Suffolk, England.\n        The Guivers are on Ian's paternal side.",
	}

	if diff := cmp.Diff(want, got, cmpopts.IgnoreFields(Config{}, "SurnameGroups", "Annotations")); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	if got.SurnameGroups == nil {
		t.Fatal("SurnameGroups is nil")
	}

	// Check Dockrell group
	g, ok := got.SurnameGroups.Lookup("Dockrell")
	if !ok {
		t.Error("expected Dockrell group")
	} else if g.Surname != "Dockrell" {
		t.Errorf("Dockrell canonical: got %q, want %q", g.Surname, "Dockrell")
	}
	for _, v := range []string{"Dockaril", "Dockarell", "Dockarill"} {
		g2, ok := got.SurnameGroups.Lookup(v)
		if !ok {
			t.Errorf("expected lookup of variant %q to succeed", v)
		} else if g2 != g {
			t.Errorf("variant %q points to different group", v)
		}
	}

	// Check Martin group
	g, ok = got.SurnameGroups.Lookup("Martin")
	if !ok {
		t.Error("expected Martin group")
	} else if g.Surname != "Martin" {
		t.Errorf("Martin canonical: got %q, want %q", g.Surname, "Martin")
	}
	if _, ok := got.SurnameGroups.Lookup("Martyn"); !ok {
		t.Error("expected Martyn to resolve to Martin group")
	}
}

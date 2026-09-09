package genbio

import "testing"

func TestVariantDeterministic(t *testing.T) {
	alts := []string{"a", "b", "c", "d"}

	v1 := NewVariant("person-123")
	v2 := NewVariant("person-123")
	for i := range 20 {
		g1 := v1.Pick(alts...)
		g2 := v2.Pick(alts...)
		if g1 != g2 {
			t.Fatalf("pick %d differs for identical seed: %q vs %q", i, g1, g2)
		}
	}
}

func TestVariantSingleAlternative(t *testing.T) {
	v := NewVariant("seed")
	if got := v.Pick("only"); got != "only" {
		t.Errorf("got %q, want %q", got, "only")
	}
}

func TestVariantEmpty(t *testing.T) {
	v := NewVariant("seed")
	if got := v.Pick(); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestVariantSpreadsAcrossAlternatives(t *testing.T) {
	alts := []string{"a", "b", "c", "d"}
	seen := map[string]bool{}
	for i := range 200 {
		v := NewVariant(string(rune('A'+i%26)) + string(rune('0'+i/26)))
		seen[v.Pick(alts...)] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected picks to spread across alternatives, only saw %v", seen)
	}
}

func TestVariantAdvancesWithinSequence(t *testing.T) {
	alts := []string{"a", "b", "c", "d"}
	differs := false
	for i := range 200 {
		v := NewVariant(string(rune('A'+i%26)) + string(rune('0'+i/26)))
		first := v.Pick(alts...)
		second := v.Pick(alts...)
		if first != second {
			differs = true
			break
		}
	}
	if !differs {
		t.Error("successive picks never differ; chooser is not advancing independently")
	}
}

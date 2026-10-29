package model

import "testing"

func TestPersonCanBeNamed(t *testing.T) {
	testCases := []struct {
		name string
		p    *Person
		want bool
	}{
		{name: "nil", p: nil, want: false},
		{name: "unknown", p: UnknownPerson(), want: false},
		{name: "named", p: &Person{PreferredFullName: "Mary Brown"}, want: true},
		{name: "unidentified", p: &Person{Unidentified: true}, want: false},
		{name: "redacted", p: &Person{Redacted: true}, want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.p.CanBeNamed(); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPersonUnknownNameDescription(t *testing.T) {
	testCases := []struct {
		gender Gender
		want   string
	}{
		{gender: GenderMale, want: "a man whose name is not known"},
		{gender: GenderFemale, want: "a woman whose name is not known"},
		{gender: GenderUnknown, want: "a person whose name is not known"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.gender), func(t *testing.T) {
			p := &Person{Gender: tc.gender, Unidentified: true}
			if got := p.UnknownNameDescription(); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

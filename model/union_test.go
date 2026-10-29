package model

import "testing"

// What is an active verb phrase in the past tense, as TimelineEvent documents.
func TestUnionTimelineEventWhat(t *testing.T) {
	testCases := []struct {
		ev   UnionTimelineEvent
		want string
	}{
		{ev: &MarriageEvent{}, want: "married"},
		{ev: &MarriageLicenseEvent{}, want: "obtained a marriage licence"},
		{ev: &MarriageBannsEvent{}, want: "had marriage banns read"},
		{ev: &DivorceEvent{}, want: "divorced"},
		{ev: &SeparationEvent{}, want: "separated"},
		{ev: &AnnulmentEvent{}, want: "had the marriage annulled"},
	}

	for _, tc := range testCases {
		t.Run(tc.ev.Type(), func(t *testing.T) {
			if got := tc.ev.What(); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUnionTimelineEventTransitiveWhat(t *testing.T) {
	const other = "[Mary Frost](/person/mf/)"

	testCases := []struct {
		ev   UnionTimelineEvent
		want string
	}{
		{ev: &MarriageEvent{}, want: "married [Mary Frost](/person/mf/)"},
		{ev: &MarriageLicenseEvent{}, want: "obtained a licence to marry [Mary Frost](/person/mf/)"},
		{ev: &MarriageBannsEvent{}, want: "had banns read to marry [Mary Frost](/person/mf/)"},
		{ev: &DivorceEvent{}, want: "divorced [Mary Frost](/person/mf/)"},
		{ev: &SeparationEvent{}, want: "separated from [Mary Frost](/person/mf/)"},
		{ev: &AnnulmentEvent{}, want: "had the marriage to [Mary Frost](/person/mf/) annulled"},
	}

	for _, tc := range testCases {
		t.Run(tc.ev.Type(), func(t *testing.T) {
			if got := tc.ev.TransitiveWhat(other); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

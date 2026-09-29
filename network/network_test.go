package network

import "testing"

func TestGetTopDescClampsRequestedCount(t *testing.T) {
	keys := []string{"first", "second"}
	got := GetTopDesc(keys, 5)
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("got %#v, want all available keys", got)
	}
}

func TestGetTopDescRejectsNegativeCount(t *testing.T) {
	if got := GetTopDesc([]string{"first"}, -1); len(got) != 0 {
		t.Fatalf("got %#v, want an empty result", got)
	}
}

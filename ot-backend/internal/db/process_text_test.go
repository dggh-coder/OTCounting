package db

import "testing"

func TestFormatProcessTextWithoutPartsReturnsNonEmptyText(t *testing.T) {
	got := formatProcessText(nil, 0, 0)
	if got != "0H,0M" {
		t.Fatalf("got %q, want %q", got, "0H,0M")
	}
}

func TestFormatProcessTextWithPartsKeepsExistingFormat(t *testing.T) {
	got := formatProcessText([]string{"(17:45-20:00)"}, 2, 15)
	want := "(17:45-20:00) = 2H,15M"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

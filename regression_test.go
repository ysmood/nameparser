package nameparser_test

import (
	"testing"

	"github.com/ysmood/nameparser"
)

func TestNonEmptyDefaultDoesNotBreakParsing(t *testing.T) {
	c := nameparser.NewConstants()
	c.EmptyAttributeDefault = "<empty>"

	h := nameparser.NewWithConstants("John Smith", c)

	if got := h.First(); got != "John" {
		t.Fatalf("first mismatch: got %q want %q", got, "John")
	}
	if got := h.Last(); got != "Smith" {
		t.Fatalf("last mismatch: got %q want %q", got, "Smith")
	}
	if got := h.Middle(); got != "<empty>" {
		t.Fatalf("middle mismatch: got %q want %q", got, "<empty>")
	}
	if got := h.Len(); got != 2 {
		t.Fatalf("len mismatch: got %d want %d", got, 2)
	}
	if h.Unparsable {
		t.Fatalf("expected parsable name, got unparsable")
	}
}

func TestEmptyNamesAreUnparsable(t *testing.T) {
	for _, input := range []string{"", "   ", "\t\n"} {
		h := nameparser.New(input)
		if !h.Unparsable {
			t.Fatalf("expected %q to be unparsable", input)
		}
	}
}

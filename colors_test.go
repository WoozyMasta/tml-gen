package main

import (
	"fmt"
	"testing"
)

func TestRGBHex(t *testing.T) {
	t.Parallel()

	cases := []struct {
		r, g, b int32
		wantHex string
	}{
		{30, 30, 30, "0xFF1E1E1E"},
		{255, 0, 0, "0xFFFF0000"},
		{0, 128, 255, "0xFF0080FF"},
	}

	for _, tc := range cases {
		got := fmt.Sprintf("0x%08X", uint32(rgb(tc.r, tc.g, tc.b)))
		if got != tc.wantHex {
			t.Fatalf("rgb(%d,%d,%d)=%s want %s", tc.r, tc.g, tc.b, got, tc.wantHex)
		}
	}
}

func TestOutlineForKeyDeterministic(t *testing.T) {
	t.Parallel()

	got := fmt.Sprintf("0x%08X", uint32(outlineForKey("a")))
	want := "0xFF3C3C3C"
	if got != want {
		t.Fatalf("outlineForKey(%q)=%s want %s", "a", got, want)
	}
}

func TestHashColorCaseInsensitive(t *testing.T) {
	t.Parallel()

	if hashColor("Test") != hashColor("test") {
		t.Fatal("hashColor should be case-insensitive")
	}
}

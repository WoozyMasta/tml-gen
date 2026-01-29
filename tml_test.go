package main

import "testing"

func TestHashP3D(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want int32
	}{
		{"", 0},
		{"a", 97},
		{"ab", 6363201},
	}

	for _, tc := range cases {
		got := hashP3D(tc.in)
		if got != tc.want {
			t.Fatalf("hashP3D(%q)=%d want %d", tc.in, got, tc.want)
		}
	}
}

func TestI32toa(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   int32
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{123456, "123456"},
	}

	for _, tc := range cases {
		got := i32toa(tc.in)
		if got != tc.want {
			t.Fatalf("i32toa(%d)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestUniqueDisplayName(t *testing.T) {
	t.Parallel()

	used := map[string]struct{}{}
	if got := uniqueDisplayName("house", "dz/structures/house.p3d", used); got != "house" {
		t.Fatalf("uniqueDisplayName base=%q got %q want %q", "house", got, "house")
	}

	used = map[string]struct{}{
		"house":       {},
		"house_wreck": {},
	}
	if got := uniqueDisplayName("house", "dz/structures/wrecks/house.p3d", used); got != "house_1" {
		t.Fatalf("uniqueDisplayName duplicate got %q want %q", got, "house_1")
	}
}

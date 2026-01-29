// Package main provides a TML generator for TerrainBuilder.
package main

import "strings"

// Default template outline color (no color).
const defaultOutline = -1

// outlinePalette provides small variations for subgroups.
var outlinePalette = []int{
	rgb(30, 30, 30),
	rgb(60, 60, 60),
	rgb(90, 90, 90),
	rgb(120, 120, 120),
	rgb(150, 150, 150),
	rgb(180, 180, 180),
}

// rgb returns ARGB color with full alpha.
func rgb(r, g, b int32) int {
	alpha := int32(-1) << 24
	return int(alpha | (r&0xFF)<<16 | (g&0xFF)<<8 | (b & 0xFF))
}

// outlineForKey picks a deterministic outline color from the subgroup key.
func outlineForKey(key string) int {
	if key == "" {
		return defaultOutline
	}
	h := int64(hashP3D(key))
	l := int64(len(outlinePalette))
	idx := int(h % l)
	if idx < 0 {
		idx += int(l)
	}

	return outlinePalette[idx]
}

// hashColor maps arbitrary names to a stable mid-tone color.
func hashColor(name string) int {
	h := hashP3D(strings.ToLower(name))
	mask := int32(0x7F)
	r := 64 + (h & mask)
	g := 64 + ((h >> 7) & mask)
	b := 64 + ((h >> 14) & mask)

	return rgb(r, g, b)
}

// colorForLibrary selects fill/outline based on library naming conventions.
func colorForLibrary(name string) (int, int) {
	lower := strings.ToLower(name)
	tokens := strings.Split(lower, "_")

	idxOf := func(tok string) int {
		for i := range tokens {
			if tokens[i] == tok {
				return i
			}
		}
		return -1
	}

	afterHas := func(idx int, tok string) bool {
		for i := idx + 1; i < len(tokens); i++ {
			if tokens[i] == tok {
				return true
			}
		}
		return false
	}

	subKey := func(idx int) string {
		if idx < 0 || idx+1 >= len(tokens) {
			return ""
		}
		return strings.Join(tokens[idx+1:], "_")
	}

	// Water: blue for rivers, turquoise for ponds.
	if idx := idxOf("water"); idx != -1 {
		fill := rgb(45, 112, 197)
		if afterHas(idx, "river") {
			fill = rgb(35, 90, 190)
		}
		if afterHas(idx, "pond") || afterHas(idx, "ponds") {
			fill = rgb(34, 160, 170)
		}

		return fill, outlineForKey(subKey(idx))
	}

	// Structures: industry/residential/etc.
	if idx := idxOf("structures"); idx != -1 {
		fill := rgb(150, 150, 150)
		switch {
		case afterHas(idx, "industrial"):
			fill = rgb(178, 132, 54)
		case afterHas(idx, "residential"):
			fill = rgb(196, 178, 146)
		case afterHas(idx, "military"):
			fill = rgb(163, 41, 41)
		case afterHas(idx, "roads") || afterHas(idx, "road"):
			fill = rgb(68, 53, 85)
		case afterHas(idx, "rail"):
			fill = rgb(107, 43, 99)
		case afterHas(idx, "ruins"):
			fill = rgb(92, 86, 82)
		case afterHas(idx, "walls"):
			fill = rgb(122, 122, 90)
		case afterHas(idx, "wrecks"):
			fill = rgb(83, 41, 14)
		case afterHas(idx, "signs"):
			fill = rgb(212, 40, 175)
		case afterHas(idx, "furniture"):
			fill = rgb(140, 110, 80)
		case afterHas(idx, "underground"):
			fill = rgb(90, 96, 110)
		}

		return fill, outlineForKey(subKey(idx))
	}

	// Nature and terrain groups.
	if idx := idxOf("plants"); idx != -1 {
		fill := rgb(78, 140, 74)

		return fill, outlineForKey(subKey(idx))
	}

	if idx := idxOf("rocks"); idx != -1 {
		fill := rgb(120, 110, 100)

		return fill, outlineForKey(subKey(idx))
	}

	if idx := idxOf("surfaces"); idx != -1 {
		fill := rgb(165, 147, 111)

		return fill, outlineForKey(subKey(idx))
	}

	if idx := idxOf("worlds"); idx != -1 {
		fill := rgb(90, 110, 140)

		return fill, outlineForKey(subKey(idx))
	}

	// Fallback for any unknown group.
	return hashColor(name), defaultOutline
}

// shapeForLibrary returns the Library shape based on group type.
func shapeForLibrary(name string) string {
	lower := strings.ToLower(name)
	tokens := strings.Split(lower, "_")

	for _, t := range tokens {
		switch t {
		case "plants", "rocks":
			return "ellipse"
		}
	}

	return "rectangle"
}

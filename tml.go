package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// hashP3D hashes a p3d model name string to an int32.
func hashP3D(s string) int32 {
	var h int32
	for i := 0; i < len(s); i++ {
		c := int32(s[i])
		h = c + (h << 6) + (h << 16) - h
	}

	return h
}

// toBackslashes converts a path to backslashes.
func toBackslashes(p string) string {
	return strings.ReplaceAll(filepath.ToSlash(p), "/", `\`)
}

// xmlEscapeText escapes text for XML.
func xmlEscapeText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// xmlEscapeAttr escapes attributes for XML.
func xmlEscapeAttr(s string) string {
	s = xmlEscapeText(s)
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// writeLibraryHeader writes the library header.
func writeLibraryHeader(b *strings.Builder, libraryName string, shape string, fill int, outline int) {
	b.WriteString(`<?xml version="1.0" ?>`)
	b.WriteString("\n")
	b.WriteString(`<Library name="`)
	b.WriteString(xmlEscapeAttr(libraryName))
	b.WriteString(`" shape="`)
	b.WriteString(xmlEscapeAttr(shape))
	b.WriteString(`" default_fill="`)
	fmt.Fprint(b, fill)
	b.WriteString(`" default_outline="`)
	fmt.Fprint(b, outline)
	b.WriteString(`" tex="0">` + "\n")
}

// writeLibraryFooter writes the library footer.
func writeLibraryFooter(b *strings.Builder) {
	b.WriteString(`</Library>`)
	b.WriteString("\n")
}

// writeTemplate writes a template.
func writeTemplate(b *strings.Builder, name string, file string, date string, fill int, outline int, hash int32) {
	b.WriteString("\t<Template>\n\t\t<Name>")
	b.WriteString(xmlEscapeText(name))
	b.WriteString("</Name>\n\t\t<File>")
	b.WriteString(xmlEscapeText(file))
	b.WriteString("</File>\n\t\t<Date>")
	b.WriteString(date)
	b.WriteString("</Date>\n\t\t<Archive></Archive>\n\t\t<Fill>")
	fmt.Fprint(b, fill)
	b.WriteString("</Fill>\n\t\t<Outline>")
	fmt.Fprint(b, outline)
	b.WriteString("</Outline>\n\t\t<Scale>1.000000</Scale>\n\t\t<Hash>")
	b.WriteString(i32toa(hash))
	b.WriteString("</Hash>\n\t\t<ScaleRandMin>0.000000</ScaleRandMin>\n\t\t<ScaleRandMax>0.000000</ScaleRandMax>\n")
	b.WriteString("\t\t<YawRandMin>0.000000</YawRandMin>\n\t\t<YawRandMax>0.000000</YawRandMax>\n")
	b.WriteString("\t\t<PitchRandMin>0.000000</PitchRandMin>\n\t\t<PitchRandMax>0.000000</PitchRandMax>\n")
	b.WriteString("\t\t<RollRandMin>0.000000</RollRandMin>\n\t\t<RollRandMax>0.000000</RollRandMax>\n")
	b.WriteString("\t\t<TexLLU>0.000000</TexLLU>\n\t\t<TexLLV>0.000000</TexLLV>\n")
	b.WriteString("\t\t<TexURU>1.000000</TexURU>\n\t\t<TexURV>1.000000</TexURV>\n")
	b.WriteString("\t\t<BBRadius>-1.000000</BBRadius>\n\t\t<BBHScale>1.000000</BBHScale>\n")
	b.WriteString("\t\t<AutoCenter>0</AutoCenter>\n")
	b.WriteString("\t\t<XShift>0.000000</XShift>\n\t\t<YShift>0.000000</YShift>\n")
	b.WriteString("\t\t<ZShift>0.000000</ZShift>\n\t\t<Height>0.000000</Height>\n")
	b.WriteString("\t\t<BoundingMin X=\"999.000000\" Y=\"999.000000\" Z=\"999.000000\" />\n")
	b.WriteString("\t\t<BoundingMax X=\"-999.000000\" Y=\"-999.000000\" Z=\"-999.000000\" />\n")
	b.WriteString("\t\t<BoundingCenter X=\"-999.000000\" Y=\"-999.000000\" Z=\"-999.000000\" />\n")
	b.WriteString("\t\t<Placement></Placement>\n\t</Template>\n")
}

// i32toa converts an int32 to a string.
func i32toa(x int32) string {
	if x == 0 {
		return "0"
	}

	// Convert negative int32 to positive uint32.
	neg := x < 0
	var u uint32
	if neg {
		u = uint32(^x) + 1
	} else {
		u = uint32(x)
	}

	// Convert uint32 to string.
	var buf [11]byte
	i := len(buf)
	for u > 0 {
		i--
		buf[i] = byte('0' + (u % 10))
		u /= 10
	}

	// Add negative sign if the original int32 was negative.
	if neg {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}

// uniqueDisplayName ensures a stable, global-unique Name across all libraries.
// It only modifies the base name when a duplicate is detected.
func uniqueDisplayName(base string, relPath string, used map[string]struct{}) string {
	baseKey := strings.ToLower(base)
	if _, ok := used[baseKey]; !ok {
		used[baseKey] = struct{}{}

		return base
	}

	lowerPath := strings.ToLower(filepath.ToSlash(relPath))
	segs := splitSegs(relPath)

	candidate := ""
	if len(segs) > 0 && strings.ToLower(segs[0]) != "dz" {
		candidate = base + "_" + segs[0]
	} else if strings.Contains(lowerPath, "wrecks") {
		candidate = base + "_wreck"
	} else if strings.Contains(lowerPath, "ruins") {
		candidate = base + "_ruin"
	} else if strings.Contains(lowerPath, "bliss") {
		candidate = base + "_bliss"
	} else if strings.Contains(lowerPath, "sakhal") {
		candidate = base + "_sakhal"
	} else if strings.Contains(lowerPath, "proxy") {
		candidate = base + "_proxy"
	} else if strings.Contains(lowerPath, "military") {
		candidate = base + "_military"
	} else if strings.Contains(lowerPath, "furniture") {
		candidate = base + "_furniture"
	} else if strings.Contains(lowerPath, "residential") {
		candidate = base + "_residential"
	} else if strings.Contains(lowerPath, "industrial") {
		candidate = base + "_industrial"
	}

	if candidate != "" {
		candKey := strings.ToLower(candidate)
		if _, ok := used[candKey]; !ok {
			used[candKey] = struct{}{}

			return candidate
		}
	}

	for i := 1; ; i++ {
		name := fmt.Sprintf("%s_%d", base, i)
		key := strings.ToLower(name)
		if _, ok := used[key]; !ok {
			used[key] = struct{}{}

			return name
		}
	}
}

// writeTML writes a tml file.
func writeTML(path string, libraryName string, relPaths []string, fill int, outline int, shape string, used map[string]struct{}) error {
	now := time.Now().Format("01/02/06 15:04:05")

	var b strings.Builder
	b.Grow(512 + len(relPaths)*900)

	writeLibraryHeader(&b, libraryName, shape, fill, outline)

	for _, rel := range relPaths {
		segs := splitSegs(rel)
		if len(segs) == 0 {
			continue
		}
		fileName := segs[len(segs)-1]
		ext := filepath.Ext(fileName)
		base := strings.TrimSuffix(fileName, ext)
		displayName := uniqueDisplayName(base, rel, used)
		outFile := toBackslashes(strings.Join(segs, "/"))
		h := hashP3D(base)

		writeTemplate(&b, displayName, outFile, now, fill, outline, h)
	}

	writeLibraryFooter(&b)
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

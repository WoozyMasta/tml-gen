package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// splitSegs splits a relative path into segments.
func splitSegs(rel string) []string {
	rel = filepath.ToSlash(rel)
	parts := strings.Split(rel, "/")
	out := parts[:0]
	for _, p := range parts {
		if p != "" && p != "." {
			out = append(out, p)
		}
	}
	return out
}

// cleanAbs cleans a path and returns it as an absolute path.
func cleanAbs(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}

	// Windows drive root normalization
	// "P:" -> "P:\"
	if len(p) == 2 && p[1] == ':' &&
		((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) {
		return strings.ToUpper(p[:1]) + `:\`
	}

	// Keep "P:\" and "P:/" as drive root
	if len(p) == 3 && p[1] == ':' && (p[2] == '\\' || p[2] == '/') {
		return strings.ToUpper(p[:1]) + `:\`
	}

	// Cross-platform cleanup
	return filepath.Clean(p)
}

// startsWithPathPrefix checks if a path starts with a prefix.
func startsWithPathPrefix(path, prefix string) bool {
	rel, err := filepath.Rel(prefix, path)
	if err != nil {
		return false
	}

	rel = filepath.ToSlash(rel)
	return rel != ".." && !strings.HasPrefix(rel, "../")
}

// normalizeRelForMatch normalizes a relative path for matching.
func normalizeRelForMatch(rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return ""
	}

	rel = filepath.ToSlash(rel)
	rel = path.Clean(rel)
	rel = strings.TrimPrefix(rel, "./")
	rel = strings.TrimLeft(rel, "/")
	if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return ""
	}

	return rel
}

// buildSkipPrefixes builds skip prefixes for matching.
func buildSkipPrefixes(gameRoot string, skip []string) ([]string, error) {
	out := make([]string, 0, len(skip))
	for _, s := range skip {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		if filepath.IsAbs(s) {
			abs := cleanAbs(s)
			if !startsWithPathPrefix(abs, gameRoot) {
				return nil, fmt.Errorf("skip path not under game-root: %s", s)
			}

			rel, err := filepath.Rel(gameRoot, abs)
			if err != nil {
				return nil, fmt.Errorf("skip path error: %s", s)
			}

			s = rel
		}

		s = normalizeRelForMatch(s)
		if s == "" {
			continue
		}
		out = append(out, strings.ToLower(s))
	}

	return out, nil
}

// matchSkip checks if a path matches a skip prefix.
func matchSkip(rel string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return false
	}

	rel = normalizeRelForMatch(rel)
	if rel == "" {
		return false
	}

	rel = strings.ToLower(rel)
	for _, p := range prefixes {
		if strings.HasSuffix(p, "*") {
			base := strings.TrimSuffix(p, "*")
			if base != "" && strings.HasPrefix(rel, base) {
				return true
			}

			continue
		}

		if rel == p || strings.HasPrefix(rel, p+"/") || strings.HasPrefix(rel, p) {
			return true
		}
	}

	return false
}

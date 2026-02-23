package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

func titleCase(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func split(name string) []string {
	return strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '.'
	})
}

func formatKitName(dir string) string {
	words := split(filepath.Base(dir))
	filtered := words[:0]
	for _, w := range words {
		if strings.ToLower(w) != "kit" {
			filtered = append(filtered, titleCase(w))
		}
	}
	return strings.Join(filtered, "-")
}

func tokenize(name string) map[string]bool {
	words := split(name)
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[strings.ToLower(w)] = true
	}
	return m
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func cleanSampleName(displayName string, dirTokens map[string]bool) string {
	words := strings.Fields(displayName)
	filtered := words[:0]
	for _, w := range words {
		if len(w) > 1 && !allDigits(w) && !dirTokens[strings.ToLower(w)] {
			filtered = append(filtered, w)
		}
	}
	return strings.Join(filtered, " ")
}

// deriveKitName builds a kit name from path components between topDir and kitPath.
// E.g. topDir="Input/ExamplePack", kitPath="Input/ExamplePack/Kit 1"
// => "ExamplePack-1"
func deriveKitName(topDir, kitPath string) string {
	rel, err := filepath.Rel(topDir, kitPath)
	if err != nil {
		return formatKitName(kitPath)
	}

	parts := strings.Split(rel, string(filepath.Separator))
	var formatted []string
	// Include the top-level dir name as prefix
	formatted = append(formatted, formatKitName(topDir))
	for _, p := range parts {
		if f := formatKitName(p); f != "" {
			formatted = append(formatted, f)
		}
	}
	joined := strings.Join(formatted, "-")
	tokens := strings.Split(joined, "-")
	deduped := tokens[:1]
	for _, t := range tokens[1:] {
		if !strings.EqualFold(t, deduped[len(deduped)-1]) {
			deduped = append(deduped, t)
		}
	}
	return strings.Join(deduped, "-")
}

// deriveSrcTokens tokenizes all path components between topDir and kitPath (inclusive).
func deriveSrcTokens(topDir, kitPath string) map[string]bool {
	tokens := make(map[string]bool)
	for tok := range tokenize(filepath.Base(topDir)) {
		tokens[tok] = true
	}
	rel, err := filepath.Rel(topDir, kitPath)
	if err != nil {
		return tokens
	}
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		for tok := range tokenize(part) {
			tokens[tok] = true
		}
	}
	return tokens
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1_000_000_000:
		return fmt.Sprintf("%.1f GB", float64(bytes)/1_000_000_000)
	case bytes >= 1_000_000:
		return fmt.Sprintf("%.1f MB", float64(bytes)/1_000_000)
	case bytes >= 1_000:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1_000)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

package sampler

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

func TitleCase(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func Split(name string) []string {
	return strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '.'
	})
}

func FormatKitName(dir string) string {
	words := Split(filepath.Base(dir))
	filtered := words[:0]
	for _, w := range words {
		if strings.ToLower(w) != "kit" {
			filtered = append(filtered, TitleCase(w))
		}
	}
	return strings.Join(filtered, "-")
}

func Tokenize(name string) map[string]bool {
	words := Split(name)
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[strings.ToLower(w)] = true
	}
	return m
}

func AllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func CleanSampleName(displayName string, dirTokens map[string]bool) string {
	words := strings.Fields(displayName)
	filtered := words[:0]
	for _, w := range words {
		if len(w) > 1 && !AllDigits(w) && !dirTokens[strings.ToLower(w)] {
			filtered = append(filtered, w)
		}
	}
	return strings.Join(filtered, " ")
}

// DedupeTokens removes consecutive duplicate tokens from a hyphen-separated name.
func DedupeTokens(name string) string {
	tokens := strings.Split(name, "-")
	deduped := tokens[:1]
	for _, t := range tokens[1:] {
		if !strings.EqualFold(t, deduped[len(deduped)-1]) {
			deduped = append(deduped, t)
		}
	}
	return strings.Join(deduped, "-")
}

// DeriveKitName builds a kit name from path components between topDir and kitPath.
// E.g. topDir="Input/ExamplePack", kitPath="Input/ExamplePack/Kit 1"
// => "ExamplePack-1"
func DeriveKitName(topDir, kitPath string) string {
	rel, err := filepath.Rel(topDir, kitPath)
	if err != nil {
		return FormatKitName(kitPath)
	}

	parts := strings.Split(rel, string(filepath.Separator))
	var formatted []string
	// Include the top-level dir name as prefix
	formatted = append(formatted, FormatKitName(topDir))
	for _, p := range parts {
		if f := FormatKitName(p); f != "" {
			formatted = append(formatted, f)
		}
	}
	return DedupeTokens(strings.Join(formatted, "-"))
}

// DeriveSrcTokens tokenizes all path components between topDir and kitPath (inclusive).
func DeriveSrcTokens(topDir, kitPath string) map[string]bool {
	tokens := make(map[string]bool)
	for tok := range Tokenize(filepath.Base(topDir)) {
		tokens[tok] = true
	}
	rel, err := filepath.Rel(topDir, kitPath)
	if err != nil {
		return tokens
	}
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		for tok := range Tokenize(part) {
			tokens[tok] = true
		}
	}
	return tokens
}

func FormatSize(bytes int64) string {
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

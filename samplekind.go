package main

import (
	"path/filepath"
	"strings"
)

type SampleKind string

func detectSampleKind(filename string) SampleKind {
	name := strings.ToLower(strings.TrimSuffix(filename, filepath.Ext(filename)))
	for _, dt := range cfg.DrumTypes {
		for _, token := range dt.Tokens {
			if strings.Contains(name, token) {
				return SampleKind(dt.Name)
			}
		}
	}
	return ""
}

type Pitch int

const (
	PitchLow Pitch = iota
	PitchMid
	PitchHigh
)

func detectPitch(filename string) Pitch {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	tokens := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return r == ' ' || r == '-' || r == '_'
	})
	for _, t := range tokens {
		switch t {
		case "low", "lo":
			return PitchLow
		case "high", "hi":
			return PitchHigh
		}
	}
	return PitchMid
}

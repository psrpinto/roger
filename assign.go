package main

import "slices"

type sample struct {
	filename   string
	extension  string
	drumKind   SampleKind
	cleanName  string // displayName with redundant directory tokens removed.
	outputName string // The final name the sample will have in the output directory.
	sourcePath string // Full path to the source WAV file.
	frameCount int
	sampleRate int
}

type kitData struct {
	name    string
	kitPath string
	samples [16]sample
}

type kitGroup struct {
	name string // formatted group name; empty string for flat packs
	dir  string // absolute path to group dir (== pack.dir for flat packs)
	kits []kitData
}

type pack struct {
	name   string // formatted display name
	dir    string // absolute path to pack top-level dir
	groups []kitGroup
}

func assignSamples(samples []sample) [16]sample {
	assigned := [16]sample{}
	used := make(map[int]bool)

	slices.SortStableFunc(samples, func(a, b sample) int {
		return int(detectPitch(a.filename)) - int(detectPitch(b.filename))
	})

	// Phase 1: multi-pass type-matched assignment by priority level.
	// All pads get their top-priority match before any pad gets a lower-priority one.
	maxDepth := 0
	for _, pad := range cfg.PadLayout {
		if len(pad) > maxDepth {
			maxDepth = len(pad)
		}
	}
	for priority := range maxDepth {
		for i, pad := range cfg.PadLayout {
			if assigned[i].filename != "" || priority >= len(pad) {
				continue
			}
			target := SampleKind(pad[priority])
			for j, s := range samples {
				if used[j] {
					continue
				}
				if s.drumKind == target {
					assigned[i] = s
					assigned[i].drumKind = SampleKind(cfg.PadLayout[i][0])
					used[j] = true
					break
				}
			}
		}
	}

	// Phase 2: collect remaining samples, fill empty pads
	var remaining []sample
	for j, s := range samples {
		if !used[j] {
			remaining = append(remaining, s)
		}
	}
	ri := 0
	for i := range assigned {
		if assigned[i].filename == "" && ri < len(remaining) {
			assigned[i] = remaining[ri]
			assigned[i].drumKind = SampleKind(cfg.PadLayout[i][0])
			ri++
		}
	}

	return assigned
}

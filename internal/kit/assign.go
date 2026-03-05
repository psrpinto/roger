package kit

import "slices"

func AssignSamples(samples []Sample, padLayout [16][]string) [16]Sample {
	assigned := [16]Sample{}
	used := make(map[int]bool)

	slices.SortStableFunc(samples, func(a, b Sample) int {
		return int(DetectPitch(a.Filename)) - int(DetectPitch(b.Filename))
	})

	// Phase 1: multi-pass type-matched assignment by priority level.
	// All pads get their top-priority match before any pad gets a lower-priority one.
	maxDepth := 0
	for _, pad := range padLayout {
		if len(pad) > maxDepth {
			maxDepth = len(pad)
		}
	}
	for priority := range maxDepth {
		for i, pad := range padLayout {
			if assigned[i].Filename != "" || priority >= len(pad) {
				continue
			}
			target := SampleKind(pad[priority])
			for j, s := range samples {
				if used[j] {
					continue
				}
				if s.DrumKind == target {
					assigned[i] = s
					assigned[i].DrumKind = SampleKind(padLayout[i][0])
					used[j] = true
					break
				}
			}
		}
	}

	// Phase 2: collect remaining samples, fill empty pads
	var remaining []Sample
	for j, s := range samples {
		if !used[j] {
			remaining = append(remaining, s)
		}
	}
	ri := 0
	for i := range assigned {
		if assigned[i].Filename == "" && ri < len(remaining) {
			assigned[i] = remaining[ri]
			assigned[i].DrumKind = SampleKind(padLayout[i][0])
			ri++
		}
	}

	return assigned
}

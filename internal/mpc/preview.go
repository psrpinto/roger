package mpc

import (
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"roger/internal/kit"
	"roger/internal/sampler"
)

// previewPattern defines the 16th-note step positions (0–31) for each drum type
// across a 2-bar boom-bap pattern at 90 BPM.
//
//	Beat:        1 . . .  2 . . .  3 . . .  4 . . .  1 . . .  2 . . .  3 . . .  4 . . .
//	Step:        0 1 2 3  4 5 6 7  8 9 A B  C D E F  0 1 2 3  4 5 6 7  8 9 A B  C D E F
//	Kick:        x . . .  . . . .  . . x .  . . . .  x . . .  . . . .  . . . x  . . . .
//	Snare:       . . . .  x . . .  . . . .  . . . .  . . . .  x . . .  . . . .  . . . .
//	ClosedHiHat: x . . x  x . . x  x . . x  x . . x  x . . x  x . . x  x . . x  x . . .
//	OpenHiHat:   . . . .  . . . .  . . . .  . . . .  . . . .  . . . .  . . . .  . . . x
//	Clap:        . . . .  x . . .  . . . .  . . . .  . . . .  x . . .  . . . .  . . . .
//	Rim:         . . . .  . . . .  . . . .  x . . .  . . . .  . . . .  . . . .  . . . x
//	Tom:         . . . .  . . . .  . . . .  . x x .  . . . .  . . . .  . . . .  . . . .
//	Shaker:      . . . x  . . . .  . . . x  . . . .  . . . x  . . . .  . . . x  . . . .
//	Tambourine:  . . . .  x . . .  . . . .  . . . .  . . . .  x . . .  . . . .  x . . .
//	Cowbell:     . . . .  . . . .  x . . .  . . . .  . . . .  . . . .  x . . .  . . . .
//	Cymbal:      x . . .  . . . .  . . . .  . . . .  . . . .  . . . .  . . . .  . . . .
//	Percussion:  . . . .  . . x .  . . . .  . . . .  . . . .  . . x .  . . . .  . . . .
//	Clave:       x . . x  . . x .  . . x .  . . . .  x . . x  . . x .  . . x .  . . . .
//	Bongo:       . . x .  . x . .  . . . .  x . . .  . . x .  . x . .  . . . .  x . . .
//	Conga:       x . . .  . . . x  . . x .  . . x .  x . . .  . . . x  . . x .  . . x .
//	Cabasa:      . x . x  . x . x  . x . x  . x . x  . x . x  . x . x  . x . x  . x . x
var previewPattern = map[string][]int{
	"Kick":        {0, 10, 16, 27},
	"Snare":       {4, 20},
	"ClosedHiHat": {0, 3, 4, 7, 8, 11, 12, 15, 16, 19, 20, 23, 24, 27, 28},
	"OpenHiHat":   {31},
	"Clap":        {4, 20},
	"Rim":         {12, 31},
	"Tom":         {13, 14},
	"Shaker":      {3, 11, 19, 27},
	"Tambourine":  {4, 20, 28},
	"Cowbell":     {8, 24},
	"Cymbal":      {0},
	"Percussion":  {6, 22},
	"Clave":       {0, 3, 6, 10, 16, 19, 22, 26},
	"Bongo":       {2, 5, 12, 18, 21, 28},
	"Conga":       {0, 7, 10, 14, 16, 23, 26, 30},
	"Cabasa":      {1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31},
}

var defaultPattern = []int{0, 8, 16, 24}

// GeneratePreview creates a drum pattern audio preview for one or more banks of samples.
func GeneratePreview(banks [][16]kit.Sample, padLayout [16][]string, outputPath string) error {
	// Determine output sample rate from the first non-empty sample.
	sampleRate := 0
	for _, bank := range banks {
		for _, s := range bank {
			if s.Filename != "" && s.SampleRate > 0 {
				sampleRate = s.SampleRate
				break
			}
		}
		if sampleRate > 0 {
			break
		}
	}
	if sampleRate == 0 {
		return fmt.Errorf("no valid samples found for preview")
	}

	const bpm = 90
	const stepsPerPattern = 32
	framesPerStep := sampleRate * 60 / bpm / 4 // 16th note duration in frames

	// Cache PCM data to avoid re-reading samples shared across banks.
	pcmCache := make(map[string]*audio.IntBuffer)

	var allSegments [][]int64 // each segment is a stereo interleaved int64 buffer

	for _, bank := range banks {
		// Group pads by their primary drum kind.
		type padInfo struct {
			index int
			s     kit.Sample
		}
		kindPads := make(map[string][]padInfo)
		for i, s := range bank {
			if s.Filename == "" {
				continue
			}
			kind := padLayout[i][0]
			kindPads[kind] = append(kindPads[kind], padInfo{index: i, s: s})
		}

		// Collect all triggers: (stepPosition, sample).
		type trigger struct {
			step   int
			sample kit.Sample
		}
		var triggers []trigger

		for kind, pads := range kindPads {
			steps, ok := previewPattern[kind]
			if !ok {
				steps = defaultPattern
			}
			for i, step := range steps {
				pad := pads[i%len(pads)]
				triggers = append(triggers, trigger{step: step, sample: pad.s})
			}
		}

		// Load PCM data for all triggered samples.
		for _, t := range triggers {
			if _, ok := pcmCache[t.sample.SourcePath]; !ok {
				buf, err := sampler.ReadPCMData(t.sample.SourcePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to read sample for preview: %s\n", t.sample.SourcePath)
					continue
				}
				pcmCache[t.sample.SourcePath] = buf
			}
		}

		// Fixed segment length: exactly 32 steps for clean looping.
		segmentFrames := stepsPerPattern * framesPerStep

		// Mix into a stereo int64 buffer (interleaved: L, R, L, R, ...).
		mix := make([]int64, segmentFrames*2)

		for _, t := range triggers {
			buf, ok := pcmCache[t.sample.SourcePath]
			if !ok {
				continue
			}
			numCh := buf.Format.NumChannels
			startFrame := t.step * framesPerStep
			numFrames := buf.NumFrames()

			for f := 0; f < numFrames; f++ {
				outIdx := (startFrame + f) * 2
				if outIdx+1 >= len(mix) {
					break
				}
				if numCh == 1 {
					v := int64(buf.Data[f])
					mix[outIdx] += v
					mix[outIdx+1] += v
				} else {
					mix[outIdx] += int64(buf.Data[f*numCh])
					mix[outIdx+1] += int64(buf.Data[f*numCh+1])
				}
			}
		}

		allSegments = append(allSegments, mix)
	}

	// Concatenate all segments.
	totalLen := 0
	for _, seg := range allSegments {
		totalLen += len(seg)
	}
	fullMix := make([]int64, 0, totalLen)
	for _, seg := range allSegments {
		fullMix = append(fullMix, seg...)
	}

	if len(fullMix) == 0 {
		return fmt.Errorf("empty preview mix")
	}

	// Normalize: scale so peak ≈ 90% of 16-bit max.
	var peak int64
	for _, v := range fullMix {
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
	}

	var target int64 = 29490 // ~90% of 32767
	scale := 1.0
	if peak > 0 {
		scale = float64(target) / float64(peak)
	}

	// Scale into int output.
	outData := make([]int, len(fullMix))
	for i, v := range fullMix {
		s := int64(float64(v) * scale)
		if s > 32767 {
			s = 32767
		} else if s < -32768 {
			s = -32768
		}
		outData[i] = int(s)
	}

	// Write as stereo 16-bit WAV.
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := wav.NewEncoder(f, sampleRate, 16, 2, 1)
	buf := &audio.IntBuffer{
		Data:           outData,
		Format:         &audio.Format{NumChannels: 2, SampleRate: sampleRate},
		SourceBitDepth: 16,
	}
	if err := enc.Write(buf); err != nil {
		return err
	}
	return enc.Close()
}

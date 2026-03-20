package gui

import (
	"fmt"
	"image/color"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"roger/internal/kit"
	"roger/internal/mpc"
)

func padTypesMatch(padIndex int, kind kit.SampleKind, padLayout [16][]string) bool {
	return slices.Contains(padLayout[padIndex], string(kind))
}

// renderPacksPreview builds the full scrollable preview for all packs,
// including warnings and legend.
func renderPacksPreview(packs []kit.Pack, padColors [16]color.NRGBA, drumTypes []kit.DrumType, padLayout [16][]string, emptyPacks, wrongSampleCount []string) fyne.CanvasObject {
	items := []fyne.CanvasObject{}

	for _, p := range packs {
		items = append(items, renderPackPreview(p, padColors, drumTypes, padLayout))
	}

	// Warnings
	warnings := renderWarnings(packs, emptyPacks, wrongSampleCount)
	if warnings != nil {
		items = append(items, warnings)
	}

	// Legend
	items = append(items, renderLegend())

	return container.NewVBox(items...)
}

// renderPackPreview builds the visual preview for one pack.
func renderPackPreview(p kit.Pack, padColors [16]color.NRGBA, drumTypes []kit.DrumType, padLayout [16][]string) fyne.CanvasObject {
	items := []fyne.CanvasObject{}

	// Pack title
	packTitle := canvas.NewText(p.Name, color.White)
	packTitle.TextStyle = fyne.TextStyle{Bold: true}
	packTitle.TextSize = 16
	items = append(items, packTitle)

	for _, group := range p.Groups {
		// Group name if multiple groups
		if len(p.Groups) > 1 && group.Name != "" {
			groupTitle := canvas.NewText(group.Name, color.NRGBA{R: 180, G: 180, B: 180, A: 255})
			groupTitle.TextSize = 14
			items = append(items, groupTitle)
		}

		for _, k := range group.Kits {
			items = append(items, renderKitGrid(k, padColors, drumTypes, padLayout))
		}
	}

	return container.NewVBox(items...)
}

// renderKitGrid renders a single kit as a 4x4 pad grid with a sample list alongside.
func renderKitGrid(k kit.KitData, padColors [16]color.NRGBA, drumTypes []kit.DrumType, padLayout [16][]string) fyne.CanvasObject {
	// Kit name header
	kitTitle := canvas.NewText(k.Name, color.White)
	kitTitle.TextStyle = fyne.TextStyle{Bold: true}
	kitTitle.TextSize = 13

	// Build 4x4 grid (bottom-to-top like MPC layout)
	grid := container.NewGridWithColumns(4)
	for row := 3; row >= 0; row-- {
		for col := 0; col < 4; col++ {
			i := row*4 + col
			grid.Add(renderPadCell(i, k.Samples[i], padColors[i], drumTypes, padLayout))
		}
	}

	// Sample list
	sampleList := renderSampleList(k.Samples)

	// Combine grid and sample list horizontally
	gridAndList := container.NewHBox(grid, sampleList)

	return container.NewVBox(kitTitle, gridAndList)
}

// renderPadCell renders a single pad cell with colored background.
func renderPadCell(index int, s kit.Sample, padColor color.NRGBA, drumTypes []kit.DrumType, padLayout [16][]string) fyne.CanvasObject {
	// Determine filename color
	var filenameColor color.Color
	filename := s.Filename
	if filename == "" {
		filename = "(empty)"
		filenameColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255} // red
	} else if padTypesMatch(index, kit.DetectSampleKind(s.Filename, drumTypes), padLayout) {
		filenameColor = color.NRGBA{R: 80, G: 220, B: 80, A: 255} // green
	} else {
		filenameColor = color.NRGBA{R: 220, G: 200, B: 60, A: 255} // yellow
	}

	// Pad header: pad number + drum type
	headerStr := fmt.Sprintf("%d %s", index+1, s.DrumKind)
	header := canvas.NewText(truncateStr(headerStr, 18), padColor)
	header.TextStyle = fyne.TextStyle{Bold: true}
	header.TextSize = 11

	// Filename
	fileText := canvas.NewText(truncateStr(filename, 18), filenameColor)
	fileText.TextSize = 10

	// Clean name
	cleanText := canvas.NewText(truncateStr(s.CleanName, 18), color.NRGBA{R: 200, G: 200, B: 200, A: 255})
	cleanText.TextSize = 10

	// Output name
	outText := canvas.NewText(truncateStr(s.OutputName, 18), color.NRGBA{R: 160, G: 160, B: 160, A: 255})
	outText.TextSize = 10

	// Background rectangle
	bg := canvas.NewRectangle(color.NRGBA{R: padColor.R / 6, G: padColor.G / 6, B: padColor.B / 6, A: 255})
	bg.SetMinSize(fyne.NewSize(140, 70))

	// Border using a slightly brighter version of pad color
	border := canvas.NewRectangle(color.NRGBA{R: padColor.R / 3, G: padColor.G / 3, B: padColor.B / 3, A: 255})
	border.SetMinSize(fyne.NewSize(144, 74))

	labels := container.NewVBox(header, fileText, cleanText, outText)
	padded := container.NewPadded(labels)

	return container.NewStack(border, container.NewPadded(container.NewStack(bg, padded)))
}

// renderSampleList renders the list of samples alongside the grid.
func renderSampleList(samples [16]kit.Sample) fyne.CanvasObject {
	items := []fyne.CanvasObject{}
	for i, s := range samples {
		var text string
		if s.Filename == "" {
			text = fmt.Sprintf("%2d  (empty)", i+1)
		} else if s.SampleRate > 0 {
			durationMs := s.FrameCount * 1000 / s.SampleRate
			text = fmt.Sprintf("%2d  %s  %dms", i+1, s.CleanName, durationMs)
		} else {
			text = fmt.Sprintf("%2d  %s", i+1, s.CleanName)
		}

		var textColor color.Color
		if s.Filename == "" {
			textColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
		} else if s.SampleRate > 0 {
			durationMs := s.FrameCount * 1000 / s.SampleRate
			if durationMs >= 2000 {
				textColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
			} else if durationMs >= 1000 {
				textColor = color.NRGBA{R: 220, G: 200, B: 60, A: 255}
			} else {
				textColor = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
			}
		} else {
			textColor = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
		}

		label := canvas.NewText(text, textColor)
		label.TextSize = 11
		items = append(items, label)
	}
	return container.NewVBox(items...)
}

func renderWarnings(packs []kit.Pack, emptyPacks, wrongSampleCount []string) fyne.CanvasObject {
	warningColor := color.NRGBA{R: 220, G: 200, B: 60, A: 255}
	var items []fyne.CanvasObject

	var missingImages []string
	for _, p := range packs {
		if imgPath, _ := mpc.FindImage(p.Dir); imgPath == "" {
			missingImages = append(missingImages, p.Name)
		}
	}

	if len(emptyPacks) > 0 {
		for _, p := range emptyPacks {
			t := canvas.NewText(fmt.Sprintf("warning: %s contains no kit directories with WAV files", p), warningColor)
			t.TextSize = 11
			items = append(items, t)
		}
	}

	if len(wrongSampleCount) > 0 {
		for _, s := range wrongSampleCount {
			t := canvas.NewText(fmt.Sprintf("warning: %s WAV files, expected 16", s), warningColor)
			t.TextSize = 11
			items = append(items, t)
		}
	}

	if len(missingImages) > 0 {
		t := canvas.NewText(fmt.Sprintf("warning: no cover image found for: %s", strings.Join(missingImages, ", ")), warningColor)
		t.TextSize = 11
		items = append(items, t)
		t2 := canvas.NewText("Place an image file in the top-level directory of each pack for the Expansion cover.", color.NRGBA{R: 160, G: 160, B: 160, A: 255})
		t2.TextSize = 11
		items = append(items, t2)
	}

	if len(items) == 0 {
		return nil
	}
	return container.NewVBox(items...)
}

func renderLegend() fyne.CanvasObject {
	dimColor := color.NRGBA{R: 160, G: 160, B: 160, A: 255}

	desc := canvas.NewText("Each grid is a 16-pad MPC kit. The left column lists samples with their duration.", dimColor)
	desc.TextSize = 11

	legendLine := container.NewHBox(
		canvas.NewText("In the grid: ", dimColor),
		newColorLabel("green", color.NRGBA{R: 80, G: 220, B: 80, A: 255}),
		canvas.NewText(" = type matched, ", dimColor),
		newColorLabel("yellow", color.NRGBA{R: 220, G: 200, B: 60, A: 255}),
		canvas.NewText(" = filled from remaining, ", dimColor),
		newColorLabel("red", color.NRGBA{R: 255, G: 80, B: 80, A: 255}),
		canvas.NewText(" = empty pad.", dimColor),
	)

	separator := widget.NewSeparator()
	return container.NewVBox(separator, desc, legendLine)
}

func newColorLabel(text string, c color.Color) *canvas.Text {
	t := canvas.NewText(text, c)
	t.TextSize = 11
	return t
}

func truncateStr(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

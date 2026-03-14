package kits

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"roger/internal/kit"
	"roger/internal/mpc"
	"roger/internal/tui/shared"
)

func padTypesMatch(padIndex int, kind kit.SampleKind, padLayout [16][]string) bool {
	return slices.Contains(padLayout[padIndex], string(kind))
}

const cellWidth = 30

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

func truncateMiddle(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	endLen := (max - 1) / 2
	startLen := max - 1 - endLen
	return string(runes[:startLen]) + "…" + string(runes[len(runes)-endLen:])
}

func padRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func topBorder() string {
	return "┏" + strings.Repeat("━", cellWidth*4) + "┓"
}

func bottomBorder() string {
	return "┗" + strings.Repeat("━", cellWidth*4) + "┛"
}

func heavySeparator() string {
	return "┣" + strings.Repeat("━", cellWidth*4) + "┫"
}

func computeLeftColWidth(p kit.Pack) int {
	w := 0
	for _, group := range p.Groups {
		for _, k := range group.Kits {
			n := utf8.RuneCountInString(k.Name)
			if n > 50 {
				n = 50
			}
			if n > w {
				w = n
			}
			for _, s := range k.Samples {
				n = 4 + utf8.RuneCountInString(s.CleanName)
				if s.SampleRate > 0 {
					n += 1 + len(fmt.Sprintf("%dms", s.FrameCount*1000/s.SampleRate))
				}
				if n > w {
					w = n
				}
			}
		}
	}
	return w + 2
}

func renderKits(p kit.Pack, leftColWidth int, padStyles [16]lipgloss.Style, drumTypes []kit.DrumType, padLayout [16][]string) string {
	var b strings.Builder

	gridWidth := leftColWidth + 4*cellWidth + 2
	packName := truncateMiddle(p.Name, gridWidth-4)
	titleLen := utf8.RuneCountInString(packName)
	totalPad := gridWidth - titleLen - 2
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	fmt.Fprintf(&b, "━%s %s %s━\n",
		strings.Repeat("━", leftPad), shared.Bold.Render(packName), strings.Repeat("━", rightPad))

	blank := strings.Repeat(" ", leftColWidth)
	multipleGroups := len(p.Groups) > 1

	if !multipleGroups {
		fmt.Fprintln(&b, blank+topBorder())
	}

	for gi, group := range p.Groups {
		if multipleGroups {
			if gi > 0 {
				fmt.Fprintln(&b, strings.Repeat("━", leftColWidth)+bottomBorder())
			}
			if group.Name != "" {
				groupName := truncateMiddle(group.Name, gridWidth-4)
				nameLen := utf8.RuneCountInString(groupName)
				gTotalPad := gridWidth - nameLen - 2
				gLeftPad := gTotalPad / 2
				gRightPad := gTotalPad - gLeftPad
				fmt.Fprintf(&b, "%s\n",
					shared.Dim.Render(fmt.Sprintf("━%s %s %s━",
						strings.Repeat("━", gLeftPad), groupName, strings.Repeat("━", gRightPad))))
			}
			fmt.Fprintln(&b, blank+topBorder())
		}

		for ki, k := range group.Kits {
			if ki > 0 {
				fmt.Fprintln(&b, strings.Repeat("━", leftColWidth)+heavySeparator())
			}

			var cleanNames [16]string
			for i, s := range k.Samples {
				left := fmt.Sprintf(" %2d %s", i+1, s.CleanName)
				if s.SampleRate > 0 {
					durationMs := s.FrameCount * 1000 / s.SampleRate
					dur := fmt.Sprintf("%dms", durationMs)
					gap := leftColWidth - utf8.RuneCountInString(left) - utf8.RuneCountInString(dur) - 1
					if gap < 1 {
						gap = 1
					}
					if durationMs >= 2000 {
						dur = shared.Red.Render(dur)
					} else if durationMs >= 1000 {
						dur = shared.Yellow.Render(dur)
					}
					cleanNames[i] = left + strings.Repeat(" ", gap) + dur + " "
				} else {
					cleanNames[i] = left
				}
			}

			lines := gridLines(k.Samples, padStyles, drumTypes, padLayout)
			total := len(lines)
			for li, line := range lines {
				if li == total-1 {
					displayName := truncateMiddle(k.Name, 50)
					boldName := " " + shared.Bold.Render(displayName)
					fmt.Fprint(&b, padRight(boldName, leftColWidth+boldExtraWidth(displayName)))
				} else if li < 16 {
					fmt.Fprint(&b, padRight(cleanNames[li], leftColWidth))
				} else {
					fmt.Fprint(&b, blank)
				}
				fmt.Fprintln(&b, line)
			}
		}
	}
	fmt.Fprintln(&b, strings.Repeat("━", leftColWidth)+bottomBorder())
	return b.String()
}

func boldExtraWidth(s string) int {
	return len(shared.Bold.Render(s)) - utf8.RuneCountInString(s)
}

func gridLines(samples [16]kit.Sample, padStyles [16]lipgloss.Style, drumTypes []kit.DrumType, padLayout [16][]string) []string {
	const innerContent = 27

	innerTop := func(i int, s kit.Sample) string {
		c := padStyles[i]
		label := fmt.Sprintf("%d %s", i+1, s.DrumKind)
		label = truncate(label, 25)
		labelLen := utf8.RuneCountInString(label)
		fill := 25 - labelLen
		return c.Render("╭─ " + label + " " + strings.Repeat("─", fill) + "╮")
	}

	innerSide := func(i int, content string) string {
		c := padStyles[i]
		return c.Render("│") + " " + padRight(truncate(content, innerContent), innerContent) + c.Render("│")
	}

	innerBottom := func(i int) string {
		return padStyles[i].Render("╰" + strings.Repeat("─", 28) + "╯")
	}

	var lines []string
	for row := 3; row >= 0; row-- {
		var top, file, clean, out, bot strings.Builder
		for col := 0; col < 4; col++ {
			i := row*4 + col
			sep := ""
			if col == 0 {
				sep = "┃"
			}
			top.WriteString(sep + innerTop(i, samples[i]))

			padded := padRight(truncate(samples[i].Filename, innerContent), innerContent)
			var fcolor lipgloss.Style
			if samples[i].Filename == "" {
				padded = padRight("(empty)", innerContent)
				fcolor = shared.Red
			} else if padTypesMatch(i, kit.DetectSampleKind(samples[i].Filename, drumTypes), padLayout) {
				fcolor = shared.Green
			} else {
				fcolor = shared.Yellow
			}
			file.WriteString(sep + padStyles[i].Render("│") + " " + fcolor.Render(padded) + padStyles[i].Render("│"))

			clean.WriteString(sep + innerSide(i, samples[i].CleanName))
			out.WriteString(sep + innerSide(i, samples[i].OutputName))
			bot.WriteString(sep + innerBottom(i))
		}
		lines = append(lines, top.String()+"┃")
		lines = append(lines, file.String()+"┃")
		lines = append(lines, clean.String()+"┃")
		lines = append(lines, out.String()+"┃")
		lines = append(lines, bot.String()+"┃")
	}
	return lines
}

func RenderGrids(packs []kit.Pack, padStyles [16]lipgloss.Style, drumTypes []kit.DrumType, padLayout [16][]string) string {
	var b strings.Builder

	globalLeftColWidth := 0
	for _, p := range packs {
		if w := computeLeftColWidth(p); w > globalLeftColWidth {
			globalLeftColWidth = w
		}
	}

	for i, p := range packs {
		if i > 0 {
			fmt.Fprintln(&b)
		}
		b.WriteString(renderKits(p, globalLeftColWidth, padStyles, drumTypes, padLayout))
	}

	return b.String()
}

func RenderLegend() string {
	var b strings.Builder
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, shared.Dim.Render("Each grid is a 16-pad MPC kit. The left column lists samples with their duration."))
	fmt.Fprintf(&b, "%sIn the grid: %s = type matched, %s = filled from remaining, %s = empty pad.%s\n",
		shared.Dim.Render(""), shared.Green.Render(" green "), shared.Yellow.Render(" yellow "), shared.Red.Render(" red "), shared.Dim.Render(""))
	return b.String()
}

func RenderWarnings(packs []kit.Pack, emptyPacks, wrongSampleCount []string) string {
	var b strings.Builder

	var missingImages []string
	for _, p := range packs {
		if imgPath, _ := mpc.FindImage(p.Dir); imgPath == "" {
			missingImages = append(missingImages, p.Name)
		}
	}

	if len(emptyPacks) > 0 {
		fmt.Fprintln(&b)
		for _, p := range emptyPacks {
			fmt.Fprintf(&b, "%s %s contains no kit directories with WAV files\n",
				shared.Yellow.Render("warning:"), p)
		}
	}
	if len(wrongSampleCount) > 0 {
		fmt.Fprintln(&b)
		for _, s := range wrongSampleCount {
			fmt.Fprintf(&b, "%s %s WAV files, expected 16\n",
				shared.Yellow.Render("warning:"), s)
		}
	}
	if len(missingImages) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "%s no cover image found for: %s\n",
			shared.Yellow.Render("warning:"), strings.Join(missingImages, ", "))
		fmt.Fprintln(&b, "Place an image file in the top-level directory of each pack so that it will be used as the cover image for the Expansion.")
	}

	return b.String()
}

func scanningStatus(done, total int) string {
	if total == 0 {
		return " Scanning samples..."
	}
	return fmt.Sprintf(" Scanning samples... (%d/%d)", done, total)
}

func generatingStatus(done, total int) string {
	if total == 0 {
		return " Generating programs..."
	}
	return fmt.Sprintf(" Generating programs... (%d/%d)", done, total)
}

func renderUsage(baseDir string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s organizes drum sample WAV files into 16-pad MPC kits.\n", shared.Bold.Render("roger"))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Workspace: %s\n", shared.Cyan.Render(baseDir))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s   Place your drum sample packs here\n", shared.Green.Render("Kits/"))
	fmt.Fprintf(&b, "  %s  Generated MPC kits and program files appear here\n", shared.Yellow.Render("Output/"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Organize your samples in Kits/ like this:")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s\n", shared.Bold.Render("PackName/"))
	fmt.Fprintf(&b, "    %s\n", shared.Dim.Render("Kit 1/"))
	fmt.Fprintln(&b, "      Kick.wav, Snare.wav, ...")
	fmt.Fprintf(&b, "    %s\n", shared.Dim.Render("Kit 2/"))
	fmt.Fprintln(&b, "      Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Kits can also be grouped:")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s\n", shared.Bold.Render("PackName/"))
	fmt.Fprintf(&b, "    %s\n", shared.Dim.Render("Group A/"))
	fmt.Fprintf(&b, "      %s\n", shared.Dim.Render("Kit 1/"))
	fmt.Fprintln(&b, "        Kick.wav, Snare.wav, ...")
	fmt.Fprintf(&b, "    %s\n", shared.Dim.Render("Group B/"))
	fmt.Fprintf(&b, "      %s\n", shared.Dim.Render("Kit 1/"))
	fmt.Fprintln(&b, "        Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Samples are auto-detected by type (kick, snare, hat, etc.) from their")
	fmt.Fprintln(&b, "filenames and assigned to pads.")

	return b.String()
}

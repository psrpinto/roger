package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// Lipgloss styles
var (
	styleBold   = lipgloss.NewStyle().Bold(true)
	styleDim    = lipgloss.NewStyle().Faint(true)
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

func padTypesMatch(padIndex int, kind SampleKind) bool {
	return slices.Contains(cfg.PadLayout[padIndex], string(kind))
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

func computeLeftColWidth(p pack) int {
	w := 0
	for _, group := range p.groups {
		for _, kit := range group.kits {
			n := utf8.RuneCountInString(kit.name)
			if n > 50 {
				n = 50
			}
			if n > w {
				w = n
			}
			for _, s := range kit.samples {
				n = 4 + utf8.RuneCountInString(s.cleanName)
				if s.sampleRate > 0 {
					n += 1 + len(fmt.Sprintf("%dms", s.frameCount*1000/s.sampleRate))
				}
				if n > w {
					w = n
				}
			}
		}
	}
	return w + 2
}

func renderKits(p pack, leftColWidth int, padStyles [16]lipgloss.Style) string {
	var b strings.Builder

	gridWidth := leftColWidth + 4*cellWidth + 2
	packName := truncateMiddle(p.name, gridWidth-4)
	titleLen := utf8.RuneCountInString(packName)
	totalPad := gridWidth - titleLen - 2
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	fmt.Fprintf(&b, "━%s %s %s━\n",
		strings.Repeat("━", leftPad), styleBold.Render(packName), strings.Repeat("━", rightPad))

	blank := strings.Repeat(" ", leftColWidth)
	multipleGroups := len(p.groups) > 1

	if !multipleGroups {
		fmt.Fprintln(&b, blank+topBorder())
	}

	for gi, group := range p.groups {
		if multipleGroups {
			if gi > 0 {
				fmt.Fprintln(&b, strings.Repeat("━", leftColWidth)+bottomBorder())
			}
			if group.name != "" {
				groupName := truncateMiddle(group.name, gridWidth-4)
				nameLen := utf8.RuneCountInString(groupName)
				gTotalPad := gridWidth - nameLen - 2
				gLeftPad := gTotalPad / 2
				gRightPad := gTotalPad - gLeftPad
				fmt.Fprintf(&b, "%s\n",
					styleDim.Render(fmt.Sprintf("━%s %s %s━",
						strings.Repeat("━", gLeftPad), groupName, strings.Repeat("━", gRightPad))))
			}
			fmt.Fprintln(&b, blank+topBorder())
		}

		for ki, kit := range group.kits {
			if ki > 0 {
				fmt.Fprintln(&b, strings.Repeat("━", leftColWidth)+heavySeparator())
			}

			var cleanNames [16]string
			for i, s := range kit.samples {
				left := fmt.Sprintf(" %2d %s", i+1, s.cleanName)
				if s.sampleRate > 0 {
					durationMs := s.frameCount * 1000 / s.sampleRate
					dur := fmt.Sprintf("%dms", durationMs)
					gap := leftColWidth - utf8.RuneCountInString(left) - utf8.RuneCountInString(dur) - 1
					if gap < 1 {
						gap = 1
					}
					if durationMs >= 2000 {
						dur = styleRed.Render(dur)
					} else if durationMs >= 1000 {
						dur = styleYellow.Render(dur)
					}
					cleanNames[i] = left + strings.Repeat(" ", gap) + dur + " "
				} else {
					cleanNames[i] = left
				}
			}

			lines := gridLines(kit.samples, padStyles)
			total := len(lines)
			for li, line := range lines {
				if li == total-1 {
					displayName := truncateMiddle(kit.name, 50)
					// Bold name needs padding that accounts for visible width only
					boldName := " " + styleBold.Render(displayName)
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

// boldExtraWidth returns the number of extra bytes that lipgloss bold styling
// adds beyond the visible text, so padRight can account for them.
func boldExtraWidth(s string) int {
	return len(styleBold.Render(s)) - utf8.RuneCountInString(s)
}

func gridLines(samples [16]sample, padStyles [16]lipgloss.Style) []string {
	const innerContent = 27
	const reset = "\033[0m"

	innerTop := func(i int, s sample) string {
		c := padStyles[i]
		label := fmt.Sprintf("%d %s", i+1, s.drumKind)
		label = truncate(label, 25)
		labelLen := utf8.RuneCountInString(label)
		fill := 25 - labelLen
		return c.Render("╭─ "+label+" "+strings.Repeat("─", fill)+"╮")
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

			padded := padRight(truncate(samples[i].filename, innerContent), innerContent)
			var fcolor lipgloss.Style
			if samples[i].filename == "" {
				padded = padRight("(empty)", innerContent)
				fcolor = styleRed
			} else if padTypesMatch(i, detectSampleKind(samples[i].filename)) {
				fcolor = styleGreen
			} else {
				fcolor = styleYellow
			}
			file.WriteString(sep + padStyles[i].Render("│") + " " + fcolor.Render(padded) + padStyles[i].Render("│"))

			clean.WriteString(sep + innerSide(i, samples[i].cleanName))
			out.WriteString(sep + innerSide(i, samples[i].outputName))
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

// extractPadStyles reads the 16 pad colors from the loaded programTemplate
// and returns them as lipgloss styles with colored foregrounds.
func extractPadStyles() [16]lipgloss.Style {
	var styles [16]lipgloss.Style
	for i := range styles {
		styles[i] = lipgloss.NewStyle()
	}

	decoder := xml.NewDecoder(bytes.NewReader(programTemplate))
	var padsJSON string
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "ProgramPads-v2.10" {
			continue
		}
		if err := decoder.DecodeElement(&padsJSON, &se); err != nil {
			break
		}
		break
	}
	if padsJSON == "" {
		return styles
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(padsJSON), &data); err != nil {
		return styles
	}
	programPads, ok := data["ProgramPads-v2.10"].(map[string]interface{})
	if !ok {
		return styles
	}
	pads, ok := programPads["pads"].(map[string]interface{})
	if !ok {
		return styles
	}

	for i := range 16 {
		val, ok := pads[fmt.Sprintf("value%d", i)]
		if !ok {
			continue
		}
		colorInt := int(val.(float64))
		r := (colorInt >> 16) & 0xFF
		g := (colorInt >> 8) & 0xFF
		b := colorInt & 0xFF
		styles[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b)))
	}
	return styles
}

func renderUsage(baseDir string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s organizes drum sample WAV files into 16-pad MPC kits.\n", styleBold.Render("roger"))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Usage: %s [PackName ...]\n", styleBold.Render("roger"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "With no arguments, all packs in Input/ are processed.")
	fmt.Fprintln(&b, "Pass one or more pack names to process only those.")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Workspace: %s\n", styleCyan.Render(baseDir))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s   Place your samples here\n", styleGreen.Render("Input/"))
	fmt.Fprintf(&b, "  %s  Generated MPC kits and program files appear here\n", styleYellow.Render("Output/"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Organize your samples in Input/ like this:")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s\n", styleBold.Render("PackName/"))
	fmt.Fprintf(&b, "    %s\n", styleDim.Render("Kit 1/"))
	fmt.Fprintln(&b, "      Kick.wav, Snare.wav, ...")
	fmt.Fprintf(&b, "    %s\n", styleDim.Render("Kit 2/"))
	fmt.Fprintln(&b, "      Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Kits can also be grouped:")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s\n", styleBold.Render("PackName/"))
	fmt.Fprintf(&b, "    %s\n", styleDim.Render("Group A/"))
	fmt.Fprintf(&b, "      %s\n", styleDim.Render("Kit 1/"))
	fmt.Fprintln(&b, "        Kick.wav, Snare.wav, ...")
	fmt.Fprintf(&b, "    %s\n", styleDim.Render("Group B/"))
	fmt.Fprintf(&b, "      %s\n", styleDim.Render("Kit 1/"))
	fmt.Fprintln(&b, "        Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Samples are auto-detected by type (kick, snare, hat, etc.) from their")
	fmt.Fprintln(&b, "filenames and assigned to pads.")

	return b.String()
}

func renderLegend() string {
	var b strings.Builder
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, styleDim.Render("Each grid is a 16-pad MPC kit. The left column lists samples with their duration."))
	fmt.Fprintf(&b, "%sIn the grid: %s = type matched, %s = filled from remaining, %s = empty pad.%s\n",
		styleDim.Render(""), styleGreen.Render(" green "), styleYellow.Render(" yellow "), styleRed.Render(" red "), styleDim.Render(""))
	return b.String()
}

func renderWarnings(packs []pack, emptyPacks, wrongSampleCount []string) string {
	var b strings.Builder

	var missingImages []string
	for _, p := range packs {
		if imgPath, _ := findImage(p.dir); imgPath == "" {
			missingImages = append(missingImages, p.name)
		}
	}

	if len(emptyPacks) > 0 {
		fmt.Fprintln(&b)
		for _, p := range emptyPacks {
			fmt.Fprintf(&b, "%s %s contains no kit directories with WAV files\n",
				styleYellow.Render("warning:"), p)
		}
	}
	if len(wrongSampleCount) > 0 {
		fmt.Fprintln(&b)
		for _, s := range wrongSampleCount {
			fmt.Fprintf(&b, "%s %s WAV files, expected 16\n",
				styleYellow.Render("warning:"), s)
		}
	}
	if len(missingImages) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "%s no cover image found for: %s\n",
			styleYellow.Render("warning:"), strings.Join(missingImages, ", "))
		fmt.Fprintln(&b, "Place an image file in the top-level directory of each pack so that it will be used as the cover image for the Expansion.")
	}

	return b.String()
}

func renderGrids(packs []pack, padStyles [16]lipgloss.Style) string {
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
		b.WriteString(renderKits(p, globalLeftColWidth, padStyles))
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

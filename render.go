package main

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

func padTypesMatch(padIndex int, kind SampleKind) bool {
	return slices.Contains(cfg.PadLayout[padIndex], string(kind))
}

// padColors holds the ANSI color for each of the 16 pad positions (0-indexed).
// Populated at startup from the program template by extractPadColorsFromTemplate().
var padColors [16]string

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

func renderKits(p pack, leftColWidth int) {
	gridWidth := leftColWidth + 4*cellWidth + 2 // 4 cells + 2 outer borders
	packName := truncateMiddle(p.name, gridWidth-4)
	titleLen := utf8.RuneCountInString(packName)
	totalPad := gridWidth - titleLen - 2 // 2 for spaces around title
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	fmt.Printf("━%s \033[1m%s\033[0m %s━\n",
		strings.Repeat("━", leftPad), packName, strings.Repeat("━", rightPad))

	blank := strings.Repeat(" ", leftColWidth)
	multipleGroups := len(p.groups) > 1

	if !multipleGroups {
		fmt.Println(blank + topBorder())
	}

	for gi, group := range p.groups {
		if multipleGroups {
			if gi > 0 {
				fmt.Println(strings.Repeat("━", leftColWidth) + bottomBorder())
			}
			if group.name != "" {
				groupName := truncateMiddle(group.name, gridWidth-4)
				nameLen := utf8.RuneCountInString(groupName)
				gTotalPad := gridWidth - nameLen - 2
				gLeftPad := gTotalPad / 2
				gRightPad := gTotalPad - gLeftPad
				fmt.Printf("\033[2m━%s %s %s━\033[0m\n",
					strings.Repeat("━", gLeftPad), groupName, strings.Repeat("━", gRightPad))
			}
			fmt.Println(blank + topBorder())
		}

		for ki := len(group.kits) - 1; ki >= 0; ki-- {
			kit := group.kits[ki]
			if ki < len(group.kits)-1 {
				fmt.Println(strings.Repeat("━", leftColWidth) + heavySeparator())
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
					var color string
					if durationMs >= 2000 {
						color = "\033[31m"
					} else if durationMs >= 1000 {
						color = "\033[33m"
					}
					if color != "" {
						dur = color + dur + "\033[0m"
					}
					cleanNames[i] = left + strings.Repeat(" ", gap) + dur + " "
				} else {
					cleanNames[i] = left
				}
			}

			lines := gridLines(kit.samples)
			total := len(lines)
			for li, line := range lines {
				if li == total-1 {
					displayName := truncateMiddle(kit.name, 50)
					fmt.Print(padRight(" \033[1m"+displayName+"\033[0m", leftColWidth+8))
				} else if li < 16 {
					fmt.Print(padRight(cleanNames[li], leftColWidth))
				} else {
					fmt.Print(blank)
				}
				fmt.Println(line)
			}
		}
	}
	fmt.Println(strings.Repeat("━", leftColWidth) + bottomBorder())
}

func gridLines(samples [16]sample) []string {
	// Each cell is cellWidth=30 chars wide (between outer separators).
	// Inside each cell we draw a colored inner box:
	//   top (30):    ╭─ {pad} {kind} ─...─╮
	//   content (3): │ {text...27 chars...} │   (colored │ bars)
	//   bottom (30): ╰────────────────────────────╯
	// Outer grid separators (┃ │) remain in default terminal color.
	const innerContent = 27 // chars available inside the inner box sides
	const reset = "\033[0m"

	// innerTop builds the 30-char colored top border of the inner box,
	// embedding the pad number and drum kind as a title.
	innerTop := func(i int, s sample) string {
		c := padColors[i]
		label := fmt.Sprintf("%d %s", i+1, s.drumKind)
		label = truncate(label, 25)
		labelLen := utf8.RuneCountInString(label)
		fill := 25 - labelLen
		return c + "╭─ " + label + " " + strings.Repeat("─", fill) + "╮" + reset
	}

	// innerSide wraps content in colored inner box side bars (│).
	innerSide := func(i int, content string) string {
		c := padColors[i]
		return c + "│" + reset + " " + padRight(truncate(content, innerContent), innerContent) + c + "│" + reset
	}

	// innerBottom builds the 30-char colored bottom border of the inner box.
	innerBottom := func(i int) string {
		return padColors[i] + "╰" + strings.Repeat("─", 28) + "╯" + reset
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
			var fcolor string
			if samples[i].filename == "" {
				padded = padRight("(empty)", innerContent)
				fcolor = "\033[31m"
			} else if padTypesMatch(i, detectSampleKind(samples[i].filename)) {
				fcolor = "\033[32m"
			} else {
				fcolor = "\033[33m"
			}
			file.WriteString(sep + padColors[i] + "│" + reset + " " + fcolor + padded + reset + padColors[i] + "│" + reset)

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

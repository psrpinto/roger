package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	baseDir := filepath.Join(desktopDir(), "roger")
	srcDir := filepath.Join(baseDir, "Input")
	destDir := filepath.Join(baseDir, "Output")

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printUsage(baseDir)
		return
	}

	for _, dir := range []string{baseDir, srcDir, destDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", dir, err)
			os.Exit(1)
		}
	}

	templatePath := filepath.Join(baseDir, "template.xpm")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		os.WriteFile(templatePath, programTemplate, 0o644)
	}
	expansionPath := filepath.Join(baseDir, "expansion.xml")
	if _, err := os.Stat(expansionPath); os.IsNotExist(err) {
		os.WriteFile(expansionPath, expansionTemplate, 0o644)
	}

	loadCustomTemplate(baseDir)
	loadCustomExpansionTemplate(baseDir)
	loadOrCreateConfig(baseDir)
	padColors = extractPadColorsFromTemplate()

	var topLevelDirs []string
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			packDir := filepath.Join(srcDir, arg)
			if info, err := os.Stat(packDir); err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "error: pack directory not found: %s\n", packDir)
				os.Exit(1)
			}
			topLevelDirs = append(topLevelDirs, packDir)
		}
	} else {
		topLevelDirs = listSubdirs(srcDir)
	}
	if len(os.Args) <= 1 && len(topLevelDirs) == 0 {
		createExampleDirs(srcDir)
		printUsage(baseDir)
		fmt.Println()
		fmt.Println("Example directories have been created in Input/ to demonstrate the structure.")
		fmt.Print("Preview example kits? [Y/n] ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer != "" && answer != "y" && answer != "yes" {
				return
			}
		}
		topLevelDirs = listSubdirs(srcDir)
	}

	// Collect and preview phase
	totalKits := 0
	for _, topDir := range topLevelDirs {
		totalKits += len(findKitDirs(topDir))
	}
	kitsDone := 0

	var packs []pack
	var wrongSampleCount []string
	var emptyPacks []string
	for _, topDir := range topLevelDirs {
		packName := formatKitName(topDir)
		kitPaths := findKitDirs(topDir)

		if len(kitPaths) == 0 {
			emptyPacks = append(emptyPacks, filepath.Base(topDir))
			continue
		}

		// Determine groups by classifying each kit path by depth relative to topDir.
		// depth 1 (Kit1/): flat pack — one group with name="" and dir=topDir.
		// depth 2 (GroupA/Kit1/): grouped pack — one group per unique first component.
		// groupOrder preserves the sort order from findKitDirs while deduplicating.
		type groupKey struct {
			dir  string
			name string
		}
		var groupOrder []groupKey
		groupSeen := make(map[string]bool)
		kitGroupDir := make(map[string]string) // kitPath → group dir

		for _, kitPath := range kitPaths {
			rel, _ := filepath.Rel(topDir, kitPath)
			parts := strings.Split(rel, string(filepath.Separator))
			var gk groupKey
			if len(parts) <= 1 {
				gk = groupKey{dir: topDir, name: ""}
			} else {
				gk = groupKey{dir: filepath.Join(topDir, parts[0]), name: formatKitName(parts[0])}
			}
			kitGroupDir[kitPath] = gk.dir
			if !groupSeen[gk.dir] {
				groupSeen[gk.dir] = true
				groupOrder = append(groupOrder, gk)
			}
		}

		kitsByGroup := make(map[string][]kitData)
		for _, kitPath := range kitPaths {
			kitsDone++
			fmt.Fprintf(os.Stderr, "\r\033[KScanning samples... (%d/%d)", kitsDone, totalKits)
			dirTokens := deriveSrcTokens(topDir, kitPath)

			kitEntries, err := os.ReadDir(kitPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: reading kit directory %s: %s\n", kitPath, err)
				continue
			}

			var samples []sample
			for _, entry := range kitEntries {
				if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".wav") && !strings.HasPrefix(entry.Name(), "._") {
					name := entry.Name()
					displayName := strings.TrimSuffix(name, filepath.Ext(name))
					sourcePath := filepath.Join(kitPath, name)

					frameCount, sampleRate, err := readSampleInfo(sourcePath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "warning: failed to read sample info for %s: %s\n", sourcePath, err)
						continue
					}

					samples = append(samples, sample{
						filename:   name,
						extension:  filepath.Ext(name),
						cleanName:  cleanSampleName(displayName, dirTokens),
						drumKind:   detectSampleKind(name),
						sourcePath: sourcePath,
						frameCount: frameCount,
						sampleRate: sampleRate,
					})
				}
			}

			if len(samples) == 0 {
				continue
			}

			if len(samples) != 16 {
				rel, _ := filepath.Rel(srcDir, kitPath)
				wrongSampleCount = append(wrongSampleCount, fmt.Sprintf("%s (%d)", rel, len(samples)))
			}

			kitName := deriveKitName(topDir, kitPath)
			assigned := assignSamples(samples)
			for i := range assigned {
				assigned[i].outputName = fmt.Sprintf("%s-%02d-%s", kitName, i+1, formatKitName(assigned[i].cleanName))
			}
			gDir := kitGroupDir[kitPath]
			kitsByGroup[gDir] = append(kitsByGroup[gDir], kitData{name: kitName, kitPath: kitPath, samples: assigned})
		}

		var groups []kitGroup
		for _, gk := range groupOrder {
			kits := kitsByGroup[gk.dir]
			if len(kits) == 0 {
				continue
			}
			groups = append(groups, kitGroup{name: gk.name, dir: gk.dir, kits: kits})
		}
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].name < groups[j].name
		})

		if len(groups) == 0 {
			continue
		}

		packs = append(packs, pack{name: packName, dir: topDir, groups: groups})
	}
	if totalKits > 0 {
		fmt.Fprintf(os.Stderr, "\r\033[K")
	}

	// Compute a global leftColWidth across all packs for consistent grid width
	globalLeftColWidth := 0
	for _, p := range packs {
		if w := computeLeftColWidth(p); w > globalLeftColWidth {
			globalLeftColWidth = w
		}
	}

	for i, p := range packs {
		if i > 0 {
			fmt.Println()
		}
		renderKits(p, globalLeftColWidth)
	}

	if len(packs) == 0 {
		return
	}

	// Legend
	dim := "\033[2m"
	reset := "\033[0m"
	fmt.Println()
	fmt.Printf("%sEach grid is a 16-pad MPC kit. The left column lists samples with their duration.%s\n", dim, reset)
	fmt.Printf("%sIn the grid: \033[32m green \033[0m%s= type matched, \033[33m yellow \033[0m%s= filled from remaining, \033[31m red \033[0m%s= empty pad.%s\n", dim, dim, dim, dim, reset)

	// Show warnings
	var missingImages []string
	for _, p := range packs {
		if imgPath, _ := findImage(p.dir); imgPath == "" {
			missingImages = append(missingImages, p.name)
		}
	}
	yellow := "\033[33m"
	if len(emptyPacks) > 0 {
		fmt.Println()
		for _, p := range emptyPacks {
			fmt.Fprintf(os.Stderr, "%swarning:%s %s contains no kit directories with WAV files\n", yellow, reset, p)
		}
	}
	if len(wrongSampleCount) > 0 {
		fmt.Println()
		for _, s := range wrongSampleCount {
			fmt.Fprintf(os.Stderr, "%swarning:%s %s WAV files, expected 16\n", yellow, reset, s)
		}
	}
	if len(missingImages) > 0 {
		fmt.Println()
		fmt.Fprintf(os.Stderr, "%swarning:%s no cover image found for: %s\n", yellow, reset, strings.Join(missingImages, ", "))
		fmt.Println("Place an image file in the top-level directory of each pack so that it will be used as the cover image for the Expansion.")
	}

	// Ask for confirmation before generating output
	fmt.Print("\nGenerate output files? [Y/n] ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return
		}
	}

	// Output phase
	var totalSize int64
	totalSampleCount := 0

	kitCount := 0
	for _, p := range packs {
		for _, group := range p.groups {
			kitCount += len(group.kits)
		}
	}
	kitsGenerated := 0

	for _, p := range packs {
		packOutDir := filepath.Join(destDir, p.name)
		previewDir := filepath.Join(packOutDir, "[Previews]")
		for _, dir := range []string{packOutDir, previewDir} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "error: creating output directory: %s\n", err)
				return
			}
		}

		// Generate program XMLs and multi-kit programs
		for _, group := range p.groups {
			for _, kit := range group.kits {
				xpm, err := renderProgramXml(kit.name, [][16]sample{kit.samples})
				if err != nil {
					return
				}
				if err := os.WriteFile(filepath.Join(packOutDir, kit.name+".xpm"), xpm, 0644); err != nil {
					return
				}
				totalSize += int64(len(xpm))

				// Generate audio preview
				previewPath := filepath.Join(previewDir, kit.name+".xpm.wav")
				if err := generatePreview([][16]sample{kit.samples}, previewPath); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to generate preview for %s: %s\n", kit.name, err)
				} else if info, err := os.Stat(previewPath); err == nil {
					totalSize += info.Size()
				}

				kitsGenerated++
				fmt.Fprintf(os.Stderr, "\r\033[KGenerating programs... (%d/%d)", kitsGenerated, kitCount)
			}

			// Generate multi-kit program for this group (if it has more than one kit)
			if len(group.kits) <= 1 {
				continue
			}
			var multiBaseName string
			if group.name == "" {
				multiBaseName = "+" + p.name + "-Multi"
			} else {
				multiBaseName = "+" + dedupeTokens(p.name+"-"+group.name) + "-Multi"
			}
			// Split into chunks of 8 banks max
			for chunkIdx := 0; chunkIdx*8 < len(group.kits); chunkIdx++ {
				start := chunkIdx * 8
				end := start + 8
				if end > len(group.kits) {
					end = len(group.kits)
				}
				chunk := group.kits[start:end]
				if len(chunk) <= 1 && chunkIdx > 0 {
					continue
				}

				var banks [][16]sample
				for _, kit := range chunk {
					banks = append(banks, kit.samples)
				}

				programName := multiBaseName
				if len(group.kits) > 8 {
					programName = fmt.Sprintf("%s-%d", multiBaseName, chunkIdx+1)
				}

				xpm, err := renderProgramXml(programName, banks)
				if err != nil {
					return
				}
				if err := os.WriteFile(filepath.Join(packOutDir, programName+".xpm"), xpm, 0644); err != nil {
					return
				}
				totalSize += int64(len(xpm))

				// Generate audio preview for multi-kit program
				previewPath := filepath.Join(previewDir, programName+".xpm.wav")
				if err := generatePreview(banks, previewPath); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to generate preview for %s: %s\n", programName, err)
				} else if info, err := os.Stat(previewPath); err == nil {
					totalSize += info.Size()
				}
			}
		}

		// Copy samples
		for _, group := range p.groups {
			for _, kit := range group.kits {
				for _, s := range kit.samples {
					if s.filename == "" {
						continue
					}
					totalSampleCount++
					destPath := filepath.Join(packOutDir, s.outputName+s.extension)
					if info, err := os.Stat(destPath); err == nil {
						totalSize += info.Size()
						continue
					}
					if err := copyFile(s.sourcePath, destPath); err != nil {
						fmt.Fprintf(os.Stderr, "error: copying %s: %s\n", s.filename, err)
						return
					}
					if info, err := os.Stat(destPath); err == nil {
						totalSize += info.Size()
					}
				}
			}
		}

		// Generate expansion XML per pack
		expansion := expansionInfo{
			Identifier:   fmt.Sprintf("org.custom.%s", p.name),
			Title:        p.name,
			Manufacturer: "Custom",
			Description:  p.name,
		}

		expansionXml, err := renderExpansionXml(expansion)
		if err != nil {
			return
		}
		if err := os.WriteFile(filepath.Join(packOutDir, "Expansion.xml"), expansionXml, 0644); err != nil {
			return
		}
		totalSize += int64(len(expansionXml))

		// Convert cover image if present
		if imgPath, _ := findImage(p.dir); imgPath != "" {
			destPath := filepath.Join(packOutDir, "Expansion.jpg")
			if err := convertCoverImage(imgPath, destPath); err != nil {
				fmt.Fprintf(os.Stderr, "error: converting cover image: %s\n", err)
			} else if info, err := os.Stat(destPath); err == nil {
				totalSize += info.Size()
			}
		}
	}

	if kitCount > 0 {
		fmt.Fprintf(os.Stderr, "\r\033[K")
	}

	fmt.Printf("\n%d kits, %d samples, %s\n", kitCount, totalSampleCount, formatSize(totalSize))
}

func printUsage(baseDir string) {
	bold := "\033[1m"
	dim := "\033[2m"
	cyan := "\033[36m"
	green := "\033[32m"
	yellow := "\033[33m"
	reset := "\033[0m"

	fmt.Printf("%sroger%s organizes drum sample WAV files into 16-pad MPC kits.\n", bold, reset)
	fmt.Println()
	fmt.Printf("Usage: %sroger%s [PackName ...]\n", bold, reset)
	fmt.Println()
	fmt.Println("With no arguments, all packs in Input/ are processed.")
	fmt.Printf("Pass one or more pack names to process only those.\n")
	fmt.Println()
	fmt.Printf("Workspace: %s%s%s\n", cyan, baseDir, reset)
	fmt.Println()
	fmt.Printf("  %sInput/%s   Place your samples here\n", green, reset)
	fmt.Printf("  %sOutput/%s  Generated MPC kits and program files appear here\n", yellow, reset)
	fmt.Println()
	fmt.Println("Organize your samples in Input/ like this:")
	fmt.Println()
	fmt.Printf("  %sPackName/%s\n", bold, reset)
	fmt.Printf("    %sKit 1/%s\n", dim, reset)
	fmt.Printf("      Kick.wav, Snare.wav, ...\n")
	fmt.Printf("    %sKit 2/%s\n", dim, reset)
	fmt.Printf("      Kick.wav, Snare.wav, ...\n")
	fmt.Println()
	fmt.Println("Kits can also be grouped:")
	fmt.Println()
	fmt.Printf("  %sPackName/%s\n", bold, reset)
	fmt.Printf("    %sGroup A/%s\n", dim, reset)
	fmt.Printf("      %sKit 1/%s\n", dim, reset)
	fmt.Printf("        Kick.wav, Snare.wav, ...\n")
	fmt.Printf("    %sGroup B/%s\n", dim, reset)
	fmt.Printf("      %sKit 1/%s\n", dim, reset)
	fmt.Printf("        Kick.wav, Snare.wav, ...\n")
	fmt.Println()
	fmt.Println("Samples are auto-detected by type (kick, snare, hat, etc.) from their")
	fmt.Println("filenames and assigned to pads.")
}

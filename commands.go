package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func waitForScanProgress(ch <-chan scanProgressMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func waitForGenProgress(ch <-chan genProgressMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func scanPacksCmd(ch chan<- scanProgressMsg, topLevelDirs []string, srcDir string) tea.Cmd {
	return func() tea.Msg {
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

			type groupKey struct {
				dir  string
				name string
			}
			var groupOrder []groupKey
			groupSeen := make(map[string]bool)
			kitGroupDir := make(map[string]string)

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
				ch <- scanProgressMsg{done: kitsDone, total: totalKits}

				dirTokens := deriveSrcTokens(topDir, kitPath)

				kitEntries, err := os.ReadDir(kitPath)
				if err != nil {
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

		return scanDoneMsg{
			packs:            packs,
			emptyPacks:       emptyPacks,
			wrongSampleCount: wrongSampleCount,
		}
	}
}

func generatePacksCmd(ch chan<- genProgressMsg, packs []pack, destDir string) tea.Cmd {
	return func() tea.Msg {
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
					return errMsg{err: fmt.Errorf("creating output directory: %w", err)}
				}
			}

			for _, group := range p.groups {
				for _, kit := range group.kits {
					xpm, err := renderProgramXml(kit.name, [][16]sample{kit.samples})
					if err != nil {
						return errMsg{err: err}
					}
					if err := os.WriteFile(filepath.Join(packOutDir, kit.name+".xpm"), xpm, 0644); err != nil {
						return errMsg{err: err}
					}
					totalSize += int64(len(xpm))

					previewPath := filepath.Join(previewDir, kit.name+".xpm.wav")
					if err := generatePreview([][16]sample{kit.samples}, previewPath); err != nil {
						// non-fatal
					} else if info, err := os.Stat(previewPath); err == nil {
						totalSize += info.Size()
					}

					kitsGenerated++
					ch <- genProgressMsg{done: kitsGenerated, total: kitCount}
				}

				if len(group.kits) <= 1 {
					continue
				}
				var multiBaseName string
				if group.name == "" {
					multiBaseName = "+" + p.name + "-Multi"
				} else {
					multiBaseName = "+" + dedupeTokens(p.name+"-"+group.name) + "-Multi"
				}
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
						return errMsg{err: err}
					}
					if err := os.WriteFile(filepath.Join(packOutDir, programName+".xpm"), xpm, 0644); err != nil {
						return errMsg{err: err}
					}
					totalSize += int64(len(xpm))

					previewPath := filepath.Join(previewDir, programName+".xpm.wav")
					if err := generatePreview(banks, previewPath); err != nil {
						// non-fatal
					} else if info, err := os.Stat(previewPath); err == nil {
						totalSize += info.Size()
					}
				}
			}

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
							return errMsg{err: fmt.Errorf("copying %s: %w", s.filename, err)}
						}
						if info, err := os.Stat(destPath); err == nil {
							totalSize += info.Size()
						}
					}
				}
			}

			expansion := expansionInfo{
				Identifier:   fmt.Sprintf("org.custom.%s", p.name),
				Title:        p.name,
				Manufacturer: "Custom",
				Description:  p.name,
			}

			expansionXml, err := renderExpansionXml(expansion)
			if err != nil {
				return errMsg{err: err}
			}
			if err := os.WriteFile(filepath.Join(packOutDir, "Expansion.xml"), expansionXml, 0644); err != nil {
				return errMsg{err: err}
			}
			totalSize += int64(len(expansionXml))

			if imgPath, _ := findImage(p.dir); imgPath != "" {
				destPath := filepath.Join(packOutDir, "Expansion.jpg")
				if err := convertCoverImage(imgPath, destPath); err != nil {
					// non-fatal
				} else if info, err := os.Stat(destPath); err == nil {
					totalSize += info.Size()
				}
			}
		}

		return genDoneMsg{
			kitCount:    kitCount,
			sampleCount: totalSampleCount,
			totalSize:   totalSize,
		}
	}
}

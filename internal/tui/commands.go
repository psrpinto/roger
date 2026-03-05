package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"roger/internal/config"
	"roger/internal/kit"
	"roger/internal/mpc"
	"roger/internal/sampler"
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

func scanPacksCmd(ch chan<- scanProgressMsg, topLevelDirs []string, srcDir string, cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		totalKits := 0
		for _, topDir := range topLevelDirs {
			totalKits += len(sampler.FindKitDirs(topDir))
		}
		kitsDone := 0

		var packs []kit.Pack
		var wrongSampleCount []string
		var emptyPacks []string

		for _, topDir := range topLevelDirs {
			packName := sampler.FormatKitName(topDir)
			kitPaths := sampler.FindKitDirs(topDir)

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
					gk = groupKey{dir: filepath.Join(topDir, parts[0]), name: sampler.FormatKitName(parts[0])}
				}
				kitGroupDir[kitPath] = gk.dir
				if !groupSeen[gk.dir] {
					groupSeen[gk.dir] = true
					groupOrder = append(groupOrder, gk)
				}
			}

			kitsByGroup := make(map[string][]kit.KitData)
			for _, kitPath := range kitPaths {
				kitsDone++
				ch <- scanProgressMsg{done: kitsDone, total: totalKits}

				dirTokens := sampler.DeriveSrcTokens(topDir, kitPath)

				kitEntries, err := os.ReadDir(kitPath)
				if err != nil {
					continue
				}

				var samples []kit.Sample
				for _, entry := range kitEntries {
					if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".wav") && !strings.HasPrefix(entry.Name(), "._") {
						name := entry.Name()
						displayName := strings.TrimSuffix(name, filepath.Ext(name))
						sourcePath := filepath.Join(kitPath, name)

						frameCount, sampleRate, err := sampler.ReadSampleInfo(sourcePath)
						if err != nil {
							continue
						}

						samples = append(samples, kit.Sample{
							Filename:   name,
							Extension:  filepath.Ext(name),
							CleanName:  sampler.CleanSampleName(displayName, dirTokens),
							DrumKind:   kit.DetectSampleKind(name, cfg.DrumTypes),
							SourcePath: sourcePath,
							FrameCount: frameCount,
							SampleRate: sampleRate,
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

				kitName := sampler.DeriveKitName(topDir, kitPath)
				assigned := kit.AssignSamples(samples, cfg.PadLayout)
				for i := range assigned {
					assigned[i].OutputName = fmt.Sprintf("%s-%02d-%s", kitName, i+1, sampler.FormatKitName(assigned[i].CleanName))
				}
				gDir := kitGroupDir[kitPath]
				kitsByGroup[gDir] = append(kitsByGroup[gDir], kit.KitData{Name: kitName, KitPath: kitPath, Samples: assigned})
			}

			var groups []kit.KitGroup
			for _, gk := range groupOrder {
				kits := kitsByGroup[gk.dir]
				if len(kits) == 0 {
					continue
				}
				groups = append(groups, kit.KitGroup{Name: gk.name, Dir: gk.dir, Kits: kits})
			}
			sort.Slice(groups, func(i, j int) bool {
				return groups[i].Name < groups[j].Name
			})

			if len(groups) == 0 {
				continue
			}

			packs = append(packs, kit.Pack{Name: packName, Dir: topDir, Groups: groups})
		}

		return scanDoneMsg{
			packs:            packs,
			emptyPacks:       emptyPacks,
			wrongSampleCount: wrongSampleCount,
		}
	}
}

func generatePacksCmd(ch chan<- genProgressMsg, packs []kit.Pack, destDir string, padLayout [16][]string) tea.Cmd {
	return func() tea.Msg {
		var totalSize int64
		totalSampleCount := 0

		kitCount := 0
		for _, p := range packs {
			for _, group := range p.Groups {
				kitCount += len(group.Kits)
			}
		}
		kitsGenerated := 0

		for _, p := range packs {
			packOutDir := filepath.Join(destDir, p.Name)
			previewDir := filepath.Join(packOutDir, "[Previews]")
			for _, dir := range []string{packOutDir, previewDir} {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return errMsg{err: fmt.Errorf("creating output directory: %w", err)}
				}
			}

			for _, group := range p.Groups {
				for _, k := range group.Kits {
					xpm, err := mpc.RenderProgramXml(k.Name, [][16]kit.Sample{k.Samples})
					if err != nil {
						return errMsg{err: err}
					}
					if err := os.WriteFile(filepath.Join(packOutDir, k.Name+".xpm"), xpm, 0644); err != nil {
						return errMsg{err: err}
					}
					totalSize += int64(len(xpm))

					previewPath := filepath.Join(previewDir, k.Name+".xpm.wav")
					if err := mpc.GeneratePreview([][16]kit.Sample{k.Samples}, padLayout, previewPath); err != nil {
						// non-fatal
					} else if info, err := os.Stat(previewPath); err == nil {
						totalSize += info.Size()
					}

					kitsGenerated++
					ch <- genProgressMsg{done: kitsGenerated, total: kitCount}
				}

				if len(group.Kits) <= 1 {
					continue
				}
				var multiBaseName string
				if group.Name == "" {
					multiBaseName = "+" + p.Name + "-Multi"
				} else {
					multiBaseName = "+" + sampler.DedupeTokens(p.Name+"-"+group.Name) + "-Multi"
				}
				for chunkIdx := 0; chunkIdx*8 < len(group.Kits); chunkIdx++ {
					start := chunkIdx * 8
					end := start + 8
					if end > len(group.Kits) {
						end = len(group.Kits)
					}
					chunk := group.Kits[start:end]
					if len(chunk) <= 1 && chunkIdx > 0 {
						continue
					}

					var banks [][16]kit.Sample
					for _, k := range chunk {
						banks = append(banks, k.Samples)
					}

					programName := multiBaseName
					if len(group.Kits) > 8 {
						programName = fmt.Sprintf("%s-%d", multiBaseName, chunkIdx+1)
					}

					xpm, err := mpc.RenderProgramXml(programName, banks)
					if err != nil {
						return errMsg{err: err}
					}
					if err := os.WriteFile(filepath.Join(packOutDir, programName+".xpm"), xpm, 0644); err != nil {
						return errMsg{err: err}
					}
					totalSize += int64(len(xpm))

					previewPath := filepath.Join(previewDir, programName+".xpm.wav")
					if err := mpc.GeneratePreview(banks, padLayout, previewPath); err != nil {
						// non-fatal
					} else if info, err := os.Stat(previewPath); err == nil {
						totalSize += info.Size()
					}
				}
			}

			for _, group := range p.Groups {
				for _, k := range group.Kits {
					for _, s := range k.Samples {
						if s.Filename == "" {
							continue
						}
						totalSampleCount++
						destPath := filepath.Join(packOutDir, s.OutputName+s.Extension)
						if info, err := os.Stat(destPath); err == nil {
							totalSize += info.Size()
							continue
						}
						if err := sampler.CopyFile(s.SourcePath, destPath); err != nil {
							return errMsg{err: fmt.Errorf("copying %s: %w", s.Filename, err)}
						}
						if info, err := os.Stat(destPath); err == nil {
							totalSize += info.Size()
						}
					}
				}
			}

			expansion := mpc.ExpansionInfo{
				Identifier:   fmt.Sprintf("org.custom.%s", p.Name),
				Title:        p.Name,
				Manufacturer: "Custom",
				Description:  p.Name,
			}

			expansionXml, err := mpc.RenderExpansionXml(expansion)
			if err != nil {
				return errMsg{err: err}
			}
			if err := os.WriteFile(filepath.Join(packOutDir, "Expansion.xml"), expansionXml, 0644); err != nil {
				return errMsg{err: err}
			}
			totalSize += int64(len(expansionXml))

			if imgPath, _ := mpc.FindImage(p.Dir); imgPath != "" {
				destPath := filepath.Join(packOutDir, "Expansion.jpg")
				if err := mpc.ConvertCoverImage(imgPath, destPath); err != nil {
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

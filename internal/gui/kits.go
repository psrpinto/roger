package gui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"roger/internal/config"
	"roger/internal/examples"
	"roger/internal/kit"
	"roger/internal/mpc"
	"roger/internal/sampler"
)

type kitsTab struct {
	container  *fyne.Container
	app        *App
	cfg        *config.Config
	kitsSrcDir string
	destDir    string
	baseDir    string
	padColors  [16]color.NRGBA
}

func newKitsTab(app *App) *container.TabItem {
	kt := &kitsTab{
		container:  container.NewStack(),
		app:        app,
		cfg:        app.cfg,
		kitsSrcDir: app.kitsSrcDir,
		destDir:    app.destDir,
		baseDir:    app.baseDir,
		padColors:  app.padColors,
	}
	kt.showSelection()
	return container.NewTabItem("Kits", kt.container)
}

func (kt *kitsTab) showSelection() {
	if err := os.MkdirAll(kt.kitsSrcDir, 0o755); err != nil {
		kt.container.Objects = []fyne.CanvasObject{widget.NewLabel(fmt.Sprintf("Error: %s", err))}
		kt.container.Refresh()
		return
	}

	topLevelDirs := sampler.ListSubdirs(kt.kitsSrcDir)

	if len(topLevelDirs) == 0 {
		kt.showEmptyState()
		return
	}

	// Build CheckGroup with pack names
	labels := make([]string, len(topLevelDirs))
	dirMap := make(map[string]string, len(topLevelDirs))
	for i, d := range topLevelDirs {
		label := sampler.FormatKitName(filepath.Base(d))
		labels[i] = label
		dirMap[label] = d
	}

	checkGroup := widget.NewCheckGroup(labels, nil)
	checkGroup.SetSelected(labels) // all checked by default

	previewBtn := widget.NewButton("Preview", func() {
		selected := checkGroup.Selected
		if len(selected) == 0 {
			dialog.NewInformation("No packs selected", "Select at least one pack to preview.", kt.app.window).Show()
			return
		}
		var selectedDirs []string
		for _, label := range selected {
			if d, ok := dirMap[label]; ok {
				selectedDirs = append(selectedDirs, d)
			}
		}
		kt.startScan(selectedDirs)
	})

	content := container.NewBorder(
		widget.NewLabel("Select packs to process:"),
		previewBtn,
		nil, nil,
		container.NewVScroll(checkGroup),
	)

	kt.container.Objects = []fyne.CanvasObject{content}
	kt.container.Refresh()
}

func (kt *kitsTab) showEmptyState() {
	genBtn := widget.NewButton("Generate example files", func() {
		examples.CreateExampleDirs(kt.kitsSrcDir)
		kt.showSelection()
	})
	instrBtn := widget.NewButton("Show instructions", func() {
		showHelp(kt.app.window, kt.baseDir)
	})

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Workspace: %s", kt.baseDir)),
		widget.NewLabel(""),
		widget.NewLabel("No kits found. What would you like to do?"),
		widget.NewLabel(""),
		genBtn,
		instrBtn,
	)

	kt.container.Objects = []fyne.CanvasObject{content}
	kt.container.Refresh()
}

func (kt *kitsTab) showPreview(packs []kit.Pack, emptyPacks, wrongSampleCount []string) {
	preview := renderPacksPreview(packs, kt.padColors, kt.cfg.DrumTypes, kt.cfg.PadLayout, emptyPacks, wrongSampleCount)

	generateBtn := widget.NewButton("Generate", func() {
		kt.startGenerate(packs)
	})
	backBtn := widget.NewButton("Back", func() {
		kt.showSelection()
	})

	buttons := container.NewHBox(generateBtn, backBtn)
	content := container.NewBorder(nil, buttons, nil, nil, container.NewVScroll(preview))

	kt.container.Objects = []fyne.CanvasObject{content}
	kt.container.Refresh()
}

func (kt *kitsTab) startScan(selectedDirs []string) {
	progress := widget.NewProgressBar()
	progressLabel := widget.NewLabel("Scanning samples...")
	progressContent := container.NewVBox(progressLabel, progress)
	dlg := dialog.NewCustomWithoutButtons("Scanning", progressContent, kt.app.window)
	dlg.Show()

	go func() {
		packs, emptyPacks, wrongSampleCount := scanPacks(selectedDirs, kt.kitsSrcDir, kt.cfg, func(done, total int) {
			if total > 0 {
				progress.SetValue(float64(done) / float64(total))
				progressLabel.SetText(fmt.Sprintf("Scanning samples... (%d/%d)", done, total))
			}
		})
		dlg.Hide()
		if len(packs) == 0 {
			dialog.NewInformation("No kits found", "No valid kits were found in the selected packs.", kt.app.window).Show()
			return
		}
		kt.showPreview(packs, emptyPacks, wrongSampleCount)
	}()
}

func (kt *kitsTab) startGenerate(packs []kit.Pack) {
	progress := widget.NewProgressBar()
	progressLabel := widget.NewLabel("Generating programs...")
	progressContent := container.NewVBox(progressLabel, progress)
	dlg := dialog.NewCustomWithoutButtons("Generating", progressContent, kt.app.window)
	dlg.Show()

	go func() {
		kitCount, sampleCount, totalSize, err := generatePacks(packs, kt.destDir, kt.cfg.PadLayout, func(done, total int) {
			if total > 0 {
				progress.SetValue(float64(done) / float64(total))
				progressLabel.SetText(fmt.Sprintf("Generating programs... (%d/%d)", done, total))
			}
		})
		dlg.Hide()
		if err != nil {
			dialog.NewError(err, kt.app.window).Show()
			return
		}
		msg := fmt.Sprintf("%d kits, %d samples, %s", kitCount, sampleCount, sampler.FormatSize(totalSize))
		dialog.NewInformation("Generation complete", msg, kt.app.window).Show()
		kt.showSelection()
	}()
}

// scanPacks scans the selected directories for kit packs.
// Extracted from tui/kits/scan.go:scanPacksCmd.
func scanPacks(topLevelDirs []string, srcDir string, cfg *config.Config, onProgress func(done, total int)) ([]kit.Pack, []string, []string) {
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
			if onProgress != nil {
				onProgress(kitsDone, totalKits)
			}

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

	return packs, emptyPacks, wrongSampleCount
}

// generatePacks generates MPC program files for the given packs.
// Extracted from tui/kits/generate.go:generatePacksCmd.
func generatePacks(packs []kit.Pack, destDir string, padLayout [16][]string, onProgress func(done, total int)) (kitCount int, sampleCount int, totalSize int64, err error) {
	kitCount = 0
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
			if err = os.MkdirAll(dir, 0o755); err != nil {
				err = fmt.Errorf("creating output directory: %w", err)
				return
			}
		}

		for _, group := range p.Groups {
			for _, k := range group.Kits {
				var xpm []byte
				xpm, err = mpc.RenderProgramXml(k.Name, [][16]kit.Sample{k.Samples})
				if err != nil {
					return
				}
				if err = os.WriteFile(filepath.Join(packOutDir, k.Name+".xpm"), xpm, 0644); err != nil {
					return
				}
				totalSize += int64(len(xpm))

				previewPath := filepath.Join(previewDir, k.Name+".xpm.wav")
				if previewErr := mpc.GeneratePreview([][16]kit.Sample{k.Samples}, padLayout, previewPath); previewErr == nil {
					if info, statErr := os.Stat(previewPath); statErr == nil {
						totalSize += info.Size()
					}
				}

				kitsGenerated++
				if onProgress != nil {
					onProgress(kitsGenerated, kitCount)
				}
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

				var xpm []byte
				xpm, err = mpc.RenderProgramXml(programName, banks)
				if err != nil {
					return
				}
				if err = os.WriteFile(filepath.Join(packOutDir, programName+".xpm"), xpm, 0644); err != nil {
					return
				}
				totalSize += int64(len(xpm))

				previewPath := filepath.Join(previewDir, programName+".xpm.wav")
				if previewErr := mpc.GeneratePreview(banks, padLayout, previewPath); previewErr == nil {
					if info, statErr := os.Stat(previewPath); statErr == nil {
						totalSize += info.Size()
					}
				}
			}
		}

		for _, group := range p.Groups {
			for _, k := range group.Kits {
				for _, s := range k.Samples {
					if s.Filename == "" {
						continue
					}
					sampleCount++
					destPath := filepath.Join(packOutDir, s.OutputName+s.Extension)
					if info, statErr := os.Stat(destPath); statErr == nil {
						totalSize += info.Size()
						continue
					}
					if err = sampler.CopyFile(s.SourcePath, destPath); err != nil {
						err = fmt.Errorf("copying %s: %w", s.Filename, err)
						return
					}
					if info, statErr := os.Stat(destPath); statErr == nil {
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

		var expansionXml []byte
		expansionXml, err = mpc.RenderExpansionXml(expansion)
		if err != nil {
			return
		}
		if err = os.WriteFile(filepath.Join(packOutDir, "Expansion.xml"), expansionXml, 0644); err != nil {
			return
		}
		totalSize += int64(len(expansionXml))

		if imgPath, _ := mpc.FindImage(p.Dir); imgPath != "" {
			imgDest := filepath.Join(packOutDir, "Expansion.jpg")
			if convertErr := mpc.ConvertCoverImage(imgPath, imgDest); convertErr == nil {
				if info, statErr := os.Stat(imgDest); statErr == nil {
					totalSize += info.Size()
				}
			}
		}
	}

	return
}

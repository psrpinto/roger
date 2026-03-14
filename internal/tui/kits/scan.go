package kits

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/config"
	"roger/internal/kit"
	"roger/internal/sampler"
	"roger/internal/tui/shared"
)

type ScanDoneMsg struct {
	Packs            []kit.Pack
	EmptyPacks       []string
	WrongSampleCount []string
}

type scanProgressMsg struct {
	done, total int
}

type ScanModel struct {
	topLevelDirs []string
	srcDir       string
	cfg          *config.Config
	scanCh       chan scanProgressMsg
	progress     int
	total        int
	spinner      spinner.Model
}

func NewScanModel(topLevelDirs []string, srcDir string, cfg *config.Config) *ScanModel {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("6"))),
	)
	return &ScanModel{
		topLevelDirs: topLevelDirs,
		srcDir:       srcDir,
		cfg:          cfg,
		scanCh:       make(chan scanProgressMsg, 1),
		spinner:      s,
	}
}

func (m *ScanModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		scanPacksCmd(m.scanCh, m.topLevelDirs, m.srcDir, m.cfg),
		waitForScanProgress(m.scanCh),
	)
}

func (m *ScanModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return nil, shared.Transition{Phase: shared.Abort}
		}
	case scanProgressMsg:
		m.progress = msg.done
		m.total = msg.total
		return waitForScanProgress(m.scanCh), shared.Transition{}
	case ScanDoneMsg:
		return nil, shared.Transition{Phase: shared.Next, Data: msg}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd, shared.Transition{}
	}
	return nil, shared.Transition{}
}

func (m *ScanModel) View() string {
	return m.spinner.View() + scanningStatus(m.progress, m.total)
}

func waitForScanProgress(ch <-chan scanProgressMsg) tea.Cmd {
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

		return ScanDoneMsg{
			Packs:            packs,
			EmptyPacks:       emptyPacks,
			WrongSampleCount: wrongSampleCount,
		}
	}
}

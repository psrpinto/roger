package kits

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/kit"
	"roger/internal/mpc"
	"roger/internal/sampler"
	"roger/internal/tui/shared"
)

type GenDoneMsg struct {
	KitCount    int
	SampleCount int
	TotalSize   int64
}

type genProgressMsg struct {
	done, total int
}

type GenModel struct {
	packs     []kit.Pack
	destDir   string
	padLayout [16][]string
	genCh     chan genProgressMsg
	progress  int
	total     int
	spinner   spinner.Model
}

func NewGenModel(packs []kit.Pack, destDir string, padLayout [16][]string) *GenModel {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("6"))),
	)
	return &GenModel{
		packs:     packs,
		destDir:   destDir,
		padLayout: padLayout,
		genCh:     make(chan genProgressMsg, 1),
		spinner:   s,
	}
}

func (m *GenModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		generatePacksCmd(m.genCh, m.packs, m.destDir, m.padLayout),
		waitForGenProgress(m.genCh),
	)
}

func (m *GenModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	switch msg := msg.(type) {
	case genProgressMsg:
		m.progress = msg.done
		m.total = msg.total
		return waitForGenProgress(m.genCh), shared.Transition{}
	case GenDoneMsg:
		return nil, shared.Transition{Phase: shared.Next, Data: msg}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd, shared.Transition{}
	}
	return nil, shared.Transition{}
}

func (m *GenModel) View() string {
	return m.spinner.View() + generatingStatus(m.progress, m.total)
}

func waitForGenProgress(ch <-chan genProgressMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
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
					return shared.ErrMsg{Err: fmt.Errorf("creating output directory: %w", err)}
				}
			}

			for _, group := range p.Groups {
				for _, k := range group.Kits {
					xpm, err := mpc.RenderProgramXml(k.Name, [][16]kit.Sample{k.Samples})
					if err != nil {
						return shared.ErrMsg{Err: err}
					}
					if err := os.WriteFile(filepath.Join(packOutDir, k.Name+".xpm"), xpm, 0644); err != nil {
						return shared.ErrMsg{Err: err}
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
						return shared.ErrMsg{Err: err}
					}
					if err := os.WriteFile(filepath.Join(packOutDir, programName+".xpm"), xpm, 0644); err != nil {
						return shared.ErrMsg{Err: err}
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
							return shared.ErrMsg{Err: fmt.Errorf("copying %s: %w", s.Filename, err)}
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
				return shared.ErrMsg{Err: err}
			}
			if err := os.WriteFile(filepath.Join(packOutDir, "Expansion.xml"), expansionXml, 0644); err != nil {
				return shared.ErrMsg{Err: err}
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

		return GenDoneMsg{
			KitCount:    kitCount,
			SampleCount: totalSampleCount,
			TotalSize:   totalSize,
		}
	}
}

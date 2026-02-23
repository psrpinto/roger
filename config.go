package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type flowStrings []string

func (f flowStrings) MarshalYAML() (interface{}, error) {
	node := &yaml.Node{
		Kind:  yaml.SequenceNode,
		Style: yaml.FlowStyle,
	}
	for _, s := range f {
		node.Content = append(node.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: s,
		})
	}
	return node, nil
}

type drumType struct {
	Name   string      `yaml:"name"`
	Tokens flowStrings `yaml:"tokens"`
}

type config struct {
	DrumTypes []drumType      `yaml:"drum_types"`
	PadLayout [16]flowStrings `yaml:"pad_layout"`
}

var cfg config

func defaultConfig() config {
	return config{
		DrumTypes: []drumType{
			{Name: "Kick", Tokens: []string{"kick", "kik", "bass drum", "bd"}},
			{Name: "Snare", Tokens: []string{"snare", "snr", "sd"}},
			{Name: "ClosedHiHat", Tokens: []string{"closed hat", "closed hi-hat", "chh", "ch"}},
			{Name: "OpenHiHat", Tokens: []string{"open hat", "open hi-hat", "ohh", "oh"}},
			{Name: "Clap", Tokens: []string{"handclap", "clap", "cp"}},
			{Name: "Rim", Tokens: []string{"rimshot", "rim shot", "rim", "rs"}},
			{Name: "Tom", Tokens: []string{"tom", "tm"}},
			{Name: "Cowbell", Tokens: []string{"cowbell", "cow bell", "cb"}},
			{Name: "Cymbal", Tokens: []string{"cymbal", "crash", "ride", "cy", "china", "splash"}},
			{Name: "Shaker", Tokens: []string{"shaker", "sh"}},
			{Name: "Tambourine", Tokens: []string{"tambourine", "tamb"}},
			{Name: "Percussion", Tokens: []string{"percussion", "perc"}},
			{Name: "Clave", Tokens: []string{"clave"}},
			{Name: "Bongo", Tokens: []string{"bongo"}},
			{Name: "Conga", Tokens: []string{"conga"}},
			{Name: "Cabasa", Tokens: []string{"cabasa"}},
		},
		PadLayout: [16]flowStrings{
			{"Kick"}, {"Snare", "Clap"}, {"ClosedHiHat"}, {"OpenHiHat"},
			{"Kick", "Tom"}, {"Clap"}, {"Rim", "Clave"}, {"Shaker", "Cabasa"},
			{"Cowbell"}, {"Tom"}, {"Tom"}, {"Tom"},
			{"Bongo", "Conga"}, {"Percussion"}, {"Tambourine"}, {"Cymbal"},
		},
	}
}

func loadOrCreateConfig(baseDir string) {
	configPath := filepath.Join(baseDir, "config.yaml")

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		cfg = defaultConfig()
		writeDefaultConfig(configPath)
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading config file: %s\n", err)
		os.Exit(1)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: parsing config file: %s\n", err)
		os.Exit(1)
	}

	validateConfig()
}

func writeDefaultConfig(path string) {
	data, err := yaml.Marshal(defaultConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: marshaling default config: %s\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: writing config file: %s\n", err)
		os.Exit(1)
	}
}

func validateConfig() {
	if len(cfg.DrumTypes) == 0 {
		fmt.Fprintf(os.Stderr, "error: config must define at least one drum type\n")
		os.Exit(1)
	}

	names := make(map[SampleKind]bool)
	for _, dt := range cfg.DrumTypes {
		names[SampleKind(dt.Name)] = true
	}

	for i, pad := range cfg.PadLayout {
		if len(pad) == 0 {
			fmt.Fprintf(os.Stderr, "error: pad_layout[%d] is empty\n", i)
			os.Exit(1)
		}
		for _, t := range pad {
			if t == "" {
				fmt.Fprintf(os.Stderr, "error: pad_layout[%d] contains an empty type\n", i)
				os.Exit(1)
			}
			if !names[SampleKind(t)] {
				fmt.Fprintf(os.Stderr, "error: pad_layout[%d] references unknown drum type %q\n", i, t)
				os.Exit(1)
			}
		}
	}
}

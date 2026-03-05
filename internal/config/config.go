package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"roger/internal/kit"
)

// flowStrings is a []string that marshals as a YAML flow sequence.
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

// yamlDrumType is the YAML representation of a drum type with flow-style tokens.
type yamlDrumType struct {
	Name   string      `yaml:"name"`
	Tokens flowStrings `yaml:"tokens"`
}

// yamlConfig is the YAML representation of the config file.
type yamlConfig struct {
	DrumTypes []yamlDrumType  `yaml:"drum_types"`
	PadLayout [16]flowStrings `yaml:"pad_layout"`
}

// Config holds the application configuration.
type Config struct {
	DrumTypes []kit.DrumType
	PadLayout [16][]string
}

func fromYAML(yc yamlConfig) *Config {
	c := &Config{}
	for _, ydt := range yc.DrumTypes {
		c.DrumTypes = append(c.DrumTypes, kit.DrumType{
			Name:   ydt.Name,
			Tokens: ydt.Tokens,
		})
	}
	for i, pe := range yc.PadLayout {
		c.PadLayout[i] = pe
	}
	return c
}

func toYAML(c *Config) yamlConfig {
	yc := yamlConfig{}
	for _, dt := range c.DrumTypes {
		yc.DrumTypes = append(yc.DrumTypes, yamlDrumType{
			Name:   dt.Name,
			Tokens: dt.Tokens,
		})
	}
	for i, pe := range c.PadLayout {
		yc.PadLayout[i] = pe
	}
	return yc
}

func defaultConfig() *Config {
	return &Config{
		DrumTypes: []kit.DrumType{
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
		PadLayout: [16][]string{
			{"Kick"}, {"Snare", "Clap"}, {"ClosedHiHat"}, {"OpenHiHat"},
			{"Kick", "Tom"}, {"Clap"}, {"Rim", "Clave"}, {"Shaker", "Cabasa"},
			{"Cowbell"}, {"Tom"}, {"Tom"}, {"Tom"},
			{"Bongo", "Conga"}, {"Percussion"}, {"Tambourine"}, {"Cymbal"},
		},
	}
}

func validateConfig(c *Config) error {
	if len(c.DrumTypes) == 0 {
		return fmt.Errorf("config must define at least one drum type")
	}

	names := make(map[string]bool)
	for _, dt := range c.DrumTypes {
		names[dt.Name] = true
	}

	for i, pad := range c.PadLayout {
		if len(pad) == 0 {
			return fmt.Errorf("pad_layout[%d] is empty", i)
		}
		for _, t := range pad {
			if t == "" {
				return fmt.Errorf("pad_layout[%d] contains an empty type", i)
			}
			if !names[t] {
				return fmt.Errorf("pad_layout[%d] references unknown drum type %q", i, t)
			}
		}
	}
	return nil
}

func writeDefaultConfig(path string, c *Config) {
	data, err := yaml.Marshal(toYAML(c))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: marshaling default config: %s\n", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: writing config file: %s\n", err)
	}
}

// LoadOrCreate loads config from baseDir/config.yaml, creating it with defaults if absent.
func LoadOrCreate(baseDir string) (*Config, error) {
	configPath := filepath.Join(baseDir, "config.yaml")

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		cfg := defaultConfig()
		writeDefaultConfig(configPath, cfg)
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	cfg := fromYAML(yc)
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

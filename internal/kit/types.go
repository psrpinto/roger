package kit

// DrumType defines a drum instrument type with detection tokens.
type DrumType struct {
	Name   string
	Tokens []string
}

// SampleKind is the drum instrument type of a sample.
type SampleKind string

// Pitch represents the pitch level of a sample.
type Pitch int

const (
	PitchLow Pitch = iota
	PitchMid
	PitchHigh
)

// Sample represents a WAV sample file.
type Sample struct {
	Filename   string
	Extension  string
	DrumKind   SampleKind
	CleanName  string
	OutputName string
	SourcePath string
	FrameCount int
	SampleRate int
}

// KitData holds the data for a 16-pad kit.
type KitData struct {
	Name    string
	KitPath string
	Samples [16]Sample
}

// KitGroup is a named grouping of kits within a pack.
type KitGroup struct {
	Name string
	Dir  string
	Kits []KitData
}

// Pack is a top-level collection of kits.
type Pack struct {
	Name   string
	Dir    string
	Groups []KitGroup
}

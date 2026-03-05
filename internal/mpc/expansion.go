package mpc

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/expansion.xml
var ExpansionTemplate []byte

// ExpansionInfo holds the metadata for an MPC expansion.
type ExpansionInfo struct {
	Name         string
	Identifier   string
	Title        string
	Manufacturer string
	Description  string
}

func LoadCustomExpansionTemplate(baseDir string) {
	path := filepath.Join(baseDir, "expansion.xml")
	if data, err := os.ReadFile(path); err == nil {
		ExpansionTemplate = data
	}
}

func RenderExpansionXml(info ExpansionInfo) (result []byte, err error) {
	tmpl, err := template.New("t").Parse(string(ExpansionTemplate))
	if err != nil {
		return
	}
	rendered := new(bytes.Buffer)
	if err = tmpl.Execute(rendered, info); err != nil {
		return
	}
	result = rendered.Bytes()
	return
}

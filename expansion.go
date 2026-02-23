package main

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/expansion.xml
var expansionTemplate []byte

type expansionInfo struct {
	Name         string
	Identifier   string
	Title        string
	Manufacturer string
	Description  string
}

func loadCustomExpansionTemplate(baseDir string) {
	path := filepath.Join(baseDir, "expansion.xml")
	if data, err := os.ReadFile(path); err == nil {
		expansionTemplate = data
	}
}

func renderExpansionXml(info expansionInfo) (result []byte, err error) {
	tmpl, err := template.New("t").Parse(string(expansionTemplate))
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

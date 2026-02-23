package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/program.xpm
var programTemplate []byte

func loadCustomTemplate(baseDir string) {
	path := filepath.Join(baseDir, "template.xpm")
	if data, err := os.ReadFile(path); err == nil {
		programTemplate = data
	}
}

func renderProgramXml(programName string, banks [][16]sample) (result []byte, err error) {
	var buf bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewBuffer(programTemplate))
	encoder := xml.NewEncoder(&buf)

	totalInstruments := len(banks) * 16
	instrumentIndex := -1
	isFirstLayer := false
	numBanks := len(banks)

	for {
		var token xml.Token
		token, err = decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		switch v := token.(type) {
		case xml.StartElement:
			if v.Name.Local == "ProgramName" {
				var name string
				if err = decoder.DecodeElement(&name, &v); err != nil {
					return
				}
				name = programName
				if err = encoder.EncodeElement(name, v); err != nil {
					return
				}
				continue
			}

			// Handle ProgramPads-v2.10 to replicate pad colors across all banks
			if v.Name.Local == "ProgramPads-v2.10" {
				var padsJSON string
				if err = decoder.DecodeElement(&padsJSON, &v); err != nil {
					return
				}

				// Only modify if we have multiple banks (multikit)
				if numBanks > 1 {
					padsJSON, err = replicatePadColors(padsJSON, numBanks)
					if err != nil {
						return
					}
				}

				if err = encoder.EncodeElement(padsJSON, v); err != nil {
					return
				}
				continue
			}

			// If we're entering the first layer of an instrument...
			if v.Name.Local == "Layer" && v.Attr[0].Value == "1" {
				instrumentIndex++
				if instrumentIndex < totalInstruments {
					isFirstLayer = true
				}
			}

			if isFirstLayer {
				bank := instrumentIndex / 16
				pad := instrumentIndex % 16

				if v.Name.Local == "SampleName" {
					// Set the sample name.
					var sampleName string
					if err = decoder.DecodeElement(&sampleName, &v); err != nil {
						return
					}

					smpl := banks[bank][pad]
					sampleName = smpl.outputName
					if err = encoder.EncodeElement(sampleName, v); err != nil {
						return
					}

					continue
				}

				if v.Name.Local == "SliceEnd" {
					// Set the SliceEnd to the length in audio samples of the sample.
					var sliceEnd string
					if err = decoder.DecodeElement(&sliceEnd, &v); err != nil {
						return
					}

					smpl := banks[bank][pad]
					sampleLength := smpl.frameCount
					sliceEnd = fmt.Sprintf("%d", sampleLength)
					if err = encoder.EncodeElement(sliceEnd, v); err != nil {
						return
					}

					continue
				}
			}

		case xml.EndElement:
			if v.Name.Local == "Layer" {
				isFirstLayer = false
			}
			break
		}

		if err = encoder.EncodeToken(xml.CopyToken(token)); err != nil {
			return
		}
	}

	if err = encoder.Flush(); err != nil {
		return
	}

	// Make sure the format is exactly the same as a program exported with the MPC.
	result = buf.Bytes()
	result = bytes.ReplaceAll(result, []byte("&#34;"), []byte("&quot;"))
	result = bytes.ReplaceAll(result, []byte("&#xA;"), []byte("\n"))
	result = bytes.ReplaceAll(result, []byte("&#x9;"), []byte("\t"))
	result = bytes.ReplaceAll(result, []byte("></DrumPadEffect>"), []byte("/>"))
	result = bytes.ReplaceAll(result, []byte("></QLinkAssignments>"), []byte("/>"))

	return
}

// extractPadColorsFromTemplate reads the 16 pad colors from the loaded programTemplate
// and returns them as ANSI 24-bit foreground color escape sequences.
func extractPadColorsFromTemplate() [16]string {
	var colors [16]string

	// Scan the XML for ProgramPads-v2.10 and grab its text content.
	decoder := xml.NewDecoder(bytes.NewReader(programTemplate))
	var padsJSON string
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "ProgramPads-v2.10" {
			continue
		}
		if err := decoder.DecodeElement(&padsJSON, &se); err != nil {
			break
		}
		break
	}
	if padsJSON == "" {
		return colors
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(padsJSON), &data); err != nil {
		return colors
	}
	programPads, ok := data["ProgramPads-v2.10"].(map[string]interface{})
	if !ok {
		return colors
	}
	pads, ok := programPads["pads"].(map[string]interface{})
	if !ok {
		return colors
	}

	for i := range 16 {
		val, ok := pads[fmt.Sprintf("value%d", i)]
		if !ok {
			continue
		}
		colorInt := int(val.(float64))
		r := (colorInt >> 16) & 0xFF
		g := (colorInt >> 8) & 0xFF
		b := colorInt & 0xFF
		colors[i] = fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
	}
	return colors
}

// replicatePadColors takes the ProgramPads JSON and replicates the first 16 pad colors
// across all banks (up to 8 banks = 128 pads total)
func replicatePadColors(padsJSON string, numBanks int) (string, error) {
	// Parse the JSON structure
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(padsJSON), &data); err != nil {
		return padsJSON, err
	}

	// Navigate to the pads object
	programPads, ok := data["ProgramPads-v2.10"].(map[string]interface{})
	if !ok {
		return padsJSON, nil
	}

	pads, ok := programPads["pads"].(map[string]interface{})
	if !ok {
		return padsJSON, nil
	}

	// Extract the first 16 pad colors
	firstBankColors := make([]int, 16)
	for i := 0; i < 16; i++ {
		key := fmt.Sprintf("value%d", i)
		if color, exists := pads[key]; exists {
			firstBankColors[i] = int(color.(float64))
		}
	}

	// Replicate colors only for banks that have samples
	padColors := make([]int, 128)
	for i := 0; i < numBanks*16 && i < 128; i++ {
		padColors[i] = firstBankColors[i%16]
	}

	// Rebuild the JSON manually to preserve numeric key ordering
	var buf strings.Builder
	buf.WriteString("{\n")
	buf.WriteString("    \"ProgramPads-v2.10\": {\n")

	// Write Universal
	if universal, ok := programPads["Universal"].(map[string]interface{}); ok {
		buf.WriteString("        \"Universal\": {\n")
		buf.WriteString(fmt.Sprintf("            \"value0\": %v\n", universal["value0"]))
		buf.WriteString("        },\n")
	}

	// Write Type
	if typ, ok := programPads["Type"].(map[string]interface{}); ok {
		buf.WriteString("        \"Type\": {\n")
		buf.WriteString(fmt.Sprintf("            \"value0\": %v\n", int(typ["value0"].(float64))))
		buf.WriteString("        },\n")
	}

	// Write universalPad
	if uPad, ok := programPads["universalPad"].(float64); ok {
		buf.WriteString(fmt.Sprintf("        \"universalPad\": %d,\n", int(uPad)))
	}

	// Write pads with numeric ordering
	buf.WriteString("        \"pads\": {\n")
	for i := 0; i < 128; i++ {
		comma := ","
		if i == 127 {
			comma = ""
		}
		buf.WriteString(fmt.Sprintf("            \"value%d\": %d%s\n", i, padColors[i], comma))
	}
	buf.WriteString("        },\n")

	// Write UnusedPads
	if unused, ok := programPads["UnusedPads"].(map[string]interface{}); ok {
		buf.WriteString("        \"UnusedPads\": {\n")
		buf.WriteString(fmt.Sprintf("            \"value0\": %v\n", int(unused["value0"].(float64))))
		buf.WriteString("        }\n")
	}

	buf.WriteString("    }\n")
	buf.WriteString("}")

	return buf.String(), nil
}

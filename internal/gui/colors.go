package gui

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image/color"

	"roger/internal/mpc"
)

// ExtractPadColors reads the 16 pad colors from the MPC program template
// and returns them as color.NRGBA values.
func ExtractPadColors() [16]color.NRGBA {
	var colors [16]color.NRGBA
	for i := range colors {
		colors[i] = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	}

	decoder := xml.NewDecoder(bytes.NewReader(mpc.ProgramTemplate))
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
		colors[i] = color.NRGBA{
			R: uint8((colorInt >> 16) & 0xFF),
			G: uint8((colorInt >> 8) & 0xFF),
			B: uint8(colorInt & 0xFF),
			A: 255,
		}
	}
	return colors
}

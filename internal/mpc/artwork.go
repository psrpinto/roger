package mpc

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
)

func FindImage(dir string) (string, string) {
	imageExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".bmp": true}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if imageExts[ext] {
			return filepath.Join(dir, e.Name()), ext
		}
	}
	return "", ""
}

func ConvertCoverImage(srcPath, dstPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	// Resize to fit within 200x200
	dst := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	// Encode as JPEG at quality 80
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80}); err != nil {
		return err
	}

	// Build JFIF APP0 marker for 72 DPI
	jpegData := buf.Bytes()
	var out bytes.Buffer
	out.Write(jpegData[:2]) // SOI marker (FF D8)

	// JFIF APP0 segment: FF E0 + 16-byte payload
	var app0 [20]byte
	app0[0] = 0xFF
	app0[1] = 0xE0
	binary.BigEndian.PutUint16(app0[2:4], 16)  // length (includes length field but not marker)
	copy(app0[4:9], "JFIF\x00")                // identifier
	app0[9] = 1                                 // major version
	app0[10] = 1                                // minor version
	app0[11] = 1                                // units: dots per inch
	binary.BigEndian.PutUint16(app0[12:14], 72) // X density
	binary.BigEndian.PutUint16(app0[14:16], 72) // Y density
	app0[16] = 0                                // thumbnail width
	app0[17] = 0                                // thumbnail height
	out.Write(app0[:18])

	// Skip existing JFIF/EXIF APP0/APP1 markers if present
	i := 2
	for i < len(jpegData)-1 {
		if jpegData[i] != 0xFF {
			break
		}
		marker := jpegData[i+1]
		if marker == 0xE0 || marker == 0xE1 { // APP0 or APP1
			segLen := int(binary.BigEndian.Uint16(jpegData[i+2 : i+4]))
			i += 2 + segLen
			continue
		}
		break
	}
	out.Write(jpegData[i:])

	return os.WriteFile(dstPath, out.Bytes(), 0644)
}

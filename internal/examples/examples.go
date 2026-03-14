package examples

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"os"
)

// createDummyWav writes a minimal valid WAV file (44100 Hz, 16-bit mono, 100 frames of silence).
func createDummyWav(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	const (
		sampleRate    = 44100
		bitsPerSample = 16
		numChannels   = 1
		numFrames     = 100
		dataSize      = numFrames * numChannels * (bitsPerSample / 8) // 200 bytes
	)

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(36+dataSize)) // file size - 8
	f.Write([]byte("WAVE"))

	// fmt chunk
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))                                     // chunk size
	binary.Write(f, binary.LittleEndian, uint16(1))                                      // PCM format
	binary.Write(f, binary.LittleEndian, uint16(numChannels))                            // channels
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))                             // sample rate
	binary.Write(f, binary.LittleEndian, uint32(sampleRate*numChannels*bitsPerSample/8)) // byte rate
	binary.Write(f, binary.LittleEndian, uint16(numChannels*bitsPerSample/8))            // block align
	binary.Write(f, binary.LittleEndian, uint16(bitsPerSample))                          // bits per sample

	// data chunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))
	silence := make([]byte, dataSize)
	f.Write(silence)

	return nil
}

// createDummyPng writes a 1000x1000 dark PNG file at 150 DPI.
func createDummyPng(path string) error {
	const size = 1000
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	dark := color.NRGBA{30, 30, 30, 255}
	for y := range size {
		for x := range size {
			img.SetNRGBA(x, y, dark)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	encoded := buf.Bytes()

	// Insert a pHYs chunk after the IHDR chunk to set 150 DPI.
	// IHDR ends at byte 33 (8-byte signature + 25-byte IHDR chunk).
	const ihdrEnd = 33
	const ppm = 5906 // 150 DPI in pixels per meter (150 / 0.0254)

	var phys [21]byte
	binary.BigEndian.PutUint32(phys[0:4], 9) // data length
	copy(phys[4:8], "pHYs")
	binary.BigEndian.PutUint32(phys[8:12], ppm)  // X pixels per unit
	binary.BigEndian.PutUint32(phys[12:16], ppm) // Y pixels per unit
	phys[16] = 1                                  // unit = meter
	crc := crc32.NewIEEE()
	crc.Write(phys[4:17]) // CRC covers type + data
	binary.BigEndian.PutUint32(phys[17:21], crc.Sum32())

	data := make([]byte, 0, len(encoded)+len(phys))
	data = append(data, encoded[:ihdrEnd]...)
	data = append(data, phys[:]...)
	data = append(data, encoded[ihdrEnd:]...)

	return os.WriteFile(path, data, 0644)
}

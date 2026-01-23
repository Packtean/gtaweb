package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// GFXFile represents a Scaleform GFX file
type GFXFile struct {
	Header  [3]byte // "GFX"
	Version byte
	Data    []byte
}

func ReadGFXFile(path string) (*GFXFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read GFX file: %w", err)
	}

	if len(data) < 4 {
		return nil, fmt.Errorf("file too small to be a valid GFX file")
	}

	gfx := &GFXFile{
		Version: data[3],
		Data:    data[4:],
	}
	copy(gfx.Header[:], data[0:3])

	header := string(gfx.Header[:])
	if header != "GFX" && header != "CFX" {
		return nil, fmt.Errorf("not a valid GFX file (header: %s)", header)
	}

	return gfx, nil
}

func (gfx *GFXFile) ConvertToSWF(outputPath string) error {
	var buf bytes.Buffer

	if string(gfx.Header[:]) == "CFX" {
		buf.WriteString("CWS") // compressed SWF
	} else {
		buf.WriteString("FWS") // uncompressed SWF
	}
	buf.WriteByte(gfx.Version)

	_, err := buf.Write(gfx.Data)
	if err != nil {
		return fmt.Errorf("failed to write SWF data: %w", err)
	}

	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write SWF file: %w", err)
	}

	return nil
}

func (gfx *GFXFile) ExtractSWFData() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("FWS")
	buf.WriteByte(gfx.Version)

	_, err := buf.Write(gfx.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to create SWF data: %w", err)
	}

	return buf.Bytes(), nil
}

func (gfx *GFXFile) CopyToWriter(w io.Writer) error {
	if _, err := w.Write([]byte("FWS")); err != nil {
		return err
	}
	if _, err := w.Write([]byte{gfx.Version}); err != nil {
		return err
	}

	if _, err := w.Write(gfx.Data); err != nil {
		return err
	}

	return nil
}

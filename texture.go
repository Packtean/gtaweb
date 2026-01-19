package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/mauserzjeh/dxt"
)

// D3DFormat represents DirectX texture formats
type D3DFormat uint32

const (
	D3DFormat_DXT1     D3DFormat = 0x31545844 // "DXT1"
	D3DFormat_DXT3     D3DFormat = 0x33545844 // "DXT3"
	D3DFormat_DXT5     D3DFormat = 0x35545844 // "DXT5"
	D3DFormat_A8R8G8B8 D3DFormat = 0x15       // ARGB8888
	D3DFormat_L8       D3DFormat = 0x32       // Luminance 8-bit
)

// TextureInfo represents a texture in the texture dictionary
type TextureInfo struct {
	Name          string
	Width         uint16
	Height        uint16
	Format        D3DFormat
	Levels        byte
	RawDataOffset uint32
	TextureData   []byte
}

// TextureDictionary represents a collection of textures
type TextureDictionary struct {
	Textures []*TextureInfo
}

// ReadTextureDictionary reads the embedded texture dictionary from WHM file
func ReadTextureDictionary(mr *MemoryReader, offset int) (*TextureDictionary, error) {
	if offset == 0 || offset >= len(mr.SystemMem) {
		return nil, nil
	}

	td := &TextureDictionary{}
	pos := offset

	// skip vtable and pgBase
	pos += 8

	// read texture count and array (pgDictionary_grcTexturePC)
	pos += 8 // skip m_pParent and m_dwUsageCount

	// read hashes array (pgArray_DWORD)
	pos += 8 // skip hash array pointer and metadata

	// read texture array (pgObjectArray_grcTexture)
	textureArrayOffset, textureArrayType := mr.ReadPointer(pos)
	pos += 4
	textureCount := int(mr.ReadUInt16At(pos))
	pos += 2
	textureSize := int(mr.ReadUInt16At(pos))

	if textureArrayType != 5 || textureArrayOffset == 0 {
		return td, nil
	}

	// read each texture pointer
	for i := 0; i < textureSize && i < textureCount; i++ {
		texPtrPos := textureArrayOffset + i*4
		texOffset, texType := mr.ReadPointer(texPtrPos)
		if texType == 5 && texOffset != 0 {
			texInfo := readTextureInfo(mr, texOffset)
			if texInfo != nil {
				td.Textures = append(td.Textures, texInfo)
			}
		}
	}

	return td, nil
}

// readTextureInfo reads a single texture info structure
func readTextureInfo(mr *MemoryReader, offset int) *TextureInfo {
	if offset == 0 || offset >= len(mr.SystemMem) {
		return nil
	}

	tex := &TextureInfo{}
	pos := offset

	// skip vtable
	pos += 4

	// skip BlockMapOffset
	pos += 4

	// skip unks
	pos += 12

	// read name pointer
	tex.Name = mr.ReadStringPtr(pos)
	pos += 4

	// skip Unknown4
	pos += 4

	// read texture data
	tex.Width = mr.ReadUInt16At(pos)
	pos += 2
	tex.Height = mr.ReadUInt16At(pos)
	pos += 2
	tex.Format = D3DFormat(mr.ReadUInt32At(pos))
	pos += 4

	// skip stride and type
	pos += 3
	tex.Levels = mr.ReadByteAt(pos)
	pos++

	// skip floats
	pos += 24

	// skip prev/next pointers
	pos += 8

	// read RawDataOffset (physical pointer)
	rawDataPtr := mr.ReadUInt32At(pos)
	tex.RawDataOffset = rawDataPtr & 0x0FFFFFFF

	return tex
}

// ReadTextureData reads the actual texture pixel data from graphics memory
func (ti *TextureInfo) ReadTextureData(graphicsMem []byte) error {
	//if ti.RawDataOffset == 0 || int(ti.RawDataOffset) >= len(graphicsMem) {
	//	return fmt.Errorf("invalid texture data offset")
	//}

	dataSize := ti.GetDataSize()
	if int(ti.RawDataOffset)+dataSize > len(graphicsMem) {
		return fmt.Errorf("texture data out of bounds")
	}

	ti.TextureData = graphicsMem[ti.RawDataOffset : ti.RawDataOffset+uint32(dataSize)]
	return nil
}

// GetDataSize calculates the size of texture data
func (ti *TextureInfo) GetDataSize() int {
	width := uint32(ti.Width)
	height := uint32(ti.Height)

	var dataSize uint32
	switch ti.Format {
	case D3DFormat_DXT1:
		dataSize = width * height / 2
	case D3DFormat_DXT3, D3DFormat_DXT5:
		dataSize = width * height
	case D3DFormat_A8R8G8B8:
		dataSize = width * height * 4
	case D3DFormat_L8:
		dataSize = width * height
	default:
		return 0
	}

	// add mipmap levels
	levels := int(ti.Levels)
	levelDataSize := dataSize
	for levels > 1 {
		dataSize += levelDataSize / 4
		levelDataSize /= 4

		// clamp to minimum size
		if levelDataSize < 16 {
			if ti.Format == D3DFormat_DXT1 && levelDataSize < 8 {
				levelDataSize = 8
			} else {
				levelDataSize = 16
			}
		}
		levels--
	}

	return int(dataSize)
}

// DecodeToPNG decodes the texture and saves it as PNG
func (ti *TextureInfo) DecodeToPNG(outputPath string) error {
	if len(ti.TextureData) == 0 {
		return fmt.Errorf("no texture data")
	}

	var rgbaData []byte
	var err error

	width := uint(ti.Width)
	height := uint(ti.Height)

	switch ti.Format {
	case D3DFormat_DXT1:
		rgbaData, err = dxt.DecodeDXT1(ti.TextureData, width, height)
	case D3DFormat_DXT3:
		rgbaData, err = dxt.DecodeDXT3(ti.TextureData, width, height)
	case D3DFormat_DXT5:
		rgbaData, err = dxt.DecodeDXT5(ti.TextureData, width, height)
	case D3DFormat_A8R8G8B8:
		// ARGB -> RGBA
		rgbaData = make([]byte, len(ti.TextureData))
		for i := 0; i < len(ti.TextureData); i += 4 {
			rgbaData[i] = ti.TextureData[i+2]   // R
			rgbaData[i+1] = ti.TextureData[i+1] // G
			rgbaData[i+2] = ti.TextureData[i]   // B
			rgbaData[i+3] = ti.TextureData[i+3] // A
		}
	case D3DFormat_L8:
		// L8 -> RGBA
		rgbaData = make([]byte, len(ti.TextureData)*4)
		for i := 0; i < len(ti.TextureData); i++ {
			rgbaData[i*4] = ti.TextureData[i]   // R
			rgbaData[i*4+1] = ti.TextureData[i] // G
			rgbaData[i*4+2] = ti.TextureData[i] // B
			rgbaData[i*4+3] = 255               // A
		}
	default:
		return fmt.Errorf("unsupported texture format: %d", ti.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to decode texture: %w", err)
	}

	// create RGBA image
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	copy(img.Pix, rgbaData)

	// create output directory
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// save as PNG
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// ReadDataOffset reads a physical memory pointer
func (mr *MemoryReader) ReadDataOffset(offset int) uint32 {
	if offset+4 > len(mr.SystemMem) {
		return 0
	}
	ptrValue := binary.LittleEndian.Uint32(mr.SystemMem[offset:])
	return ptrValue & 0x0FFFFFFF
}

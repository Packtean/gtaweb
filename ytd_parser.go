package main

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/zeozeozeo/dxt"
)

type RSC7Header struct {
	Magic         [4]byte // "RSC7"
	Version       uint32
	SystemFlags   uint32
	GraphicsFlags uint32
}

type RSC7File struct {
	Header       RSC7Header
	SystemData   []byte
	GraphicsData []byte
}

const (
	BASE_SIZE = 0x2000
)

func ParseRSC7(filePath string) (*RSC7File, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) < 16 {
		return nil, fmt.Errorf("file too small to contain RSC7 header")
	}

	rsc := &RSC7File{}

	// parse header
	copy(rsc.Header.Magic[:], data[0:4])
	if string(rsc.Header.Magic[:]) != "RSC7" {
		return nil, fmt.Errorf("invalid magic: expected RSC7, got %s", string(rsc.Header.Magic[:]))
	}

	rsc.Header.Version = binary.LittleEndian.Uint32(data[4:8])
	rsc.Header.SystemFlags = binary.LittleEndian.Uint32(data[8:12])
	rsc.Header.GraphicsFlags = binary.LittleEndian.Uint32(data[12:16])

	// calculate memory sizes from flags
	systemSize := calculateMemorySize(rsc.Header.SystemFlags)
	graphicsSize := calculateMemorySize(rsc.Header.GraphicsFlags)

	fmt.Printf("RSC7 Header:\n")
	fmt.Printf("  Version: 0x%08X\n", rsc.Header.Version)
	fmt.Printf("  SystemFlags: 0x%08X -> %d bytes\n", rsc.Header.SystemFlags, systemSize)
	fmt.Printf("  GraphicsFlags: 0x%08X -> %d bytes\n", rsc.Header.GraphicsFlags, graphicsSize)

	// decompress data
	compressedData := data[16:]

	reader := flate.NewReader(bytes.NewReader(compressedData))
	defer reader.Close()

	//read system memory
	rsc.SystemData = make([]byte, systemSize)
	if systemSize > 0 {
		n, err := io.ReadFull(reader, rsc.SystemData)
		if err != nil {
			return nil, fmt.Errorf("failed to read system data: %w (read %d/%d bytes)", err, n, systemSize)
		}
	}

	// read graphics memory
	rsc.GraphicsData = make([]byte, graphicsSize)
	if graphicsSize > 0 {
		n, err := io.ReadFull(reader, rsc.GraphicsData)
		if err != nil {
			return nil, fmt.Errorf("failed to read graphics data: %w (read %d/%d bytes)", err, n, graphicsSize)
		}
	}

	fmt.Printf("  Decompressed: %d system bytes, %d graphics bytes\n", len(rsc.SystemData), len(rsc.GraphicsData))

	return rsc, nil
}

// calculateMemorySize calculates memory size from flags
func calculateMemorySize(flags uint32) int {
	pagesDiv16 := int((flags >> 27) & 0x1)
	pagesDiv8 := int((flags >> 26) & 0x1)
	pagesDiv4 := int((flags >> 25) & 0x1)
	pagesDiv2 := int((flags >> 24) & 0x1)
	pagesMul1 := int((flags >> 17) & 0x7F)
	pagesMul2 := int((flags >> 11) & 0x3F)
	pagesMul4 := int((flags >> 7) & 0xF)
	pagesMul8 := int((flags >> 5) & 0x3)
	pagesMul16 := int((flags >> 4) & 0x1)
	pagesSizeShift := int((flags >> 0) & 0xF)

	baseSize := BASE_SIZE << pagesSizeShift
	size := baseSize*pagesDiv16/16 +
		baseSize*pagesDiv8/8 +
		baseSize*pagesDiv4/4 +
		baseSize*pagesDiv2/2 +
		baseSize*pagesMul1*1 +
		baseSize*pagesMul2*2 +
		baseSize*pagesMul4*4 +
		baseSize*pagesMul8*8 +
		baseSize*pagesMul16*16

	return size
}

// YTDTexture represents a texture in the texture dictionary
type YTDTexture struct {
	Name          string
	Width         uint16
	Height        uint16
	Format        uint32
	MipMapLevels  uint8
	TextureData   []byte
	TextureOffset uint32 // offset in graphics memory
	DetectedDXT   string // DXT1/DXT5
}

// ParseYTDTextures extracts texture information from decompressed RSC7 data
func (rsc *RSC7File) ParseYTDTextures() ([]*YTDTexture, error) {
	if len(rsc.SystemData) < 0x40 {
		return nil, fmt.Errorf("system data too small for texture dictionary")
	}

	textures := []*YTDTexture{}

	// TextureDictionary at SystemData+0
	// 0x00-0x0F: FileBase64_GTA5_pc (vtable + pgBase)
	// 0x10-0x1F: unk
	// 0x20-0x2F: TextureNameHashes (ResourceSimpleList64<uint_r>)
	// 0x30-0x3F: Textures (ResourcePointerList64<TextureDX11>)

	textureListPtr := binary.LittleEndian.Uint64(rsc.SystemData[0x30:0x38])
	textureCount := binary.LittleEndian.Uint16(rsc.SystemData[0x38:0x3A])
	textureCapacity := binary.LittleEndian.Uint16(rsc.SystemData[0x3A:0x3C])

	fmt.Printf("\nTexture Dictionary:\n")
	fmt.Printf("  Texture list pointer: 0x%016X\n", textureListPtr)
	fmt.Printf("  Texture count: %d\n", textureCount)
	fmt.Printf("  Texture capacity: %d\n", textureCapacity)

	if textureListPtr == 0 || textureCount == 0 {
		return textures, nil
	}

	listOffset := resolvePointer(textureListPtr)
	if listOffset < 0 || listOffset >= len(rsc.SystemData) {
		return nil, fmt.Errorf("texture list pointer out of bounds: 0x%X -> offset %d", textureListPtr, listOffset)
	}

	fmt.Printf("  Texture list offset: 0x%X\n", listOffset)

	// read each texture pointer
	for i := 0; i < int(textureCount); i++ {
		ptrOffset := listOffset + i*8
		if ptrOffset+8 > len(rsc.SystemData) {
			break
		}

		texPtr := binary.LittleEndian.Uint64(rsc.SystemData[ptrOffset : ptrOffset+8])
		if texPtr == 0 {
			continue
		}

		texOffset := resolvePointer(texPtr)
		if texOffset < 0 || texOffset >= len(rsc.SystemData) {
			fmt.Printf("  Texture %d: pointer 0x%016X out of bounds\n", i, texPtr)
			continue
		}

		tex := rsc.parseTextureAt(texOffset)
		if tex != nil {
			textures = append(textures, tex)
			fmt.Printf("  Texture %d: %s (%dx%d, format 0x%X, %d mipmaps)\n",
				i, tex.Name, tex.Width, tex.Height, tex.Format, tex.MipMapLevels)
		}
	}

	return textures, nil
}

// resolvePointer converts a virtual pointer to an offset in memory
func resolvePointer(ptr uint64) int {
	// 0-27: offset (28 bits)
	// 28-31: memory type (4 bits) - 5 = system, 6 = graphics
	// 32-63: unused (always 0)
	return int(ptr & 0x0FFFFFFF)
}

// getPointerType returns the memory type of a pointer (5=system, 6=graphics, 0=invalid)
func getPointerType(ptr uint64) int {
	return int((ptr >> 28) & 0xF)
}

// parseTextureAt reads a texture structure at the given offset
func (rsc *RSC7File) parseTextureAt(offset int) *YTDTexture {
	// Note that this is a little different from the regular .ytd texture dictionaries
	if offset+0x80 > len(rsc.SystemData) {
		return nil
	}

	tex := &YTDTexture{}

	// 0x00-0x0F: unk (vtable/flags?)
	// 0x10-0x17: unk
	// 0x18: width (ushort)
	// 0x1A: height (ushort)
	// 0x20+: name pointer (somewhere in there)
	// 0x30+: graphics data pointer (somewhere in there)

	tex.Width = binary.LittleEndian.Uint16(rsc.SystemData[offset+0x18 : offset+0x1A])
	tex.Height = binary.LittleEndian.Uint16(rsc.SystemData[offset+0x1A : offset+0x1C])

	for off := 0; off < 0x80; off += 8 {
		if offset+off+8 > len(rsc.SystemData) {
			break
		}
		ptr := binary.LittleEndian.Uint64(rsc.SystemData[offset+off : offset+off+8])
		// check type nibble
		if ptr != 0 && getPointerType(ptr) == 5 {
			nameOffset := resolvePointer(ptr)
			if nameOffset >= 0 && nameOffset < len(rsc.SystemData) {
				name := readNullTerminatedString(rsc.SystemData[nameOffset:])
				if len(name) > 0 && len(name) < 100 && isValidTextureName(name) {
					tex.Name = name
					break
				}
			}
		}
	}

	for off := 0; off < 0x80; off += 8 {
		if offset+off+8 > len(rsc.SystemData) {
			break
		}
		ptr := binary.LittleEndian.Uint64(rsc.SystemData[offset+off : offset+off+8])
		if ptr != 0 && getPointerType(ptr) == 6 {
			tex.TextureOffset = uint32(ptr & 0x0FFFFFFF)
			break
		}
	}

	tex.Format = 0 // unk
	tex.MipMapLevels = 1

	return tex
}
func isValidTextureName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func readNullTerminatedString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}

func DumpTextures(ytdPath string) error {
	fmt.Printf("Parsing: %s\n", ytdPath)
	fmt.Println(string(bytes.Repeat([]byte("="), 80)))

	rsc, err := ParseRSC7(ytdPath)
	if err != nil {
		return fmt.Errorf("failed to parse RSC7: %w", err)
	}

	textures, err := rsc.ParseYTDTextures()
	if err != nil {
		return fmt.Errorf("failed to parse textures: %w", err)
	}

	fmt.Printf("\nFound %d textures:\n", len(textures))
	for i, tex := range textures {
		fmt.Printf("  [%d] %s (%dx%d)\n", i, tex.Name, tex.Width, tex.Height)
		fmt.Printf("      Graphics offset: 0x%X\n", tex.TextureOffset)
	}

	return nil
}
func ExtractYTDTextures(ytdPath, outputDir string) error {
	rsc, err := ParseRSC7(ytdPath)
	if err != nil {
		return fmt.Errorf("failed to parse RSC7: %w", err)
	}

	textures, err := rsc.ParseYTDTextures()
	if err != nil {
		return fmt.Errorf("failed to parse textures: %w", err)
	}

	if len(textures) == 0 {
		return nil
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, tex := range textures {
		if tex.TextureOffset < uint32(len(rsc.GraphicsData)) {
			detectTextureFormat(tex, rsc.GraphicsData)

			// Convert to PNG
			pngPath := filepath.Join(outputDir, tex.Name+".png")
			if err := tex.DecodeToPNG(pngPath); err != nil {
				fmt.Printf("Warning: failed to convert %s to PNG: %v\n", tex.Name, err)
			}
		}
	}

	return nil
}

func detectTextureFormat(tex *YTDTexture, graphicsData []byte) {
	if int(tex.TextureOffset) >= len(graphicsData) {
		return
	}

	w := int(tex.Width)
	h := int(tex.Height)

	dxt1Size := (w / 4) * (h / 4) * 8
	dxt5Size := (w / 4) * (h / 4) * 16

	available := len(graphicsData) - int(tex.TextureOffset)

	if available >= dxt5Size {
		tex.DetectedDXT = "DXT5"
		tex.TextureData = graphicsData[tex.TextureOffset : tex.TextureOffset+uint32(dxt5Size)]
	} else if available >= dxt1Size {
		tex.DetectedDXT = "DXT1"
		tex.TextureData = graphicsData[tex.TextureOffset : tex.TextureOffset+uint32(dxt1Size)]
	} else {
		tex.DetectedDXT = "DXT1"
		tex.TextureData = graphicsData[tex.TextureOffset:]
	}
}

// DecodeToPNG decodes the texture and saves it as PNG
func (tex *YTDTexture) DecodeToPNG(outputPath string) error {
	if len(tex.TextureData) == 0 {
		return fmt.Errorf("no texture data")
	}

	var rgbaData []byte
	var err error

	width := uint(tex.Width)
	height := uint(tex.Height)

	switch tex.DetectedDXT {
	case "DXT1":
		rgbaData, err = dxt.DecodeDXT1(tex.TextureData, width, height)
	case "DXT5":
		rgbaData, err = dxt.DecodeDXT5(tex.TextureData, width, height)
	default:
		return fmt.Errorf("unsupported texture format: %d", tex.Format)
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

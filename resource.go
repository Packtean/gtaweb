package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// ResourceFile represents a GTA IV resource file (RSC format)
type ResourceFile struct {
	Version         uint32
	SystemMemSize   uint32
	GraphicsMemSize uint32
	SystemMemData   []byte
	GraphicsMemData []byte
}

// ReadResourceFile reads a RSC file
func ReadResourceFile(filename string) (*ResourceFile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rf ResourceFile

	// read header
	header := make([]byte, 12)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// check magic
	if header[0] != 'R' || header[1] != 'S' || header[2] != 'C' {
		return nil, fmt.Errorf("invalid RSC magic")
	}

	rf.Version = binary.LittleEndian.Uint32(header[0:4])
	resourceType := binary.LittleEndian.Uint32(header[4:8])
	flags := binary.LittleEndian.Uint32(header[8:12])

	_ = resourceType

	// calculate sizes from flags
	rf.SystemMemSize = (flags & 0x7FF) << (((flags >> 11) & 0xF) + 8)
	rf.GraphicsMemSize = ((flags >> 15) & 0x7FF) << (((flags >> 26) & 0xF) + 8)

	// read compressed data
	compressedData, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %w", err)
	}

	// decompress
	reader, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	// split into system and graphics memory
	if len(decompressed) < int(rf.SystemMemSize) {
		return nil, fmt.Errorf("decompressed data too small: got %d, need %d", len(decompressed), rf.SystemMemSize)
	}

	rf.SystemMemData = decompressed[:rf.SystemMemSize]
	if len(decompressed) > int(rf.SystemMemSize) {
		rf.GraphicsMemData = decompressed[rf.SystemMemSize:]
	}

	return &rf, nil
}

// MemoryReader helps read from memory segments with pointer resolution
type MemoryReader struct {
	SystemMem   []byte
	GraphicsMem []byte
	currentPos  int
}

func NewMemoryReader(systemMem, graphicsMem []byte) *MemoryReader {
	return &MemoryReader{
		SystemMem:   systemMem,
		GraphicsMem: graphicsMem,
	}
}

func (mr *MemoryReader) ReadUInt32At(offset int) uint32 {
	if offset+4 > len(mr.SystemMem) {
		return 0
	}
	return binary.LittleEndian.Uint32(mr.SystemMem[offset:])
}

func (mr *MemoryReader) ReadInt32At(offset int) int32 {
	return int32(mr.ReadUInt32At(offset))
}

func (mr *MemoryReader) ReadUInt16At(offset int) uint16 {
	if offset+2 > len(mr.SystemMem) {
		return 0
	}
	return binary.LittleEndian.Uint16(mr.SystemMem[offset:])
}

func (mr *MemoryReader) ReadByteAt(offset int) byte {
	if offset >= len(mr.SystemMem) {
		return 0
	}
	return mr.SystemMem[offset]
}

func (mr *MemoryReader) ReadFloat32At(offset int) float32 {
	bits := mr.ReadUInt32At(offset)
	return float32frombits(bits)
}

// ReadPointer reads a relocated pointer (pgPtr)
func (mr *MemoryReader) ReadPointer(offset int) (int, int) {
	if offset+4 > len(mr.SystemMem) {
		return 0, 0
	}
	ptrValue := binary.LittleEndian.Uint32(mr.SystemMem[offset:])
	ptrOffset := int(ptrValue & 0x0FFFFFFF)
	ptrType := int(ptrValue >> 28)
	return ptrOffset, ptrType
}

// ReadString reads a null-terminated string at the given offset
func (mr *MemoryReader) ReadString(offset int) string {
	if offset >= len(mr.SystemMem) {
		return ""
	}

	end := offset
	for end < len(mr.SystemMem) && mr.SystemMem[end] != 0 {
		end++
	}

	return string(mr.SystemMem[offset:end])
}

// ReadStringPtr reads a pointer to a string
func (mr *MemoryReader) ReadStringPtr(offset int) string {
	ptrOffset, ptrType := mr.ReadPointer(offset)
	if ptrType != 5 { // 5 = valid RAM pointer
		return ""
	}
	return mr.ReadString(ptrOffset)
}

func float32frombits(bits uint32) float32 {
	return math.Float32frombits(bits)
}

package main

import (
	"fmt"
)

// ReadWHMFile reads and parses a WHM file
func ReadWHMFile(filename string) (*HtmlDocument, error) {
	rf, err := ReadResourceFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource file: %w", err)
	}

	mr := NewMemoryReader(rf.SystemMemData, rf.GraphicsMemData)
	doc := &HtmlDocument{}

	if len(rf.SystemMemData) < 64 {
		return nil, fmt.Errorf("system memory too small: %d bytes", len(rf.SystemMemData))
	}

	// read root element pointer at offset 0
	rootOffset, rootType := mr.ReadPointer(0)

	// pointer type 0 might also be valid
	if rootOffset != 0 && rootOffset < len(mr.SystemMem) {
		doc.RootElement = parseNode(mr, rootOffset)
	}

	if doc.RootElement == nil {
		return nil, fmt.Errorf("failed to parse root element (offset: %d, type: %d, sysmem: %d bytes)",
			rootOffset, rootType, len(rf.SystemMemData))
	}

	// read texture dictionary offset (CHtmlDocument+12)
	tdOffset, tdType := mr.ReadPointer(12)
	if tdType == 5 && tdOffset != 0 {
		doc.TextureDictOffset = uint32(tdOffset)
		texDict, err := ReadTextureDictionary(mr, tdOffset)
		if err == nil && texDict != nil {
			doc.TextureDictionary = texDict

			for _, tex := range texDict.Textures {
				if err := tex.ReadTextureData(rf.GraphicsMemData); err != nil {
					fmt.Printf("  ERROR: failed to read texture data for texture %s: %v\n", tex.Name, err)
					continue
				}
			}
		}
	}

	return doc, nil
}

// parseNode parses an HtmlNode from memory
func parseNode(mr *MemoryReader, offset int) *HtmlNode {
	if offset == 0 || offset >= len(mr.SystemMem) {
		return nil
	}

	node := &HtmlNode{}

	pos := offset

	// read vtable and node type
	node.VTable = mr.ReadUInt32At(pos)
	pos += 4
	node.NodeType = HtmlNodeType(mr.ReadUInt32At(pos))
	pos += 4

	// skip parent node pointer
	pos += 4

	// read child nodes (pgObjectArray)
	childArrayOffset, childArrayType := mr.ReadPointer(pos)
	pos += 4
	_ = int(mr.ReadUInt16At(pos)) // childCount
	pos += 2
	childSize := int(mr.ReadUInt16At(pos))
	pos += 2

	if childArrayType == 5 && childArrayOffset != 0 {
		node.ChildNodes = parseChildNodes(mr, childArrayOffset, childSize)
	}

	// read render state
	node.RenderState = parseRenderState(mr, pos)
	pos += 0xC4

	// parse based on node type
	if node.NodeType == HtmlDataNode {
		// data node, read string pointer
		node.Data = mr.ReadStringPtr(pos)
	} else {
		// element node
		node.Tag = HtmlTag(mr.ReadUInt32At(pos))
		pos += 4

		// tag name pointer (m_pszTagName)
		tagName := mr.ReadStringPtr(pos)
		pos += 4

		if node.Tag == TagText || node.Tag == TagScriptObj {
			node.Data = tagName
		}

		// read link address (SimpleCollection of bytes)
		linkOffset, linkType := mr.ReadPointer(pos)
		pos += 4
		linkCount := int(mr.ReadUInt16At(pos))
		pos += 2
		_ = int(mr.ReadUInt16At(pos)) // linkSize
		pos += 2

		if linkType == 5 && linkOffset != 0 && linkCount > 0 {
			linkBytes := make([]byte, linkCount)
			for i := range linkCount {
				linkBytes[i] = mr.ReadByteAt(linkOffset + i)
			}
			node.LinkAddress = string(linkBytes)
		}

		// read background image name if present
		if node.RenderState.BackgroundImageOffset != 0 {
			node.RenderState.BackgroundImageName = readTextureName(mr, int(node.RenderState.BackgroundImageOffset))
		}
	}

	return node
}

// parseChildNodes parses an array of child node pointers
func parseChildNodes(mr *MemoryReader, arrayOffset, count int) []*HtmlNode {
	nodes := make([]*HtmlNode, 0, count)

	for i := range count {
		ptrPos := arrayOffset + i*4
		childOffset, childType := mr.ReadPointer(ptrPos)
		if childType == 5 && childOffset != 0 {
			child := parseNode(mr, childOffset)
			if child != nil {
				nodes = append(nodes, child)
			}
		}
	}

	return nodes
}

// parseRenderState parses the HtmlRenderState structure
func parseRenderState(mr *MemoryReader, offset int) HtmlRenderState {
	rs := HtmlRenderState{}
	pos := offset

	rs.Display = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.Width = mr.ReadFloat32At(pos)
	pos += 4
	rs.Height = mr.ReadFloat32At(pos)
	pos += 4
	rs._fC = mr.ReadFloat32At(pos)
	pos += 4
	rs._f10 = mr.ReadFloat32At(pos)
	pos += 4
	for i := range 4 {
		rs._f14[i] = mr.ReadByteAt(pos)
		pos++
	}
	rs._f18 = mr.ReadFloat32At(pos)
	pos += 4
	rs._f1C = mr.ReadFloat32At(pos)
	pos += 4
	rs.BackgroundColor = mr.ReadUInt32At(pos)
	pos += 4
	bgImgOffset, _ := mr.ReadPointer(pos)
	rs.BackgroundImageOffset = uint32(bgImgOffset)
	pos += 4
	rs._f28h = mr.ReadUInt32At(pos)
	pos += 4
	rs._f28l = mr.ReadUInt32At(pos)
	pos += 4
	rs.BackgroundRepeat = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.Color = mr.ReadUInt32At(pos)
	pos += 4
	rs.HorizontalAlign = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.VerticalAlign = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.TextDecoration = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs._f44 = mr.ReadUInt32At(pos)
	pos += 4
	rs.FontSize = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.FontStyle = mr.ReadInt32At(pos)
	pos += 4
	rs.FontWeight = mr.ReadInt32At(pos)
	pos += 4
	rs._f54 = mr.ReadUInt32At(pos)
	pos += 4
	rs.BorderBottomColor = mr.ReadUInt32At(pos)
	pos += 4
	rs.BorderBottomStyle = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.BorderBottomWidth = mr.ReadFloat32At(pos)
	pos += 4
	rs.BorderLeftColor = mr.ReadUInt32At(pos)
	pos += 4
	rs.BorderLeftStyle = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.BorderLeftWidth = mr.ReadFloat32At(pos)
	pos += 4
	rs.BorderRightColor = mr.ReadUInt32At(pos)
	pos += 4
	rs.BorderRightStyle = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.BorderRightWidth = mr.ReadFloat32At(pos)
	pos += 4
	rs.BorderTopColor = mr.ReadUInt32At(pos)
	pos += 4
	rs.BorderTopStyle = HtmlAttributeValue(mr.ReadUInt32At(pos))
	pos += 4
	rs.BorderTopWidth = mr.ReadFloat32At(pos)
	pos += 4
	rs.MarginBottom = mr.ReadFloat32At(pos)
	pos += 4
	rs.MarginLeft = mr.ReadFloat32At(pos)
	pos += 4
	rs.MarginRight = mr.ReadFloat32At(pos)
	pos += 4
	rs.MarginTop = mr.ReadFloat32At(pos)
	pos += 4
	rs.PaddingBottom = mr.ReadFloat32At(pos)
	pos += 4
	rs.PaddingLeft = mr.ReadFloat32At(pos)
	pos += 4
	rs.PaddingRight = mr.ReadFloat32At(pos)
	pos += 4
	rs.PaddingTop = mr.ReadFloat32At(pos)
	pos += 4
	rs.CellPadding = mr.ReadFloat32At(pos)
	pos += 4
	rs.CellSpacing = mr.ReadFloat32At(pos)
	pos += 4
	rs.ColSpan = mr.ReadInt32At(pos)
	pos += 4
	rs.RowSpan = mr.ReadInt32At(pos)
	pos += 4
	rs.HasBackground = mr.ReadByteAt(pos)
	pos++
	rs._fB9 = mr.ReadByteAt(pos)
	pos++
	rs._fBA[0] = mr.ReadByteAt(pos)
	pos++
	rs._fBA[1] = mr.ReadByteAt(pos)
	pos++
	rs.ALinkColor = mr.ReadUInt32At(pos)
	pos += 4
	rs._fC0 = mr.ReadInt32At(pos)

	return rs
}

// readTextureName reads texture name from TextureInfo structure
func readTextureName(mr *MemoryReader, offset int) string {
	if offset == 0 || offset >= len(mr.SystemMem) {
		return ""
	}

	// TextureInfo structure
	// skip VTable (4), Unknown1 (4), Unknown2 (2), Unknown3 (2), Unknown4 (4), Unknown5 (4)
	pos := offset + 20 // TextureNameOffset
	nameOffset, nameType := mr.ReadPointer(pos)
	if nameType == 5 && nameOffset != 0 {
		return mr.ReadString(nameOffset)
	}

	return ""
}

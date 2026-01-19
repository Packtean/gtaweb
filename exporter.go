package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tdewolff/minify/v2"
	minifycss "github.com/tdewolff/minify/v2/css"
	minifyhtml "github.com/tdewolff/minify/v2/html"
)

var globalMinifier = minify.New()

func init() {
	globalMinifier.Add("text/html", &minifyhtml.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
	})
	globalMinifier.Add("text/css", &minifycss.Minifier{})
}

// ExportHTML exports the HTML document to a file
func ExportHTML(doc *HtmlDocument, filename string, depth int) error {
	// iv/www.example.com/page.html -> www.example.com
	currentSite := extractSiteName(filename)

	var buf bytes.Buffer

	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html>\n<head>\n")
	buf.WriteString("<meta charset=\"UTF-8\">\n")
	buf.WriteString("<style>\n")

	fmt.Fprintf(&buf, `@font-face {
  font-family: 'DIN Medium';
  font-style: normal;
  font-weight: normal;
  src: local('DIN Medium'), url('%sttf/DIN-Medium.woff2') format('woff2');
}`, strings.Repeat("../", depth))

	buf.WriteString(`body {
  font-family: 'DIN Medium', Helvetica, Arial, sans-serif;
  margin: 0;
  padding: 0;
  background-color: #000;
  zoom: 150%;
}
</style>
</head><body>`)

	if doc.RootElement != nil {
		exportNodeWithContext(doc.RootElement, &buf, "", currentSite)
	}

	buf.WriteString("</body></html>")

	minified, err := globalMinifier.Bytes("text/html", buf.Bytes())
	if err != nil {
		minified = buf.Bytes()
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(minified)
	return err
}

// extractSiteName extracts the site name from a file path, e.g. iv/www.example.com/page.html -> www.example.com
func extractSiteName(filepath string) string {
	parts := strings.SplitSeq(strings.ReplaceAll(filepath, "\\", "/"), "/")
	for part := range parts {
		if strings.HasPrefix(part, "www.") {
			return part
		}
	}
	return ""
}

// exportNodeWithContext exports a single node and its children with site context
func exportNodeWithContext(node *HtmlNode, w io.Writer, indent string, currentSite string) {
	if node == nil {
		return
	}

	// handle data nodes
	if node.NodeType == HtmlDataNode {
		if node.Data != "" {
			// try to localize the text
			text := LocalizeText(node.Data)
			// escape HTML entities
			escaped := html.EscapeString(text)
			w.Write([]byte(escaped))
		}
		return
	}

	// handle element nodes
	tagName := node.Tag.String()
	attributes := getAttributesWithContext(node, currentSite)
	emptyNode := len(node.ChildNodes) == 0

	// skip style tag since we add our own
	if node.Tag == TagStyle {
		return
	}

	if node.Tag == TagText || node.Tag == TagScriptObj {
		if node.Data != "" {
			text := LocalizeText(node.Data)
			w.Write([]byte(html.EscapeString(text)))
			return
		}
		if node.LinkAddress != "" {
			text := LocalizeText(node.LinkAddress)
			w.Write([]byte(html.EscapeString(text)))
			return
		}
		if len(node.ChildNodes) > 0 {
			for _, child := range node.ChildNodes {
				exportNodeWithContext(child, w, indent, currentSite)
			}
			return
		}
		return
	}

	if node.Tag == TagHtml || node.Tag == TagHead {
		for _, child := range node.ChildNodes {
			exportNodeWithContext(child, w, indent, currentSite)
		}
		return
	}

	// for body tag, apply its background color and create a centered container
	if node.Tag == TagBody {
		bgColor := getColor(node.RenderState.BackgroundColor)

		w.Write(fmt.Appendf(nil, "<div style=\"background-color: %s; min-height: 100vh; display: flex; justify-content: center; align-items: flex-start;\">", bgColor))
		w.Write([]byte("<div style=\"max-width: 100%; margin: 0 auto;\">"))

		for _, child := range node.ChildNodes {
			exportNodeWithContext(child, w, indent, currentSite)
		}

		w.Write([]byte("</div></div>"))
		return
	}

	// Write opening tag
	if emptyNode && node.Tag != TagStyle {
		w.Write(fmt.Appendf(nil, "<%s%s/>", tagName, attributes))
	} else {
		w.Write(fmt.Appendf(nil, "<%s%s>", tagName, attributes))

		// Write children
		for _, child := range node.ChildNodes {
			exportNodeWithContext(child, w, indent+"   ", currentSite)
		}

		// Write closing tag
		w.Write(fmt.Appendf(nil, "</%s>", tagName))
	}
}

// getAttributesWithContext generates HTML attributes from node render state
func getAttributesWithContext(node *HtmlNode, currentSite string) string {
	if node.Tag == TagHtml || node.Tag == TagHead || node.Tag == TagTitle || node.Tag == TagStyle {
		return ""
	}

	attrs := make(map[string]string)
	styleAttrs := make(map[string]string)
	rs := node.RenderState

	// Background
	if rs.HasBackground == 1 {
		if rs.BackgroundImageOffset == 0 {
			styleAttrs["background-color"] = getColor(rs.BackgroundColor)
		} else if rs.BackgroundImageName != "" {
			styleAttrs["background-image"] = fmt.Sprintf("url(%s.png)", rs.BackgroundImageName)
			styleAttrs["background-repeat"] = getAttributeValue(rs.BackgroundRepeat)
		}
	}

	// Width and Height
	if rs.Width > -1 {
		if node.NodeType == HtmlTableNode || node.NodeType == HtmlTableElementNode {
			attrs["width"] = fmt.Sprintf("%.0f", rs.Width)
		} else {
			styleAttrs["width"] = getSize(rs.Width)
		}
	}
	if rs.Height > -1 {
		if node.NodeType == HtmlTableNode || node.NodeType == HtmlTableElementNode {
			attrs["height"] = fmt.Sprintf("%.0f", rs.Height)
		} else {
			styleAttrs["height"] = getSize(rs.Height)
		}
	}

	// Margins - only if non-zero
	if rs.MarginLeft != 0 {
		styleAttrs["margin-left"] = getSize(rs.MarginLeft)
	}
	if rs.MarginRight != 0 {
		styleAttrs["margin-right"] = getSize(rs.MarginRight)
	}
	if rs.MarginTop != 0 {
		styleAttrs["margin-top"] = getSize(rs.MarginTop)
	}
	if rs.MarginBottom != 0 {
		styleAttrs["margin-bottom"] = getSize(rs.MarginBottom)
	}

	// Padding - only if non-zero
	if rs.PaddingLeft != 0 {
		styleAttrs["padding-left"] = getSize(rs.PaddingLeft)
	}
	if rs.PaddingRight != 0 {
		styleAttrs["padding-right"] = getSize(rs.PaddingRight)
	}
	if rs.PaddingTop != 0 {
		styleAttrs["padding-top"] = getSize(rs.PaddingTop)
	}
	if rs.PaddingBottom != 0 {
		styleAttrs["padding-bottom"] = getSize(rs.PaddingBottom)
	}

	// Borders - only if style is not 'none'
	if rs.BorderLeftStyle != AttrNone && rs.BorderLeftWidth > 0 {
		styleAttrs["border-left"] = fmt.Sprintf("%s %s %s",
			getAttributeValue(rs.BorderLeftStyle),
			getSize(rs.BorderLeftWidth),
			getColor(rs.BorderLeftColor))
	}
	if rs.BorderRightStyle != AttrNone && rs.BorderRightWidth > 0 {
		styleAttrs["border-right"] = fmt.Sprintf("%s %s %s",
			getAttributeValue(rs.BorderRightStyle),
			getSize(rs.BorderRightWidth),
			getColor(rs.BorderRightColor))
	}
	if rs.BorderTopStyle != AttrNone && rs.BorderTopWidth > 0 {
		styleAttrs["border-top"] = fmt.Sprintf("%s %s %s",
			getAttributeValue(rs.BorderTopStyle),
			getSize(rs.BorderTopWidth),
			getColor(rs.BorderTopColor))
	}
	if rs.BorderBottomStyle != AttrNone && rs.BorderBottomWidth > 0 {
		styleAttrs["border-bottom"] = fmt.Sprintf("%s %s %s",
			getAttributeValue(rs.BorderBottomStyle),
			getSize(rs.BorderBottomWidth),
			getColor(rs.BorderBottomColor))
	}

	// Text styling
	if textDeco := getAttributeValue(rs.TextDecoration); textDeco != "" {
		styleAttrs["text-decoration"] = textDeco
	}
	if fontSize := getAttributeValue(rs.FontSize); fontSize != "" {
		styleAttrs["font-size"] = fontSize
	}
	// Handle display property carefully - skip for table elements entirely
	// Tables should use their native HTML display behavior
	if node.Tag != TagTable && node.Tag != TagTr && node.Tag != TagTd && node.Tag != TagTh && node.NodeType != HtmlTableNode && node.NodeType != HtmlTableElementNode {
		if display := getAttributeValue(rs.Display); display != "" && display != "inline" {
			styleAttrs["display"] = display
		}
	}
	if rs.Color != 0 {
		styleAttrs["color"] = getColor(rs.Color)
	}

	// table-specific attributes
	switch node.NodeType {
	case HtmlTableNode:
		attrs["cellpadding"] = fmt.Sprintf("%.0f", rs.CellPadding)
		attrs["cellspacing"] = fmt.Sprintf("%.0f", rs.CellSpacing)
		if rs.CellSpacing == 0 {
			styleAttrs["border-collapse"] = "collapse"
		}
	case HtmlTableElementNode:
		if rs.ColSpan > 1 {
			attrs["colspan"] = fmt.Sprintf("%d", rs.ColSpan)
		}
		if rs.RowSpan > 1 {
			attrs["rowspan"] = fmt.Sprintf("%d", rs.RowSpan)
		}
		if int(rs.VerticalAlign) > -1 {
			attrs["valign"] = getAttributeValue(rs.VerticalAlign)
		}
		if int(rs.HorizontalAlign) > -1 {
			attrs["align"] = getAttributeValue(rs.HorizontalAlign)
		}
	default:
		// only add alignment if it's explicitly set
		if int(rs.VerticalAlign) > -1 && int(rs.VerticalAlign) != int(AttrInherit) {
			if valign := getAttributeValue(rs.VerticalAlign); valign != "" {
				styleAttrs["vertical-align"] = valign
			}
		}
		if int(rs.HorizontalAlign) > -1 && int(rs.HorizontalAlign) != int(AttrInherit) {
			if halign := getAttributeValue(rs.HorizontalAlign); halign != "" {
				styleAttrs["text-align"] = halign
			}
		}
	}

	// tag-specific attributes
	switch node.Tag {
	case TagA:
		if node.LinkAddress != "" {
			// convert internal links to proper relative paths
			link := node.LinkAddress

			link = strings.ReplaceAll(link, "\\", "/")

			if after, ok := strings.CutPrefix(link, "http://"); ok {
				link = after
			}

			if strings.HasPrefix(link, "www.") {
				// extract target site
				parts := strings.SplitN(link, "/", 2)
				targetSite := parts[0]
				targetPath := ""
				if len(parts) == 2 {
					targetPath = parts[1]
				}

				// check if we should use relative path
				if targetSite == currentSite {
					if targetPath != "" {
						link = targetPath
					} else {
						link = "index.html"
					}
				} else {
					if targetPath != "" {
						link = "../" + targetSite + "/" + targetPath
					} else {
						link = "../" + targetSite + "/index.html"
					}
				}
			}

			// add .html extension if missing
			if !strings.HasSuffix(link, ".html") && !strings.HasSuffix(link, ".htm") {
				baseName := filepath.Base(link)
				if !strings.Contains(baseName, ".") {
					link = link + ".html"
				}
			}

			attrs["href"] = link
		} else {
			attrs["href"] = "#"
		}
	case TagImg:
		if node.LinkAddress != "" {
			path := node.LinkAddress
			// remove extension if present
			if idx := strings.LastIndex(path, "."); idx != -1 {
				path = path[:idx]
			}
			attrs["src"] = path + ".png"
		}
	}

	// build attribute string
	var result strings.Builder

	// add style attribute if we have style properties
	if len(styleAttrs) > 0 {
		var styleBuilder strings.Builder
		for key, value := range styleAttrs {
			fmt.Fprintf(&styleBuilder, "%s: %s; ", key, value)
		}
		attrs["style"] = styleBuilder.String()
	}

	// add all attributes
	for key, value := range attrs {
		fmt.Fprintf(&result, " %s=\"%s\"", key, value)
	}

	return result.String()
}

// getColor converts a uint32 color to CSS hex format
func getColor(color uint32) string {
	return fmt.Sprintf("#%06x", color&0xFFFFFF)
}

// getSize converts a float size to CSS px format
func getSize(size float32) string {
	return fmt.Sprintf("%.0fpx", size)
}

// getAttributeValue converts an HtmlAttributeValue to CSS string
func getAttributeValue(value HtmlAttributeValue) string {
	switch value {
	case AttrLeft:
		return "left"
	case AttrRight:
		return "right"
	case AttrCenter:
		return "center"
	case AttrJustify:
		return "justify"
	case AttrTop:
		return "top"
	case AttrBottom:
		return "bottom"
	case AttrMiddle:
		return "middle"
	case AttrInherit:
		return "inherit"
	case AttrXXSmall:
		return "6px"
	case AttrXSmall:
		return "7px"
	case AttrSmall:
		return "8px"
	case AttrMedium:
		return "9px"
	case AttrLarge:
		return "11px"
	case AttrXLarge:
		return "12px"
	case AttrXXLarge:
		return "14px"
	case AttrBlock:
		return "block"
	case AttrInline:
		return "inline"
	case AttrTable:
		return "table"
	case AttrTableCell:
		return "table-cell"
	case AttrNone:
		return "none"
	case AttrSolid:
		return "solid"
	case AttrUnderline:
		return "underline"
	case AttrOverline:
		return "overline"
	case AttrLineThrough:
		return "line-through"
	case AttrBlink:
		return "blink"
	case AttrRepeat:
		return "repeat"
	case AttrNoRepeat:
		return "no-repeat"
	case AttrRepeatX:
		return "repeat-x"
	case AttrRepeatY:
		return "repeat-y"
	case AttrCollapse:
		return "collapse"
	case AttrSeparate:
		return "separate"
	default:
		return ""
	}
}

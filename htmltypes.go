package main

// HtmlNodeType represents the type of HTML node
type HtmlNodeType uint32

const (
	HtmlElementNode      HtmlNodeType = 0
	HtmlDataNode         HtmlNodeType = 1
	HtmlTableNode        HtmlNodeType = 2
	HtmlTableElementNode HtmlNodeType = 3
)

// HtmlTag represents HTML tag types
type HtmlTag uint32

const (
	TagHtml      HtmlTag = 0
	TagTitle     HtmlTag = 1
	TagA         HtmlTag = 2
	TagBody      HtmlTag = 3
	TagB         HtmlTag = 4
	TagBr        HtmlTag = 5
	TagCenter    HtmlTag = 6
	TagCode      HtmlTag = 7
	TagDl        HtmlTag = 8
	TagDt        HtmlTag = 9
	TagDd        HtmlTag = 10
	TagDiv       HtmlTag = 11
	TagEmbed     HtmlTag = 12
	TagEm        HtmlTag = 13
	TagHead      HtmlTag = 14
	TagH1        HtmlTag = 15
	TagH2        HtmlTag = 16
	TagH3        HtmlTag = 17
	TagH4        HtmlTag = 18
	TagH5        HtmlTag = 19
	TagH6        HtmlTag = 20
	TagImg       HtmlTag = 21
	TagI         HtmlTag = 22
	TagLink      HtmlTag = 23
	TagLi        HtmlTag = 24
	TagMeta      HtmlTag = 25
	TagObject    HtmlTag = 26
	TagOl        HtmlTag = 27
	TagP         HtmlTag = 28
	TagParam     HtmlTag = 29
	TagSpan      HtmlTag = 30
	TagStrong    HtmlTag = 31
	TagStyle     HtmlTag = 32
	TagTable     HtmlTag = 33
	TagTr        HtmlTag = 34
	TagTh        HtmlTag = 35
	TagTd        HtmlTag = 36
	TagUl        HtmlTag = 37
	TagText      HtmlTag = 38
	TagScriptObj HtmlTag = 39
)

var tagNames = map[HtmlTag]string{
	TagHtml: "html", TagTitle: "title", TagA: "a", TagBody: "body",
	TagB: "b", TagBr: "br", TagCenter: "center", TagCode: "code",
	TagDl: "dl", TagDt: "dt", TagDd: "dd", TagDiv: "div",
	TagEmbed: "embed", TagEm: "em", TagHead: "head", TagH1: "h1",
	TagH2: "h2", TagH3: "h3", TagH4: "h4", TagH5: "h5", TagH6: "h6",
	TagImg: "img", TagI: "i", TagLink: "link", TagLi: "li",
	TagMeta: "meta", TagObject: "object", TagOl: "ol", TagP: "p",
	TagParam: "param", TagSpan: "span", TagStrong: "strong", TagStyle: "style",
	TagTable: "table", TagTr: "tr", TagTh: "th", TagTd: "td", TagUl: "ul",
	TagText:      "span",
	TagScriptObj: "span", // TODO?
}

func (t HtmlTag) String() string {
	if name, ok := tagNames[t]; ok {
		return name
	}
	return "unknown"
}

// HtmlAttributeValue represents CSS/HTML attribute values
type HtmlAttributeValue uint32

const (
	AttrLeft        HtmlAttributeValue = 0
	AttrRight       HtmlAttributeValue = 1
	AttrCenter      HtmlAttributeValue = 2
	AttrJustify     HtmlAttributeValue = 3
	AttrTop         HtmlAttributeValue = 4
	AttrBottom      HtmlAttributeValue = 5
	AttrMiddle      HtmlAttributeValue = 6
	AttrInherit     HtmlAttributeValue = 7
	AttrXXSmall     HtmlAttributeValue = 8
	AttrXSmall      HtmlAttributeValue = 9
	AttrSmall       HtmlAttributeValue = 10
	AttrMedium      HtmlAttributeValue = 11
	AttrLarge       HtmlAttributeValue = 12
	AttrXLarge      HtmlAttributeValue = 13
	AttrXXLarge     HtmlAttributeValue = 14
	AttrBlock       HtmlAttributeValue = 15
	AttrTable       HtmlAttributeValue = 16
	AttrTableCell   HtmlAttributeValue = 17
	AttrInline      HtmlAttributeValue = 18
	AttrNone        HtmlAttributeValue = 19
	AttrSolid       HtmlAttributeValue = 20
	AttrUnderline   HtmlAttributeValue = 21
	AttrOverline    HtmlAttributeValue = 22
	AttrLineThrough HtmlAttributeValue = 23
	AttrBlink       HtmlAttributeValue = 24
	AttrRepeat      HtmlAttributeValue = 25
	AttrNoRepeat    HtmlAttributeValue = 26
	AttrRepeatX     HtmlAttributeValue = 27
	AttrRepeatY     HtmlAttributeValue = 28
	AttrCollapse    HtmlAttributeValue = 29
	AttrSeparate    HtmlAttributeValue = 30
)

// HtmlRenderState contains CSS styling information
type HtmlRenderState struct {
	Display               HtmlAttributeValue
	Width                 float32
	Height                float32
	_fC                   float32
	_f10                  float32
	_f14                  [4]byte
	_f18                  float32
	_f1C                  float32
	BackgroundColor       uint32
	BackgroundImageOffset uint32
	_f28h                 uint32
	_f28l                 uint32
	BackgroundRepeat      HtmlAttributeValue
	Color                 uint32
	HorizontalAlign       HtmlAttributeValue
	VerticalAlign         HtmlAttributeValue
	TextDecoration        HtmlAttributeValue
	_f44                  uint32
	FontSize              HtmlAttributeValue
	FontStyle             int32
	FontWeight            int32
	_f54                  uint32
	BorderBottomColor     uint32
	BorderBottomStyle     HtmlAttributeValue
	BorderBottomWidth     float32
	BorderLeftColor       uint32
	BorderLeftStyle       HtmlAttributeValue
	BorderLeftWidth       float32
	BorderRightColor      uint32
	BorderRightStyle      HtmlAttributeValue
	BorderRightWidth      float32
	BorderTopColor        uint32
	BorderTopStyle        HtmlAttributeValue
	BorderTopWidth        float32
	MarginBottom          float32
	MarginLeft            float32
	MarginRight           float32
	MarginTop             float32
	PaddingBottom         float32
	PaddingLeft           float32
	PaddingRight          float32
	PaddingTop            float32
	CellPadding           float32
	CellSpacing           float32
	ColSpan               int32
	RowSpan               int32
	HasBackground         byte
	_fB9                  byte
	_fBA                  [2]byte
	ALinkColor            uint32
	_fC0                  int32

	// Resolved data
	BackgroundImageName string
}

// HtmlNode represents a node in the HTML document tree
type HtmlNode struct {
	VTable      uint32
	NodeType    HtmlNodeType
	ParentNode  *HtmlNode
	ChildNodes  []*HtmlNode
	RenderState HtmlRenderState

	// For ElementNodes
	Tag         HtmlTag
	LinkAddress string

	// For DataNodes
	Data string
}

// HtmlDocument represents the parsed WHM document
type HtmlDocument struct {
	RootElement       *HtmlNode
	TextureDictionary *TextureDictionary
	TextureDictOffset uint32
}

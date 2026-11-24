package markdown

// Publiczny widok drzewa dokumentu dla walidatorów (ADR 0010). Drzewo jest tylko do odczytu: Node to mały
// uchwyt z metodami, a wewnętrzna struktura węzłów pozostaje prywatna.

// Kind is the kind of a document node.
type Kind uint8

// Node kinds. Blocks:
const (
	KindDocument      = Kind(kindDocument)
	KindBlockQuote    = Kind(kindBlockQuote) // also an alert: Node.Alert then returns its kind
	KindList          = Kind(kindList)
	KindItem          = Kind(kindItem)
	KindParagraph     = Kind(kindParagraph)
	KindHeading       = Kind(kindHeading)
	KindThematicBreak = Kind(kindThematicBreak)
	KindCodeBlock     = Kind(kindCodeBlock)
	KindHTMLBlock     = Kind(kindHTMLBlock)
	KindTable         = Kind(kindTable)
	KindTableRow      = Kind(kindTableRow)
	KindTableCell     = Kind(kindTableCell)
	KindFootnote      = Kind(kindFootnoteDef) // a footnote definition, at the place where it stands in the source
)

// Inline content:
const (
	KindText         = Kind(kindText)
	KindSoftBreak    = Kind(kindSoftBreak)
	KindHardBreak    = Kind(kindHardBreak)
	KindCode         = Kind(kindCode)
	KindEmphasis     = Kind(kindEmph)
	KindStrong       = Kind(kindStrong)
	KindLink         = Kind(kindLink)
	KindImage        = Kind(kindImage)
	KindHTMLInline   = Kind(kindHTMLInline)
	KindStrike       = Kind(kindStrike)
	KindTaskCheckbox = Kind(kindTaskCheckbox)
	KindReference    = Kind(kindReference)   // an inline [[target]] reference
	KindFootnoteRef  = Kind(kindFootnoteRef) // a footnote reference [^label]
)

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "unknown"
}

// Node is a document node as a validator sees it. It is valid as long as the document is; it does not
// allow anything to be changed.
type Node struct {
	n   *node
	doc *document
}

// Kind returns the kind of the node.
func (n Node) Kind() Kind { return Kind(n.n.kind) }

// Position returns the line and column where the node starts in the source, both from 1; the column is
// counted in bytes.
func (n Node) Position() (line, column int) { return n.doc.position(n.n.pos) }

// Text returns the text of the node: for text, code, a code block, raw HTML and a [[target]] reference
// its content, for a footnote and a footnote reference the label, and for other nodes the text of all
// descendants without markup (for example the text of a heading or the description of an image).
func (n Node) Text() string {
	switch n.n.kind {
	case kindText, kindCode, kindCodeBlock, kindHTMLBlock, kindHTMLInline, kindReference:
		return string(n.n.literal)
	case kindFootnoteDef:
		return n.n.info
	case kindFootnoteRef:
		return n.n.ref.info
	}
	return string(plainText(n.n))
}

// Destination returns the destination of a link or image after backslash escapes and character
// references are resolved, before percent-encoding and regardless of the list of allowed schemes.
func (n Node) Destination() string {
	if n.n.kind == kindLink || n.n.kind == kindImage {
		return n.n.info
	}
	return ""
}

// Title returns the title of a link or image.
func (n Node) Title() string {
	if n.n.kind == kindLink || n.n.kind == kindImage {
		return n.n.title
	}
	return ""
}

// Level returns the level of a heading (1–6) or 0.
func (n Node) Level() int {
	if n.n.kind == kindHeading {
		return int(n.n.level)
	}
	return 0
}

// Info returns the info string of a fenced code block and Language its first word; both after backslash
// escapes and character references are resolved.
func (n Node) Info() string {
	if n.n.kind != kindCodeBlock {
		return ""
	}
	_, info := codeInfo(n.n)
	return info
}

// Language: see Info.
func (n Node) Language() string {
	if n.n.kind != kindCodeBlock {
		return ""
	}
	lang, _ := codeInfo(n.n)
	return lang
}

// Ordered reports whether a list is ordered; Start returns its first number; Tight reports whether it is tight.
func (n Node) Ordered() bool { return n.n.kind == kindList && n.n.list.ordered }

// Start: see Ordered.
func (n Node) Start() int {
	if n.n.kind == kindList && n.n.list.ordered {
		return n.n.list.start
	}
	return 0
}

// Tight: see Ordered.
func (n Node) Tight() bool { return n.n.kind == kindList && n.n.list.tight }

// Checked reports whether a task list checkbox is checked.
func (n Node) Checked() bool { return n.n.kind == kindTaskCheckbox && n.n.level == 1 }

// Header reports whether a table row is the header row.
func (n Node) Header() bool { return n.n.kind == kindTableRow && n.n.level == 0 }

// Column returns the column number of a table cell, from 0, and Align the alignment of that column: "",
// "left", "center" or "right". Padded reports that the cell was not in the source: the row was shorter than the header.
func (n Node) Column() int {
	if n.n.kind == kindTableCell {
		return int(n.n.level)
	}
	return 0
}

// Align: see Column.
func (n Node) Align() string {
	if n.n.kind == kindTableCell {
		return alignNames[n.n.parent.parent.aligns[n.n.level]]
	}
	return ""
}

// Padded: see Column.
func (n Node) Padded() bool { return n.n.kind == kindTableCell && n.n.padded }

// Alert returns the kind of an alert ("note", "tip", "important", "warning", "caution") or "" for an
// ordinary block quote.
func (n Node) Alert() string {
	if n.n.kind == kindBlockQuote {
		return alertNames[n.n.level]
	}
	return ""
}

// Walk visits the nodes of the document in source order. The function is called twice for each node:
// with entering = true before its children and with entering = false after them (also for nodes with no
// children). The value returned on entering says whether to visit the children; on leaving it is ignored.
//
// A paragraph of which only link reference definitions remained is not visited.
func (d *Document) Walk(fn func(n Node, entering bool) bool) {
	root := d.doc.root
	n, entering := root, true
	for {
		visible := n.kind != kindRefDefs
		if entering && visible && fn(Node{n, d.doc}, true) && n.first != nil {
			n = n.first
			continue
		}
		if visible {
			fn(Node{n, d.doc}, false)
		}
		if n == root {
			return
		}
		if n.next != nil {
			n, entering = n.next, true
		} else {
			n, entering = n.parent, false
		}
	}
}

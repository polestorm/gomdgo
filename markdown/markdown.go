package markdown

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

// RawHTML says what to do with raw HTML from the source (HTML blocks and inline tags).
type RawHTML uint8

const (
	// RawHTMLEscape writes raw HTML as text: the reader sees the tags, the browser does not act on them.
	// This is the default.
	RawHTMLEscape RawHTML = iota
	// RawHTMLOmit leaves raw HTML out of the output.
	RawHTMLOmit
	// RawHTMLPass passes raw HTML through unchanged. The output is then only as safe as the source.
	RawHTMLPass
)

// Config holds the settings of a converter. The zero value gives the safe defaults.
type Config struct {
	// RawHTML is the policy for raw HTML; RawHTMLEscape by default.
	RawHTML RawHTML

	// Schemes lists the allowed schemes of link and image destinations, without the colon; case is ignored.
	// nil means the default list (http, https, mailto, tel); an empty list allows only destinations with no
	// scheme. Destinations with no scheme (relative ones, #fragment) are always allowed.
	Schemes []string

	// AnyScheme turns the scheme check off. The output is then only as safe as the source.
	AnyScheme bool

	// Disable lists syntax extensions to turn off. All GitHub Flavored Markdown extensions are on by
	// default (tables, strikethrough, task lists, extended autolinks, the tag filter);
	// Disable: GFM gives plain CommonMark.
	Disable Extension

	// Enable lists extensions from outside the GFM specification to turn on: Alerts and Footnotes.
	Enable Extension

	// AlertTitles gives alert titles by kind ("note", "tip", "important", "warning", "caution").
	// For a kind with a non-empty title the alert starts with a <p class="alert-title"> paragraph. The
	// library has no titles of its own, because it does not know the language of the document.
	AlertTitles map[string]string

	// SingleTilde (compatibility with goldmark and with what GitHub does): text in single tildes, ~like
	// this~, is struck through as well. It has no effect when Strikethrough is disabled. The GFM
	// specification requires two tildes.
	SingleTilde bool

	// TableAlignStyle (compatibility with goldmark): the alignment of a table column is written as
	// style="text-align:…" instead of the align attribute of the GFM specification.
	TableAlignStyle bool

	// CodeBlocks maps a language (the first word of the info string, case-sensitive), for example
	// "mermaid", to a function that renders fenced code blocks in that language. A block in any other
	// language is written the usual way. The result of the function goes into the document unchanged.
	CodeBlocks map[string]CodeBlockFunc

	// References resolves inline [[target]] references. When it is nil, such text is ordinary CommonMark
	// text.
	References ReferenceFunc

	// HeadingIDs adds to each heading an id attribute made from its text, unique within the document.
	HeadingIDs bool
}

// DefaultSchemes are the destination schemes allowed when Config.Schemes is nil.
var DefaultSchemes = []string{"http", "https", "mailto", "tel"}

// Converter converts Markdown to HTML with fixed settings. It is immutable once created and may be used
// from many goroutines at the same time.
type Converter struct {
	cfg     Config
	schemes map[string]bool
	ext     Extension // rozszerzenia przekazywane parserowi, z bitami wewnętrznymi

	codeBlocks  map[string]CodeBlockFunc
	alertTitles [len(alertNames)]string
}

// New creates a converter. An unknown value of Config.RawHTML behaves like RawHTMLEscape.
func New(cfg Config) *Converter {
	c := &Converter{cfg: cfg, schemes: map[string]bool{}, ext: GFM &^ cfg.Disable & GFM}
	if cfg.SingleTilde && c.ext&Strikethrough != 0 {
		c.ext |= extSingleTilde
	}
	list := cfg.Schemes
	if list == nil {
		list = DefaultSchemes
	}
	for _, s := range list {
		c.schemes[strings.ToLower(strings.TrimSuffix(s, ":"))] = true
	}
	c.cfg.Schemes = nil // lista wywołującego nie jest przechowywana; obowiązuje kopia w schemes
	if len(cfg.CodeBlocks) > 0 {
		c.codeBlocks = make(map[string]CodeBlockFunc, len(cfg.CodeBlocks))
		for lang, fn := range cfg.CodeBlocks {
			c.codeBlocks[lang] = fn
		}
	}
	c.cfg.CodeBlocks = nil
	c.ext |= cfg.Enable & (Alerts | Footnotes)
	for i, name := range alertNames {
		c.alertTitles[i] = cfg.AlertTitles[name]
	}
	c.cfg.AlertTitles = nil
	if cfg.References != nil {
		c.ext |= extReferences
	}
	return c
}

var defaultConverter = New(Config{})

// Render converts Markdown source to an HTML fragment with the default settings: CommonMark with the GFM
// extensions. It does not fail for any input: raw HTML becomes text, and links with a scheme that is not
// allowed lose their destination.
func Render(src []byte) []byte {
	out, _ := defaultConverter.Parse(src).HTML() // bez punktów rozszerzeń błąd nie może powstać
	return out
}

// Parse reads Markdown source and returns a document. It does not fail for any input: every sequence of
// bytes is some Markdown document.
func (c *Converter) Parse(src []byte) *Document {
	return &Document{conv: c, doc: parse(sanitize(src), c.ext)}
}

// Render is a shortcut for Parse followed by Document.HTML.
func (c *Converter) Render(src []byte) ([]byte, error) {
	return c.Parse(src).HTML()
}

// Document is a parsed document: the tree and the link reference definitions. It is immutable; HTML may
// be called many times and from many goroutines.
//
// Walk gives validators read-only access to the tree.
type Document struct {
	conv *Converter
	doc  *document
}

// HTML writes the document as an HTML fragment. An error can only come from a caller-supplied function
// plugged into an extension point (Config.CodeBlocks, Config.References) and has the type *Error; without
// such functions it is always nil.
func (d *Document) HTML() ([]byte, error) {
	r := renderer{doc: d.doc, conv: d.conv}
	// Wynik jest zwykle 1,2–1,6 raza dłuższy od źródła; rezerwacja z góry oszczędza kolejne powiększenia bufora.
	r.out.Grow(len(d.doc.src) + len(d.doc.src)/2)
	r.render(d.doc.root)
	r.footnotes()
	if r.err != nil {
		return nil, r.err // wynik jest kompletny albo nie ma go wcale (ADR 0005)
	}
	return r.out.Bytes(), nil
}

// sanitize zwraca źródło, w którym każdy znak U+0000 (wymóg specyfikacji) i każda seria bajtów niebędących
// poprawnym UTF-8 (ADR 0004) są zastąpione znakiem U+FFFD. Poprawne źródło wraca bez kopiowania.
func sanitize(src []byte) []byte {
	if !utf8.Valid(src) {
		src = bytes.ToValidUTF8(src, []byte("�"))
	}
	if bytes.IndexByte(src, 0) >= 0 {
		src = bytes.ReplaceAll(src, []byte{0}, []byte("�"))
	}
	return src
}

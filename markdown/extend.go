package markdown

import (
	"fmt"
	"strings"
)

// Punkty rozszerzeń: miejsca, w których wynik dla fragmentu dokumentu dostarcza funkcja wywołującego
// (ADR 0001 punkt 2, ADR 0009). Są to jedyne miejsca, z których Document.HTML może zwrócić błąd.

// CodeBlock is a fenced code block handed to the caller's function.
type CodeBlock struct {
	Lang string // the first word of the info string, escapes and character references resolved
	Info string // the whole info string, resolved
	Code []byte // the content of the block; a copy the function may modify
	Line int    // the line of the opening fence in the source, from 1
}

// CodeBlockFunc returns an HTML fragment for a code block. The result goes into the document unchanged,
// so the function is responsible for its safety: the guarantees of safe output do not cover it.
type CodeBlockFunc func(CodeBlock) ([]byte, error)

// Reference is an inline [[target]] reference handed to the caller's function.
type Reference struct {
	Target string // the text between the brackets, trimmed
	Line   int    // the position of the reference in the source, from 1
	Column int
}

// ReferenceLink is the answer of the caller's function to an inline reference. The renderer builds a link
// from it in the same way as from a link in the source: the destination is checked against the allowed
// schemes, the text and the title are escaped. The output therefore stays safe.
type ReferenceLink struct {
	Href  string // the destination; empty means the reference is not a link and stays plain text
	Text  string // the link text; empty means the target of the reference
	Title string // an optional title
}

// ReferenceFunc resolves an inline reference.
type ReferenceFunc func(Reference) (ReferenceLink, error)

// Stable error codes, so that the caller can translate the messages.
const (
	// ErrCodeBlock: the function rendering a code block returned an error.
	ErrCodeBlock = "code-block-renderer-failed"
	// ErrReference: the function resolving an inline reference returned an error.
	ErrReference = "reference-resolver-failed"
)

// Error is a rendering error: a stable code, the place in the source and the error of the caller's function.
type Error struct {
	Code   string // one of the Err… constants
	Line   int    // the position in the source, from 1
	Column int
	Detail string // the language of the code block or the target of the reference
	Err    error  // the error returned by the caller's function
}

func (e *Error) Error() string {
	what := "code block renderer for language"
	if e.Code == ErrReference {
		what = "resolver for reference"
	}
	return fmt.Sprintf("gomdgo: line %d, column %d: %s %q failed: %v", e.Line, e.Column, what, e.Detail, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// codeInfo zwraca napis informacyjny bloku kodu po rozwinięciu i jego pierwsze słowo.
func codeInfo(n *node) (lang, info string) {
	info = unescape(n.info)
	lang, _, _ = strings.Cut(strings.ReplaceAll(info, "\t", " "), " ")
	return lang, info
}

// customCodeBlock zapisuje blok kodu przez funkcję wywołującego, jeśli dla jego języka jest zarejestrowana.
func (r *renderer) customCodeBlock(n *node, lang, info string) bool {
	fn := r.conv.codeBlocks[lang]
	if fn == nil || !n.fenced {
		return false
	}
	line, col := r.doc.position(n.pos)
	html, err := fn(CodeBlock{Lang: lang, Info: info, Code: append([]byte(nil), n.literal...), Line: line})
	if err != nil {
		r.err = &Error{Code: ErrCodeBlock, Line: line, Column: col, Detail: lang, Err: err}
		return true
	}
	r.cr()
	r.out.Write(html)
	r.cr()
	return true
}

// reference zapisuje odwołanie w wierszu według odpowiedzi funkcji wywołującego.
func (r *renderer) reference(n *node) {
	line, col := r.doc.position(n.pos)
	target := string(n.literal)
	link, err := r.conv.cfg.References(Reference{Target: target, Line: line, Column: col})
	if err != nil {
		r.err = &Error{Code: ErrReference, Line: line, Column: col, Detail: target, Err: err}
		return
	}
	if link.Text == "" {
		link.Text = target
	}
	if link.Href == "" || !r.conv.allowed(link.Href) {
		escapeHTML(&r.out, []byte(link.Text))
		return
	}
	r.out.WriteString(`<a href="`)
	escapeHTML(&r.out, encodeURL(link.Href))
	r.out.WriteByte('"')
	if link.Title != "" {
		r.out.WriteString(` title="`)
		escapeHTML(&r.out, []byte(link.Title))
		r.out.WriteByte('"')
	}
	r.out.WriteByte('>')
	escapeHTML(&r.out, []byte(link.Text))
	r.out.WriteString("</a>")
}

// scanReference rozpoznaje odwołanie [[cel]] zaczynające się w s[i]: cel jest niepusty, w jednym wierszu
// i bez nawiasów kwadratowych. Zwraca położenie celu i pozycję za odwołaniem albo end = -1.
func scanReference(s []byte, i int) (target span, end int) {
	if i+1 >= len(s) || s[i] != '[' || s[i+1] != '[' {
		return span{}, -1
	}
	j := i + 2
	for j < len(s) && s[j] != '[' && s[j] != ']' && s[j] != '\n' {
		j++
	}
	if j+1 >= len(s) || s[j] != ']' || s[j+1] != ']' {
		return span{}, -1
	}
	a, b := i+2, j
	for a < b && isSpaceOrTab(s[a]) {
		a++
	}
	for b > a && isSpaceOrTab(s[b-1]) {
		b--
	}
	if a == b {
		return span{}, -1
	}
	return span{a, b}, j + 2
}

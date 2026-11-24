package markdown

import (
	"fmt"
	"strings"
)

// Punkty rozszerzeń: miejsca, w których wynik dla fragmentu dokumentu dostarcza funkcja wywołującego
// (ADR 0001 punkt 2, ADR 0009). Są to jedyne miejsca, z których Document.HTML może zwrócić błąd.

// CodeBlock to blok kodu ogrodzonego przekazywany funkcji wywołującego.
type CodeBlock struct {
	Lang string // pierwsze słowo napisu informacyjnego, po rozwinięciu znaków ucieczki i odwołań znakowych
	Info string // cały napis informacyjny, po rozwinięciu
	Code []byte // treść bloku; kopia, którą funkcja może zmieniać
	Line int    // wiersz otwierającego ogrodzenia w źródle, od 1
}

// CodeBlockFunc zwraca fragment HTML dla bloku kodu. Wynik trafia do dokumentu bez zmian, więc za jego
// bezpieczeństwo odpowiada funkcja: gwarancje wyniku bezpiecznego z konstrukcji go nie obejmują.
type CodeBlockFunc func(CodeBlock) ([]byte, error)

// Reference to odwołanie w wierszu [[cel]] przekazywane funkcji wywołującego.
type Reference struct {
	Target string // tekst między nawiasami, bez białych znaków na brzegach
	Line   int    // pozycja odwołania w źródle, od 1
	Column int
}

// ReferenceLink to odpowiedź funkcji wywołującego na odwołanie w wierszu. Renderer buduje z niej odnośnik
// tak samo jak z odnośnika w źródle: adres przechodzi przez listę dozwolonych schematów, tekst i tytuł są
// escapowane. Wynik pozostaje więc bezpieczny z konstrukcji.
type ReferenceLink struct {
	Href  string // adres; pusty oznacza, że odwołanie nie jest odnośnikiem i zostaje samym tekstem
	Text  string // tekst odnośnika; pusty oznacza cel odwołania
	Title string // opcjonalny tytuł
}

// ReferenceFunc rozwiązuje odwołanie w wierszu.
type ReferenceFunc func(Reference) (ReferenceLink, error)

// Stałe kody błędów, do tłumaczenia komunikatów przez wywołującego.
const (
	// ErrCodeBlock: funkcja renderująca blok kodu zwróciła błąd.
	ErrCodeBlock = "code-block-renderer-failed"
	// ErrReference: funkcja rozwiązująca odwołanie w wierszu zwróciła błąd.
	ErrReference = "reference-resolver-failed"
)

// Error to błąd renderowania: stały kod, miejsce w źródle i błąd funkcji wywołującego.
type Error struct {
	Code   string // jedna ze stałych Err…
	Line   int    // pozycja w źródle, od 1
	Column int
	Detail string // język bloku kodu albo cel odwołania
	Err    error  // błąd zwrócony przez funkcję wywołującego
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

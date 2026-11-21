package gomdgo

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

// RawHTML mówi, co zrobić z surowym HTML ze źródła (bloki HTML i znaczniki w wierszu).
type RawHTML uint8

const (
	// RawHTMLEscape zapisuje surowy HTML jako tekst: czytelnik widzi znaczniki, przeglądarka ich nie wykonuje.
	// To ustawienie domyślne.
	RawHTMLEscape RawHTML = iota
	// RawHTMLOmit pomija surowy HTML w wyniku.
	RawHTMLOmit
	// RawHTMLPass przepuszcza surowy HTML bez zmian. Wynik jest wtedy tak bezpieczny jak źródło.
	RawHTMLPass
)

// Config to ustawienia konwertera. Wartość zerowa daje ustawienia domyślne, bezpieczne.
type Config struct {
	// RawHTML: polityka surowego HTML; domyślnie RawHTMLEscape.
	RawHTML RawHTML

	// Schemes: dozwolone schematy adresów odnośników i obrazów, bez dwukropka; wielkość liter nie ma znaczenia.
	// nil oznacza listę domyślną (http, https, mailto, tel); pusta lista dopuszcza tylko adresy bez schematu.
	// Adresy bez schematu (względne, #fragment) są dozwolone zawsze.
	Schemes []string

	// AnyScheme wyłącza sprawdzanie schematów. Wynik jest wtedy tak bezpieczny jak źródło.
	AnyScheme bool

	// Disable: rozszerzenia składni do wyłączenia. Domyślnie włączone są wszystkie rozszerzenia GitHub
	// Flavored Markdown (tabele, przekreślenie, listy zadań, rozszerzone autolinki, filtr znaczników);
	// Disable: GFM daje czysty CommonMark.
	Disable Extension

	// SingleTilde (zgodność z parserem odniesienia i z praktyką GitHub): przekreślenie obejmuje także tekst
	// w pojedynczych tyldach, ~tak~. Bez znaczenia, gdy Strikethrough jest wyłączone. Specyfikacja GFM mówi
	// o dwóch tyldach.
	SingleTilde bool

	// TableAlignStyle (zgodność z parserem odniesienia): wyrównanie kolumny tabeli jest zapisywane jako
	// style="text-align:…" zamiast atrybutu align ze specyfikacji GFM.
	TableAlignStyle bool

	// HeadingIDs dodaje nagłówkom atrybut id utworzony z ich tekstu, unikalny w dokumencie.
	HeadingIDs bool
}

// DefaultSchemes to schematy adresów dozwolone, gdy Config.Schemes jest nil.
var DefaultSchemes = []string{"http", "https", "mailto", "tel"}

// Converter zamienia Markdown na HTML według ustalonych ustawień. Po utworzeniu jest niezmienny i można
// go używać z wielu gorutyn jednocześnie.
type Converter struct {
	cfg     Config
	schemes map[string]bool
	ext     Extension // rozszerzenia przekazywane parserowi, z bitami wewnętrznymi
}

// New tworzy konwerter. Nieznana wartość Config.RawHTML działa jak RawHTMLEscape.
func New(cfg Config) *Converter {
	c := &Converter{cfg: cfg, schemes: map[string]bool{}, ext: GFM &^ cfg.Disable}
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
	return c
}

var defaultConverter = New(Config{})

// Render zamienia źródło Markdown na fragment HTML przy ustawieniach domyślnych: CommonMark z rozszerzeniami
// GFM. Nie zawodzi dla żadnego wejścia: surowy HTML staje się tekstem, a odnośniki o niedozwolonym schemacie
// tracą adres.
func Render(src []byte) []byte {
	out, _ := defaultConverter.Parse(src).HTML() // bez punktów rozszerzeń błąd nie może powstać
	return out
}

// Parse czyta źródło Markdown i zwraca dokument. Nie zawodzi dla żadnego wejścia: każdy ciąg bajtów jest
// jakimś dokumentem Markdown.
func (c *Converter) Parse(src []byte) *Document {
	return &Document{conv: c, doc: parse(sanitize(src), c.ext)}
}

// Render to skrót: Parse, a potem Document.HTML.
func (c *Converter) Render(src []byte) ([]byte, error) {
	return c.Parse(src).HTML()
}

// Document to sparsowany dokument: drzewo i definicje odsyłaczy. Jest niezmienny; HTML można wołać
// wielokrotnie i z wielu gorutyn.
//
// Dostęp do drzewa dla walidatorów dojdzie w etapie 5 (pytanie 3 w docs/adr/open-questions.md).
type Document struct {
	conv *Converter
	doc  *document
}

// HTML zapisuje dokument jako fragment HTML. Błąd może pochodzić wyłącznie z funkcji wywołującego
// podłączonej w punkcie rozszerzenia (etap 5); do tego czasu jest zawsze nil.
func (d *Document) HTML() ([]byte, error) {
	r := renderer{doc: d.doc, conv: d.conv}
	r.render(d.doc.root)
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

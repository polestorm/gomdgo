package gomdgo

import "sort"

// kind to rodzaj węzła drzewa dokumentu.
type kind uint8

const (
	kindDocument kind = iota
	kindBlockQuote
	kindList
	kindItem
	kindParagraph
	kindHeading
	kindThematicBreak
	kindCodeBlock
	kindHTMLBlock
	kindRefDefs // akapit, z którego zostały same definicje odsyłaczy; nie daje wyniku
)

var kindNames = [...]string{
	"document", "block_quote", "list", "item", "paragraph", "heading", "thematic_break", "code_block",
	"html_block", "ref_defs",
}

func (k kind) String() string { return kindNames[k] }

// isContainer mówi, czy blok może zawierać inne bloki.
func (k kind) isContainer() bool { return k <= kindItem }

// span to fragment źródła: przesunięcia bajtowe [start, end).
type span struct{ start, end int }

// listData opisuje listę albo znacznik jej elementu.
type listData struct {
	ordered bool
	delim   byte // znak wypunktowania (- + *) albo ogranicznik numeru (. albo ))
	start   int  // numer pierwszego elementu listy numerowanej
	tight   bool
}

// node to węzeł drzewa dokumentu. Jedna struktura obsługuje wszystkie rodzaje bloków; pola spoza rodzaju
// węzła zostają zerowe.
type node struct {
	kind kind
	open bool // blok przyjmuje jeszcze wiersze
	// pos to przesunięcie bajtowe początku węzła w źródle (ADR 0003); wiersz i kolumnę daje document.position.
	pos                       int
	parent, first, last, next *node
	startLine, endLine        int // pierwszy i ostatni wiersz źródła należący do bloku, od 1

	list          listData // lista i element listy
	contentIndent int      // element listy: szerokość znacznika z odstępem (wcięcie treści)
	markerLine    int      // cytat: ostatni wiersz ze znacznikiem >

	level int    // nagłówek: 1–6
	spans []span // akapit i nagłówek: wiersze treści do parsowania w wierszu

	literal     []byte // kod i blok HTML: treść dosłowna
	trim        int    // kod przez wcięcie: długość treści bez końcowych pustych wierszy
	fenced      bool
	fenceChar   byte
	fenceLen    int
	fenceIndent int
	info        string // kod ogrodzony: napis informacyjny, surowy
	htmlType    int    // blok HTML: warunek początku 1–7
}

func (n *node) appendChild(c *node) {
	c.parent = n
	if n.last == nil {
		n.first = c
	} else {
		n.last.next = c
	}
	n.last = c
}

// reference to definicja odsyłacza; cel i tytuł są surowe (ze znakami ucieczki i odwołaniami znakowymi).
type reference struct {
	dest, title string
	hasTitle    bool
	pos         int
}

// document to wynik parsowania: drzewo, definicje odsyłaczy i źródło.
type document struct {
	src        []byte
	root       *node
	refs       map[string]*reference
	lineStarts []int
}

// position zamienia przesunięcie bajtowe na wiersz i kolumnę, oba od 1; kolumna liczona w bajtach (ADR 0003).
func (d *document) position(offset int) (line, col int) {
	i := sort.Search(len(d.lineStarts), func(i int) bool { return d.lineStarts[i] > offset })
	if i == 0 {
		return 1, offset + 1
	}
	return i, offset - d.lineStarts[i-1] + 1
}

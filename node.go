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

	// treść w wierszu
	kindText
	kindSoftBreak
	kindHardBreak
	kindCode
	kindEmph
	kindStrong
	kindLink
	kindImage
	kindHTMLInline
	kindStrike       // rozszerzenie: przekreślenie
	kindTaskCheckbox // rozszerzenie: pole wyboru na początku akapitu elementu listy; level 1 = zaznaczone
	kindReference    // punkt rozszerzenia: odwołanie w wierszu [[cel]]; literal to cel
	kindFootnoteRef  // rozszerzenie: odwołanie do przypisu; ref wskazuje definicję

	// rozszerzenie: tabela (blok liściowy), jej wiersze i komórki
	kindTable
	kindTableRow  // level 0 = wiersz nagłówka
	kindTableCell // level = numer kolumny

	kindFootnoteDef // rozszerzenie: definicja przypisu (kontener); info to etykieta po normalizacji
)

var kindNames = [...]string{
	"document", "block_quote", "list", "item", "paragraph", "heading", "thematic_break", "code_block",
	"html_block", "ref_defs",
	"text", "soft_break", "hard_break", "code", "emph", "strong", "link", "image", "html_inline",
	"strike", "task_checkbox", "reference", "footnote_ref", "table", "table_row", "table_cell", "footnote",
}

func (k kind) String() string { return kindNames[k] }

// isContainer mówi, czy blok może zawierać inne bloki.
func (k kind) isContainer() bool { return k <= kindItem || k == kindFootnoteDef }

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
	pos                             int
	parent, first, last, prev, next *node
	startLine, endLine              int // pierwszy i ostatni wiersz źródła należący do bloku, od 1

	list          listData // lista i element listy
	contentIndent int      // element listy: szerokość znacznika z odstępem (wcięcie treści)
	markerLine    int      // cytat: ostatni wiersz ze znacznikiem >

	level  int     // nagłówek: 1–6; cytat: rodzaj bloku uwagi; wiersz i komórka tabeli, pole wyboru: patrz rodzaje
	ref    *node   // odwołanie do przypisu: jego definicja
	spans  []span  // akapit, nagłówek i komórka tabeli: wiersze treści do parsowania w wierszu
	aligns []uint8 // tabela: wyrównania kolumn
	padded bool    // komórka tabeli: dopełniona, bo wiersz źródła był krótszy niż nagłówek

	literal     []byte // kod, blok HTML, tekst i HTML w wierszu: treść dosłowna
	trim        int    // kod przez wcięcie: długość treści bez końcowych pustych wierszy
	fenced      bool
	fenceChar   byte
	fenceLen    int
	fenceIndent int
	info        string // kod ogrodzony: napis informacyjny, surowy; odnośnik i obraz: cel po rozwinięciu
	title       string // odnośnik i obraz: tytuł po rozwinięciu
	hasTitle    bool
	htmlType    int // blok HTML: warunek początku 1–7
}

func (n *node) appendChild(c *node) {
	c.parent = n
	c.prev = n.last
	if n.last == nil {
		n.first = c
	} else {
		n.last.next = c
	}
	n.last = c
}

// prependChild wstawia c jako pierwsze dziecko n.
func (n *node) prependChild(c *node) {
	if n.first == nil {
		n.appendChild(c)
		return
	}
	c.parent, c.prev, c.next = n, nil, n.first
	n.first.prev = c
	n.first = c
}

// insertAfter wstawia w jako następne rodzeństwo n.
func (n *node) insertAfter(w *node) {
	w.parent, w.prev, w.next = n.parent, n, n.next
	if n.next != nil {
		n.next.prev = w
	} else {
		n.parent.last = w
	}
	n.next = w
}

// unlink wyjmuje węzeł z drzewa.
func (n *node) unlink() {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		n.parent.first = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		n.parent.last = n.prev
	}
	n.parent, n.prev, n.next = nil, nil, nil
}

// adopt przenosi ciąg rodzeństwa od first do last włącznie pod węzeł w, który nie ma jeszcze dzieci.
func adopt(w, first, last *node) {
	parent := first.parent
	if first.prev != nil {
		first.prev.next = last.next
	} else {
		parent.first = last.next
	}
	if last.next != nil {
		last.next.prev = first.prev
	} else {
		parent.last = first.prev
	}
	first.prev, last.next = nil, nil
	w.first, w.last = first, last
	for c := first; c != nil; c = c.next {
		c.parent = w
	}
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
	ext        Extension
	footnotes  map[string]*node // definicje przypisów według etykiety; obowiązuje pierwsza
}

// position zamienia przesunięcie bajtowe na wiersz i kolumnę, oba od 1; kolumna liczona w bajtach (ADR 0003).
func (d *document) position(offset int) (line, col int) {
	i := sort.Search(len(d.lineStarts), func(i int) bool { return d.lineStarts[i] > offset })
	if i == 0 {
		return 1, offset + 1
	}
	return i, offset - d.lineStarts[i-1] + 1
}

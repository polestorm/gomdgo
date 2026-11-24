package gomdgo

// Publiczny widok drzewa dokumentu dla walidatorów (ADR 0010). Drzewo jest tylko do odczytu: Node to mały
// uchwyt z metodami, a wewnętrzna struktura węzłów pozostaje prywatna.

// Kind to rodzaj węzła dokumentu.
type Kind uint8

// Rodzaje węzłów. Bloki:
const (
	KindDocument      = Kind(kindDocument)
	KindBlockQuote    = Kind(kindBlockQuote) // także blok uwagi: Node.Alert zwraca wtedy jego rodzaj
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
	KindFootnote      = Kind(kindFootnoteDef) // definicja przypisu, w miejscu, w którym stoi w źródle
)

// Treść w wierszu:
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
	KindReference    = Kind(kindReference)   // odwołanie w wierszu [[cel]]
	KindFootnoteRef  = Kind(kindFootnoteRef) // odwołanie do przypisu [^etykieta]
)

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "unknown"
}

// Node to węzeł dokumentu widziany przez walidator. Jest ważny tak długo jak dokument; nie pozwala niczego
// zmienić.
type Node struct {
	n   *node
	doc *document
}

// Kind zwraca rodzaj węzła.
func (n Node) Kind() Kind { return Kind(n.n.kind) }

// Position zwraca wiersz i kolumnę początku węzła w źródle, oba od 1; kolumna jest liczona w bajtach
// (ADR 0003, ADR 0004).
func (n Node) Position() (line, column int) { return n.doc.position(n.n.pos) }

// Text zwraca tekst węzła: dla tekstu, kodu, bloku kodu, surowego HTML i odwołania [[cel]] jego treść,
// dla przypisu i odwołania do przypisu etykietę, a dla pozostałych węzłów tekst wszystkich potomków
// bez znaczników (np. tekst nagłówka albo opis obrazu).
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

// Destination zwraca adres odnośnika albo obrazu po rozwinięciu znaków ucieczki i odwołań znakowych,
// przed kodowaniem procentowym i niezależnie od listy dozwolonych schematów.
func (n Node) Destination() string {
	if n.n.kind == kindLink || n.n.kind == kindImage {
		return n.n.info
	}
	return ""
}

// Title zwraca tytuł odnośnika albo obrazu.
func (n Node) Title() string {
	if n.n.kind == kindLink || n.n.kind == kindImage {
		return n.n.title
	}
	return ""
}

// Level zwraca poziom nagłówka (1–6) albo 0.
func (n Node) Level() int {
	if n.n.kind == kindHeading {
		return int(n.n.level)
	}
	return 0
}

// Info zwraca napis informacyjny bloku kodu ogrodzonego, a Language jego pierwsze słowo; oba po rozwinięciu
// znaków ucieczki i odwołań znakowych.
func (n Node) Info() string {
	if n.n.kind != kindCodeBlock {
		return ""
	}
	_, info := codeInfo(n.n)
	return info
}

// Language: patrz Info.
func (n Node) Language() string {
	if n.n.kind != kindCodeBlock {
		return ""
	}
	lang, _ := codeInfo(n.n)
	return lang
}

// Ordered mówi, czy lista jest numerowana; Start zwraca jej numer początkowy; Tight mówi, czy jest zwarta.
func (n Node) Ordered() bool { return n.n.kind == kindList && n.n.list.ordered }

// Start: patrz Ordered.
func (n Node) Start() int {
	if n.n.kind == kindList && n.n.list.ordered {
		return n.n.list.start
	}
	return 0
}

// Tight: patrz Ordered.
func (n Node) Tight() bool { return n.n.kind == kindList && n.n.list.tight }

// Checked mówi, czy pole wyboru listy zadań jest zaznaczone.
func (n Node) Checked() bool { return n.n.kind == kindTaskCheckbox && n.n.level == 1 }

// Header mówi, czy wiersz tabeli jest wierszem nagłówka.
func (n Node) Header() bool { return n.n.kind == kindTableRow && n.n.level == 0 }

// Column zwraca numer kolumny komórki tabeli, od 0, a Align wyrównanie tej kolumny: "", "left", "center"
// albo "right". Padded mówi, że komórki nie było w źródle: wiersz był krótszy niż nagłówek.
func (n Node) Column() int {
	if n.n.kind == kindTableCell {
		return int(n.n.level)
	}
	return 0
}

// Align: patrz Column.
func (n Node) Align() string {
	if n.n.kind == kindTableCell {
		return alignNames[n.n.parent.parent.aligns[n.n.level]]
	}
	return ""
}

// Padded: patrz Column.
func (n Node) Padded() bool { return n.n.kind == kindTableCell && n.n.padded }

// Alert zwraca rodzaj bloku uwagi ("note", "tip", "important", "warning", "caution") albo "" dla zwykłego
// cytatu.
func (n Node) Alert() string {
	if n.n.kind == kindBlockQuote {
		return alertNames[n.n.level]
	}
	return ""
}

// Walk odwiedza węzły dokumentu w kolejności źródła. Dla każdego węzła funkcja jest wołana dwa razy:
// z entering = true przed jego dziećmi i z entering = false po nich (także dla węzłów bez dzieci).
// Wartość zwrócona przy wejściu mówi, czy odwiedzić dzieci; przy wyjściu jest pomijana.
//
// Akapit, z którego zostały same definicje odsyłaczy, nie jest odwiedzany.
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

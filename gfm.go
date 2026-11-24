package gomdgo

import (
	"bytes"
	"strings"
)

// Rozszerzenia GFM 0.29: tabele, listy zadań, rozszerzone autolinki i filtr znaczników. Przekreślenie jest
// w inline.go, bo korzysta ze stosu ograniczników.

// Extension to zbiór rozszerzeń składni. Rozszerzenia GFM są domyślnie włączone; Config.Disable wyłącza
// wskazane.
type Extension uint16

const (
	// Tables: tabele z wierszem nagłówka i wierszem ograniczników.
	Tables Extension = 1 << iota
	// Strikethrough: przekreślenie ~~tekst~~.
	Strikethrough
	// TaskLists: elementy list zaczynające się od [ ] albo [x].
	TaskLists
	// Autolinks: adresy www., http://, https://, ftp:// i adresy e-mail rozpoznawane w zwykłym tekście.
	Autolinks
	// TagFilter: przy przepuszczaniu surowego HTML dziewięć znaczników zmieniających sposób czytania
	// dokumentu przez przeglądarkę (m.in. script, style, iframe) jest zapisywanych jako tekst.
	TagFilter

	// GFM to wszystkie rozszerzenia specyfikacji GitHub Flavored Markdown.
	GFM = Tables | Strikethrough | TaskLists | Autolinks | TagFilter

	// Alerts: bloki uwag, cytat zaczynający się od [!NOTE], [!TIP], [!IMPORTANT], [!WARNING] albo
	// [!CAUTION]. Spoza specyfikacji GFM; włączane przez Config.Enable.
	Alerts Extension = 1 << 5
	// Footnotes: przypisy [^etykieta] z definicjami [^etykieta]: treść. Spoza specyfikacji GFM; włączane
	// przez Config.Enable.
	Footnotes Extension = 1 << 6

	// Bity wewnętrzne, ustawiane przez Config.SingleTilde i Config.References.
	extSingleTilde Extension = 1 << 15
	extReferences  Extension = 1 << 14
)

// maxPaddedCells ogranicza liczbę komórek, którymi w jednym dokumencie dopełnia się wiersze tabel krótsze
// niż nagłówek (ADR 0008). Bez limitu tabela o wielu kolumnach i wielu krótkich wierszach dawałaby wynik
// o rozmiarze kolumny × wiersze. Po wyczerpaniu limitu krótsze wiersze zostają krótsze.
const maxPaddedCells = 100_000

// Wyrównanie kolumny tabeli.
const (
	alignNone = iota
	alignLeft
	alignCenter
	alignRight
)

var alignNames = [...]string{"", "left", "center", "right"}

// splitTableRow dzieli wiersz tabeli na komórki: zwraca położenia treści komórek w s, bez białych znaków
// na brzegach. Znak | poprzedzony ukośnikiem nie dzieli; skrajne znaki | są opcjonalne.
func splitTableRow(s []byte) []span {
	start, end := 0, len(s)
	for start < end && isSpaceOrTab(s[start]) {
		start++
	}
	for end > start && isSpaceOrTab(s[end-1]) {
		end--
	}
	if start < end && s[start] == '|' {
		start++
	}
	var cells []span
	cellStart := start
	for i := start; i <= end; i++ {
		if i < end {
			if s[i] == '\\' {
				i++
				continue
			}
			if s[i] != '|' {
				continue
			}
		} else if cellStart >= end {
			break // wiersz kończy się znakiem |: po nim nie ma komórki
		}
		a, b := cellStart, min(i, end)
		for a < b && isSpaceOrTab(s[a]) {
			a++
		}
		for b > a && isSpaceOrTab(s[b-1]) {
			b--
		}
		cells = append(cells, span{a, b})
		cellStart = i + 1
	}
	return cells
}

// tableDelimiterRow rozpoznaje wiersz ograniczników tabeli i zwraca wyrównania kolumn albo nil.
func tableDelimiterRow(s []byte) []uint8 {
	for _, c := range s {
		if c != '|' && c != '-' && c != ':' && c != ' ' && c != '\t' {
			return nil
		}
	}
	cells := splitTableRow(s)
	if len(cells) == 0 {
		return nil
	}
	aligns := make([]uint8, len(cells))
	for k, cell := range cells {
		c := s[cell.start:cell.end]
		left := len(c) > 0 && c[0] == ':'
		right := len(c) > 1 && c[len(c)-1] == ':'
		c = bytes.TrimSuffix(bytes.TrimPrefix(c, []byte(":")), []byte(":"))
		if len(c) == 0 || len(bytes.Trim(c, "-")) != 0 {
			return nil
		}
		switch {
		case left && right:
			aligns[k] = alignCenter
		case left:
			aligns[k] = alignLeft
		case right:
			aligns[k] = alignRight
		}
	}
	return aligns
}

// addTableRow dopisuje do tabeli wiersz z komórkami z podanego fragmentu źródła; level 0 to nagłówek.
// Komórek jest tyle, ile kolumn: brakujące są puste (do limitu maxPaddedCells), nadmiarowe pomijane.
func (p *parser) addTableRow(table *node, row span, level int) {
	src := p.doc.src
	r := newBlock(kindTableRow, row.start, p.lineNo)
	r.level = int32(level)
	table.appendChild(r)
	cells := splitTableRow(src[row.start:row.end])
	for col := range table.aligns {
		if col >= len(cells) {
			if p.padded >= maxPaddedCells {
				break
			}
			p.padded++
		}
		cell := newBlock(kindTableCell, row.end, p.lineNo)
		cell.level = int32(col)
		cell.padded = col >= len(cells)
		if col < len(cells) {
			cell.pos = row.start + cells[col].start
			if cells[col].end > cells[col].start {
				cell.spans = []span{{row.start + cells[col].start, row.start + cells[col].end}}
			}
		}
		r.appendChild(cell)
	}
}

// taskMarker rozpoznaje znacznik zadania na początku pierwszego akapitu elementu listy: [ ], [x] albo [X],
// a po nim biały znak. Zwraca 0 (brak), 1 (niezaznaczone) albo 2 (zaznaczone).
func taskMarker(src []byte, n *node) int {
	if n.parent.kind != kindItem || n.prev != nil || len(n.spans) == 0 {
		return 0
	}
	s := src[n.spans[0].start:n.spans[0].end]
	if len(s) < 3 || s[0] != '[' || s[2] != ']' {
		return 0
	}
	if len(s) == 3 && len(n.spans) == 1 || len(s) > 3 && !isSpaceOrTab(s[3]) {
		return 0
	}
	switch s[1] {
	case ' ', '\t':
		return 1
	case 'x', 'X':
		return 2
	}
	return 0
}

// Znaczniki zapisywane jako tekst przez filtr znaczników.
var filteredTags = map[string]bool{
	"title": true, "textarea": true, "style": true, "xmp": true, "iframe": true, "noembed": true,
	"noframes": true, "script": true, "plaintext": true,
}

// writeTagFiltered zapisuje surowy HTML, zamieniając < na &lt; w znacznikach z listy filtra.
func writeTagFiltered(out *bytes.Buffer, s []byte) {
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] != '<' {
			continue
		}
		j := i + 1
		if j < len(s) && s[j] == '/' {
			j++
		}
		k := j
		for k < len(s) && k-j < 10 && isLetter(s[k]) {
			k++
		}
		if k == j || k < len(s) && (isLetter(s[k]) || isDigit(s[k]) || s[k] == '-') {
			continue
		}
		if filteredTags[strings.ToLower(string(s[j:k]))] {
			out.Write(s[start:i])
			out.WriteString("&lt;")
			start = i + 1
		}
	}
	out.Write(s[start:])
}

// Rozszerzone autolinki (GFM 6.9) są rozpoznawane po zakończeniu parsowania treści w wierszu, w węzłach
// zwykłego tekstu poza odnośnikami. Sąsiednie węzły tekstu są najpierw łączone, bo adres może zawierać
// znaki, które parser zapisał osobno (np. podkreślenie, które nie zostało wyróżnieniem).
func (p *inlineParser) autolinks() {
	root := p.parent
	n := root.first
	if n == nil {
		return
	}
	for {
		if n.kind == kindText {
			mergeTextRun(n)
			if bytes.IndexByte(n.literal, '.') >= 0 { // każdy rozszerzony autolink zawiera kropkę
				n = linkify(n)
			}
		}
		if n.first != nil && n.kind != kindLink && n.kind != kindImage {
			n = n.first
			continue
		}
		for n.next == nil {
			if n = n.parent; n == root {
				return
			}
		}
		n = n.next
	}
}

// mergeTextRun łączy węzeł tekstu z następującymi bezpośrednio po nim węzłami tekstu.
func mergeTextRun(n *node) {
	if n.next == nil || n.next.kind != kindText {
		return
	}
	size := 0
	for m := n; m != nil && m.kind == kindText; m = m.next {
		size += len(m.literal)
	}
	buf := make([]byte, 0, size)
	buf = append(buf, n.literal...)
	for n.next != nil && n.next.kind == kindText {
		buf = append(buf, n.next.literal...)
		n.next.unlink()
	}
	n.literal = buf
}

// linkify zamienia adresy w węźle tekstu na odnośniki, dzieląc węzeł; zwraca ostatni z powstałych węzłów.
// Pozycje nowych węzłów są liczone od początku węzła tekstu; po połączeniu węzłów albo po znaku ucieczki
// są przybliżone w obrębie wiersza.
func linkify(n *node) *node {
	text, base := n.literal, n.pos
	cur, last := n, 0 // cur to węzeł z tekstem text[last:]
	for j := 0; j < len(text); j++ {
		a, b, dest := 0, -1, ""
		switch text[j] {
		case 'w':
			if bytes.HasPrefix(text[j:], []byte("www.")) && autolinkBoundary(text, j, n) {
				if a, b = j, urlEnd(text, j, j); b > 0 {
					dest = "http://" + string(text[a:b])
				}
			}
		case ':':
			if !bytes.HasPrefix(text[j:], []byte("://")) {
				continue
			}
			for _, scheme := range []string{"https", "http", "ftp"} {
				if s := j - len(scheme); s >= last && string(text[s:j]) == scheme && autolinkBoundary(text, s, n) {
					if a, b = s, urlEnd(text, s, j+3); b > 0 {
						dest = string(text[a:b])
					}
					break
				}
			}
		case '@':
			if a, b = emailAt(text, j, last); b > 0 {
				dest = "mailto:" + string(text[a:b])
			}
		default:
			continue
		}
		if b <= a {
			continue
		}
		link := &node{kind: kindLink, pos: base + a, info: dest}
		link.appendChild(&node{kind: kindText, pos: base + a, literal: text[a:b]})
		cur.insertAfter(link)
		if a > last {
			cur.literal = text[last:a]
		} else {
			cur.unlink()
		}
		if b == len(text) {
			return link
		}
		cur = &node{kind: kindText, pos: base + b, literal: text[b:]}
		link.insertAfter(cur)
		last, j = b, b-1
	}
	return cur
}

// autolinkBoundary: autolink może zacząć się na początku wiersza, po białym znaku albo po jednym ze znaków
// * _ ~ (. Na początku węzła tekstu rozstrzyga to, co go poprzedza w drzewie.
func autolinkBoundary(text []byte, j int, n *node) bool {
	if j > 0 {
		c := text[j-1]
		return c == ' ' || c == '\t' || c == '\n' || c == '*' || c == '_' || c == '~' || c == '('
	}
	if n.prev == nil {
		return true
	}
	switch n.prev.kind {
	case kindSoftBreak, kindHardBreak, kindEmph, kindStrong, kindStrike:
		return true
	}
	return false
}

func isDomainByte(c byte) bool {
	return isLetter(c) || isDigit(c) || c == '_' || c == '-' || c == '.'
}

// validDomain: segmenty rozdzielone kropkami, co najmniej jedna kropka, bez podkreśleń w dwóch ostatnich
// segmentach. Kropki na końcu nie należą do domeny (są interpunkcją zdania).
func validDomain(s []byte) bool {
	s = bytes.TrimRight(s, ".")
	last := bytes.LastIndexByte(s, '.')
	if last < 0 || last == len(s)-1 {
		return false
	}
	prev := bytes.LastIndexByte(s[:last], '.')
	return bytes.IndexByte(s[prev+1:], '_') < 0
}

// urlEnd wyznacza koniec rozszerzonego autolinku zaczynającego się w text[start], którego domena zaczyna
// się w text[domain]; zwraca -1, gdy domena jest niepoprawna.
func urlEnd(text []byte, start, domain int) int {
	d := domain
	for d < len(text) && isDomainByte(text[d]) {
		d++
	}
	if !validDomain(text[domain:d]) {
		return -1
	}
	e := d
	opens, closes := 0, 0
	for ; e < len(text); e++ {
		c := text[e]
		if c == ' ' || c == '\t' || c == '\n' || c == '<' {
			break
		}
		if c == '(' {
			opens++
		} else if c == ')' {
			closes++
		}
	}
	// Interpunkcja na końcu nie należy do adresu; nawias zamykający tylko wtedy, gdy ma parę w adresie;
	// końcówka wyglądająca jak odwołanie znakowe też nie.
trim:
	for e > domain {
		switch c := text[e-1]; c {
		case '?', '!', '.', ',', ':', '*', '_', '~':
			e--
		case ')':
			if closes <= opens {
				break trim
			}
			closes--
			e--
		case ';':
			k := e - 2
			for k > domain && (isLetter(text[k]) || isDigit(text[k])) {
				k--
			}
			if k == e-2 || text[k] != '&' {
				break trim
			}
			e = k
		default:
			break trim
		}
	}
	if e < d && !validDomain(text[domain:e]) {
		return -1
	}
	return e
}

// emailAt rozpoznaje adres e-mail wokół znaku @ w text[at]; część lokalna nie sięga przed pozycję limit.
func emailAt(text []byte, at, limit int) (a, b int) {
	a = at
	for a > limit && (isLetter(text[a-1]) || isDigit(text[a-1]) || bytes.IndexByte([]byte(".-_+"), text[a-1]) >= 0) {
		a--
	}
	if a == at {
		return 0, -1
	}
	b = at + 1
	dots := 0
	for b < len(text) && isDomainByte(text[b]) {
		if text[b] == '.' {
			dots++
		}
		b++
	}
	for b > at+1 && text[b-1] == '.' {
		b--
		dots--
	}
	if b == at+1 || dots == 0 || text[b-1] == '-' || text[b-1] == '_' {
		return 0, -1
	}
	return a, b
}

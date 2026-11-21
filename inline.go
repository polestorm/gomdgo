package gomdgo

import (
	"bytes"
	"sort"
	"unicode"
	"unicode/utf8"
)

// Parser treści w wierszu: druga faza strategii opisanej w dodatku do specyfikacji CommonMark. Tekst akapitu
// albo nagłówka jest czytany raz, od lewej do prawej. Ograniczniki wyróżnień trafiają na stos i są łączone
// w pary dopiero na końcu albo przy zamknięciu odnośnika; nawiasy kwadratowe mają osobny stos.
//
// Czas liniowy: każde wyszukiwanie w przód, które może sięgnąć końca tekstu i zawieść, zapisuje ten fakt,
// więc nie jest powtarzane (zamknięcie kodu w wierszu, komentarza, cudzysłowu w znaczniku).
type inlineParser struct {
	doc      *document
	ext      Extension
	parent   *node
	text     []byte
	spans    []span
	lineOffs []int // dla tekstu z kilku wierszy: początek każdego wiersza w text
	removed  []int // dla komórki tabeli: pozycje w text, przed którymi usunięto ukośnik ze źródła

	delimTop *delim
	brackets []bracket
	// inactiveDepth: otwarcia odnośników (nie obrazów) leżące na stosie poniżej tej głębokości są nieaktywne,
	// bo odnośnik nie może zawierać odnośnika.
	inactiveDepth int

	btScanned bool        // tekst za bieżącą pozycją został przejrzany do końca w poszukiwaniu serii `
	btLast    map[int]int // ostatnia znana pozycja serii ` o danej długości
	noMatch   [searchKinds]bool
}

// Rodzaje wyszukiwań w przód, których niepowodzenie jest zapamiętywane.
const (
	searchComment = iota
	searchInstruction
	searchDeclaration
	searchCDATA
	searchDoubleQuote
	searchSingleQuote
	searchKinds
)

// delim to seria ograniczników wyróżnienia na stosie.
type delim struct {
	node              *node // węzeł tekstu z niezużytymi znakami serii
	ch                byte
	count, orig       int
	canOpen, canClose bool
	prev, next        *delim
}

// bracket to otwarcie odnośnika [ albo obrazu ![ na stosie.
type bracket struct {
	node        *node
	image       bool
	textPos     int    // pozycja w text tuż za nawiasem [
	delimBottom *delim // szczyt stosu ograniczników w chwili otwarcia
}

var inlineSpecial = func() (t [256]bool) {
	for _, c := range "\n\\`*_[]!<&~" {
		t[c] = true
	}
	return
}()

// parse zamienia wiersze treści węzła na jego dzieci.
func (p *inlineParser) parse(n *node) {
	src := p.doc.src
	p.parent, p.spans = n, n.spans
	p.lineOffs, p.removed = p.lineOffs[:0], p.removed[:0]
	if len(n.spans) == 0 {
		return // pusta komórka tabeli
	}
	if len(n.spans) == 1 {
		p.text = src[n.spans[0].start:n.spans[0].end]
		if n.kind == kindTableCell && bytes.Contains(p.text, []byte(`\|`)) {
			// W komórce tabeli \| oznacza znak |, także wewnątrz kodu w wierszu.
			text := make([]byte, 0, len(p.text))
			for i := 0; i < len(p.text); i++ {
				if p.text[i] == '\\' && i+1 < len(p.text) && p.text[i+1] == '|' {
					p.removed = append(p.removed, len(text))
					continue
				}
				text = append(text, p.text[i])
			}
			p.text = text
		}
	} else {
		var text []byte
		for i, s := range n.spans {
			if i > 0 {
				text = append(text, '\n')
			}
			p.lineOffs = append(p.lineOffs, len(text))
			text = append(text, src[s.start:s.end]...)
		}
		p.text = text
	}
	p.text = bytes.TrimRight(p.text, " \t")
	p.delimTop, p.brackets, p.inactiveDepth = nil, p.brackets[:0], 0
	p.btScanned, p.noMatch = false, [searchKinds]bool{}
	clear(p.btLast)

	text := p.text
	start := 0 // początek zwykłego tekstu jeszcze niezapisanego jako węzeł
	flush := func(end int) {
		if end > start {
			p.add(kindText, start).literal = text[start:end]
		}
	}
	for i := 0; i < len(text); {
		c := text[i]
		if !inlineSpecial[c] {
			i++
			continue
		}
		switch c {
		case '\n':
			end := i
			for end > start && text[end-1] == ' ' {
				end--
			}
			flush(end)
			if i-end >= 2 {
				p.add(kindHardBreak, end)
			} else {
				p.add(kindSoftBreak, i)
			}
			i++
			start = i

		case '\\':
			switch {
			case i+1 < len(text) && isPunct(text[i+1]):
				flush(i)
				start = i + 1 // znak po ukośniku zaczyna następny kawałek zwykłego tekstu
				i += 2
			case i+1 < len(text) && text[i+1] == '\n':
				flush(i)
				p.add(kindHardBreak, i)
				i += 2
				start = i
			default:
				i++
			}

		case '`':
			n := 1
			for i+n < len(text) && text[i+n] == '`' {
				n++
			}
			closer := p.backtickCloser(i+n, n)
			if closer < 0 {
				i += n
				continue
			}
			flush(i)
			p.add(kindCode, i).literal = codeSpanContent(text[i+n : closer])
			i = closer + n
			start = i

		case '~':
			n := 1
			for i+n < len(text) && text[i+n] == c {
				n++
			}
			// Przekreślenie to dokładnie dwie tyldy; z Config.SingleTilde także jedna.
			if p.ext&Strikethrough == 0 || n > 2 || n == 1 && p.ext&extSingleTilde == 0 {
				i += n
				continue
			}
			flush(i)
			p.pushDelim(i, n)
			i += n
			start = i

		case '*', '_':
			flush(i)
			n := 1
			for i+n < len(text) && text[i+n] == c {
				n++
			}
			p.pushDelim(i, n)
			i += n
			start = i

		case '[':
			if p.ext&Footnotes != 0 {
				if label, end := footnoteLabel(text[i:]); end > 0 {
					if def := p.doc.footnotes[normalizeLabel(label)]; def != nil {
						flush(i)
						p.add(kindFootnoteRef, i).ref = def
						i += end
						start = i
						continue
					}
				}
			}
			if p.ext&extReferences != 0 {
				if target, end := scanReference(text, i); end > 0 {
					flush(i)
					p.add(kindReference, i).literal = text[target.start:target.end]
					i = end
					start = i
					continue
				}
			}
			flush(i)
			node := p.add(kindText, i)
			node.literal = text[i : i+1]
			p.brackets = append(p.brackets, bracket{node: node, textPos: i + 1, delimBottom: p.delimTop})
			i++
			start = i

		case '!':
			if i+1 < len(text) && text[i+1] == '[' {
				flush(i)
				node := p.add(kindText, i)
				node.literal = text[i : i+2]
				p.brackets = append(p.brackets, bracket{node: node, image: true, textPos: i + 2, delimBottom: p.delimTop})
				i += 2
				start = i
			} else {
				i++
			}

		case ']':
			flush(i)
			start = i
			if end := p.closeBracket(i); end > 0 {
				i = end
				start = i
			} else {
				i++
			}

		case '<':
			if dest, end := scanAutolink(text, i); end > 0 {
				flush(i)
				link := p.add(kindLink, i)
				link.info = dest
				label := &node{kind: kindText, pos: p.pos(i + 1), literal: text[i+1 : end-1]}
				link.appendChild(label)
				i = end
				start = i
			} else if end := p.inlineHTML(i); end > 0 {
				flush(i)
				p.add(kindHTMLInline, i).literal = text[i:end]
				i = end
				start = i
			} else {
				i++
			}

		case '&':
			if decoded, n := entity(string(text[i:min(i+34, len(text))])); n > 0 {
				flush(i)
				p.add(kindText, i).literal = []byte(decoded)
				i += n
				start = i
			} else {
				i++
			}
		}
	}
	flush(len(text))
	p.processEmphasis(nil)
	if p.ext&Autolinks != 0 {
		p.autolinks()
	}
}

// pos zamienia pozycję w tekście akapitu na przesunięcie w źródle.
func (p *inlineParser) pos(off int) int {
	if len(p.spans) == 1 {
		if len(p.removed) > 0 {
			off += sort.SearchInts(p.removed, off+1)
		}
		return p.spans[0].start + off
	}
	k := sort.Search(len(p.lineOffs), func(k int) bool { return p.lineOffs[k] > off }) - 1
	return p.spans[k].start + off - p.lineOffs[k]
}

func (p *inlineParser) add(k kind, off int) *node {
	n := &node{kind: k, pos: p.pos(off)}
	p.parent.appendChild(n)
	return n
}

// backtickCloser szuka od pozycji from serii dokładnie n znaków ` i zwraca jej początek albo -1.
func (p *inlineParser) backtickCloser(from, n int) int {
	if p.btScanned {
		if last, ok := p.btLast[n]; !ok || last < from {
			return -1
		}
	}
	text := p.text
	for j := from; j < len(text); {
		k := bytes.IndexByte(text[j:], '`')
		if k < 0 {
			break
		}
		j += k
		m := 1
		for j+m < len(text) && text[j+m] == '`' {
			m++
		}
		if m == n {
			return j
		}
		if p.btLast == nil {
			p.btLast = map[int]int{}
		}
		p.btLast[m] = j
		j += m
	}
	p.btScanned = true
	return -1
}

// codeSpanContent: końce wierszy stają się spacjami; jedna spacja z każdej strony jest zdejmowana, gdy są
// obie, a treść nie składa się z samych spacji.
func codeSpanContent(s []byte) []byte {
	if bytes.IndexByte(s, '\n') >= 0 {
		s = bytes.ReplaceAll(s, []byte{'\n'}, []byte{' '})
	}
	if len(s) >= 2 && s[0] == ' ' && s[len(s)-1] == ' ' && len(bytes.Trim(s, " ")) > 0 {
		s = s[1 : len(s)-1]
	}
	return s
}

// charClass: 0 biały znak (także początek i koniec tekstu), 1 interpunkcja, 2 pozostałe.
func charClass(r rune, size int) int {
	switch {
	case size == 0 || r == ' ' || r == '\t' || r == '\n' || r == '\f' || r == '\r':
		return 0
	case r < utf8.RuneSelf:
		if isPunct(byte(r)) {
			return 1
		}
		return 2
	case unicode.Is(unicode.Zs, r):
		return 0
	case unicode.IsPunct(r) || unicode.IsSymbol(r):
		return 1
	}
	return 2
}

// pushDelim zapisuje serię n ograniczników zaczynającą się w text[i] jako tekst i odkłada ją na stos.
func (p *inlineParser) pushDelim(i, n int) {
	text := p.text
	before := charClass(utf8.DecodeLastRune(text[:i]))
	after := charClass(utf8.DecodeRune(text[i+n:]))
	left := after != 0 && (after != 1 || before != 2)
	right := before != 0 && (before != 1 || after != 2)
	d := &delim{ch: text[i], count: n, orig: n, prev: p.delimTop}
	if d.ch != '_' {
		d.canOpen, d.canClose = left, right
	} else {
		d.canOpen = left && (!right || before == 1)
		d.canClose = right && (!left || after == 1)
	}
	d.node = p.add(kindText, i)
	d.node.literal = text[i : i+n]
	if !d.canOpen && !d.canClose {
		return
	}
	if p.delimTop != nil {
		p.delimTop.next = d
	}
	p.delimTop = d
}

func (p *inlineParser) removeDelim(d *delim) {
	if d.prev != nil {
		d.prev.next = d.next
	}
	if d.next != nil {
		d.next.prev = d.prev
	} else {
		p.delimTop = d.prev
	}
}

// processEmphasis łączy w pary ograniczniki leżące na stosie powyżej bottom (procedura „process emphasis”
// ze specyfikacji) i zdejmuje je ze stosu.
func (p *inlineParser) processEmphasis(bottom *delim) {
	// Dolne granice szukania otwarcia, osobno dla znaku, zdolności zamknięcia do otwierania i długości
	// serii modulo 3: poniżej granicy na pewno nie ma pasującego otwarcia.
	var openersBottom [3][2][3]*delim
	for a := range openersBottom {
		for b := range openersBottom[a] {
			for c := range openersBottom[a][b] {
				openersBottom[a][b][c] = bottom
			}
		}
	}
	closer := p.delimTop
	if closer == bottom {
		return
	}
	for closer.prev != bottom {
		closer = closer.prev
	}
	for closer != nil {
		if !closer.canClose {
			closer = closer.next
			continue
		}
		a, b := 0, 0
		switch closer.ch {
		case '_':
			a = 1
		case '~':
			a = 2
		}
		if closer.canOpen {
			b = 1
		}
		limit := &openersBottom[a][b][closer.orig%3]
		opener := closer.prev
		for ; opener != bottom && opener != *limit; opener = opener.prev {
			if opener.ch != closer.ch || !opener.canOpen {
				continue
			}
			// Reguła „sumy podzielnej przez 3” dla serii, które mogą i otwierać, i zamykać.
			if closer.ch != '~' && (opener.canClose || closer.canOpen) && (opener.orig+closer.orig)%3 == 0 &&
				(opener.orig%3 != 0 || closer.orig%3 != 0) {
				continue
			}
			break
		}
		if opener == bottom || opener == *limit {
			*limit = closer.prev
			next := closer.next
			if !closer.canOpen {
				p.removeDelim(closer)
			}
			closer = next
			continue
		}

		use, k := 1, kindEmph
		if opener.count >= 2 && closer.count >= 2 {
			use, k = 2, kindStrong
		}
		if closer.ch == '~' {
			k = kindStrike // jedna tylda (Config.SingleTilde) i dwie dają to samo przekreślenie
		}
		opener.count -= use
		closer.count -= use
		opener.node.literal = opener.node.literal[:opener.count]
		closer.node.literal = closer.node.literal[use:]
		closer.node.pos += use
		w := &node{kind: k, pos: opener.node.pos + opener.count}
		if opener.node.next != closer.node {
			adopt(w, opener.node.next, closer.node.prev)
		}
		opener.node.insertAfter(w)
		// Ograniczniki między otwarciem a zamknięciem nie mają już z kim się połączyć.
		opener.next = closer
		closer.prev = opener
		if opener.count == 0 {
			opener.node.unlink()
			p.removeDelim(opener)
		}
		if closer.count == 0 {
			closer.node.unlink()
			next := closer.next
			p.removeDelim(closer)
			closer = next
		}
	}
	p.delimTop = bottom
	if bottom != nil {
		bottom.next = nil
	}
}

// closeBracket obsługuje ] w text[i]: jeśli z ostatnim otwarciem tworzy odnośnik albo obraz, buduje węzeł
// i zwraca pozycję za całą konstrukcją; inaczej zwraca -1, a ] zostaje zwykłym tekstem.
func (p *inlineParser) closeBracket(i int) int {
	depth := len(p.brackets) - 1
	if depth < 0 {
		return -1
	}
	b := p.brackets[depth]
	pop := func() {
		p.brackets = p.brackets[:depth]
		p.inactiveDepth = min(p.inactiveDepth, depth)
	}
	if !b.image && depth < p.inactiveDepth {
		pop()
		return -1
	}
	text := p.text
	after := i + 1
	dest, title, hasTitle, end := p.inlineLinkTail(after)
	if end < 0 {
		var label []byte
		if labelEnd := scanLinkLabel(text, after); labelEnd > 0 {
			label, end = text[after+1:labelEnd-1], labelEnd // odsyłacz pełny: [tekst][etykieta]
		} else if scanLinkLabel(text, b.textPos-1) == after {
			label, end = text[b.textPos:i], after // skrócony: [etykieta]
			if bytes.HasPrefix(text[after:], []byte("[]")) {
				end = after + 2 // zwinięty: [etykieta][]
			}
		}
		ref := p.doc.refs[normalizeLabel(label)]
		if label == nil || ref == nil {
			pop()
			return -1
		}
		dest, title, hasTitle = ref.dest, ref.title, ref.hasTitle
	}

	p.processEmphasis(b.delimBottom)
	w := &node{kind: kindLink, pos: b.node.pos, info: unescape(dest), hasTitle: hasTitle}
	if b.image {
		w.kind = kindImage
	}
	if hasTitle {
		w.title = unescape(title)
	}
	if b.node.next != nil {
		adopt(w, b.node.next, p.parent.last)
	}
	b.node.insertAfter(w)
	b.node.unlink()
	pop()
	if !b.image {
		p.inactiveDepth = depth
	}
	return end
}

func skipLinkSpace(s []byte, i int) int {
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n') {
		i++
	}
	return i
}

// inlineLinkTail rozpoznaje ogon odnośnika w wierszu, (cel "tytuł"), zaczynający się w text[i].
func (p *inlineParser) inlineLinkTail(i int) (dest, title string, hasTitle bool, end int) {
	text := p.text
	if i >= len(text) || text[i] != '(' {
		return "", "", false, -1
	}
	j := skipLinkSpace(text, i+1)
	if j < len(text) && text[j] == ')' {
		return "", "", false, j + 1
	}
	d, e := scanLinkDestination(text, j)
	if e < 0 {
		return "", "", false, -1
	}
	k := skipLinkSpace(text, e)
	if k > e && k < len(text) && (text[k] == '"' || text[k] == '\'' || text[k] == '(') {
		t, te := scanLinkTitle(text, k)
		if te < 0 {
			return "", "", false, -1
		}
		title, hasTitle = string(t), true
		k = skipLinkSpace(text, te)
	}
	if k >= len(text) || text[k] != ')' {
		return "", "", false, -1
	}
	return string(d), title, hasTitle, k + 1
}

// search szuka wzorca od pozycji from; niepowodzenie zapamiętuje, bo dotyczy też każdej dalszej pozycji.
func (p *inlineParser) search(kind int, from int, pattern string) int {
	if p.noMatch[kind] || from > len(p.text) {
		return -1
	}
	k := bytes.Index(p.text[from:], []byte(pattern))
	if k < 0 {
		p.noMatch[kind] = true
		return -1
	}
	return from + k + len(pattern)
}

// inlineHTML rozpoznaje surowy HTML zaczynający się w text[i] i zwraca pozycję za nim albo -1.
func (p *inlineParser) inlineHTML(i int) int {
	text := p.text
	if i+2 >= len(text) {
		return -1
	}
	rest := text[i:]
	switch text[i+1] {
	case '/':
		return scanClosingTag(text, i)
	case '?':
		return p.search(searchInstruction, i+2, "?>")
	case '!':
		switch {
		case bytes.HasPrefix(rest, []byte("<!--")):
			if bytes.HasPrefix(rest[4:], []byte(">")) {
				return i + 5
			}
			if bytes.HasPrefix(rest[4:], []byte("->")) {
				return i + 6
			}
			return p.search(searchComment, i+4, "-->")
		case bytes.HasPrefix(rest, []byte("<![CDATA[")):
			return p.search(searchCDATA, i+9, "]]>")
		case isLetter(text[i+2]):
			return p.search(searchDeclaration, i+2, ">")
		}
		return -1
	}
	return scanOpenTag(text, i, &p.noMatch)
}

// scanAutolink rozpoznaje autolink <adres> albo <e-mail> w s[i]; zwraca cel odnośnika i pozycję za >.
func scanAutolink(s []byte, i int) (dest string, end int) {
	j := i + 1
	// adres bezwzględny: schemat z 2–32 znaków, dwukropek, potem bez spacji, znaków sterujących, < i >
	if j < len(s) && isLetter(s[j]) {
		k := j + 1
		for k < len(s) && k-j < 32 && (isLetter(s[k]) || isDigit(s[k]) || s[k] == '+' || s[k] == '.' || s[k] == '-') {
			k++
		}
		if k-j >= 2 && k < len(s) && s[k] == ':' {
			for k++; k < len(s); k++ {
				if c := s[k]; c == '>' {
					return string(s[j:k]), k + 1
				} else if c <= ' ' || c == '<' || c == 0x7f {
					break
				}
			}
		}
	}
	// adres e-mail
	k := j
	for k < len(s) && (isLetter(s[k]) || isDigit(s[k]) || bytes.IndexByte([]byte(".!#$%&'*+/=?^_`{|}~-"), s[k]) >= 0) {
		k++
	}
	if k == j || k >= len(s) || s[k] != '@' {
		return "", -1
	}
	for {
		k++ // za @ albo kropką: etykieta domeny, 1–63 znaki, litery, cyfry i łączniki nie na brzegach
		labelStart := k
		for k < len(s) && (isLetter(s[k]) || isDigit(s[k]) || s[k] == '-') {
			k++
		}
		if k == labelStart || k-labelStart > 63 || s[labelStart] == '-' || s[k-1] == '-' || k >= len(s) {
			return "", -1
		}
		if s[k] == '>' {
			return "mailto:" + string(s[j:k]), k + 1
		}
		if s[k] != '.' {
			return "", -1
		}
	}
}

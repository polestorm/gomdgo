package gomdgo

import "bytes"

// Parser bloków: pierwsza faza strategii opisanej w dodatku do specyfikacji CommonMark. Źródło jest czytane
// wiersz po wierszu; każdy wiersz najpierw dopasowuje otwarte bloki od korzenia w dół, potem może otworzyć
// nowe, a to, co zostanie, jest treścią ostatniego otwartego bloku. Treść w wierszu to faza druga.
type parser struct {
	doc     document
	tip     *node // najgłębszy otwarty blok
	lineNo  int   // numer bieżącego wiersza, od 1
	lineOff int   // przesunięcie bieżącego wiersza w źródle
}

func parse(src []byte) *document {
	p := &parser{doc: document{src: src, refs: map[string]*reference{}}}
	p.doc.root = &node{kind: kindDocument, open: true, startLine: 1, endLine: 1}
	p.tip = p.doc.root
	for off := 0; off < len(src); {
		end := off
		for end < len(src) && src[end] != '\n' && src[end] != '\r' {
			end++
		}
		next := end
		if next < len(src) {
			next++
			if src[end] == '\r' && next < len(src) && src[next] == '\n' {
				next++
			}
		}
		p.lineNo++
		p.lineOff = off
		p.doc.lineStarts = append(p.doc.lineStarts, off)
		p.line(src[off:end])
		off = next
	}
	p.closeDescendants(p.doc.root)
	p.close(p.doc.root)
	p.inlines()
	return &p.doc
}

// inlines uruchamia fazę drugą: treść każdego akapitu i nagłówka. Dopiero teraz znane są wszystkie
// definicje odsyłaczy. Przejście po drzewie jest iteracyjne.
func (p *parser) inlines() {
	ip := inlineParser{doc: &p.doc}
	root := p.doc.root
	for n := root; ; {
		if n.kind == kindParagraph || n.kind == kindHeading {
			ip.parse(n)
		} else if n.first != nil {
			n = n.first
			continue
		}
		for n != root && n.next == nil {
			n = n.parent
		}
		if n == root {
			return
		}
		n = n.next
	}
}

// line przetwarza jeden wiersz źródła.
func (p *parser) line(s []byte) {
	c := &cursor{s: s}

	// 1. Dopasowanie otwartych kontenerów.
	container := p.doc.root
	for {
		child := container.last
		if child == nil || !child.open || !child.kind.isContainer() || !p.continues(child, c) {
			break
		}
		container = child
	}
	next := container.last
	if next != nil && !next.open {
		next = nil
	}
	allMatched := next == nil || !next.kind.isContainer()

	// para to otwarty akapit, którego ten wiersz może być dalszym ciągiem; lazy: przez kontynuację leniwą,
	// czyli mimo niedopasowania kontenerów akapitu.
	var para *node
	lazy := false
	if allMatched && next != nil {
		switch next.kind {
		case kindCodeBlock:
			if p.continueCode(next, c) {
				return
			}
		case kindHTMLBlock:
			if p.continueHTML(next, c) {
				return
			}
		case kindParagraph:
			if !c.blank() {
				para = next
			}
		}
	} else if !allMatched && !c.blank() {
		if p.tip.kind == kindParagraph {
			para, lazy = p.tip, true
		}
	}

	closed := false
	closeUnmatched := func() {
		if !closed {
			p.closeDescendants(container)
			closed = true
		}
		para, lazy = nil, false
	}

	// Nieudana próba linii poziomej zostawia pozycję bajtu, który ją wykluczył; kolejne próby w tym samym
	// wierszu (po znacznikach zagnieżdżonych list) nie czytają go od nowa.
	var noBreak [4]int

	// 2. Początki nowych bloków.
starts:
	for {
		indent, j := c.indent()
		if j == len(s) {
			break
		}
		if indent >= 4 {
			if para != nil {
				break // kod przez wcięcie nie przerywa akapitu
			}
			closeUnmatched()
			n := p.add(container, kindCodeBlock, j)
			c.skipCols(4)
			n.literal = append(c.appendRest(n.literal), '\n')
			n.trim = len(n.literal)
			return
		}
		rest := s[j:]
		pos := j
		switch ch := rest[0]; {
		case ch == '>':
			closeUnmatched()
			container = p.add(container, kindBlockQuote, pos)
			container.markerLine = p.lineNo
			c.skipCols(indent)
			c.advance(1)
			c.skipCols(1)
			continue starts

		case ch == '#':
			if level, content, ok := atxHeading(rest); ok {
				closeUnmatched()
				n := p.add(container, kindHeading, pos)
				n.level = level
				n.spans = []span{{p.lineOff + j + content.start, p.lineOff + j + content.end}}
				p.close(n)
				return
			}

		case ch == '`' || ch == '~':
			if length, info, ok := fenceOpen(rest); ok {
				closeUnmatched()
				n := p.add(container, kindCodeBlock, pos)
				n.fenced, n.fenceChar, n.fenceLen, n.fenceIndent, n.info = true, ch, length, indent, info
				return
			}

		case ch == '<':
			if t := htmlBlockStart(rest); t != 0 && (t != 7 || para == nil) {
				closeUnmatched()
				n := p.add(container, kindHTMLBlock, pos)
				n.htmlType = t
				p.addHTMLLine(n, c)
				return
			}
		}

		if para != nil && !lazy && (rest[0] == '=' || rest[0] == '-') && setextUnderline(rest) {
			p.extractRefs(para)
			if len(para.spans) > 0 {
				para.kind = kindHeading
				para.level = 1
				if rest[0] == '-' {
					para.level = 2
				}
				para.endLine = p.lineNo
				para.open = false
				p.tip = para.parent
				return
			}
			closeUnmatched() // w akapicie były same definicje odsyłaczy: to nie jest podkreślenie nagłówka
		}

		if ch := rest[0]; ch == '*' || ch == '-' || ch == '_' {
			slot := &noBreak[ch&3] // trzy znaki mają różne dwa najmłodsze bity
			if j >= *slot {
				stop := thematicBreak(rest)
				if stop < 0 {
					closeUnmatched()
					p.close(p.add(container, kindThematicBreak, pos))
					return
				}
				*slot = j + stop + 1
			}
		}

		if data, width, ok := listMarker(rest); ok {
			emptyItem := true
			for _, b := range rest[width:] {
				if b != ' ' && b != '\t' {
					emptyItem = false
					break
				}
			}
			// Element listy przerywa akapit tylko wtedy, gdy nie jest pusty, a lista numerowana zaczyna się od 1.
			if para != nil && !lazy && (emptyItem || data.ordered && data.start != 1) {
				break starts
			}
			closeUnmatched()
			if container.kind != kindList || container.list.ordered != data.ordered || container.list.delim != data.delim {
				container = p.add(container, kindList, pos)
				container.list = data
			}
			container = p.add(container, kindItem, pos)
			container.list = data
			c.skipCols(indent)
			c.advance(width)
			gap := 1
			if !emptyItem {
				if cols, _ := c.indent(); cols <= 4 {
					gap = cols
				}
				c.skipCols(gap)
			}
			container.contentIndent = indent + width + gap
			continue starts
		}
		break
	}

	// 3. Reszta wiersza: pusty wiersz, dalszy ciąg akapitu albo nowy akapit.
	if c.blank() {
		closeUnmatched()
		return
	}
	if para == nil {
		closeUnmatched()
		para = p.add(container, kindParagraph, 0)
		para.spans = make([]span, 0, 1)
	}
	c.skipSpaces()
	if len(para.spans) == 0 {
		para.pos = p.lineOff + c.i
	}
	para.spans = append(para.spans, span{p.lineOff + c.i, p.lineOff + len(s)})
	para.endLine = p.lineNo
}

// continues sprawdza, czy otwarty kontener obejmuje bieżący wiersz, i zużywa jego znacznik albo wcięcie.
func (p *parser) continues(n *node, c *cursor) bool {
	switch n.kind {
	case kindBlockQuote:
		indent, j := c.indent()
		if indent > 3 || j == len(c.s) || c.s[j] != '>' {
			return false
		}
		c.skipCols(indent)
		c.advance(1)
		c.skipCols(1)
		n.markerLine = p.lineNo
		return true
	case kindItem:
		indent, j := c.indent()
		if j == len(c.s) {
			// Pusty wiersz należy do elementu, który ma już treść; element zaczęty pustym wierszem
			// kończy się na drugim pustym.
			return n.first != nil
		}
		if indent < n.contentIndent {
			return false
		}
		c.skipCols(n.contentIndent)
		return true
	}
	return true // lista trwa, dopóki nie pojawi się blok, którego nie może zawierać
}

// continueCode dopisuje wiersz do otwartego bloku kodu; false znaczy, że kod przez wcięcie się skończył.
func (p *parser) continueCode(n *node, c *cursor) bool {
	indent, j := c.indent()
	if n.fenced {
		if indent <= 3 && j < len(c.s) && c.s[j] == n.fenceChar {
			k := j
			for k < len(c.s) && c.s[k] == n.fenceChar {
				k++
			}
			if k-j >= n.fenceLen && len(bytes.Trim(c.s[k:], " \t")) == 0 {
				n.endLine = p.lineNo
				p.close(n)
				return true
			}
		}
		c.skipCols(n.fenceIndent)
		n.literal = append(c.appendRest(n.literal), '\n')
		n.endLine = p.lineNo
		return true
	}
	blank := j == len(c.s)
	if indent < 4 && !blank {
		return false
	}
	c.skipCols(4)
	n.literal = append(c.appendRest(n.literal), '\n')
	if !blank {
		n.trim = len(n.literal)
		n.endLine = p.lineNo
	}
	return true
}

// continueHTML dopisuje wiersz do otwartego bloku HTML; false znaczy, że blok kończy się pustym wierszem.
func (p *parser) continueHTML(n *node, c *cursor) bool {
	if n.htmlType >= 6 && c.blank() {
		return false
	}
	p.addHTMLLine(n, c)
	return true
}

func (p *parser) addHTMLLine(n *node, c *cursor) {
	start := len(n.literal)
	n.literal = append(c.appendRest(n.literal), '\n')
	n.endLine = p.lineNo
	if n.htmlType <= 5 && htmlBlockEnd(n.htmlType, n.literal[start:]) {
		p.close(n)
	}
}

// add dopisuje nowy otwarty blok do kontenera; kontenery, które nie mogą go zawierać, są zamykane.
func (p *parser) add(parent *node, k kind, col int) *node {
	for (parent.kind == kindList) != (k == kindItem) {
		p.close(parent)
		parent = parent.parent
	}
	n := &node{kind: k, open: true, pos: p.lineOff + col, startLine: p.lineNo, endLine: p.lineNo}
	parent.appendChild(n)
	p.tip = n
	return n
}

// closeDescendants zamyka wszystkie otwarte bloki poniżej n.
func (p *parser) closeDescendants(n *node) {
	for p.tip != n {
		p.close(p.tip)
	}
}

// close zamyka blok i ustala to, co wiadomo dopiero na jego końcu.
func (p *parser) close(n *node) {
	if !n.open {
		return
	}
	n.open = false
	if p.tip == n {
		p.tip = n.parent
	}
	switch n.kind {
	case kindParagraph:
		p.extractRefs(n)
		if len(n.spans) == 0 {
			n.kind = kindRefDefs
		}
	case kindCodeBlock:
		if !n.fenced {
			n.literal = n.literal[:n.trim]
		}
	case kindBlockQuote:
		n.endLine = n.markerLine
	case kindList:
		n.list.tight = listIsTight(n)
	}
	if n.kind.isContainer() && n.last != nil && n.last.endLine > n.endLine {
		n.endLine = n.last.endLine
	}
}

// listIsTight: lista jest luźna, gdy pusty wiersz dzieli dwa jej elementy albo dwa bloki będące bezpośrednio
// w jednym elemencie.
func listIsTight(list *node) bool {
	for item := list.first; item != nil; item = item.next {
		if item.next != nil && item.next.startLine > item.endLine+1 {
			return false
		}
		for b := item.first; b != nil && b.next != nil; b = b.next {
			if b.next.startLine > b.endLine+1 {
				return false
			}
		}
	}
	return true
}

// extractRefs zdejmuje z początku akapitu definicje odsyłaczy i zapisuje je w dokumencie. Definicja kończy się
// zawsze na końcu wiersza, więc z akapitu ubywa całych wierszy.
func (p *parser) extractRefs(n *node) {
	if len(n.spans) == 0 || p.doc.src[n.spans[0].start] != '[' {
		return
	}
	text := joinSpans(p.doc.src, n.spans)
	lines := 0
	for off := 0; off < len(text) && text[off] == '['; {
		label, ref, used, ok := scanRefDef(text[off:])
		if !ok {
			break
		}
		ref.pos = n.spans[lines].start
		if _, dup := p.doc.refs[label]; !dup {
			p.doc.refs[label] = &ref
		}
		lines += bytes.Count(text[off:off+used], []byte{'\n'})
		off += used
		if off == len(text) {
			lines = len(n.spans)
		}
	}
	n.spans = n.spans[lines:]
	n.startLine += lines
	if len(n.spans) > 0 {
		n.pos = n.spans[0].start
	}
}

// joinSpans składa wiersze treści w jeden tekst, rozdzielone znakiem nowego wiersza.
func joinSpans(src []byte, spans []span) []byte {
	if len(spans) == 1 {
		return src[spans[0].start:spans[0].end]
	}
	var out []byte
	for i, s := range spans {
		if i > 0 {
			out = append(out, '\n')
		}
		out = append(out, src[s.start:s.end]...)
	}
	return out
}

// atxHeading rozpoznaje nagłówek ATX na początku s (bez wcięcia) i zwraca poziom oraz położenie treści w s.
func atxHeading(s []byte) (level int, content span, ok bool) {
	for level < len(s) && s[level] == '#' {
		level++
	}
	if level > 6 || level < len(s) && s[level] != ' ' && s[level] != '\t' {
		return 0, span{}, false
	}
	start, end := level, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	// Sekwencja zamykająca: znaki # na końcu, poprzedzone spacją albo stanowiące całą treść.
	k := end
	for k > start && s[k-1] == '#' {
		k--
	}
	if k == start {
		end = start
	} else if k < end && (s[k-1] == ' ' || s[k-1] == '\t') {
		end = k
		for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
			end--
		}
	}
	return level, span{start, end}, true
}

// setextUnderline: same znaki = albo same -, potem najwyżej spacje i tabulatory.
func setextUnderline(s []byte) bool {
	i := 0
	for i < len(s) && s[i] == s[0] {
		i++
	}
	for ; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return false
		}
	}
	return true
}

// thematicBreak sprawdza linię poziomą: co najmniej trzy jednakowe znaki * - albo _ (pierwszy to s[0]),
// między nimi najwyżej spacje i tabulatory. Zwraca -1 dla linii poziomej, a inaczej pozycję bajtu, który
// ją wyklucza (długość s, gdy znaków jest za mało).
func thematicBreak(s []byte) int {
	count := 0
	for i, b := range s {
		switch b {
		case s[0]:
			count++
		case ' ', '\t':
		default:
			return i
		}
	}
	if count >= 3 {
		return -1
	}
	return len(s)
}

// fenceOpen rozpoznaje otwarcie kodu ogrodzonego; zwraca długość ogrodzenia i napis informacyjny.
func fenceOpen(s []byte) (length int, info string, ok bool) {
	for length < len(s) && s[length] == s[0] {
		length++
	}
	if length < 3 {
		return 0, "", false
	}
	rest := bytes.Trim(s[length:], " \t")
	if s[0] == '`' && bytes.IndexByte(rest, '`') >= 0 {
		return 0, "", false
	}
	return length, string(rest), true
}

// listMarker rozpoznaje znacznik elementu listy na początku s; width to długość samego znacznika.
func listMarker(s []byte) (data listData, width int, ok bool) {
	switch ch := s[0]; {
	case ch == '-' || ch == '+' || ch == '*':
		data.delim = ch
		width = 1
	case ch >= '0' && ch <= '9':
		for width < len(s) && width < 9 && s[width] >= '0' && s[width] <= '9' {
			data.start = data.start*10 + int(s[width]-'0')
			width++
		}
		if width == len(s) || s[width] != '.' && s[width] != ')' {
			return data, 0, false
		}
		data.ordered = true
		data.delim = s[width]
		width++
	default:
		return data, 0, false
	}
	if width < len(s) && s[width] != ' ' && s[width] != '\t' {
		return data, 0, false
	}
	return data, width, true
}

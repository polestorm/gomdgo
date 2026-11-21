package gomdgo

import (
	"bytes"
	"strconv"
	"strings"
	"unicode"
)

type renderer struct {
	doc  *document
	conv *Converter
	out  bytes.Buffer
	ids  map[string]bool // identyfikatory nagłówków użyte w tym przebiegu
}

// cr zaczyna nowy wiersz wyniku, jeśli bieżący nie jest pusty.
func (r *renderer) cr() {
	if b := r.out.Bytes(); len(b) > 0 && b[len(b)-1] != '\n' {
		r.out.WriteByte('\n')
	}
}

// render zapisuje całe drzewo. Przejście jest iteracyjne, po wskaźnikach węzłów, żeby głębokość
// zagnieżdżenia w źródle nie przekładała się na głębokość stosu.
func (r *renderer) render(root *node) {
	n, entering := root, true
	for {
		if entering {
			if descend := r.enter(n); descend && n.first != nil {
				n = n.first
				continue
			}
		}
		r.leave(n)
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

// tightParagraph: w liście zwartej akapit będący bezpośrednio w elemencie nie ma znaczników <p>.
func tightParagraph(n *node) bool {
	return n.parent.kind == kindItem && n.parent.parent.list.tight
}

func listTag(n *node) string {
	if n.list.ordered {
		return "ol"
	}
	return "ul"
}

// enter zapisuje początek węzła, a dla węzłów bez dzieci całą treść. Zwraca false, gdy dzieci węzła
// nie należy odwiedzać.
func (r *renderer) enter(n *node) bool {
	out := &r.out
	switch n.kind {
	case kindBlockQuote:
		r.cr()
		out.WriteString("<blockquote>\n")
	case kindList:
		r.cr()
		out.WriteString("<" + listTag(n))
		if n.list.ordered && n.list.start != 1 {
			out.WriteString(` start="` + strconv.Itoa(n.list.start) + `"`)
		}
		out.WriteString(">\n")
	case kindItem:
		out.WriteString("<li>")
	case kindParagraph:
		if !tightParagraph(n) {
			r.cr()
			out.WriteString("<p>")
		}
	case kindHeading:
		r.cr()
		out.WriteString("<h" + strconv.Itoa(n.level))
		if r.conv.cfg.HeadingIDs {
			out.WriteString(` id="`)
			escapeHTML(out, []byte(r.headingID(n)))
			out.WriteByte('"')
		}
		out.WriteByte('>')
	case kindThematicBreak:
		r.cr()
		out.WriteString("<hr />\n")
	case kindCodeBlock:
		r.cr()
		out.WriteString("<pre><code")
		if lang, _, _ := strings.Cut(strings.ReplaceAll(n.info, "\t", " "), " "); lang != "" {
			out.WriteString(` class="language-`)
			escapeHTML(out, []byte(unescape(lang)))
			out.WriteString(`"`)
		}
		out.WriteString(">")
		escapeHTML(out, n.literal)
		out.WriteString("</code></pre>\n")
	case kindHTMLBlock:
		switch r.conv.cfg.RawHTML {
		case RawHTMLPass:
			r.cr()
			out.Write(n.literal)
		case RawHTMLOmit:
		default:
			// Blok HTML jako tekst zachowuje się jak akapit, także w liście zwartej.
			text := bytes.TrimRight(n.literal, "\n")
			if tightParagraph(n) {
				if n.prev != nil {
					r.cr()
				}
				escapeHTML(out, text)
			} else {
				r.cr()
				out.WriteString("<p>")
				escapeHTML(out, text)
				out.WriteString("</p>\n")
			}
		}

	case kindText:
		escapeHTML(out, n.literal)
	case kindSoftBreak:
		out.WriteByte('\n')
	case kindHardBreak:
		out.WriteString("<br />\n")
	case kindCode:
		out.WriteString("<code>")
		escapeHTML(out, n.literal)
		out.WriteString("</code>")
	case kindEmph:
		out.WriteString("<em>")
	case kindStrong:
		out.WriteString("<strong>")
	case kindLink:
		// Odnośnik o niedozwolonym schemacie zostaje samą swoją treścią.
		if r.conv.allowed(n.info) {
			out.WriteString(`<a href="`)
			escapeHTML(out, encodeURL(n.info))
			r.titleAttr(n)
			out.WriteString(">")
		}
	case kindImage:
		// Obraz o niedozwolonym schemacie zostaje samym swoim opisem.
		alt := plainText(n)
		if r.conv.allowed(n.info) {
			out.WriteString(`<img src="`)
			escapeHTML(out, encodeURL(n.info))
			out.WriteString(`" alt="`)
			escapeHTML(out, alt)
			r.titleAttr(n)
			out.WriteString(" />")
		} else {
			escapeHTML(out, alt)
		}
		return false
	case kindHTMLInline:
		switch r.conv.cfg.RawHTML {
		case RawHTMLPass:
			out.Write(n.literal)
		case RawHTMLOmit:
		default:
			escapeHTML(out, n.literal)
		}
	}
	return true
}

// titleAttr domyka atrybut adresu i dopisuje atrybut title, jeśli odnośnik albo obraz ma tytuł.
func (r *renderer) titleAttr(n *node) {
	r.out.WriteByte('"')
	if n.hasTitle {
		r.out.WriteString(` title="`)
		escapeHTML(&r.out, []byte(n.title))
		r.out.WriteByte('"')
	}
}

// plainText zwraca sam tekst poddrzewa, bez znaczników: opis obrazu i podstawa identyfikatora nagłówka.
func plainText(root *node) []byte {
	var out []byte
	if root.first == nil {
		return out
	}
	for n := root.first; ; {
		switch n.kind {
		case kindText, kindCode:
			out = append(out, n.literal...)
		case kindSoftBreak, kindHardBreak:
			out = append(out, '\n')
		}
		if n.first != nil {
			n = n.first
			continue
		}
		for n.next == nil {
			if n = n.parent; n == root {
				return out
			}
		}
		n = n.next
	}
}

// headingID tworzy identyfikator nagłówka z jego tekstu: litery i cyfry małymi literami, odstępy jako
// łączniki, reszta pominięta. Powtórzony identyfikator dostaje kolejny numer.
func (r *renderer) headingID(n *node) string {
	var b strings.Builder
	for _, c := range string(plainText(n)) {
		switch {
		case unicode.IsLetter(c) || unicode.IsDigit(c):
			b.WriteRune(unicode.ToLower(c))
		case c == '-' || c == '_':
			b.WriteRune(c)
		case c == ' ' || c == '\t' || c == '\n':
			b.WriteByte('-')
		}
	}
	base := b.String()
	if base == "" {
		base = "section"
	}
	if r.ids == nil {
		r.ids = map[string]bool{}
	}
	id := base
	for i := 1; r.ids[id]; i++ {
		id = base + "-" + strconv.Itoa(i)
	}
	r.ids[id] = true
	return id
}

// allowed mówi, czy adres odnośnika albo obrazu może trafić do wyniku. Adres bez schematu jest dozwolony
// zawsze; schemat to litera, a po niej litery, cyfry i znaki + . - aż do pierwszego dwukropka. Każdy inny
// znak przed dwukropkiem (także taki, który zostanie zakodowany procentowo) oznacza adres względny.
func (c *Converter) allowed(url string) bool {
	if c.cfg.AnyScheme {
		return true
	}
	for i := 0; i < len(url); i++ {
		switch b := url[i]; {
		case b == ':':
			return i == 0 || c.schemes[strings.ToLower(url[:i])]
		case isLetter(b):
		case i > 0 && (isDigit(b) || b == '+' || b == '.' || b == '-'):
		default:
			return true
		}
	}
	return true
}

// leave zapisuje koniec węzła, który ma znacznik zamykający.
func (r *renderer) leave(n *node) {
	out := &r.out
	switch n.kind {
	case kindBlockQuote:
		out.WriteString("</blockquote>\n")
	case kindList:
		out.WriteString("</" + listTag(n) + ">\n")
	case kindItem:
		out.WriteString("</li>\n")
	case kindParagraph:
		if !tightParagraph(n) {
			out.WriteString("</p>\n")
		}
	case kindHeading:
		out.WriteString("</h" + strconv.Itoa(n.level) + ">\n")
	case kindEmph:
		out.WriteString("</em>")
	case kindStrong:
		out.WriteString("</strong>")
	case kindLink:
		if r.conv.allowed(n.info) {
			out.WriteString("</a>")
		}
	}
}

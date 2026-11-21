package gomdgo

import (
	"bytes"
	"strconv"
	"strings"
)

type renderer struct {
	doc  *document
	opts Options
	out  bytes.Buffer
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
		out.WriteString("<h" + strconv.Itoa(n.level) + ">")
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
		r.cr()
		r.rawHTML(n.literal)

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
		out.WriteString(`<a href="`)
		escapeHTML(out, encodeURL(n.info))
		r.titleAttr(n)
		out.WriteString(">")
	case kindImage:
		out.WriteString(`<img src="`)
		escapeHTML(out, encodeURL(n.info))
		out.WriteString(`" alt="`)
		r.plainText(n)
		r.titleAttr(n)
		out.WriteString(" />")
		return false
	case kindHTMLInline:
		r.rawHTML(n.literal)
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

// rawHTML zapisuje surowy HTML ze źródła zgodnie z ustawieniami.
func (r *renderer) rawHTML(literal []byte) {
	if r.opts.UnsafeHTML {
		r.out.Write(literal)
	} else {
		// Tymczasowo: jako tekst. Politykę surowego HTML ustala etap 3.
		escapeHTML(&r.out, literal)
	}
}

// plainText zapisuje sam tekst poddrzewa (opis obrazu w atrybucie alt), bez znaczników.
func (r *renderer) plainText(root *node) {
	if root.first == nil {
		return
	}
	for n := root.first; ; {
		switch n.kind {
		case kindText, kindCode:
			escapeHTML(&r.out, n.literal)
		case kindSoftBreak, kindHardBreak:
			r.out.WriteByte('\n')
		}
		if n.first != nil {
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
		out.WriteString("</a>")
	}
}

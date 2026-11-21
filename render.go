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
			r.enter(n)
			if n.first != nil {
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

func listTag(n *node) string {
	if n.list.ordered {
		return "ol"
	}
	return "ul"
}

// enter zapisuje początek bloku, a dla bloków bez dzieci całą treść.
func (r *renderer) enter(n *node) {
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
		// W liście zwartej akapit będący bezpośrednio w elemencie nie ma znaczników <p>.
		if n.parent.kind == kindItem && n.parent.parent.list.tight {
			r.inline(n)
			return
		}
		r.cr()
		out.WriteString("<p>")
		r.inline(n)
		out.WriteString("</p>\n")
	case kindHeading:
		r.cr()
		tag := "h" + strconv.Itoa(n.level)
		out.WriteString("<" + tag + ">")
		r.inline(n)
		out.WriteString("</" + tag + ">\n")
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
		if r.opts.UnsafeHTML {
			out.Write(n.literal)
		} else {
			// Tymczasowo: treść bloku jako tekst. Politykę surowego HTML ustala etap 3.
			escapeHTML(out, n.literal)
		}
	}
}

// leave zapisuje koniec bloku, który ma znacznik zamykający.
func (r *renderer) leave(n *node) {
	switch n.kind {
	case kindBlockQuote:
		r.out.WriteString("</blockquote>\n")
	case kindList:
		r.out.WriteString("</" + listTag(n) + ">\n")
	case kindItem:
		r.out.WriteString("</li>\n")
	}
}

// inline zapisuje treść akapitu albo nagłówka. ZAŚLEPKA etapu 1: treść w wierszu nie jest jeszcze
// parsowana (etap 2); wiersze są łączone znakiem nowego wiersza, bez białych znaków na końcach.
func (r *renderer) inline(n *node) {
	for i, s := range n.spans {
		if i > 0 {
			r.out.WriteByte('\n')
		}
		escapeHTML(&r.out, bytes.TrimRight(r.doc.src[s.start:s.end], " \t"))
	}
}

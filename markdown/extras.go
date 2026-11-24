package markdown

import (
	"bytes"
	"strconv"
)

// Rozszerzenia bez specyfikacji: bloki uwag i przypisy, według dokumentacji użytkowej GitHub i własnych
// wymagań (ADR 0011, requirements/extensions.md). Włączane przez Config.Enable.

var alertNames = [...]string{"", "note", "tip", "important", "warning", "caution"}

// detectAlert zamienia zamykany cytat w blok uwagi, jeśli pierwszym wierszem jego pierwszego akapitu jest
// znacznik [!RODZAJ]. Znacznik znika z treści.
func (p *parser) detectAlert(quote *node) {
	first := quote.first
	if first == nil || first.kind != kindParagraph || len(first.spans) == 0 {
		return
	}
	line := bytes.TrimRight(p.doc.src[first.spans[0].start:first.spans[0].end], " \t")
	if len(line) < 4 || len(line) > 16 || line[0] != '[' || line[1] != '!' || line[len(line)-1] != ']' {
		return
	}
	name := string(bytes.ToLower(line[2 : len(line)-1]))
	for i := 1; i < len(alertNames); i++ {
		if alertNames[i] != name {
			continue
		}
		quote.level = int32(i)
		first.spans = first.spans[1:]
		first.startLine++
		if len(first.spans) == 0 {
			first.unlink()
		} else {
			first.pos = first.spans[0].start
		}
		return
	}
}

// footnoteLabel rozpoznaje [^etykieta na początku s i zwraca etykietę oraz pozycję za nawiasem zamykającym.
// Etykieta jest niepusta, ma najwyżej 999 bajtów i nie zawiera białych znaków ani nawiasów kwadratowych.
func footnoteLabel(s []byte) (label []byte, end int) {
	if len(s) < 4 || s[0] != '[' || s[1] != '^' {
		return nil, -1
	}
	for i := 2; i < len(s) && i < 1002; i++ {
		switch s[i] {
		case ']':
			if i == 2 {
				return nil, -1
			}
			return s[2:i], i + 1
		case ' ', '\t', '\n', '[':
			return nil, -1
		}
	}
	return nil, -1
}

// footnoteState to stan przypisów jednego przebiegu renderowania.
type footnoteState struct {
	order   []*node       // przypisy w kolejności pierwszego odwołania
	number  map[*node]int // numer przypisu, od 1
	refs    map[*node]int // liczba odwołań do przypisu zapisanych do tej pory
	current *node         // przypis, którego treść jest właśnie zapisywana
}

// footnoteRef zapisuje odwołanie do przypisu i nadaje mu numer przy pierwszym odwołaniu.
func (r *renderer) footnoteRef(def *node) {
	f := &r.fn
	if f.number == nil {
		f.number, f.refs = map[*node]int{}, map[*node]int{}
	}
	num := f.number[def]
	if num == 0 {
		f.order = append(f.order, def)
		num = len(f.order)
		f.number[def] = num
	}
	f.refs[def]++
	id := strconv.Itoa(num)
	r.out.WriteString(`<sup><a href="#fn:` + id + `" id="fnref:` + id)
	if k := f.refs[def]; k > 1 {
		r.out.WriteString(":" + strconv.Itoa(k))
	}
	r.out.WriteString(`">` + id + `</a></sup>`)
}

// backrefs zapisuje odnośniki powrotne z przypisu do wszystkich odwołań do niego.
func (r *renderer) backrefs(def *node, inParagraph bool) {
	id := strconv.Itoa(r.fn.number[def])
	for k := 1; k <= r.fn.refs[def]; k++ {
		if inParagraph || k > 1 {
			r.out.WriteByte(' ')
		}
		r.out.WriteString(`<a href="#fnref:` + id)
		if k > 1 {
			r.out.WriteString(":" + strconv.Itoa(k))
		}
		r.out.WriteString(`" class="footnote-backref">↩</a>`)
	}
}

// footnotes zapisuje na końcu dokumentu treść przypisów, do których były odwołania, w kolejności pierwszego
// odwołania. Treść przypisu może odwoływać się do kolejnych przypisów; lista rośnie wtedy w trakcie.
func (r *renderer) footnotes() {
	if len(r.fn.order) == 0 {
		return
	}
	r.cr()
	r.out.WriteString("<section class=\"footnotes\">\n<ol>\n")
	for i := 0; i < len(r.fn.order) && r.err == nil; i++ {
		r.fn.current = r.fn.order[i]
		r.render(r.fn.current)
	}
	r.out.WriteString("</ol>\n</section>\n")
}

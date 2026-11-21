// Package htmlnorm sprowadza fragment HTML do postaci kanonicznej, żeby dwa wyniki różniące się tylko zapisem
// (a nie znaczeniem) porównywały się jako równe.
//
// Specyfikacja CommonMark zastrzega, że jej przykłady opisują wynik z dokładnością do nieistotnych różnic
// zapisu HTML. Tutaj za nieistotne uznajemy:
//   - białe znaki na początku i końcu dokumentu oraz przy znacznikach blokowych; serie białych znaków w tekście
//     (poza <pre>) to jedna spacja;
//   - wielkość liter nazw znaczników i atrybutów, kolejność atrybutów, rodzaj cudzysłowu, ukośnik w <br />;
//   - atrybut bez wartości i z wartością pustą (checked ≡ checked="");
//   - zapis znaku wprost albo odwołaniem znakowym (&quot; ≡ ", &#39; ≡ ', &ouml; ≡ ö).
//
// Wszystko inne, w tym kodowanie procentowe adresów, jest różnicą rzeczywistą. Pakiet nie jest parserem HTML:
// komentarze, deklaracje i wszystko, co nie wygląda jak znacznik, przechodzi bez zmian.
package htmlnorm

import (
	"html"
	"sort"
	"strings"
)

var blockTags = map[string]bool{
	"blockquote": true, "br": true, "dd": true, "div": true, "dl": true, "dt": true, "h1": true, "h2": true,
	"h3": true, "h4": true, "h5": true, "h6": true, "hr": true, "li": true, "ol": true, "p": true, "pre": true,
	"section": true, "table": true, "tbody": true, "td": true, "tfoot": true, "th": true, "thead": true,
	"tr": true, "ul": true,
}

type token struct {
	text  string // postać kanoniczna
	tag   bool
	block bool // znacznik blokowy: białe znaki obok niego są nieistotne
	raw   bool // tekst wewnątrz <pre>: białe znaki znaczące
}

// Normalize zwraca postać kanoniczną fragmentu HTML.
func Normalize(src string) string {
	var tokens []token
	pre := 0
	for i := 0; i < len(src); {
		if src[i] == '<' {
			if end := verbatimEnd(src, i); end > 0 {
				tokens = append(tokens, token{text: src[i:end], tag: true})
				i = end
				continue
			}
			if t, name, closing, end := parseTag(src, i); end > 0 {
				if name == "pre" {
					if closing && pre > 0 {
						pre--
					} else if !closing {
						pre++
					}
				}
				tokens = append(tokens, token{text: t, tag: true, block: blockTags[name]})
				i = end
				continue
			}
		}
		end := i + 1
		for end < len(src) && src[end] != '<' {
			end++
		}
		text := escapeText(html.UnescapeString(src[i:end]))
		if n := len(tokens); n > 0 && !tokens[n-1].tag && tokens[n-1].raw == (pre > 0) {
			tokens[n-1].text += text
		} else {
			tokens = append(tokens, token{text: text, raw: pre > 0})
		}
		i = end
	}

	var out strings.Builder
	for i, t := range tokens {
		if t.tag || t.raw {
			out.WriteString(t.text)
			continue
		}
		text := strings.Join(strings.Fields(t.text), " ")
		if text == "" {
			// sam odstęp: znaczący tylko między dwoma elementami w wierszu
			if i > 0 && i < len(tokens)-1 && !tokens[i-1].block && !tokens[i+1].block {
				out.WriteByte(' ')
			}
			continue
		}
		if startsWithSpace(t.text) && i > 0 && !tokens[i-1].block {
			out.WriteByte(' ')
		}
		out.WriteString(text)
		if endsWithSpace(t.text) && i < len(tokens)-1 && !tokens[i+1].block {
			out.WriteByte(' ')
		}
	}
	return out.String()
}

func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' }

func startsWithSpace(s string) bool { return s != "" && isSpace(s[0]) }
func endsWithSpace(s string) bool   { return s != "" && isSpace(s[len(s)-1]) }

func escapeText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

func escapeAttr(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// verbatimEnd rozpoznaje komentarz, instrukcję przetwarzania, deklarację i sekcję CDATA; zwraca pozycję za nimi
// albo 0.
func verbatimEnd(s string, i int) int {
	for _, d := range [][2]string{{"<!--", "-->"}, {"<![CDATA[", "]]>"}, {"<?", "?>"}, {"<!", ">"}} {
		if strings.HasPrefix(s[i:], d[0]) {
			if end := strings.Index(s[i+len(d[0]):], d[1]); end >= 0 {
				return i + len(d[0]) + end + len(d[1])
			}
			return 0
		}
	}
	return 0
}

func isNameByte(c byte, first bool) bool {
	if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
		return true
	}
	return !first && (c >= '0' && c <= '9' || c == '-' || c == '_' || c == ':' || c == '.')
}

// parseTag rozpoznaje znacznik otwierający albo zamykający w s[i:] i zwraca jego postać kanoniczną, nazwę,
// informację, czy jest zamykający, i pozycję za nim. Gdy to nie znacznik, end jest równe 0.
func parseTag(s string, i int) (canon, name string, closing bool, end int) {
	j := i + 1
	if j < len(s) && s[j] == '/' {
		closing = true
		j++
	}
	start := j
	for j < len(s) && isNameByte(s[j], j == start) {
		j++
	}
	if j == start {
		return "", "", false, 0
	}
	name = strings.ToLower(s[start:j])
	type attr struct{ name, value string }
	var attrs []attr
	for {
		for j < len(s) && isSpace(s[j]) {
			j++
		}
		if j >= len(s) {
			return "", "", false, 0
		}
		if s[j] == '>' {
			j++
			break
		}
		if s[j] == '/' {
			j++
			continue
		}
		start = j
		for j < len(s) && !isSpace(s[j]) && s[j] != '=' && s[j] != '>' && s[j] != '/' {
			j++
		}
		if j == start {
			return "", "", false, 0
		}
		a := attr{name: strings.ToLower(s[start:j])}
		for j < len(s) && isSpace(s[j]) {
			j++
		}
		if j < len(s) && s[j] == '=' {
			j++
			for j < len(s) && isSpace(s[j]) {
				j++
			}
			if j >= len(s) {
				return "", "", false, 0
			}
			if q := s[j]; q == '"' || q == '\'' {
				close := strings.IndexByte(s[j+1:], q)
				if close < 0 {
					return "", "", false, 0
				}
				a.value = s[j+1 : j+1+close]
				j += close + 2
			} else {
				start = j
				for j < len(s) && !isSpace(s[j]) && s[j] != '>' {
					j++
				}
				a.value = s[start:j]
			}
		}
		attrs = append(attrs, a)
	}
	if closing {
		return "</" + name + ">", name, true, j
	}
	sort.SliceStable(attrs, func(a, b int) bool { return attrs[a].name < attrs[b].name })
	var b strings.Builder
	b.WriteString("<" + name)
	for _, a := range attrs {
		b.WriteString(" " + a.name + `="` + escapeAttr(html.UnescapeString(a.value)) + `"`)
	}
	b.WriteString(">")
	return b.String(), name, false, j
}

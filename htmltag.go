package gomdgo

import "bytes"

// Rozpoznawanie znaczników HTML: warunki początku i końca bloków HTML (CommonMark 4.6) oraz skanery
// znacznika otwierającego i zamykającego, wspólne z surowym HTML w wierszu (faza druga).

func isLetter(c byte) bool { return c|0x20 >= 'a' && c|0x20 <= 'z' }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }

// Znaczniki, których treść jest dosłowna (warunek początku 1).
var verbatimTags = map[string]bool{"pre": true, "script": true, "style": true, "textarea": true}

// Znaczniki blokowe (warunek początku 6), lista ze specyfikacji 0.31.2.
var blockTags = func() map[string]bool {
	m := map[string]bool{}
	for _, name := range bytes.Fields([]byte(`address article aside base basefont blockquote body caption center
		col colgroup dd details dialog dir div dl dt fieldset figcaption figure footer form frame frameset
		h1 h2 h3 h4 h5 h6 head header hr html iframe legend li link main menu menuitem nav noframes ol
		optgroup option p param search section summary table tbody td tfoot th thead title tr track ul`)) {
		m[string(name)] = true
	}
	return m
}()

// htmlBlockStart zwraca numer warunku początku bloku HTML (1–7) dla wiersza zaczynającego się od <, albo 0.
func htmlBlockStart(s []byte) int {
	if len(s) < 2 {
		return 0
	}
	switch {
	case bytes.HasPrefix(s, []byte("<!--")):
		return 2
	case s[1] == '?':
		return 3
	case bytes.HasPrefix(s, []byte("<![CDATA[")):
		return 5
	case s[1] == '!':
		if len(s) > 2 && isLetter(s[2]) {
			return 4
		}
		return 0
	}
	i := 1
	closing := s[1] == '/'
	if closing {
		i++
	}
	j := i
	for j < len(s) && (isLetter(s[j]) || isDigit(s[j])) {
		j++
	}
	name := string(bytes.ToLower(s[i:j]))
	if verbatimTags[name] {
		if !closing && (j == len(s) || s[j] == ' ' || s[j] == '\t' || s[j] == '>') {
			return 1
		}
		return 0
	}
	if blockTags[name] &&
		(j == len(s) || s[j] == ' ' || s[j] == '\t' || s[j] == '>' || bytes.HasPrefix(s[j:], []byte("/>"))) {
		return 6
	}
	end := scanOpenTag(s, 0)
	if end < 0 {
		end = scanClosingTag(s, 0)
	}
	if end >= 0 && len(bytes.Trim(s[end:], " \t")) == 0 {
		return 7
	}
	return 0
}

// htmlBlockEnd sprawdza warunek końca bloku HTML rodzaju 1–5 w wierszu.
func htmlBlockEnd(htmlType int, line []byte) bool {
	switch htmlType {
	case 1:
		lower := bytes.ToLower(line)
		for name := range verbatimTags {
			if bytes.Contains(lower, []byte("</"+name+">")) {
				return true
			}
		}
		return false
	case 2:
		return bytes.Contains(line, []byte("-->"))
	case 3:
		return bytes.Contains(line, []byte("?>"))
	case 4:
		return bytes.IndexByte(line, '>') >= 0
	}
	return bytes.Contains(line, []byte("]]>"))
}

func isTagSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' }

// scanTagName rozpoznaje nazwę znacznika: litera, potem litery, cyfry i łączniki.
func scanTagName(s []byte, i int) int {
	if i >= len(s) || !isLetter(s[i]) {
		return -1
	}
	for i < len(s) && (isLetter(s[i]) || isDigit(s[i]) || s[i] == '-') {
		i++
	}
	return i
}

// scanOpenTag rozpoznaje znacznik otwierający zaczynający się w s[i]; zwraca pozycję za nim albo -1.
func scanOpenTag(s []byte, i int) int {
	if i >= len(s) || s[i] != '<' {
		return -1
	}
	i = scanTagName(s, i+1)
	if i < 0 {
		return -1
	}
	for {
		spaces := i
		for i < len(s) && isTagSpace(s[i]) {
			i++
		}
		if i >= len(s) {
			return -1
		}
		if s[i] == '>' {
			return i + 1
		}
		if s[i] == '/' {
			if i+1 < len(s) && s[i+1] == '>' {
				return i + 2
			}
			return -1
		}
		if i == spaces {
			return -1 // atrybut musi być poprzedzony białym znakiem
		}
		// nazwa atrybutu
		if !(isLetter(s[i]) || s[i] == '_' || s[i] == ':') {
			return -1
		}
		for i < len(s) && (isLetter(s[i]) || isDigit(s[i]) || s[i] == '_' || s[i] == ':' || s[i] == '.' || s[i] == '-') {
			i++
		}
		// opcjonalna wartość
		j := i
		for j < len(s) && isTagSpace(s[j]) {
			j++
		}
		if j >= len(s) || s[j] != '=' {
			continue
		}
		j++
		for j < len(s) && isTagSpace(s[j]) {
			j++
		}
		if j >= len(s) {
			return -1
		}
		switch q := s[j]; q {
		case '"', '\'':
			k := bytes.IndexByte(s[j+1:], q)
			if k < 0 {
				return -1
			}
			i = j + k + 2
		default:
			k := j
			for k < len(s) && !isTagSpace(s[k]) && !bytes.ContainsRune([]byte("\"'=<>`"), rune(s[k])) {
				k++
			}
			if k == j {
				return -1
			}
			i = k
		}
	}
}

// scanClosingTag rozpoznaje znacznik zamykający zaczynający się w s[i]; zwraca pozycję za nim albo -1.
func scanClosingTag(s []byte, i int) int {
	if i+1 >= len(s) || s[i] != '<' || s[i+1] != '/' {
		return -1
	}
	i = scanTagName(s, i+2)
	if i < 0 {
		return -1
	}
	for i < len(s) && isTagSpace(s[i]) {
		i++
	}
	if i < len(s) && s[i] == '>' {
		return i + 1
	}
	return -1
}

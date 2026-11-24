package markdown

import (
	"bytes"
	"html"
	"strconv"
	"strings"
)

// escapeHTML zapisuje tekst z zamianą czterech znaków, które specyfikacja każe escapować w treści i atrybutach.
func escapeHTML(out *bytes.Buffer, text []byte) {
	start := 0
	for i, c := range text {
		var rep string
		switch c {
		case '&':
			rep = "&amp;"
		case '<':
			rep = "&lt;"
		case '>':
			rep = "&gt;"
		case '"':
			rep = "&quot;"
		default:
			continue
		}
		out.Write(text[start:i])
		out.WriteString(rep)
		start = i + 1
	}
	out.Write(text[start:])
}

// unescape usuwa znaki ucieczki przed interpunkcją ASCII i rozwija odwołania znakowe. Stosowane tam, gdzie
// specyfikacja nie parsuje treści w wierszu, a mimo to honoruje oba zapisy: napis informacyjny kodu,
// cel i tytuł odnośnika.
func unescape(s string) string {
	if !strings.ContainsAny(s, `\&`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\\' && i+1 < len(s) && isPunct(s[i+1]):
			i++
			b.WriteByte(s[i])
		case c == '&':
			if text, n := entity(s[i:]); n > 0 {
				b.WriteString(text)
				i += n - 1
			} else {
				b.WriteByte(c)
			}
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// entity rozpoznaje odwołanie znakowe na początku s (które zaczyna się od &) i zwraca jego znaczenie oraz
// długość zapisu; n = 0, gdy to nie jest poprawne odwołanie.
//
// Nazwy encji HTML5 zna biblioteka standardowa (pakiet html), ale przyjmuje też zapisy bez średnika
// i dopasowuje przedrostki; tu ważny jest tylko zapis pełny, ze średnikiem.
func entity(s string) (text string, n int) {
	end := strings.IndexByte(s, ';')
	if end < 2 || end > 33 {
		return "", 0
	}
	body := s[1:end]
	if body[0] == '#' {
		digits, base, maxLen := body[1:], 10, 7
		if len(digits) > 0 && (digits[0] == 'x' || digits[0] == 'X') {
			digits, base, maxLen = digits[1:], 16, 6
		}
		if len(digits) == 0 || len(digits) > maxLen {
			return "", 0
		}
		code, err := strconv.ParseUint(digits, base, 32)
		if err != nil {
			return "", 0
		}
		if code == 0 || code > 0x10FFFF || code >= 0xD800 && code <= 0xDFFF {
			return "�", end + 1
		}
		return string(rune(code)), end + 1
	}
	for i := 0; i < len(body); i++ {
		if !isLetter(body[i]) && !isDigit(body[i]) {
			return "", 0
		}
	}
	text = html.UnescapeString(s[:end+1])
	// Dopasowanie samego przedrostka (np. &copyx;) zostawia średnik na końcu; jedyna encja, której
	// znaczeniem jest średnik, to &semi;.
	if text == s[:end+1] || strings.HasSuffix(text, ";") && body != "semi" {
		return "", 0
	}
	return text, end + 1
}

// urlSafe: znaki, które w adresie zostają bez zmian (litery, cyfry, znaki zastrzeżone i niezastrzeżone
// według RFC 3986 oraz znak procentu, żeby nie kodować dwa razy).
var urlSafe = func() (t [256]bool) {
	for _, c := range "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789;/?:@&=+$,-_.!~*'()#%" {
		t[c] = true
	}
	return
}()

// encodeURL koduje procentowo bajty, które nie mogą wystąpić w adresie wprost.
func encodeURL(s string) []byte {
	clean := true
	for i := 0; i < len(s); i++ {
		if !urlSafe[s[i]] {
			clean = false
			break
		}
	}
	if clean {
		return []byte(s)
	}
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(s)+16)
	for i := 0; i < len(s); i++ {
		if c := s[i]; urlSafe[c] {
			out = append(out, c)
		} else {
			out = append(out, '%', hex[c>>4], hex[c&15])
		}
	}
	return out
}

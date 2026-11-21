package gomdgo

import (
	"strings"
)

// Elementy składni odsyłaczy wspólne dla definicji (faza bloków) i odnośników w wierszu (faza druga).

func isSpaceOrTab(c byte) bool { return c == ' ' || c == '\t' }

// isPunct: znak interpunkcyjny ASCII, czyli taki, który można poprzedzić znakiem ucieczki.
func isPunct(c byte) bool {
	return c >= '!' && c <= '/' || c >= ':' && c <= '@' || c >= '[' && c <= '`' || c >= '{' && c <= '~'
}

// scanLinkLabel rozpoznaje etykietę [..] zaczynającą się w s[i] i zwraca pozycję za nawiasem zamykającym
// albo -1. Etykieta ma najwyżej 999 znaków, co najmniej jeden niebiały i żadnych nawiasów kwadratowych
// bez znaku ucieczki.
func scanLinkLabel(s []byte, i int) int {
	if i >= len(s) || s[i] != '[' {
		return -1
	}
	nonSpace := false
	for j := i + 1; j < len(s) && j-i <= 1000; j++ {
		switch c := s[j]; c {
		case '\\':
			if j+1 < len(s) && isPunct(s[j+1]) {
				j++
			}
			nonSpace = true
		case '[':
			return -1
		case ']':
			if !nonSpace {
				return -1
			}
			return j + 1
		case ' ', '\t', '\n':
		default:
			nonSpace = true
		}
	}
	return -1
}

// normalizeLabel sprowadza etykietę do postaci, w której porównuje się etykiety: bez białych znaków
// na brzegach, z seriami białych znaków zwiniętymi do spacji i po zwinięciu wielkości liter.
func normalizeLabel(label []byte) string {
	s := strings.Join(strings.Fields(string(label)), " ")
	s = strings.ToLower(strings.ToUpper(s))
	// Pełne zwijanie wielkości liter Unicode zamienia ß i ẞ na „ss”; ToLower tego nie robi.
	return strings.ReplaceAll(s, "ß", "ss")
}

// maxParenDepth ogranicza zagnieżdżenie nawiasów w celu odnośnika. Specyfikacja pozwala na limit (nie mniejszy
// niż 3), bo bez niego seria niedomkniętych nawiasów daje czas kwadratowy.
const maxParenDepth = 32

// scanLinkDestination rozpoznaje cel odnośnika zaczynający się w s[i]. Zwraca surowy cel (bez nawiasów
// kątowych) i pozycję za nim albo end = -1.
func scanLinkDestination(s []byte, i int) (dest []byte, end int) {
	if i >= len(s) {
		return nil, -1
	}
	if s[i] == '<' {
		for j := i + 1; j < len(s); j++ {
			switch s[j] {
			case '\\':
				if j+1 < len(s) && isPunct(s[j+1]) {
					j++
				}
			case '\n', '<':
				return nil, -1
			case '>':
				return s[i+1 : j], j + 1
			}
		}
		return nil, -1
	}
	depth := 0
	j := i
loop:
	for ; j < len(s); j++ {
		switch c := s[j]; {
		case c == '\\':
			if j+1 < len(s) && isPunct(s[j+1]) {
				j++
			}
		case c == '(':
			if depth++; depth > maxParenDepth {
				return nil, -1
			}
		case c == ')':
			if depth == 0 {
				break loop
			}
			depth--
		case c <= ' ' || c == 0x7f:
			break loop
		}
	}
	if j == i || depth != 0 {
		return nil, -1
	}
	return s[i:j], j
}

// scanLinkTitle rozpoznaje tytuł odnośnika zaczynający się w s[i]. Zwraca surowy tytuł (bez ograniczników)
// i pozycję za nim albo end = -1.
func scanLinkTitle(s []byte, i int) (title []byte, end int) {
	if i >= len(s) {
		return nil, -1
	}
	closer := s[i]
	switch closer {
	case '"', '\'':
	case '(':
		closer = ')'
	default:
		return nil, -1
	}
	for j := i + 1; j < len(s); j++ {
		switch c := s[j]; {
		case c == '\\':
			if j+1 < len(s) && isPunct(s[j+1]) {
				j++
			}
		case c == closer:
			return s[i+1 : j], j + 1
		case c == '(' && closer == ')':
			return nil, -1
		}
	}
	return nil, -1
}

// lineEnd zwraca pozycję za końcem wiersza, jeśli od s[i] do końca wiersza są tylko spacje i tabulatory;
// inaczej -1.
func lineEnd(s []byte, i int) int {
	for i < len(s) && isSpaceOrTab(s[i]) {
		i++
	}
	switch {
	case i == len(s):
		return i
	case s[i] == '\n':
		return i + 1
	}
	return -1
}

// scanRefDef rozpoznaje definicję odsyłacza na początku tekstu akapitu. used to liczba zużytych bajtów,
// zawsze do końca wiersza włącznie.
func scanRefDef(s []byte) (label string, ref reference, used int, ok bool) {
	i := scanLinkLabel(s, 0)
	if i < 0 || i >= len(s) || s[i] != ':' {
		return "", ref, 0, false
	}
	label = normalizeLabel(s[1 : i-1])
	i++
	for i < len(s) && isSpaceOrTab(s[i]) {
		i++
	}
	if i < len(s) && s[i] == '\n' {
		i++
		for i < len(s) && isSpaceOrTab(s[i]) {
			i++
		}
	}
	dest, i := scanLinkDestination(s, i)
	if i < 0 {
		return "", ref, 0, false
	}
	ref.dest = string(dest)

	// Tytuł jest oddzielony od celu białym znakiem i może zacząć się w następnym wierszu.
	afterDest := lineEnd(s, i)
	j := i
	for j < len(s) && isSpaceOrTab(s[j]) {
		j++
	}
	if afterDest >= 0 {
		j = afterDest
		for j < len(s) && isSpaceOrTab(s[j]) {
			j++
		}
	} else if j == i {
		return "", ref, 0, false
	}
	if title, end := scanLinkTitle(s, j); end >= 0 {
		if used = lineEnd(s, end); used >= 0 {
			ref.title, ref.hasTitle = string(title), true
			return label, ref, used, true
		}
	}
	// Bez poprawnego tytułu definicja jest ważna tylko wtedy, gdy cel kończy swój wiersz.
	if afterDest < 0 {
		return "", ref, 0, false
	}
	return label, ref, afterDest, true
}

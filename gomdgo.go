package gomdgo

import (
	"bytes"
)

// Render zamienia źródło Markdown na fragment HTML.
//
// ZAŚLEPKA etapu 0: parsera jeszcze nie ma. Niepuste wejście trafia w całości, escapowane, do jednego akapitu.
// Funkcja istnieje po to, żeby testy przykładów specyfikacji, fuzzing i tools/compare miały co wołać, a pomiar
// zerowy był uczciwy. Sygnatura jest tymczasowa: opcje i błędy punktów rozszerzeń dojdą w etapach 3 i 5.
func Render(src []byte) []byte {
	text := bytes.Trim(src, " \t\r\n")
	if len(text) == 0 {
		return nil
	}
	var out bytes.Buffer
	out.Grow(len(text) + 16)
	out.WriteString("<p>")
	escapeHTML(&out, text)
	out.WriteString("</p>\n")
	return out.Bytes()
}

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

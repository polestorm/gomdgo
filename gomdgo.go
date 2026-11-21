package gomdgo

import "bytes"

// Options to ustawienia renderowania.
//
// Kształt tymczasowy: ostateczne API (opcje, punkty rozszerzeń, błędy funkcji wywołującego) wymaga decyzji
// zapisanej jako pytanie 7 w docs/adr/open-questions.md.
type Options struct {
	// UnsafeHTML przepuszcza surowy HTML ze źródła bez zmian. Domyślnie HTML jest escapowany.
	UnsafeHTML bool
}

// Render zamienia źródło Markdown na fragment HTML przy ustawieniach domyślnych (bezpiecznych).
//
// Stan po etapie 2: pełny CommonMark 0.31.2. Surowy HTML ze źródła jest escapowany; adresy odnośników nie są
// jeszcze sprawdzane względem listy dozwolonych schematów (etap 3).
func Render(src []byte) []byte {
	return RenderWith(src, Options{})
}

// RenderWith zamienia źródło Markdown na fragment HTML z podanymi ustawieniami.
func RenderWith(src []byte, opts Options) []byte {
	r := renderer{doc: parse(src), opts: opts}
	r.render(r.doc.root)
	out := r.out.Bytes()
	if bytes.IndexByte(out, 0) >= 0 {
		// Specyfikacja: znak U+0000 jest ze względów bezpieczeństwa zamieniany na U+FFFD.
		out = bytes.ReplaceAll(out, []byte{0}, []byte("�"))
	}
	return out
}

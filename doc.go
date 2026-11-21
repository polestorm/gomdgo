// Package gomdgo zamienia Markdown (CommonMark z wybranymi rozszerzeniami GFM) na HTML.
//
// Najprostsze użycie to Render(src): CommonMark z rozszerzeniami GFM, wynik bezpieczny domyślnie; ustawienia przyjmuje New(Config), a Converter.Parse i Document.HTML
// rozdzielają parsowanie od zapisu (docs/adr/0005-api-shape.md).
//
// Pakiet jest w budowie (gotowy jest pełny CommonMark, rozszerzenia GFM i wynik bezpieczny domyślnie; punkty rozszerzeń jeszcze nie): plan etapów jest w docs/roadmap.md, a kryteria, po których zastąpi parser
// odniesienia, w docs/adr/0001-scope-and-acceptance.md.
package gomdgo

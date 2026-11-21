// Package gomdgo zamienia Markdown (CommonMark z rozszerzeniami GFM) na HTML.
//
// Najprostsze użycie to Render(src): CommonMark z rozszerzeniami GFM, wynik bezpieczny domyślnie.
// Ustawienia przyjmuje New(Config), a Converter.Parse i Document.HTML rozdzielają parsowanie od zapisu
// (docs/adr/0005-api-shape.md).
//
// Pakiet jest w budowie. Gotowe są: pełny CommonMark, rozszerzenia GFM, wynik bezpieczny domyślnie, punkty
// rozszerzeń, bloki uwag, przypisy i widok drzewa dla walidatorów (Document.Walk); przed nim utwardzanie
// i pomiar wydajności. Plan etapów jest
// w docs/roadmap.md, a kryteria, po których pakiet zastąpi parser odniesienia,
// w docs/adr/0001-scope-and-acceptance.md.
package gomdgo

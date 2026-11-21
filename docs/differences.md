# Różnice wobec parsera odniesienia

Każda różnica wyniku `gomdgo` i `goldmark` w korpusie różnicowym (`tools/compare`), z przykładem, punktem
specyfikacji i uzasadnieniem. Gdy `goldmark` odbiega od specyfikacji, wygrywa specyfikacja.

| # | Wejście | `gomdgo` | `goldmark` | Specyfikacja | Uzasadnienie |
|---|---|---|---|---|---|
| 1 | tryb bezpieczny (`tools/compare -safe`): odnośnik albo autolink ze schematem spoza `http`, `https`, `mailto`, np. `<irc://foo.bar:2233/baz>`, `<a+b+c:d>`, `<made-up-scheme://foo,bar>`, `<localhost:5001/foo>` (przykłady CommonMark 596, 598, 599, 601) | sama treść, bez `<a>` | `<a href="…">` z pełnym adresem | nie dotyczy: specyfikacja nie opisuje filtrowania adresów | lista dozwolonych zamiast listy zakazanych (ADR 0006, `SAFE-003`); przy przepuszczaniu wszystkich schematów (`AnyScheme`) wyniki są identyczne |

Różnice trybu bezpiecznego, których korpus dziś nie zawiera, a które wynikają z tych samych decyzji:
`goldmark` dla schematów `javascript:`, `vbscript:`, `file:`, `data:` zostawia odnośnik z pustym `href`, `gomdgo`
zostawia samą treść; `goldmark` w miejscu pominiętego surowego HTML wstawia komentarz `<!-- raw HTML omitted -->`,
`gomdgo` (`RawHTMLOmit`) nie wstawia nic (porównanie usuwa ten komentarz przed normalizacją).

## Zapowiedziane (odstępstwa odniesienia od specyfikacji, zanim powstał nasz kod)

Pomiar z 2026-11-21, `goldmark` v1.8.7 z `extension.GFM` i ustawieniami domyślnymi rozszerzeń, `tools/compare`.
Wejdą do tabeli powyżej, gdy etap 4 zaimplementuje te elementy.

| Przykład GFM | Wejście | `goldmark` | Specyfikacja |
|---|---|---|---|
| 199, Tables | kolumna z wyrównaniem (`:-:`, `---:`) | `<th style="text-align:center">` | `<th align="center">` |
| 653, Disallowed Raw HTML | `<title>`, `<style>`, `<xmp>` w surowym HTML | znaczniki przepuszczone bez zmian | `<` zamienione na `&lt;` w dziewięciu niedozwolonych znacznikach |

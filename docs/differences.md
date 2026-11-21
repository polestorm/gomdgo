# Różnice wobec parsera odniesienia

Każda różnica wyniku `gomdgo` i `goldmark` w korpusie różnicowym (`tools/compare`), z przykładem, punktem
specyfikacji i uzasadnieniem. Gdy `goldmark` odbiega od specyfikacji, wygrywa specyfikacja.

| # | Wejście | `gomdgo` | `goldmark` | Specyfikacja | Uzasadnienie |
|---|---|---|---|---|---|

## Zapowiedziane (odstępstwa odniesienia od specyfikacji, zanim powstał nasz kod)

Pomiar z 2026-11-21, `goldmark` v1.8.7 z `extension.GFM` i ustawieniami domyślnymi rozszerzeń, `tools/compare`.
Wejdą do tabeli powyżej, gdy etap 4 zaimplementuje te elementy.

| Przykład GFM | Wejście | `goldmark` | Specyfikacja |
|---|---|---|---|
| 199, Tables | kolumna z wyrównaniem (`:-:`, `---:`) | `<th style="text-align:center">` | `<th align="center">` |
| 653, Disallowed Raw HTML | `<title>`, `<style>`, `<xmp>` w surowym HTML | znaczniki przepuszczone bez zmian | `<` zamienione na `&lt;` w dziewięciu niedozwolonych znacznikach |

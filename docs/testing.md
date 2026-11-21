# Testy i porównania

## Warstwy

| Warstwa | Gdzie | Co sprawdza |
|---|---|---|
| przykłady specyfikacji | `testdata/spec/` (pytanie 1) | CommonMark 0.31.2 i rozszerzenia GFM 0.29-gfm: wynik HTML co do znaku, po normalizacji opisanej w specyfikacji |
| wymagania | `requirements/*.md`, `TestRequirementsCovered` | każde wymaganie ma test albo plik korpusu; każde odwołanie `REQ:` wskazuje istniejące wymaganie |
| korpus własny | `testdata/corpus/*.md` + `*.html` | dokumenty z polskimi znakami, rozszerzenia, punkty rozszerzeń; zmiana wzorca przez `go test -update` i przegląd różnicy |
| różnicowe | `tools/compare` (osobny moduł) | ten sam korpus przez `gomdgo` i `goldmark`; raport różnic po normalizacji HTML; każda różnica w `differences.md` |
| bezpieczeństwo | `TestSafeByDefault` | przy ustawieniach domyślnych w wyniku nie ma surowego HTML ze źródła ani adresów spoza listy schematów |
| powtarzalność | `TestDeterministic` | kilka renderowań daje ten sam wynik |
| odporność | `FuzzRender` | dowolne wejście: bez paniki, poprawny fragment HTML, czas ograniczony |
| patologiczne | `TestPathological` | głębokie zagnieżdżenia, długie serie ograniczników i nawiasów, wiele odsyłaczy: czas rośnie liniowo |
| wydajność | `BenchmarkCorpus` | porównanie z `goldmark` (kryterium K7) |

## Pomiary

Wyniki zapisujemy tutaj z datą, wersją i sposobem pomiaru.

| Data | Commit | CommonMark | GFM | Zgodność z `goldmark` | Uwagi |
|---|---|---|---|---|---|
| — | — | — | — | — | jeszcze bez kodu |

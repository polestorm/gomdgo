# Testy i porównania

## Warstwy

| Warstwa | Gdzie | Co sprawdza |
|---|---|---|
| przykłady specyfikacji | `testdata/spec/*.txt`, `TestSpecCommonMark`, `TestSpecGFM` | CommonMark 0.31.2 (652 przykłady) i rozdziały rozszerzeń GFM 0.29-gfm (24 przykłady): wynik HTML **co do bajta** |
| wymagania | `requirements/*.md`, `TestRequirementsCovered` | każde wymaganie ma test albo plik korpusu; każde odwołanie `REQ:` wskazuje istniejące wymaganie |
| korpus własny | `testdata/corpus/*.md` + `*.html`, `TestCorpus` | dokumenty z polskimi znakami, rozszerzenia, punkty rozszerzeń; zmiana wzorca przez `go test -update` i przegląd różnicy; wymagania wskazuje pierwszy wiersz `[//]: # (REQ: …)`, który jest definicją odsyłacza i nie daje wyniku |
| różnicowe | `tools/compare` (osobny moduł) | ten sam korpus przez `gomdgo` i `goldmark`; raport różnic po normalizacji HTML; każda różnica w `differences.md` |
| bezpieczeństwo | `TestSafeByDefault` | przy ustawieniach domyślnych w wyniku nie ma surowego HTML ze źródła ani adresów spoza listy schematów |
| powtarzalność | `TestDeterministic` | kilka renderowań daje ten sam wynik |
| odporność | `FuzzRender` | dowolne wejście: bez paniki, poprawny fragment HTML (`checkFragment`), wynik powtarzalny; ziarna: wszystkie przykłady obu specyfikacji |
| patologiczne | `TestPathological` | głębokie zagnieżdżenia, długie serie ograniczników i nawiasów, wiele odsyłaczy: czas rośnie liniowo (porównanie dwóch rozmiarów różniących się 16 razy) |
| drzewo i pozycje | `TestBlockPositions`, `TestReferenceDefinitions` | rodzaje węzłów z wierszem i kolumną (ADR 0003); zapisane definicje odsyłaczy |
| wydajność | `BenchmarkCorpus` | porównanie z `goldmark` (kryterium K7) |

Stan po etapie 1: nie ma jeszcze warstw „bezpieczeństwo” (etap 3) i „wydajność” (etap 6). `TestPathological`
obejmuje na razie strukturę bloków (15 przypadków) i rośnie z etapami.

`TestPathological` mierzy czas, więc zależy od maszyny: odśmiecanie pamięci jest na czas pomiaru wyłączone,
liczy się najlepszy z trzech przebiegów, a przekroczenie progu jest sprawdzane drugi raz. Pomiar kontrolny
z 2026-11-21 (160 tys. zagnieżdżonych cytatów w jednym wierszu): parsowanie 41–62 ns na poziom, renderowanie
147–205 ns na poziom przy rozmiarach od 10 tys. do 640 tys., czyli wzrost liniowy; różnica między małym a dużym
rozmiarem to koszt pamięci podręcznej procesora, nie algorytmu. `go test -short` pomija ten test.

## Przykłady specyfikacji i próg

Teksty specyfikacji leżą w `testdata/spec/` (pochodzenie i sumy: `testdata/spec/README.md`, decyzja: ADR 0002).
Przykłady wyciąga `internal/spectest`; jego test pilnuje liczby 652 dla CommonMark i 24 w pięciu rozdziałach
rozszerzeń GFM.

`testdata/spec/passing-commonmark.txt` i `passing-gfm.txt` to listy numerów zaliczonych przykładów:

- przykład spoza listy, który nie przechodzi, nie psuje testu (parser dopiero powstaje);
- przykład z listy, który przestał przechodzić, psuje test — liczba zaliczonych nie spada niezauważenie;
- przykład, który zaczął przechodzić, a nie ma go na liście, też psuje test, żeby lista nie zostawała w tyle.

```sh
go test -update .          # przepisuje obie listy; różnicę ogląda się w git diff
go test -v -run TestSpec . # liczby zaliczonych na rozdział
```

Usunięcie numeru z listy wymaga wpisu w opisie commita, dlaczego.

## Fuzzing

```sh
go test -run xxx -fuzz FuzzRender -fuzztime 10m .
```

Czas fuzzingu sumujemy w tabeli pomiarów (kryterium K5: 24 h łącznie). Wejścia, które wywołały błąd, Go zapisuje
w `testdata/fuzz/FuzzRender/`; trafiają do repozytorium jako testy regresji.

## Porównanie z parserem odniesienia

`tools/compare` to osobny moduł (`tools/compare/go.mod`, jedyne miejsce z zależnością od `goldmark`). Renderuje
każde wejście przez `gomdgo` i `goldmark`, normalizuje oba wyniki (`internal/htmlnorm`) i liczy zgodne.
Dla przykładów specyfikacji podaje też zgodność każdego z parserów ze specyfikacją.

```sh
cd tools/compare
go run . -spec ../../testdata/spec/commonmark-0.31.2.txt -spec ../../testdata/spec/gfm-0.29.txt \
    -out ../../out/compare ../../testdata/corpus/*.md
```

Normalizacja uznaje za nieistotne: białe znaki przy znacznikach blokowych i serie białych znaków poza `<pre>`,
kolejność i zapis atrybutów, `<br />` wobec `<br>`, znak zapisany wprost albo odwołaniem znakowym. Jej poprawność
potwierdza pomiar: `goldmark` po normalizacji zalicza 652 z 652 przykładów CommonMark.

Ustawienia odniesienia: `goldmark` z przepuszczaniem surowego HTML (`html.WithUnsafe()`), bo tego oczekują przykłady
specyfikacji; dla rozdziałów rozszerzeń GFM z `extension.GFM`. `gomdgo` jest wołany z `Options{UnsafeHTML: true}`,
żeby porównywać to samo. Porównanie w trybie domyślnym `gomdgo` (HTML escapowany) dojdzie w etapie 3.

## Pomiary

Wyniki zapisujemy tutaj z datą, wersją i sposobem pomiaru. CommonMark i GFM: `go test -v -run TestSpec .`,
porównanie co do bajta. Zgodność z `goldmark`: `tools/compare` na przykładach obu specyfikacji (676 wejść),
po normalizacji.

| Data | Commit | CommonMark | GFM | Zgodność z `goldmark` | Uwagi |
|---|---|---|---|---|---|
| 2026-11-21 | `cc7051d` (etap 0) | 80 / 652 | 1 / 24 | 88 / 676 (13,0%), `goldmark` v1.8.7 | zaślepka `Render` bez parsera: wynik przypadkowy, poziom zerowy miary; fuzzing 20 s (2,9 mln wywołań) bez błędu |
| 2026-11-21 | etap 1, przed commitem | 343 / 652 | 2 / 24 | 345 / 676 (51,0%) na przykładach specyfikacji; korpus własny 4 / 4; `goldmark` v1.8.7 | struktura bloków bez treści w wierszu; rozdziały Tabs, Precedence, Indented code blocks, Blank lines, Block quotes, List items i Lists zaliczone w całości (124 / 124), w pozostałych rozdziałach blokowych brakuje tylko przykładów wymagających treści w wierszu; fuzzing łącznie 125 s bez błędu |

Pomiar samego odniesienia, 2026-11-21, `goldmark` v1.8.7, po normalizacji: CommonMark 652 / 652; rozszerzenia GFM
22 / 24. Dwa odstępstwa od specyfikacji GFM są zapisane w `differences.md` jako zapowiedziane różnice.

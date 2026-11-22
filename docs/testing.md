# Testy i porównania

## Warstwy

| Warstwa | Gdzie | Co sprawdza |
|---|---|---|
| przykłady specyfikacji | `testdata/spec/*.txt`, `TestSpecCommonMark`, `TestSpecGFM` | CommonMark 0.31.2 (652 przykłady) i rozdziały rozszerzeń GFM 0.29-gfm (24 przykłady): wynik HTML **co do bajta** |
| wymagania | `requirements/*.md`, `TestRequirementsCovered` | każde wymaganie ma test albo plik korpusu; każde odwołanie `REQ:` wskazuje istniejące wymaganie |
| korpus własny | `testdata/corpus/*.md` + `*.html`, `TestCorpus` | dokumenty z polskimi znakami, rozszerzenia, punkty rozszerzeń; zmiana wzorca przez `go test -update` i przegląd różnicy; wymagania wskazuje pierwszy wiersz `[//]: # (REQ: …)`, który jest definicją odsyłacza i nie daje wyniku |
| różnicowe | `tools/compare` (osobny moduł) | ten sam korpus przez `gomdgo` i `goldmark`; raport różnic po normalizacji HTML; każda różnica w `differences.md` |
| bezpieczeństwo | `TestSafeByDefault`, `TestRawHTMLPolicy`, `TestSchemes` | przy ustawieniach domyślnych w wyniku są wyłącznie znaczniki i atrybuty z zamkniętej listy i adresy z dozwolonym schematem; sprawdzian `checkSafe` jest niezależny od renderera i działa na przykładach specyfikacji, korpusie, liście ataków oraz w `FuzzRender` |
| API | `TestConverterAPI`, `TestHeadingIDs` | dwa kroki `Parse` i `HTML`, powtarzalność i równoległość, identyfikatory nagłówków |
| punkty rozszerzeń | `TestCodeBlockRendererInput`, `TestCodeBlockRendererOutput`, `TestReferences`, `TestExtensionErrors` | co dostaje funkcja wywołującego, gdzie trafia jej wynik, bezpieczeństwo odnośników z odwołań, przerwanie na pierwszym błędzie ze stałym kodem i pozycją |
| rozszerzenia spoza GFM | `TestAlerts`, `TestFootnotes`, `testdata/corpus-extras/` | bloki uwag i przypisy według `EXT-005` i `EXT-006`; ten katalog korpusu ma własny konwerter i nie wchodzi do porównania z odniesieniem |
| publiczne drzewo | `TestWalk` | każdy rodzaj węzła z pozycją i cechami widziany przez `Document.Walk`; pary wejście–wyjście; równoległość z renderowaniem |
| powtarzalność | `TestDeterministic` | kilka renderowań daje ten sam wynik |
| odporność | `FuzzRender` | dowolne wejście: bez paniki, poprawny fragment HTML (`checkFragment`), wynik powtarzalny; ziarna: wszystkie przykłady obu specyfikacji |
| patologiczne | `TestPathological` | głębokie zagnieżdżenia, długie serie ograniczników i nawiasów, wiele odsyłaczy: czas rośnie liniowo (porównanie dwóch rozmiarów różniących się 16 razy) |
| drzewo i pozycje | `TestBlockPositions`, `TestInlinePositions`, `TestReferenceDefinitions` | rodzaje węzłów z wierszem i kolumną (ADR 0003); zapisane definicje odsyłaczy |
| wydajność | `BenchmarkCorpus` w `tools/compare` | porównanie z `goldmark` (kryterium K7) na przykładach specyfikacji, korpusie własnym i dużym dokumencie |
| pakiet dowodów | `tools/evidence` | wymagania → testy → wynik ostatniego przebiegu (kryterium K9) |

Stan po etapie 5 i pierwszej części etapu 6: istnieją wszystkie warstwy. `TestPathological`
obejmuje strukturę bloków (15 przypadków), treść w wierszu (42, lista `pathologicalInlines`) i rozszerzenia
GFM, odwołania w wierszu, bloki uwag i przypisy (35, lista `pathologicalGFM`); wszystkie są mierzone
z włączonymi wszystkimi rozszerzeniami.

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

Czas fuzzingu zapisujemy w tabeli pomiarów. Do kryterium K5 liczy się 24 h na kandydacie do wydania; każda zmiana
kodu biblioteki zeruje licznik, a wcześniejsze przebiegi są pomiarami pośrednimi (ADR 0012). Wejścia, które wywołały błąd, Go zapisuje
w `testdata/fuzz/FuzzRender/`; trafiają do repozytorium jako testy regresji.

Długi przebieg poza maszyną roboczą (Linux x86-64, bez instalowania Go): budujemy binarkę testową, kopiujemy ją
z katalogiem `testdata/` i uruchamiamy z katalogu, w którym leży `testdata/`:

```sh
GOOS=linux GOARCH=amd64 go test -c -fuzz FuzzRender -o gomdgo.test .
# na maszynie docelowej; katalog pamięci podręcznej musi być podany jawnie
nohup nice -n 10 ./gomdgo.test -test.run xxx -test.fuzz FuzzRender -test.fuzztime 24h -test.parallel 3 \
  -test.fuzzcachedir "$PWD/cache" > fuzz.log 2>&1 < /dev/null &
```

Wynik to ostatnie wiersze `fuzz.log` (`PASS` albo opis błędu); wejście, które wywołało błąd, binarka zapisuje
w `testdata/fuzz/FuzzRender/` obok siebie i trzeba je przenieść do repozytorium.

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

Dwa tryby porównania:

- zwykły: `goldmark` z `html.WithUnsafe()`, `gomdgo` z `Config{RawHTML: RawHTMLPass, AnyScheme: true}`, bo tego
  oczekują przykłady specyfikacji; dla przykładów CommonMark oba parsery są bez rozszerzeń (`Disable: GFM`),
  dla rozdziałów rozszerzeń GFM i dla korpusu własnego oba je mają. Odniesienie zapisuje wtedy wyrównanie kolumn w atrybucie `align`
  (`-refalign=attribute`, domyślnie); z `-refalign=style` odniesienie ma ustawienia domyślne, a `gomdgo`
  opcje zgodności `SingleTilde` i `TableAlignStyle` (ADR 0008);
- `-safe`: `goldmark` z ustawieniami domyślnymi, `gomdgo` z `Config{RawHTML: RawHTMLOmit}`; komentarz, który
  `goldmark` wstawia w miejscu pominiętego HTML, jest usuwany przed normalizacją. Zgodności ze specyfikacją
  w tym trybie się nie liczy.

Polityki domyślnej `gomdgo` (`RawHTMLEscape`) nie da się porównać z `goldmark`, bo ten jej nie ma.

## Pomiary

Wyniki zapisujemy tutaj z datą, wersją i sposobem pomiaru. CommonMark i GFM: `go test -v -run TestSpec .`,
porównanie co do bajta. Zgodność z `goldmark`: `tools/compare` na przykładach obu specyfikacji (676 wejść),
po normalizacji.

| Data | Commit | CommonMark | GFM | Zgodność z `goldmark` | Uwagi |
|---|---|---|---|---|---|
| 2026-11-21 | `cc7051d` (etap 0) | 80 / 652 | 1 / 24 | 88 / 676 (13,0%), `goldmark` v1.8.7 | zaślepka `Render` bez parsera: wynik przypadkowy, poziom zerowy miary; fuzzing 20 s (2,9 mln wywołań) bez błędu |
| 2026-11-21 | `41b1f76` (etap 1) | 343 / 652 | 2 / 24 | 345 / 676 (51,0%) na przykładach specyfikacji; korpus własny 4 / 4; `goldmark` v1.8.7 | struktura bloków bez treści w wierszu; rozdziały Tabs, Precedence, Indented code blocks, Blank lines, Block quotes, List items i Lists zaliczone w całości (124 / 124), w pozostałych rozdziałach blokowych brakuje tylko przykładów wymagających treści w wierszu; fuzzing łącznie 125 s bez błędu |
| 2026-11-21 | `a38663e` (etap 2) | **652 / 652** | 2 / 24 | 654 / 676 (96,7%) na przykładach specyfikacji, w tym CommonMark 652 / 652; korpus własny 6 / 6; `goldmark` v1.8.7 | treść w wierszu; K1 spełnione przy porównaniu co do bajta; brakujące 22 przykłady to rozszerzenia GFM (etap 4); fuzzing łącznie 215 s (ok. 5 mln wywołań w ostatnim przebiegu) bez błędu |
| 2026-11-21 | `29b7f7d` (etap 3) | 652 / 652 | 2 / 24 | tryb zwykły: 654 / 676 (96,7%), CommonMark 652 / 652, korpus 6 / 6; tryb `-safe`: CommonMark 648 / 652 (99,4%), korpus 6 / 6; `goldmark` v1.8.7 | API z ADR 0005, polityka surowego HTML, lista schematów, identyfikatory nagłówków, niepoprawny UTF-8 (ADR 0004); 4 różnice trybu bezpiecznego to wpis 1 w `differences.md`; fuzzing łącznie 315 s, od tego etapu ze sprawdzianem bezpieczeństwa (7 mln wywołań w ostatnim przebiegu), bez błędu |
| 2026-11-21 | `6189ebb` (etap 4) | 652 / 652 | **24 / 24** | tryb zwykły, 683 wejścia (676 przykładów + 7 plików korpusu): 681 (99,7%) z `-refalign=attribute`; 681 (99,7%) z `-refalign=style`, czyli odniesienie domyślne wobec `gomdgo` z opcjami zgodności (bez nich 680); tryb `-safe`: 677 (99,1%); `goldmark` v1.8.7 | lista schematów z `tel`; rozszerzenia GFM (domyślnie włączone, `Config.Disable` je wyłącza), opcje zgodności i limit dopełniania tabel (ADR 0007, 0008); CommonMark mierzony z `Disable: GFM`; K2 spełnione przy porównaniu co do bajta; wszystkie różnice opisane w `differences.md` (wiersze 1–8); `TestPathological` ma 80 przypadków; fuzzing łącznie 465 s, bez błędu |
| 2026-11-21 | `6ebd995` (etap 5) | 652 / 652 | 24 / 24 | bez zmian: 681 / 683 (99,7%); tryb `-safe` 677 / 683 | punkty rozszerzeń (ADR 0009), publiczne drzewo `Document.Walk` (ADR 0010), bloki uwag i przypisy (ADR 0011); żadne z nich nie zmienia wyniku przy ustawieniach domyślnych; `TestPathological` ma 92 przypadki; fuzzing łącznie 615 s, także z punktami rozszerzeń, blokami uwag i przypisami, bez błędu |
| 2026-11-21 | `3acafeb` (etap 6, część 1) | 652 / 652 | 24 / 24 | przykłady specyfikacji i korpus własny: 681 / 683 (99,7%), tryb `-safe` 677 / 683; dokumenty rzeczywiste: 122 / 123 (99,2%), tryb `-safe` 123 / 123 | K7 zmierzone: stosunek czasu 0,42–0,72 wobec `goldmark` (sekcja „Wydajność”); K9: generator pakietu dowodów; komórki dopełnione bez wyrównania (ADR 0007); fuzzing: przebiegi 30 min (125,9 mln wywołań) i 60 min (127,1 mln wywołań, 2026-11-22) na 4 rdzeniach oraz 8 h (3366,9 mln wywołań, 2026-11-22, osobna sesja), bez błędu; łącznie 34 815 s (9 h 40 min), pomiar pośredni (ADR 0012) |

Pomiar samego odniesienia, 2026-11-21, `goldmark` v1.8.7, po normalizacji: CommonMark 652 / 652; rozszerzenia GFM
22 / 24. Dwa odstępstwa od specyfikacji GFM są zapisane w `differences.md` jako zapowiedziane różnice.

## Dokumenty rzeczywiste

Korpus własny jest pisany pod wymagania, więc nie zastąpi dokumentów pisanych przez ludzi w innym celu. Dopóki
nie ma dokumentów odbiorców, rolę tę pełnią pliki Markdown z projektów rodziny: `../gopdfgo`, `../gomermaidgo`
i samo `gomdgo` (dokumentacja, decyzje, wymagania, przeglądy), bez katalogów `testdata`, `tests`, `out` i kopii
roboczych, po odrzuceniu plików o identycznej treści. Pliki nie są kopiowane do repozytorium; listę podaje się
do `tools/compare` jako argumenty.

Pomiar z 2026-11-21, `goldmark` v1.8.7, 123 dokumenty, 948 kB:

| Tryb | Identyczne | Różnice |
|---|---|---|
| zwykły (`-refalign=attribute`) | 122 (99,2%) | `0033-tagged-pdf.md`: `<title>` w zdaniu, filtr znaczników (wiersz 3 w `differences.md`) |
| zwykły, odniesienie domyślne i opcje zgodności (`-refalign=style`) | 122 (99,2%) | ta sama |
| bezpieczny (`-safe`) | 123 (100%) | brak |

Przed tym pomiarem identycznych było 121: drugą różnicą było wyrównanie w dopełnionych komórkach tabeli
(README jednego z projektów). Specyfikacja tego nie rozstrzyga, więc przyjęliśmy zachowanie odniesienia
(wiersz 4 w `differences.md`, ADR 0007). Pomiar nie wykrył błędu w `gomdgo`.

## Wydajność (K7)

```sh
cd tools/compare
go test -run xxx -bench BenchmarkCorpus -benchtime 2s -count 5 .
GOMDGO_BENCH_FILES="$(cat lista-plików.txt)" go test -run xxx -bench BenchmarkCorpus .   # zestaw extra
```

Oba parsery mają ustawienia z porównania różnicowego. Wynik podajemy jako **przewagę `gomdgo` w procentach**:
`(goldmark − gomdgo) / goldmark × 100%` dla miar, w których mniej znaczy lepiej (czas, pamięć, liczba alokacji,
rozmiar wyniku, rozmiar kodu). Wartość dodatnia oznacza, że `gomdgo` wypada lepiej, ujemna, że gorzej. Tak samo
zapisujemy porównania w `README.md`.

Pomiar z 2026-11-22, commit `d8ed0d5` (kod biblioteki jak w `3acafeb`), Apple M-series (10 rdzeni), Go 1.26.6,
`goldmark` v1.8.7, `-benchtime 2s -count 5`, mediana, średnie obciążenie maszyny 3–5:

| Zestaw | Miara | `gomdgo` | `goldmark` | Przewaga `gomdgo` |
|---|---|---:|---:|---:|
| spec: 652 przykłady CommonMark, osobno | czas | 0,52 ms | 1,24 ms | **+58,0%** |
| | pamięć | 1,53 MB | 4,42 MB | +65,4% |
| | alokacje | 12 742 | 18 450 | +30,9% |
| | rozmiar wyniku | 27 572 B | 27 444 B | −0,5% |
| corpus: 7 plików korpusu własnego, osobno | czas | 0,091 ms | 0,180 ms | **+49,6%** |
| | pamięć | 0,290 MB | 0,265 MB | −9,2% |
| | alokacje | 1 550 | 1 753 | +11,6% |
| | rozmiar wyniku | 8 725 B | 8 888 B | +1,8% |
| large: jeden dokument 258 kB | czas | 4,98 ms | 8,44 ms | **+40,9%** |
| | pamięć | 13,45 MB | 10,52 MB | −27,9% |
| | alokacje | 65 898 | 74 952 | +12,1% |
| | rozmiar wyniku | 395 400 B | 403 550 B | +2,0% |
| extra: 123 dokumenty rzeczywiste, ok. 1 MB | czas | 9,36 ms | 15,91 ms | **+41,2%** |
| | pamięć | 21,58 MB | 18,58 MB | −16,1% |
| | alokacje | 78 223 | 101 696 | +23,1% |
| | rozmiar wyniku | 1 202 562 B | 1 204 674 B | +0,2% |

K7 wymaga czasu najwyżej 2× `goldmark`, czyli przewagi nie mniejszej niż −100%; we wszystkich zestawach przewaga
jest dodatnia. W trzech przebiegach z tego dnia przewaga czasu wahała się w granicach: spec 58–61%, corpus 48–51%,
large 39–43%, extra 41–42%; pamięć, alokacje i rozmiar wyniku są powtarzalne. Słabą stroną jest pamięć na dużych
dokumentach (o 16–28% więcej); kryterium tego nie obejmuje.

Rozmiar wyniku to miara `out-bytes` w wyniku wzorca: łączna liczba bajtów HTML dla zestawu. W zestawach z tabelami
część różnicy bierze się z zapisu wyrównania (`gomdgo` atrybutem `align`, odniesienie atrybutem `style`,
`differences.md`); pozostałych źródeł różnicy w bajtach nie rozbieraliśmy, bo porównanie różnicowe działa na
wyniku znormalizowanym.

Pomiar z 2026-11-21 był robiony przy fuzzingu na 4 rdzeniach w tle i dał przewagę czasu 58%, 42%, 28% i 35%;
obciążenie zaniżało wynik `gomdgo`. Dokumenty rzeczywiste nie są w repozytorium i zmieniają się, stąd zmienny
rozmiar zestawu extra.

Rozmiar kodu, pomiar z 2026-11-22 (wiersze i bajty plików `*.go` bez testów; przyrost binarki to program
„stdin → HTML” z GFM, `-ldflags="-s -w"`, macOS arm64, minus ten sam program bez parsera):

| Miara | `gomdgo` | `goldmark` | Przewaga `gomdgo` |
|---|---:|---:|---:|
| wiersze kodu biblioteki | 3 561 | 13 293 | +73,2% |
| bajty kodu biblioteki | 98 765 | 461 450 | +78,6% |
| pliki | 14 | 53 | +73,6% |
| przyrost binarki | 466 kB | 1 289 kB | +63,8% |

## Pakiet dowodów (K9)

```sh
go run ./tools/evidence -o out/evidence.md
```

Narzędzie czyta `requirements/`, odwołania `REQ:` w testach i plikach korpusu, uruchamia `go test -count=1 -json
./...` i zapisuje dla każdego wymagania jego treść, testy i wynik, z datą, commitem i wersją Go. Kończy się
kodem błędu, gdy któreś wymaganie nie ma testu albo test nie przeszedł. Wynik jest poza repozytorium (`out/`),
bo opisuje konkretny przebieg; generuje się go dla commita, który ma być przekazany odbiorcy.

Przebieg z 2026-11-21: 54 wymagania, wszystkie z dowodem; 44 testy, wszystkie przeszły; K1 652 z 652,
K2 24 z 24.

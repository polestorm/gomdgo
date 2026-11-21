# gomdgo — instrukcje dla Claude Code

Projekt: parser i renderer Markdown → HTML w czystym Go. Cele: zgodność z CommonMark i wybranymi rozszerzeniami GFM,
zero zależności poza biblioteką standardową, wynik bezpieczny z konstrukcji (bez osobnego sanityzatora), powtarzalny
co do bajta, czas liniowy także dla danych złośliwych, mały kod, który da się przeczytać i zwalidować. Projekt jest
agnostyczny wobec odbiorców.

Zacznij od `README.md`, potem `docs/roadmap.md`, `docs/adr/0001-scope-and-acceptance.md` i `docs/provenance.md`.

## Po co i w jakim układzie (stan 2026-11-21)

- Typowy łańcuch u odbiorcy: Markdown → `gomdgo` → HTML (z SVG diagramów od funkcji wywołującego, np.
  `gomermaidgo.Render`) → konwerter HTML + CSS + SVG → PDF (np. `gopdfgo`). `gomdgo` nie wie nic o PDF ani
  o diagramach: zna tylko punkt rozszerzenia „blok kodu z językiem X” i odwołania w wierszu.
- Dziś odbiorcy używają `goldmark` za własnym interfejsem. `gomdgo` zastąpi go za tym interfejsem dopiero po spełnieniu
  **wszystkich** kryteriów K1–K9 (ADR 0001). Nie ma terminu, jest próg jakości.
- Rodzina: `../gopdfgo` i `../gomermaidgo` powstają tak samo (od zera, czarna skrzynka jako odniesienie, korpus,
  wymagania, datowane pomiary). Wzoruj układ testów i dokumentów na `../gomermaidgo/docs/testing.md`
  i `../gomermaidgo/requirements/`. Kodu innych parserów Markdown nie czytaj (`docs/provenance.md`).

## Kryteria zastąpienia (skrót ADR 0001)

| # | Kryterium | Gdzie mierzone |
|---|---|---|
| K1 | CommonMark 0.31.2: 652/652 przykładów | test przykładów specyfikacji |
| K2 | GFM 0.29-gfm: wszystkie przykłady rozszerzeń | test przykładów specyfikacji |
| K3 | ≥ 99,5% wyników identycznych z `goldmark` po normalizacji; każda różnica w `docs/differences.md` | `tools/compare` (osobny moduł) |
| K4 | każde wymaganie z `requirements/` ma test; testy przechodzą | `TestRequirementsCovered` |
| K5 | fuzzing ≥ 24 h łącznie bez paniki; przypadki patologiczne w czasie liniowym | `FuzzRender`, `TestPathological` |
| K6 | domyślnie bez surowego HTML i niedozwolonych schematów adresów | `TestSafeByDefault` |
| K7 | nie wolniej niż 2× `goldmark` | `BenchmarkCorpus` |
| K8 | pliki wzorcowe odbiorcy przechodzą bez zmian po przepięciu | u odbiorcy, etap 7 |
| K9 | pakiet dowodów: wymagania → testy → wynik ostatniego przebiegu | etap 7 |

Po każdym etapie dopisz wiersz do tabeli pomiarów w `docs/testing.md` (data, commit, liczby K1–K3) i zaktualizuj
kolumnę „Stan” w `docs/roadmap.md`.

## Jak pracujemy

- etapy w kolejności z `docs/roadmap.md`; **pierwszy krok: etap 0** (wczytanie przykładów specyfikacji, test pokrycia
  wymagań, szkielet fuzzingu, `tools/compare` z własnym `go.mod`, w którym wolno zależeć od `goldmark`);
- przed etapem 0 potrzebna decyzja właściciela w pytaniu 1 (`docs/adr/open-questions.md`: przykłady specyfikacji
  na CC BY-SA 4.0 w repozytorium czy pobierane); pytanie 2 (licencja) nie blokuje pracy;
- wymagania piszemy razem z etapem, w plikach według prefiksów z `requirements/README.md`;
- najpierw test (przykład specyfikacji albo plik korpusu), potem kod; liczba zaliczonych przykładów nie spada między
  commitami bez wpisu, dlaczego;
- repozytorium `github.com/polestorm/gomdgo` (prywatne); commit i push tylko na prośbę właściciela.

Zasady:
- **czysty rodowód** (`docs/provenance.md`): piszemy ze specyfikacji CommonMark i GFM. Nie czytaj, nie kopiuj i nie
  przenoś kodu innych parserów Markdown (goldmark, blackfriday, gomarkdown, cmark, commonmark.js, markdown-it, pulldown-cmark).
  Wolno uruchamiać je jako czarną skrzynkę, żeby porównać wynik;
- **zero zależności**: `go.mod` w katalogu głównym nie ma sekcji `require`; porównanie z parserem odniesienia żyje
  w osobnym module `tools/compare` i nie jest importowane przez bibliotekę;
- wzorcem są specyfikacje, nie zachowanie parsera odniesienia: gdy goldmark różni się od specyfikacji, wygrywa
  specyfikacja, a różnicę wpisujemy do `docs/differences.md`;
- nazwy plików, pakietów i identyfikatorów po angielsku; komentarze i dokumentacja po polsku; `README.md` po angielsku;
  komunikaty błędów biblioteki po angielsku, ze stałym kodem do tłumaczenia przez wywołującego;
- każdy obsługiwany element składni ma wymaganie w `requirements/` i test, który je wskazuje;
- każda decyzja = plik w `docs/adr/`; pytania bez decyzji w `docs/adr/open-questions.md`;
- liczby (zaliczone przykłady, zgodność z odniesieniem) są datowane i podpisane sposobem pomiaru;
- opisy commitów po angielsku; przed commitem muszą przechodzić `go vet ./...` i `go test ./...`.

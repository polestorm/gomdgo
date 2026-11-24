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
| K5 | fuzzing ≥ 24 h bez paniki na kandydacie do wydania (ADR 0012); przypadki patologiczne w czasie liniowym | `FuzzRender`, `TestPathological` |
| K6 | domyślnie bez surowego HTML i niedozwolonych schematów adresów | `TestSafeByDefault` |
| K7 | nie wolniej niż 2× `goldmark` | `BenchmarkCorpus` w `tools/compare` |
| K8 | pliki wzorcowe odbiorcy przechodzą bez zmian po przepięciu | u odbiorcy, etap 7 |
| K9 | pakiet dowodów: wymagania → testy → wynik ostatniego przebiegu | `tools/evidence`, przekazanie w etapie 7 |

Po każdym etapie dopisz wiersz do tabeli pomiarów w `docs/testing.md` (data, commit, liczby K1–K3) i zaktualizuj
kolumnę „Stan” w `docs/roadmap.md`.

## Jak pracujemy

- etapy w kolejności z `docs/roadmap.md`; etapy 0–4 są gotowe (K1: 652/652; K2: 24/24; K6: `TestSafeByDefault`),
  etap 5 też: punkty rozszerzeń (`extend.go`, ADR 0009), `Document.Walk` (`walk.go`, ADR 0010), bloki uwag
  i przypisy (`extras.go`, ADR 0011); **etap 6 w toku**: K7 zmierzone (`BenchmarkCorpus`), pakiet dowodów
  generuje `tools/evidence`, korpus dokumentów rzeczywistych opisuje `docs/testing.md`; K5 liczymy na kandydacie
  do wydania (ADR 0012): 24 h bez błędu na kodzie biblioteki z `9c8577b` (2026-11-23), potem zmiana węzła drzewa
  (`blockData` w `node.go`, mniej pamięci) i nowy przebieg 24 h na Synology od 2026-11-24 19:53, wynik w
  `~/gomdgo-fuzz/fuzz.log` na serwerze, do wpisania w `docs/testing.md`; każda zmiana plików `*.go`
  pakietu `markdown/` poza testami zeruje licznik, więc poprawki po dokumentach odbiorcy (K8) zbieraj w paczki;
  API według ADR 0005 (`New(Config).Parse(src).HTML()`), polityka bezpieczeństwa według ADR 0006, rozszerzenia
  według ADR 0007 i 0008 (`gfm.go`): GFM domyślnie włączone, `Config.Disable` wyłącza;
- układ repozytorium (ADR 0016): w katalogu głównym tylko `README.md`, `LICENSE`, `NOTICE`, `go.mod`; biblioteka to
  pakiet `markdown` w `markdown/` (import `github.com/polestorm/gomdgo/markdown`), ten plik leży w `.claude/`;
  wydana wersja: v0.1.0 (2026-11-24); `go.mod` wymaga Go 1.22 (sprawdzone toolchainem 1.22.12), a CI
  (`.github/workflows/test.yml`) puszcza testy na Go 1.22 i bieżącym, na Linuksie i macOS, także w obu
  zagnieżdżonych modułach;
- układ kodu (wszystko w `markdown/`): `block.go` (faza 1, wiersz po wierszu), `inline.go` (faza 2, stos ograniczników i stos nawiasów),
  `links.go`, `htmltag.go`, `escape.go` (skanery wspólne), `render.go` (przejście iteracyjne); każde wyszukiwanie
  w przód, które może zawieść, musi zapamiętać niepowodzenie, a każdy nowy przypadek złośliwy trafia do
  `TestPathological`;
- `cmd/gomdgo` to proste polecenie Markdown → HTML (ADR 0014): flagi odpowiadają polom `Config`, nie uruchamia
  innych programów; kod poleceń jest w `internal/cli`; `cmd/gomdgo-mermaid` to osobny moduł, który dokłada
  diagramy `gomermaidgo` przez `Config.CodeBlocks` (ADR 0015; testy puszczaj w jego katalogu); łączenie z `gomermaidgo` i `gopdfgo` należy do czwartego projektu, `../gopressgo`
  (założony 2026-11-23). Jest pierwszym odbiorcą `gomdgo`: od niego przyjdą pliki wzorcowe do K8, a uwagi
  przekazane z tej strony leżą w `../gopressgo/docs/notes-from-gomdgo.md`;
- miara postępu: `go test -v -run TestSpec ./markdown`; listy zaliczonych przykładów `testdata/spec/passing-*.txt` przepisuje
  `go test -update ./markdown` (opis w `docs/testing.md`); porównanie z odniesieniem: `tools/compare`;
- pytań otwartych nie ma (`docs/adr/open-questions.md`; wybory robocze potwierdzone 2026-11-22); licencja to Apache 2.0 (ADR 0013), a porównanie z `goldmark` w `README.md` trzeba
  uaktualniać razem z liczbami w `docs/testing.md`;
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
- porównania z odniesieniem (rozmiar, szybkość, pamięć, rozmiar wyniku) podajemy jako przewagę `gomdgo`
  w procentach, `(goldmark − gomdgo) / goldmark × 100%`; wartość ujemna oznacza, że `gomdgo` wypada gorzej,
  i też ją zapisujemy (`docs/testing.md`, `README.md`);
- opisy commitów po angielsku; przed commitem muszą przechodzić `go vet ./...` i `go test ./...`;
- opis commita kończy się jedną linią `Made with the help of Claude Code` (po pustym wierszu); nie dopisuj
  `Co-Authored-By` ani `Claude-Session` (decyzja właściciela z 2026-11-24, historia przepisana tego dnia).

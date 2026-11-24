# 0015 Jedna binarka z diagramami: osobny moduł `cmd/gomdgo-mermaid`

Status: przyjęta · 2026-11-24 (polecenie właściciela; uzupełnia ADR 0014)

## Kontekst

Właściciel chce mieć jeden program, który zamienia Markdown z blokami `mermaid` na HTML, i chce, żeby dodanie
diagramów było opcjonalne przy kompilacji. Biblioteka `gomermaidgo` jest gotowa (`mermaid.Render`), a `gomdgo`
ma punkt rozszerzenia `Config.CodeBlocks`. ADR 0014 odsyła łączenie bibliotek do osobnego projektu
(`gopressgo`), ale tamten program zawiera też `gopdfgo` i ma ponad 20 MB.

Rozważone sposoby:

1. **Znacznik budowania w `cmd/gomdgo`** (`go build -tags mermaid`, plik z `//go:build mermaid`). Go nie ma
   zależności opcjonalnych: import pod znacznikiem i tak wymaga wpisu `require` w `go.mod` katalogu głównego.
   Każdy, kto importuje `gomdgo`, dostałby `gomermaidgo` w swoim grafie modułów. Odrzucone, bo łamie zasadę
   zera zależności. (`gomermaidgo` używa znacznika `nopdf` do wyłączania formatów, ale tam oba pakiety są
   w jednym module.)
2. **Wtyczki ładowane w czasie działania** (`plugin`, zewnętrzny program). Odrzucone: `plugin` nie działa na
   wszystkich systemach, a uruchamianie programów odrzuca ADR 0014.
3. **Potok tekstowy** (`gomermaidgo/markdown.Embed` przed `gomdgo`). Odrzucone w ADR 0014: wymaga
   przepuszczania całego surowego HTML.
4. **Osobny moduł w tym repozytorium**, z własnym `go.mod`, tak jak `tools/compare`. Wybrane.

## Decyzja

1. Wspólna część poleceń (flagi, odczyt, zapis) mieszka w pakiecie `internal/cli`: `cli.Run(args, stdin,
   stdout, stderr, cli.Options)`. `Options` niesie nazwę polecenia, funkcje dla bloków kodu
   (`Config.CodeBlocks`) i HTML dopisywany do `<head>` przy `-standalone`.
2. `cmd/gomdgo` woła `cli.Run` z pustymi `Options` i pozostaje bez zależności.
3. `cmd/gomdgo-mermaid` to osobny moduł (`cmd/gomdgo-mermaid/go.mod`) z zależnością od `gomermaidgo`
   (przypiętą do wydanej wersji) i z `replace github.com/polestorm/gomdgo => ../..`. Rejestruje jedną funkcję:
   blok `mermaid` → `mermaid.Render` → `<figure class="diagram">` z SVG. Przy `-standalone` dopisuje arkusz
   stylów `gomermaidgo/markdown.CSS`.
4. „Przełącznikiem kompilacji” jest wybór katalogu: `go build ./cmd/gomdgo` daje program bez diagramów,
   `cd cmd/gomdgo-mermaid && go build .` daje program z diagramami. Flagi obu są takie same.
5. Błąd w diagramie przerywa pracę kodem wyjścia 1, z wierszem i kolumną bloku w dokumencie. Wynik niepełny nie
   powstaje.
6. Kolejna biblioteka (inny język bloku kodu) to kolejny wpis w mapie `CodeBlocks` w osobnym module albo
   własny plik `main.go` u odbiorcy, kilkanaście wierszy na publicznym API (przykład w `README.md`).
   `internal/cli` nie jest publicznym API.
7. Pełny łańcuch do PDF pozostaje w `gopressgo`.
8. `internal/cli` i `cmd/` leżą poza kodem biblioteki w rozumieniu ADR 0012: zmiany nie zerują licznika K5.

## Pomiar (2026-11-24, macOS arm64, `-ldflags="-s -w"`, `gomermaidgo` v0.1.0)

| Program | Rozmiar |
|---|---:|
| `gomdgo` | 2,0 MB |
| `gomdgo-mermaid` | 6,3 MB |

## Konsekwencje

+ `go.mod` katalogu głównego nadal nie ma sekcji `require`; biblioteka nie wie nic o diagramach;
+ surowy HTML dokumentu pozostaje zapisany jako tekst także w programie z diagramami;
+ oba polecenia mają ten sam kod flag, więc nie rozjadą się;
− `go vet ./...` i `go test ./...` z katalogu głównego nie wchodzą do zagnieżdżonego modułu; trzeba je puścić
  osobno w `cmd/gomdgo-mermaid` (tak jak w `tools/compare`);
− przez dyrektywę `replace` nie działa `go install …/cmd/gomdgo-mermaid@latest`; program buduje się z klonu
  repozytorium. Po oznaczeniu wersji `gomdgo` można `replace` zastąpić numerem wersji;
− `gomdgo-mermaid` nie ma flag `gomermaidgo` (motyw, krój, zwijanie); diagramy powstają z ustawieniami
  domyślnymi. Flagi dojdą, gdy będą potrzebne.

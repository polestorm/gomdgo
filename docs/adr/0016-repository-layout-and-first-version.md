# 0016 Układ repozytorium przed publikacją i wersja v0.1.0

Status: przyjęta · 2026-11-24 (polecenie właściciela)

## Kontekst

Repozytorium ma zostać upublicznione, tak jak `gomermaidgo` i `gopdfgo`. W katalogu głównym leżało 30 plików:
kod biblioteki, testy i instrukcje. Projekty rodziny mają w katalogu głównym cztery pliki, a bibliotekę
w podkatalogu (`gomermaidgo/mermaid`, `gopdfgo/htmlpdf`); oba mają wersję v0.1.0.

## Decyzja

1. W katalogu głównym zostają `README.md`, `LICENSE`, `NOTICE` i `go.mod` (oraz `.gitignore`).
2. Biblioteka to pakiet `markdown` w katalogu `markdown/`; importuje się ją jako
   `github.com/polestorm/gomdgo/markdown` i woła `markdown.Render(src)`. Ścieżka modułu się nie zmienia.
   Komunikaty błędów zachowują przedrostek `gomdgo:`.
3. Instrukcje dla Claude Code leżą w `.claude/CLAUDE.md`; reszta `.claude/` jest ignorowana.
4. Dane testowe (`testdata/`) i wymagania (`requirements/`) zostają w katalogu głównym; testy pakietu czytają je
   przez `../`.
5. Skompilowane programy nie trafiają do repozytorium. Plik `tools/compare/compare` (4,5 MB, dodany omyłkowo
   na etapie 3) został usunięty z drzewa i, poleceniem właściciela z 2026-11-24, z całej historii
   (`git filter-repo`). Zmieniło to identyfikatory dziesięciu commitów; identyfikatory w dokumentach są już nowe,
   a znacznik v0.1.0 wskazuje commit po przepisaniu. Treść plików w każdym commicie pozostała bez zmian.
   Tego samego dnia przepisane zostały opisy wszystkich commitów: linie `Co-Authored-By` i `Claude-Session`
   zastąpiła jedna, `Made with the help of Claude Code`; identyfikatory commitów zmieniły się po raz drugi.
6. Pierwsza wersja to v0.1.0. Oznacza stan „funkcje gotowe, utwardzanie w toku”: kryteria K5 (24 h fuzzingu
   na tym kodzie) i K8 (pliki wzorcowe odbiorcy) nie są jeszcze spełnione, więc wersja 0.x nie obiecuje
   stałości API.
7. Przeniesienie plików nie zmienia treści kodu biblioteki poza nazwą pakietu, więc nie zeruje licznika K5
   (ADR 0012).
8. Najniższa obsługiwana wersja Go to 1.22, tak jak w `gomermaidgo`: wszystkie trzy pliki `go.mod` mają
   `go 1.22`. Sprawdzone 2026-11-24 toolchainem 1.22.12 (`go vet` i `go test` w trzech modułach); Go 1.21 nie
   było sprawdzane. CI (`.github/workflows/test.yml`) powtarza to na Go 1.22 i bieżącym, na Linuksie i macOS.

## Przegląd przed publikacją (2026-11-24)

Przeszukane pliki śledzone i opisy commitów: adresy IP, nazwy hostów i użytkowników, ścieżki katalogów domowych,
adresy e-mail, hasła, klucze, żetony. W plikach tekstowych nic takiego nie ma; adresy e-mail w testach są
w domenie `example.com`. Znalezione:

- `tools/compare/compare`: binarka z wbudowanymi ścieżkami katalogu domowego autora, usunięta z historii (punkt 5);
- opisy commitów zawierają adres e-mail autora i odnośniki do sesji Claude Code, tak jak w opublikowanych już
  `gomermaidgo` i `gopdfgo`;
- `docs/testing.md` podaje model serwera, na którym biegnie fuzzing, bez adresu i nazwy użytkownika.

## Konsekwencje

+ katalog główny jest czytelny, układ zgodny z rodziną;
− zmienia się ścieżka importu: odbiorcy przypięci do starszego commita (`gopressgo`, przypięty do commita sprzed przepisania historii) muszą przy
  podbiciu wersji zmienić import z `github.com/polestorm/gomdgo` na `github.com/polestorm/gomdgo/markdown`;
− dokumenty sprzed 2026-11-24 podają nazwy plików bez katalogu `markdown/`.

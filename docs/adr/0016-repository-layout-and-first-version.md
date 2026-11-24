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
   w commicie `29b7f7d`) został usunięty z drzewa; zostaje w historii, dopóki właściciel nie zdecyduje o jej
   przepisaniu.
6. Pierwsza wersja to v0.1.0. Oznacza stan „funkcje gotowe, utwardzanie w toku”: kryteria K5 (24 h fuzzingu
   na tym kodzie) i K8 (pliki wzorcowe odbiorcy) nie są jeszcze spełnione, więc wersja 0.x nie obiecuje
   stałości API.
7. Przeniesienie plików nie zmienia treści kodu biblioteki poza nazwą pakietu, więc nie zeruje licznika K5
   (ADR 0012).

## Przegląd przed publikacją (2026-11-24)

Przeszukane pliki śledzone i opisy commitów: adresy IP, nazwy hostów i użytkowników, ścieżki katalogów domowych,
adresy e-mail, hasła, klucze, żetony. W plikach tekstowych nic takiego nie ma; adresy e-mail w testach są
w domenie `example.com`. Znalezione:

- `tools/compare/compare`: binarka z wbudowanymi ścieżkami katalogu domowego autora (punkt 5);
- opisy commitów zawierają adres e-mail autora i odnośniki do sesji Claude Code, tak jak w opublikowanych już
  `gomermaidgo` i `gopdfgo`;
- `docs/testing.md` podaje model serwera, na którym biegnie fuzzing, bez adresu i nazwy użytkownika.

## Konsekwencje

+ katalog główny jest czytelny, układ zgodny z rodziną;
− zmienia się ścieżka importu: odbiorcy przypięci do starszego commita (`gopressgo` na `6298d40`) muszą przy
  podbiciu wersji zmienić import z `github.com/polestorm/gomdgo` na `github.com/polestorm/gomdgo/markdown`;
− dokumenty sprzed 2026-11-24 podają nazwy plików bez katalogu `markdown/`.

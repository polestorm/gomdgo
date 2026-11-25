# 0018 Język bloku kodu w `Config.CodeBlocks` bez rozróżniania wielkości liter

Status: przyjęta · 2026-11-25 (decyzja właściciela; zmienia ADR 0009 w punkcie wyszukiwania funkcji)

## Kontekst

`Config.CodeBlocks` szukał funkcji po pierwszym słowie napisu informacyjnego dokładnie, więc blok ```` ```Go ````
nie trafiał do funkcji zarejestrowanej pod `go`. Wyszło to przy kolorowaniu składni (`gohighlightgo`), które
porównuje nazwy języków bez rozróżniania wielkości liter. Specyfikacje nie rozstrzygają sprawy: CommonMark mówi
tylko, że pierwsze słowo napisu informacyjnego zwykle oznacza język i trafia do `class`, i niczego nie narzuca;
HTML proponuje zwyczaj klasy `language-…` bez reguł dla nazw. W praktyce narzędzia (GitHub, highlight.js)
traktują ```` ```Go ```` jak ```` ```go ````.

## Decyzja

1. Język bloku i klucze `Config.CodeBlocks` są porównywane **bez rozróżniania wielkości liter** (`strings.ToLower`,
   także dla liter spoza ASCII). Klucze są zamieniane raz, przy tworzeniu konwertera.
2. Funkcja dostaje język **w zapisie z dokumentu** (`CodeBlock.Lang == "Go"`), a zwykły blok zachowuje klasę
   `language-Go`: zmienia się tylko wybór funkcji.
3. Klucze różniące się tylko wielkością liter (`Go` i `go`) dają jeden wpis; **wygrywa pierwszy w porządku
   bajtowym**, żeby wybór nie zależał od losowej kolejności przechodzenia mapy w Go (wynik powtarzalny).
4. Rozważona funkcja domyślna dla bloków spoza mapy została odłożona: porównanie bez rozróżniania wielkości liter
   wystarcza odbiorcom, a nowe pole `Config` to większa zmiana API.

## Skutki

+ adapter kolorowania wpisuje nazwy z `Languages()` małymi literami i obsługuje każdy zapis;
+ zachowanie zgodne z praktyką GitHub;
− zmiana zachowania wobec v0.1.1: blok ```` ```Mermaid ```` trafia teraz do funkcji `mermaid` (wcześniej był
  zwykłym blokiem); do opisania przy następnej wersji;
− zmiana kodu biblioteki, więc licznik K5 biegnie od nowa (ADR 0012); przebieg 24 h z 2026-11-24 na Synology
  dotyczy już poprzedniego kodu.

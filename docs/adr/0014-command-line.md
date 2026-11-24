# 0014 Polecenie `gomdgo`: proste, bez uruchamiania innych programów

Status: przyjęta · 2026-11-22 (decyzja właściciela); uzupełnia ją ADR 0015 (polecenie z diagramami w osobnym module)

## Kontekst

Biblioteka nie miała programu wiersza poleceń. Rozważane były trzy sposoby łączenia z `gomermaidgo`: potok
tekstowy (diagramy wstawiane do Markdownu przed parserem), flaga uruchamiająca zewnętrzny program dla bloków
kodu w danym języku, oraz jeden program łączący biblioteki w procesie. Potok wymaga przepuszczania całego
surowego HTML, więc znosi wynik bezpieczny z konstrukcji. Program łączący biblioteki nie może mieszkać
w `gomdgo`, bo biblioteka jest agnostyczna wobec odbiorców i nie ma zależności.

## Decyzja

1. `cmd/gomdgo` zamienia Markdown na HTML i nic więcej: czyta plik albo standardowe wejście, pisze na standardowe
   wyjście albo do pliku (`-o`). Ustawienia domyślne są ustawieniami biblioteki.
2. Flagi odpowiadają polom `Config`: `-raw-html`, `-schemes`, `-any-scheme`, `-commonmark`, `-alerts`,
   `-footnotes`, `-heading-ids`, `-single-tilde`, `-table-align-style`. `-standalone` z `-title` owija wynik
   w kompletny dokument HTML, żeby dało się go otworzyć w przeglądarce.
3. Polecenie **nie uruchamia innych programów** i nie ma punktów rozszerzeń: blok kodu `mermaid` wychodzi jako
   zwykły blok kodu. Punkty rozszerzeń (`Config.CodeBlocks`, `Config.References`) są dostępne tylko z kodu Go.
4. Łączenie z diagramami i z PDF (Markdown z Mermaid albo HTML na wejściu, HTML albo PDF na wyjściu) należy do
   osobnego projektu, który importuje `gomdgo`, `gomermaidgo` i `gopdfgo`. Pomiar próbny z 2026-11-22
   (macOS arm64, `-ldflags="-s -w"`): Markdown z diagramami do HTML 5,8 MB, pełny łańcuch do PDF 21,4 MB.
5. Kody wyjścia: 0 powodzenie, 1 błąd odczytu, zapisu lub renderowania, 2 błędne argumenty. Komunikaty po angielsku.
6. Polecenie leży poza kodem biblioteki w rozumieniu ADR 0012: jego zmiany nie zerują licznika K5.

## Konsekwencje

+ polecenie jest małe (jeden plik) i nie zmienia zasady zera zależności;
+ nie ma flagi, która wykonuje dowolne polecenie z wiersza poleceń;
− diagramów i PDF nie da się uzyskać samym `gomdgo`; do tego służy osobny projekt.

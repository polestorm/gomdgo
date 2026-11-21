# Rodowód kodu

`gomdgo` powstaje od zera, ze specyfikacji. Ten plik mówi, z czego wolno korzystać, a z czego nie.

## Źródła, z których piszemy

- specyfikacja CommonMark (tekst i przykłady), w wersji wskazanej w `docs/adr/0001-scope-and-acceptance.md`;
- specyfikacja GitHub Flavored Markdown: rozdziały o rozszerzeniach;
- normy, do których te specyfikacje odsyłają: HTML (nazwy encji, bloki HTML), Unicode (klasy znaków, zwijanie
  wielkości liter w etykietach odsyłaczy), RFC 3986 (adresy).

Specyfikacja CommonMark opisuje w dodatku strategię parsowania (dwie fazy: struktura bloków, potem treść w wierszu;
algorytm ograniczników dla wyróżnień). To jest część specyfikacji i wolno ją stosować.

## Czego nie robimy

Nie czytamy, nie kopiujemy i nie przenosimy kodu innych parserów Markdown: goldmark, blackfriday, gomarkdown, cmark,
commonmark.js, markdown-it, pulldown-cmark ani ich portów.

## Co wolno

- uruchamiać cudzy parser jako czarną skrzynkę i porównywać jego wynik z naszym (`tools/compare`);
- czytać publiczną dokumentację użytkową cudzych parserów (opis opcji, opis rozszerzeń bez specyfikacji, np. przypisów).

## Dane testowe

Teksty specyfikacji z przykładami są danymi testowymi, nie kodem. Leżą niezmienione w `testdata/spec/`, na licencji
CC BY-SA 4.0, z przypisaniem autorstwa i sumami kontrolnymi w `testdata/spec/README.md`
(decyzja: `docs/adr/0002-spec-examples-in-repo.md`).

## Parser odniesienia

`tools/compare` wywołuje `goldmark` wyłącznie przez publiczne API (`goldmark.New`, opcje, `Convert`), w wersji
przypiętej w `tools/compare/go.mod`. Jego kod źródłowy nie jest czytany; kopia w pamięci podręcznej modułów Go
podlega tej samej zasadzie.

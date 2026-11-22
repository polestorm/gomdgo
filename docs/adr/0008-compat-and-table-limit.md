# 0008 Opcje zgodności z parserem odniesienia i limit dopełniania tabel

Status: przyjęta · 2026-11-21 (decyzje właściciela; zamyka pytania otwarte 9 i 10)

## Kontekst

Dwie sprawy z etapu 4 (ADR 0007):

- `gomdgo` czyta specyfikację GFM ściśle, a parser odniesienia w dwóch miejscach zachowuje się inaczej: zapisuje
  wyrównanie kolumn tabeli jako `style="text-align:…"` i przekreśla tekst objęty pojedynczymi tyldami. Pliki
  wzorcowe odbiorcy powstały z parsera odniesienia, więc bez opcji zgodności kryterium K8 by nie przeszło;
- wiersze tabeli krótsze niż nagłówek są dopełniane pustymi komórkami, więc tabela o wielu kolumnach i wielu
  krótkich wierszach dawała wynik o rozmiarze kolumny × wiersze: z 1 MB wejścia rzędu setek gigabajtów.

## Decyzja

1. **`Config.TableAlignStyle`**: wyrównanie kolumny jako `style="text-align:left|center|right"` zamiast `align`.
   Wartość atrybutu pochodzi wyłącznie z renderera, więc lista dozwolonych atrybutów (`SAFE-005`) rośnie o `style`
   z trzema dopuszczalnymi wartościami.
2. **`Config.SingleTilde`** (wymaga `Strikethrough`): serie jednej i dwóch tyld są ogranicznikami przekreślenia;
   przy różnych długościach zużywany jest jeden znak z każdej strony, tak jak w wyróżnieniu. Trzy i więcej tyld
   to tekst. Zachowanie odwzorowuje parser odniesienia uruchomiony jako czarna skrzynka; oczekiwania są w teście
   `TestCompatibilityOptions`.
3. Obie opcje są **domyślnie wyłączone**: wzorcem pozostaje specyfikacja (ADR 0007). Nie ma opcji zbiorczej
   „zachowuj się jak odniesienie”: każda zgodność jest nazwana i ma własne wymaganie.
4. **Limit dopełniania**: w jednym dokumencie dopełnianych jest łącznie najwyżej 100 000 komórek
   (`maxPaddedCells`). Po wyczerpaniu limitu krótsze wiersze zachowują tylko własne komórki; tabela i jej wiersze
   zostają, HTML pozostaje poprawny. Limit jest na dokument, nie na tabelę, żeby wiele tabel nie omijało go.
   Limitu liczby kolumn nie ma: sam by nie wystarczył (256 kolumn × pół miliona wierszy to nadal ponad gigabajt).

## Konsekwencje

+ K8 da się spełnić bez zmiany plików wzorcowych odbiorcy w dwóch najczęstszych przypadkach;
+ rozmiar wyniku jest ograniczony przez rozmiar wejścia razy stała plus ok. 1 MB na dopełnione komórki;
− pozostają różnice wobec odniesienia, na które opcji nie ma (`../differences.md`, wiersze 3, 5, 7): filtr
  znaczników, pole wyboru bez odstępu, znaki spoza ASCII w rozszerzonym autolinku.
  Jeśli pliki wzorcowe odbiorcy na nich polegają, wymagają osobnej decyzji;
− dokument, który uczciwie potrzebuje ponad 100 000 dopełnionych komórek, dostanie krótsze wiersze bez ostrzeżenia.
  Walidator odbiorcy może to wykryć na drzewie (wiersz z mniejszą liczbą komórek niż nagłówek).

# 0003 Pozycja w źródle na każdym węźle

Status: przyjęta · 2026-11-21 (decyzja właściciela; zamyka pytanie otwarte 5)

## Kontekst

Walidatory odbiorców działają na drzewie dokumentu i muszą wskazać miejsce w źródle („wiersz 12, kolumna 5:
niedozwolony obraz”). Pytanie brzmiało, czy płacić za to pamięcią na każdym węźle.

## Decyzja

1. **Każdy węzeł** drzewa, blokowy i w wierszu, niesie pozycję swojego początku w źródle.
2. Węzeł przechowuje jedno pole: przesunięcie bajtowe w źródle. Wiersz i kolumnę wylicza się z niego na żądanie
   z tablicy początków wierszy, którą parser i tak buduje w pierwszej fazie.
3. Wiersze i kolumny liczymy od 1; kolumna jest liczona w bajtach wiersza źródła (tabulator to jeden bajt).
   Jeśli odbiorca będzie potrzebował kolumn w znakach, dojdzie funkcja przeliczająca, bez zmiany węzła.

## Konsekwencje

+ komunikaty walidatorów są dokładne także dla elementów w wierszu;
+ koszt to jedno pole całkowite na węzeł, niezależnie od tego, czy ktoś pyta o pozycję;
− treść w wierszu jest parsowana z tekstu złożonego z kilku wierszy bloku (po zdjęciu znaczników cytatu i wcięć),
  więc parser bloków musi zachować odwzorowanie tego tekstu na przesunięcia w źródle. To wymaganie projektowe
  dla etapów 1 i 2, do pokrycia testem pozycji.

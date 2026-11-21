# 0004 Niepoprawny UTF-8 na wejściu

Status: przyjęta · 2026-11-21 (decyzja właściciela; zamyka pytanie otwarte 6)

## Kontekst

Specyfikacja CommonMark mówi tylko o znaku U+0000: ze względów bezpieczeństwa ma być zastąpiony znakiem U+FFFD.
O bajtach, które nie są poprawnym UTF-8, milczy. Do etapu 2 takie bajty przechodziły do wyniku bez zmian,
więc wynik mógł nie być poprawnym UTF-8, a konwerter HTML → PDF albo przeglądarka musiałyby zgadywać.

## Decyzja

1. Każda nieprzerwana seria bajtów, które nie tworzą poprawnie zakodowanego znaku, jest zastępowana **jednym**
   znakiem U+FFFD. Dotyczy to także sekwencji urwanych, nadmiarowo długich i zakodowanych surogatów.
2. Zamiana następuje **na wejściu, przed parsowaniem**, razem z zamianą U+0000. Zamiana dopiero w wyniku
   pozwalałaby dwóm niepoprawnym bajtom złożyć się w poprawny znak ponad usuniętym znacznikiem (np. w opisie
   obrazu), czyli wytworzyć znak, którego w źródle nie było.
3. Źródło poprawne nie jest kopiowane: sprawdzenie to jeden przebieg po bajtach.
4. Opcji wyłączającej nie ma. Wynik jest zawsze poprawnym UTF-8 bez U+0000, w każdym trybie.

## Konsekwencje

+ „bezpieczny z konstrukcji” obejmuje kodowanie: odbiorca wyniku nie musi go sprawdzać;
+ klasy znaków (biały znak, interpunkcja) są wyznaczane na tekście, który faktycznie trafia do wyniku;
− pozycje węzłów (ADR 0003) odnoszą się do źródła po zamianie. Numery wierszy są zawsze takie same jak
  w oryginale; kolumny w wierszu, w którym coś zamieniono, są przesunięte za miejscem zamiany (U+FFFD to 3 bajty).
  Dla źródła poprawnego różnicy nie ma;
− w adresie odnośnika zamieniony bajt daje `%EF%BF%BD`, a nie kod oryginalnego bajtu.

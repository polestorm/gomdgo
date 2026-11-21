# 0010 Publiczne drzewo dokumentu: funkcja odwiedzająca

Status: przyjęta · 2026-11-21 (decyzja właściciela; zamyka pytanie otwarte 3)

## Kontekst

Walidatory odbiorców działają na drzewie dokumentu między parsowaniem a renderowaniem (ADR 0001, ADR 0005).
Do wyboru były: funkcja odwiedzająca z małym widokiem węzła, publiczne typy węzłów z polami, albo płaska lista
elementów.

## Decyzja

1. `Document.Walk(func(n Node, entering bool) bool)` odwiedza węzły w kolejności źródła. Każdy węzeł jest
   zgłaszany dwa razy, przy wejściu i przy wyjściu, zawsze parami, także gdy nie ma dzieci. Wartość zwrócona przy
   wejściu mówi, czy odwiedzić dzieci.
2. `Node` jest uchwytem tylko do odczytu: `Kind`, `Position`, `Text` i metody dla cech poszczególnych rodzajów
   (`Destination`, `Title`, `Level`, `Language`, `Info`, `Ordered`, `Start`, `Tight`, `Checked`, `Header`, `Column`,
   `Align`, `Alert`). Metoda wołana na węźle innego rodzaju zwraca wartość zerową.
3. Wewnętrzna struktura węzłów zostaje prywatna. `Kind` ma stałe publiczne (`KindHeading`, `KindImage`, …)
   równe wartościom wewnętrznym; nowe rodzaje mogą dochodzić, więc walidator powinien mieć gałąź domyślną.
4. Walidator widzi to, co jest w źródle, a nie to, co trafi do wyniku: adres odnośnika jest podawany niezależnie
   od listy dozwolonych schematów, surowy HTML niezależnie od polityki, a definicja przypisu w miejscu, w którym stoi.
5. Zależności od zagnieżdżenia (np. obraz wewnątrz tabeli) walidator śledzi sam, na stosie budowanym z wywołań
   wejścia i wyjścia; `Node` nie daje dostępu do rodzica ani rodzeństwa.

## Konsekwencje

+ dokument pozostaje niezmienny, więc odwiedzanie i renderowanie mogą biec równolegle;
+ zmiana reprezentacji węzłów nie łamie odbiorców;
− odbiorca nie może przebudować drzewa przed renderowaniem; kto tego potrzebuje, zmienia źródło albo używa punktów
  rozszerzeń (ADR 0009);
− akapit złożony z samych definicji odsyłaczy nie jest odwiedzany; definicje odsyłaczy nie mają dziś widoku
  publicznego.

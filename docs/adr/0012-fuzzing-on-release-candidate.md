# 0012 Fuzzing do kryterium K5 liczymy na kandydacie do wydania

Status: przyjęta · 2026-11-22 (decyzja właściciela; doprecyzowuje K5 z ADR 0001)

## Kontekst

ADR 0001 wymaga „fuzzingu co najmniej 24 h łącznie” i nie mówi, czy godziny mają dotyczyć jednej wersji kodu.
Do 2026-11-22 czas sumowaliśmy przez etapy, choć kod zmieniał się między przebiegami. Godziny fuzzingu mówią
coś tylko o kodzie, na którym biegły, a kod będzie się jeszcze zmieniał: etap 7 (dokumenty odbiorcy), wybory
robocze z pytania 8 (potwierdzone bez zmian 2026-11-22) i poprawki błędów znalezionych przez sam fuzzing.

## Decyzja

1. K5 jest spełnione, gdy `FuzzRender` przebiegł co najmniej 24 h bez błędu na **kandydacie do wydania**, czyli
   na commicie, którego kod biblioteki (pliki `*.go` pakietu biblioteki poza testami; do 2026-11-24 w katalogu głównym, potem w `markdown/`,
   ADR 0016; samo przeniesienie plików i zmiana nazwy pakietu bez zmiany treści nie zeruje licznika) jest tym, który idzie do
   odbiorcy. Czas z kilku przebiegów na tym samym kodzie biblioteki sumuje się.
2. Każda zmiana kodu biblioteki po rozpoczęciu liczenia zeruje licznik K5. Zmiany w dokumentach, testach
   i narzędziach go nie zerują.
3. Przebiegi na wcześniejszych wersjach zostają w tabeli pomiarów jako pomiary pośrednie. Służą do wczesnego
   znajdowania błędów i nie wliczają się do K5.
4. Po każdej zmianie kodu biblioteki robimy przebieg kontrolny, co najmniej 30 minut.
5. Wiersz tabeli pomiarów podaje dla każdego przebiegu commit, czas, liczbę wywołań i maszynę.
6. Długie przebiegi wolno prowadzić poza maszyną roboczą. Trafia tam binarka testowa (`go test -c -fuzz
   FuzzRender`) i katalog `testdata/`; opis w `docs/testing.md`.

## Konsekwencje

+ K5 dowodzi odporności kodu, który faktycznie trafia do odbiorcy;
+ pomiary pośrednie nadal mają wartość, bo błąd znaleziony wcześniej jest tańszy;
− pełne 24 h trzeba powtórzyć po ostatniej zmianie w bibliotece, więc K5 zamyka się dopiero na końcu etapu 7;
− poprawka znaleziona w trakcie liczenia wydłuża drogę o kolejne 24 h.

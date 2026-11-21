# 0009 Punkty rozszerzeń: bloki kodu według języka i odwołania w wierszu

Status: przyjęta (robocza: kształt funkcji to wybór wykonawcy, do potwierdzenia przez właściciela) · 2026-11-21

## Kontekst

ADR 0001 (punkt 2) ustala, że zamiast surowego HTML odbiorca dostaje punkty rozszerzeń: blok kodu
z zarejestrowanym językiem (np. `mermaid`) renderuje jego funkcja, tak samo odwołania w wierszu
(np. `[[NC-2026-017]]`). ADR 0005 ustala, że błąd takiej funkcji przerywa renderowanie. Zostawał kształt funkcji.

## Decyzja

1. **Blok kodu**: `Config.CodeBlocks map[string]CodeBlockFunc`, `func(CodeBlock) ([]byte, error)`. Funkcja dostaje
   strukturę (język, napis informacyjny, treść, wiersz), żeby dało się dodawać pola bez zmiany sygnatury.
   **Jej wynik trafia do dokumentu bez zmian.** To jedyne miejsce, w którym do wyniku wchodzi HTML spoza
   renderera, i jedyne, którego nie obejmuje gwarancja „bezpieczny z konstrukcji”: za wynik odpowiada funkcja
   odbiorcy (np. `gomermaidgo.Render`, który sam gwarantuje zamknięty zbiór elementów SVG).
2. **Odwołanie w wierszu**: `Config.References func(Reference) (ReferenceLink, error)`. Funkcja **nie zwraca
   HTML**, tylko adres, tekst i tytuł; odnośnik buduje renderer tą samą drogą co odnośnik ze źródła (lista
   schematów, kodowanie adresu, escapowanie). Odwołania nie wyprowadzają więc wyniku poza gwarancję bezpieczeństwa.
3. **Składnia odwołania to `[[cel]]`** i działa tylko wtedy, gdy funkcja jest ustawiona; inaczej zapis jest zwykłym
   tekstem CommonMark, więc sama obecność tej możliwości nie zmienia znaczenia istniejących dokumentów.
4. **Błędy** mają typ `*Error` ze stałym kodem, pozycją w źródle i oryginalnym błędem. Komunikat jest po
   angielsku; kod służy wywołującemu do własnego tłumaczenia.
5. **Brak odpowiedzi to nie błąd**: funkcja odwołań, która nie zna celu, zwraca pusty adres i odwołanie zostaje
   tekstem. Czy nieznany cel ma zatrzymać dokument, decyduje odbiorca, zwracając błąd.

## Konsekwencje

+ droga Markdown → HTML z SVG diagramów nie wymaga `RawHTMLPass`;
+ walidator odbiorcy i funkcje rozszerzeń widzą te same pozycje w źródle;
− wynik z `Config.CodeBlocks` jest tak bezpieczny jak zarejestrowane funkcje; test bezpieczeństwa (`checkSafe`)
  sprawdza konwertery bez nich oraz z funkcją, która sama escapuje;
− funkcje są wołane przy każdym `Document.HTML`; kosztowne wyniki odbiorca buforuje po swojej stronie.

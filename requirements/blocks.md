# Wymagania: struktura bloków

Źródło: CommonMark 0.31.2, rozdziały 2–5. Przykłady z danego rozdziału specyfikacji są testami wymagania
(`TestSpecCommonMark`, tabela `sectionRequirements`). Dopóki nie ma treści w wierszu (etap 2), przykłady,
które jej wymagają, nie przechodzą; lista zaliczonych jest w `testdata/spec/passing-commonmark.txt`.

| Id | Wymaganie |
|---|---|
| BLK-001 | Wiersz kończy się znakiem nowego wiersza, powrotu karetki albo parą obu; koniec źródła kończy ostatni wiersz. Znak U+0000 jest w wyniku zamieniany na U+FFFD. |
| BLK-002 | Tabulator w miejscu, w którym białe znaki wyznaczają strukturę bloków, działa jak dopełnienie do kolumny podzielnej przez 4; w treści zostaje tabulatorem. Część tabulatora niezużyta przez strukturę przechodzi do treści jako spacje (rozdział „Tabs”). |
| BLK-003 | Struktura bloków ma pierwszeństwo przed strukturą treści w wierszu (rozdział „Precedence”). |
| BLK-004 | Linia pozioma: co najmniej trzy jednakowe znaki `*`, `-` albo `_` z wcięciem do 3 spacji, z dowolnymi spacjami i tabulatorami między nimi; może przerwać akapit; ma pierwszeństwo przed elementem listy (rozdział „Thematic breaks”). |
| BLK-005 | Nagłówek ATX: 1–6 znaków `#` z wcięciem do 3 spacji, potem spacja, tabulator albo koniec wiersza; opcjonalna sekwencja zamykająca jest pomijana; treść bez białych znaków na brzegach; może być pusty (rozdział „ATX headings”). |
| BLK-006 | Nagłówek setext: akapit, po którym następuje wiersz z samych `=` (poziom 1) albo `-` (poziom 2); podkreślenie nie może być kontynuacją leniwą; akapit złożony z samych definicji odsyłaczy nie staje się nagłówkiem (rozdział „Setext headings”). |
| BLK-007 | Kod przez wcięcie: wiersze z wcięciem co najmniej 4 kolumn; nie przerywa akapitu; puste wiersze w środku zostają, początkowe i końcowe są pomijane (rozdział „Indented code blocks”). |
| BLK-008 | Kod ogrodzony: co najmniej trzy znaki `` ` `` albo `~`; zamknięcie tym samym znakiem, nie krótsze niż otwarcie; wcięcie otwarcia jest zdejmowane z wierszy treści; pierwsze słowo napisu informacyjnego daje klasę `language-…`; blok niezamknięty trwa do końca kontenera (rozdział „Fenced code blocks”). |
| BLK-009 | Blok HTML: siedem warunków początku i odpowiadające im warunki końca; treść przechodzi bez parsowania; warunek 7 nie przerywa akapitu (rozdział „HTML blocks”). |
| BLK-010 | Definicja odsyłacza: etykieta, dwukropek, cel i opcjonalny tytuł, także w kilku wierszach; nie przerywa akapitu; nie daje wyniku; przy powtórzonej etykiecie obowiązuje pierwsza definicja; etykiety porównuje się bez rozróżniania wielkości liter i po zwinięciu białych znaków (rozdział „Link reference definitions”). |
| BLK-011 | Akapit: kolejne niepuste wiersze, których nie da się odczytać jako innego bloku; wiodące białe znaki wierszy są pomijane (rozdział „Paragraphs”). |
| BLK-012 | Puste wiersze między blokami są pomijane; rozstrzygają tylko o tym, czy lista jest zwarta (rozdział „Blank lines”). |
| BLK-013 | Cytat: znacznik `>` z wcięciem do 3 spacji i opcjonalną spacją po nim; kontynuacja leniwa obejmuje tylko dalszy ciąg akapitu (rozdział „Block quotes”). |
| BLK-014 | Element listy: znacznik wypunktowania `-`, `+`, `*` albo numer z 1–9 cyfr i znakiem `.` albo `)`; wcięcie treści wynika z szerokości znacznika i odstępu po nim; element może zaczynać się pustym wierszem, ale najwyżej jednym; akapit przerywa tylko element niepusty, a numerowany tylko zaczynający się od 1 (rozdział „List items”). |
| BLK-015 | Lista: kolejne elementy tego samego rodzaju; zmiana znaku wypunktowania albo ogranicznika numeru zaczyna nową listę; numer początkowy pochodzi z pierwszego elementu; lista jest luźna, gdy pusty wiersz dzieli jej elementy albo bloki będące bezpośrednio w elemencie, a w liście zwartej akapity nie mają znaczników `<p>` (rozdział „Lists”). |
| BLK-016 | Każdy węzeł drzewa niesie pozycję swojego początku w źródle; da się ją zamienić na wiersz i kolumnę liczone od 1 (ADR 0003). |

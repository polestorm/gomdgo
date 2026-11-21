# Wymagania: treść w wierszu

Źródło: CommonMark 0.31.2, rozdział 6. Przykłady z danego rozdziału specyfikacji są testami wymagania
(`TestSpecCommonMark`, tabela `sectionRequirements`).

| Id | Wymaganie |
|---|---|
| INL-001 | Znak ucieczki `\` przed znakiem interpunkcyjnym ASCII daje ten znak dosłownie; przed innym znakiem jest zwykłym ukośnikiem; nie działa w kodzie, autolinkach i surowym HTML; działa w celach i tytułach odnośników oraz w napisie informacyjnym kodu (rozdział „Backslash escapes”). |
| INL-002 | Odwołania znakowe: nazwane encje HTML5 ze średnikiem, dziesiętne (1–7 cyfr) i szesnastkowe (1–6 cyfr); numer niepoprawny albo zero daje U+FFFD; odwołanie nie tworzy struktury (rozdział „Entity and numeric character references”). |
| INL-003 | Kod w wierszu: seria znaków `` ` `` zamknięta serią tej samej długości; końce wierszy stają się spacjami; po jednej spacji z obu stron jest zdejmowane, gdy treść nie składa się z samych spacji; ma pierwszeństwo przed pozostałą treścią w wierszu poza autolinkami i surowym HTML zaczętymi wcześniej (rozdziały „Inlines” i „Code spans”). |
| INL-004 | Wyróżnienie i mocne wyróżnienie znakami `*` i `_` według 17 reguł specyfikacji: serie lewo- i prawostronne, ograniczenia `_` wewnątrz słów, reguła sumy długości podzielnej przez 3, pierwszeństwo krótszego zasięgu (rozdział „Emphasis and strong emphasis”). |
| INL-005 | Odnośniki: w wierszu `[tekst](cel "tytuł")` oraz referencyjne pełne, zwinięte i skrócone; cel w nawiasach kątowych albo bez, z nawiasami zrównoważonymi; odnośnik nie zawiera odnośnika; adres jest kodowany procentowo, tytuł escapowany (rozdział „Links”). |
| INL-006 | Obrazy: składnia odnośnika poprzedzona `!`; opis w `alt` to sam tekst treści, bez znaczników (rozdział „Images”). |
| INL-007 | Autolinki `<adres>` ze schematem 2–32 znaków i `<adres e-mail>`; bez spacji w środku; znak ucieczki w nich nie działa (rozdział „Autolinks”). |
| INL-008 | Surowy HTML w wierszu: znacznik otwierający i zamykający, komentarz, instrukcja przetwarzania, deklaracja, sekcja CDATA; treść przechodzi bez zmian, gdy przepuszczanie HTML jest włączone (rozdział „Raw HTML”). |
| INL-009 | Twarde łamanie wiersza: co najmniej dwie spacje albo ukośnik przed końcem wiersza; nie na końcu bloku; nie w kodzie ani w znaczniku HTML (rozdział „Hard line breaks”). |
| INL-010 | Miękkie łamanie wiersza: koniec wiersza w akapicie daje znak nowego wiersza w wyniku; spacje na końcu wiersza i na początku następnego są pomijane (rozdział „Soft line breaks”). |
| INL-011 | Wszystko, co nie jest inną konstrukcją, jest tekstem; znaki spoza ASCII przechodzą bez zmian, spacje wewnątrz wiersza są zachowane (rozdział „Textual content”). |
| INL-012 | Każdy węzeł treści w wierszu niesie pozycję swojego początku w źródle, także w akapicie złożonym z kilku wierszy i wewnątrz kontenerów (ADR 0003). |

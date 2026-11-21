# Różnice wobec parsera odniesienia

Każda różnica wyniku `gomdgo` i `goldmark` w korpusie różnicowym (`tools/compare`), z przykładem, punktem
specyfikacji i uzasadnieniem. Gdy `goldmark` odbiega od specyfikacji, wygrywa specyfikacja.

| # | Wejście | `gomdgo` | `goldmark` | Specyfikacja | Uzasadnienie |
|---|---|---|---|---|---|
| 1 | tryb bezpieczny (`tools/compare -safe`): odnośnik albo autolink ze schematem spoza `http`, `https`, `mailto`, `tel`, np. `<irc://foo.bar:2233/baz>`, `<a+b+c:d>`, `<made-up-scheme://foo,bar>`, `<localhost:5001/foo>` (przykłady CommonMark 596, 598, 599, 601) | sama treść, bez `<a>` | `<a href="…">` z pełnym adresem | nie dotyczy: specyfikacja nie opisuje filtrowania adresów | lista dozwolonych zamiast listy zakazanych (ADR 0006, `SAFE-003`); przy przepuszczaniu wszystkich schematów (`AnyScheme`) wyniki są identyczne |
| 2 | tabela z wyrównaniem kolumn, np. `\|:-:\|` (przykład GFM 199; korpus `007-gfm.md`) | `<th align="center">` | `<th style="text-align:center">` przy ustawieniach domyślnych; z opcją `TableCellAlignAttribute` tak jak `gomdgo` | GFM, „Tables”: przykład 199 podaje `align="center"` | specyfikacja; `Config.TableAlignStyle` daje zapis odniesienia (ADR 0008); `tools/compare -refalign=attribute` przełącza odniesienie na zapis specyfikacji, `-refalign=style` porównuje odniesienie domyślne z `gomdgo` z opcjami zgodności |
| 3 | znaczniki z listy filtra przy przepuszczaniu surowego HTML, np. `<title>`, `<xmp>` (przykład GFM 653) | `&lt;title>` | `<title>` bez zmian: filtr nie jest zaimplementowany | GFM, „Disallowed Raw HTML” | specyfikacja |
| 4 | wiersz tabeli krótszy niż nagłówek w kolumnie z wyrównaniem (korpus `007-gfm.md`) | dopełniona komórka ma `align` kolumny | dopełniona komórka nie ma wyrównania | GFM mówi tylko, że „wstawiane są puste komórki” | wyrównanie jest cechą kolumny; komórka pusta w źródle (`\|\|`) dostaje je w obu parserach, więc dopełniona też powinna |
| 5 | `- [ ]bez odstępu` (korpus `007-gfm.md`) | zwykły tekst `[ ]bez odstępu` | pole wyboru | GFM, „Task list items”: po znaczniku „co najmniej jeden biały znak” | specyfikacja |
| 6 | `~jedna tylda~` (korpus `007-gfm.md`) | zwykły tekst | `<del>jedna tylda</del>` | GFM, „Strikethrough”: „tekst objęty dwiema tyldami” | specyfikacja; `Config.SingleTilde` daje zachowanie odniesienia (ADR 0008) |
| 7 | rozszerzony autolink ze znakiem spoza ASCII w ścieżce, `https://example.com/ścieżka?a=1` (korpus `007-gfm.md`) | cały adres jest odnośnikiem, ścieżka zakodowana procentowo | odnośnik kończy się przed `ś`, reszta jest tekstem | GFM, „Autolinks”: po domenie „zero lub więcej znaków niebędących spacją ani `<`” | specyfikacja |
| 8 | tryb bezpieczny: rozszerzony autolink `ftp://foo.bar.baz` (przykład GFM 628) | zwykły tekst | odnośnik | nie dotyczy | ta sama decyzja co w wierszu 1: `ftp` nie jest na domyślnej liście schematów |

Różnice trybu bezpiecznego, których korpus dziś nie zawiera, a które wynikają z tych samych decyzji:
`goldmark` dla schematów `javascript:`, `vbscript:`, `file:`, `data:` zostawia odnośnik z pustym `href`, `gomdgo`
zostawia samą treść; `goldmark` w miejscu pominiętego surowego HTML wstawia komentarz `<!-- raw HTML omitted -->`,
`gomdgo` (`RawHTMLOmit`) nie wstawia nic (porównanie usuwa ten komentarz przed normalizacją).

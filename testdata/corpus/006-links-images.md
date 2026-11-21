[//]: # (REQ: INL-005 INL-006 INL-007)

Odnośnik [w wierszu](https://example.com/ścieżka?q=zażółć&x=1 "Tytuł z „cudzysłowem”") i [pusty]().
Cel w nawiasach kątowych: [spacja](<adres ze spacją> 'tytuł') oraz [nawiasy](/a(b)c).
Tekst z wyróżnieniem: [**mocny** i `kod`](/cel), z ucieczką: [a\]b](/c\)d).

Referencyjne: [pełny][Etykieta], [zwinięty][], [skrócony], [Wielkość  LITER][etykieta].
Nieznane: [brak][nieznana], [też brak][] i [samotny].

Obraz: ![opis *z kursywą* i [odnośnikiem](/x)](/obraz.png "Tytuł obrazu") oraz ![referencyjny][etykieta].
Obraz w odnośniku: [![logo](/logo.svg)](/start).

Odnośnik w odnośniku nie powstaje: [zewnętrzny [wewnętrzny](/b) koniec](/a).

Autolinki: <https://example.com/a?b=c&d=e>, <mailto:jan@example.com>, <jan.kowalski@example.com>,
ale nie <to nie adres> ani <https://example.com/a b>.

Nawiasy bez odnośnika: [tekst] (nie cel), [tekst](nie cel bo spacje), [a](b "niedomknięty tytuł).

[etykieta]: /cel-etykiety "Tytuł etykiety"
[zwinięty]: </cel zwinięty>
[skrócony]: /cel-skrócony
    'tytuł w następnym wierszu'

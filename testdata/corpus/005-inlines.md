[//]: # (REQ: INL-001 INL-002 INL-003 INL-004 INL-009 INL-010 INL-011)

Zażółć *gęślą* jaźń, a potem **mocno** i ***bardzo mocno***.
Podkreślenia: _tak_, __mocno__, ale nie w_środku_słowa ani w nazwie plik_z_danymi.txt.

Cudzysłowy: „*kursywa*” i «**pogrubienie**», nawiasy (*tak*) oraz myślnik —*też*—.
Gwiazdki z odstępem * nie wyróżniają *, a 2*3*4 już tak.

Zagnieżdżenie: *zewnętrzne **wewnętrzne** dalej*, **mocne *zwykłe* mocne**, *a**b**c*.
Niedomknięte: **bez końca, *też bez końca i __jeszcze jedno.

Kod: `go test ./...`, z odwrotnym apostrofem `` a`b ``, ze spacjami `  x  ` i pusty `` ` ``.
Kod ma pierwszeństwo: `*nie kursywa*` oraz *kursywa `z kodem*` dalej*.

Znaki ucieczki: \*nie kursywa\*, \[nie odnośnik\], \# nie nagłówek, 100\% i ukośnik \ą.
Encje: &copy; &amp; &lt;tag&gt; &#261; &#x105; &nbsp;| &nieznana; &#0; &amp bez średnika.

Twarde łamanie dwiema spacjami  
i ukośnikiem\
a tu miękkie
łamanie wiersza.

# Nagłówek z *kursywą* i `kodem`

Tytuł **setext** z \*gwiazdką\*
===============================

- element z *kursywą*
- element z `kodem` i **pogrubieniem**
  w dwóch wierszach

> cytat z _podkreśleniem_
> i drugim wierszem  
> po twardym łamaniu

# 0011 Bloki uwag i przypisy

Status: przyjęta · 2026-11-21 (wzorzec: decyzja właściciela, zamyka pytanie otwarte 4; postać wyniku: wybór
wykonawcy, potwierdzony przez właściciela 2026-11-22)

## Kontekst

ADR 0001 obejmuje zakresem przypisy i bloki uwag `> [!NOTE]`, które nie mają specyfikacji. Właściciel ustalił
wzorzec: składnia z publicznej dokumentacji użytkowej GitHub („Alerts”, „Footnotes”), a wszystko, czego ona nie
rozstrzyga, jako własne wymagania (`requirements/extensions.md`, `EXT-005`, `EXT-006`). Dokumentacja użytkowa nie
opisuje HTML, więc postać wyniku jest nasza.

## Decyzja

1. Oba rozszerzenia są **domyślnie wyłączone** i włączane osobno przez `Config.Enable` (`Alerts`, `Footnotes`),
   zgodnie z ADR 0001. Bez nich `[^1]: /url` pozostaje definicją odsyłacza CommonMark, a `> [!NOTE]` zwykłym cytatem.
2. **Blok uwagi** to cytat ze znacznikiem w pierwszym wierszu; wynik: `<blockquote class="alert alert-RODZAJ">`.
   Biblioteka **nie wstawia własnego tytułu** („Note”, „Warning”), bo nie zna języka dokumentu: tytuły podaje
   odbiorca w `Config.AlertTitles`, a bez nich blok rozróżnia sama klasa. Blok uwagi może stać wszędzie tam,
   gdzie cytat, także zagnieżdżony (GitHub tego nie dopuszcza; ograniczenie nie ma u nas uzasadnienia).
3. **Przypis** jest blokiem-kontenerem z wcięciem treści 4 kolumn. Numer dostaje według pierwszego odwołania;
   treść trafia na koniec dokumentu do `<section class="footnotes">`. Przypis bez odwołań nie trafia do wyniku.
4. **Identyfikatory przypisów powstają z numerów, nie z etykiet**: `fn:1`, `fnref:1`, `fnref:1:2`. Etykieta jest
   tekstem autora i nie trafia do atrybutów; dwukropek wyklucza kolizję z identyfikatorami nagłówków, które go
   nie zawierają (`OUT-002`).
5. Odnośnik powrotny to znak `↩` z klasą `footnote-backref`, po jednym na każde odwołanie.
6. Zgodności HTML z parserem odniesienia ani z GitHub nie obiecujemy; plik korpusu z tymi rozszerzeniami
   (`testdata/corpus-extras/`) nie wchodzi do porównania różnicowego.

## Konsekwencje

+ lista dozwolonych znaczników i atrybutów (`SAFE-005`) rośnie o `sup`, `section`, atrybuty `class` i `id`
  o wartościach wytwarzanych wyłącznie przez renderer; sprawdza to `checkSafe`, także w fuzzingu;
− odbiorca, którego pliki wzorcowe zawierają przypisy z parsera odniesienia, zobaczy inny HTML przypisów;
  jeśli to przeszkoda dla K8, potrzebna będzie opcja zgodności, jak w ADR 0008;
− bez `Config.AlertTitles` i bez arkusza stylów blok uwagi wygląda w wyniku jak zwykły cytat.

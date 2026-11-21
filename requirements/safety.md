# Wymagania: bezpieczeństwo wyniku

Decyzje: `docs/adr/0006-safe-output.md`. „Ustawienia domyślne” to `Render(src)` albo `New(Config{})`,
czyli z rozszerzeniami GFM.

| Id | Wymaganie |
|---|---|
| SAFE-001 | Tekst ze źródła trafia do treści i do wartości atrybutów z zamianą znaków `&`, `<`, `>`, `"` na odwołania znakowe; wartości atrybutów są zawsze w cudzysłowach. Dotyczy także opisu obrazu, tytułu odnośnika, napisu informacyjnego kodu i identyfikatora nagłówka. |
| SAFE-002 | Surowy HTML ze źródła (bloki HTML i HTML w wierszu) przy ustawieniach domyślnych jest zapisywany jako tekst (`RawHTMLEscape`): blok HTML jak akapit, HTML w wierszu w miejscu wystąpienia. `RawHTMLOmit` pomija go w wyniku, `RawHTMLPass` przepuszcza bez zmian. Nieznana wartość ustawienia działa jak `RawHTMLEscape`. |
| SAFE-003 | Adres odnośnika i obrazu trafia do wyniku tylko wtedy, gdy nie ma schematu albo jego schemat jest na liście dozwolonych (domyślnie `http`, `https`, `mailto`, `tel`; porównanie bez rozróżniania wielkości liter, po rozwinięciu znaków ucieczki i odwołań znakowych). Odnośnik o niedozwolonym schemacie zostaje samą swoją treścią, obraz samym swoim opisem; dotyczy to także autolinków i odsyłaczy referencyjnych. Listę można zastąpić własną (pusta dopuszcza tylko adresy bez schematu) albo wyłączyć sprawdzanie (`AnyScheme`). |
| SAFE-004 | Adres w wyniku składa się wyłącznie z liter, cyfr i znaków `;/?:@&=+$,-_.!~*'()#%`; każdy inny bajt jest zakodowany procentowo. Znak, który nie może być częścią schematu, nie zniknie więc z adresu po drodze do przeglądarki. |
| SAFE-005 | Bez `RawHTMLPass` wynik zawiera wyłącznie znaczniki i atrybuty z zamkniętej listy: `p`, `h1`–`h6` (`id`), `ul`, `ol` (`start`), `li`, `blockquote`, `pre`, `code` (`class` zaczynające się od `language-`), `hr`, `br`, `em`, `strong`, `a` (`href`, `title`), `img` (`src`, `alt`, `title`), a z rozszerzeniami GFM także `table`, `thead`, `tbody`, `tr`, `th` i `td` (`align` albo `style` z wartością `text-align:left|center|right`), `del`, `input` (`checked`, `disabled`, `type="checkbox"`). Sprawdza to test niezależny od renderera, także w fuzzingu. |

# 0007 Rozszerzenia GFM: domyślnie włączone, wyłączane jawnie

Status: przyjęta · 2026-11-21 (punkt 1: decyzja właściciela; punkty 3–6: wybory wykonawcy, do potwierdzenia)

## Kontekst

ADR 0001 obejmuje zakresem rozszerzenia GFM 0.29-gfm: tabele, przekreślenie, listy zadań, rozszerzone autolinki
i filtr znaczników. Trzeba było ustalić, czy działają domyślnie i jak ściśle trzymamy się tekstu specyfikacji tam,
gdzie parser odniesienia i sam GitHub przyjmują więcej.

## Decyzja

1. **Domyślnie włączone są wszystkie rozszerzenia GFM.** `Config{}` i `Render(src)` to CommonMark z tabelami,
   przekreśleniem, listami zadań, rozszerzonymi autolinkami i filtrem znaczników. Uzasadnienie właściciela:
   GitHub jest powszechnie znany i to jego odmiany autorzy dokumentów oczekują pod nazwą „Markdown”.
   (Pierwsza, robocza wersja tej decyzji miała rozszerzenia domyślnie wyłączone, jak parser odniesienia.)
2. **Wyłączanie jest jawne i osobne dla każdego rozszerzenia:** `Config.Disable` przyjmuje bity `Tables`,
   `Strikethrough`, `TaskLists`, `Autolinks`, `TagFilter`; `Disable: GFM` daje czysty CommonMark, dla którego
   wynik przykładów CommonMark jest dokładnie taki jak w specyfikacji. Rozszerzenia spoza GFM (przypisy, bloki
   uwag, etap 5) będą włączane osobno, zgodnie z ADR 0001.
3. **Wzorcem jest tekst specyfikacji, czytany ściśle:**
   - przekreślenie to dokładnie dwie tyldy z każdej strony; jedna tylda nie przekreśla;
   - znacznik zadania wymaga białego znaku po `]`;
   - rozszerzony autolink kończy się na białym znaku albo `<`, więc obejmuje znaki spoza ASCII w ścieżce;
   - wyrównanie kolumny trafia do atrybutu `align`.
   Tam, gdzie specyfikacja milczy, przyjmujemy zachowanie parsera odniesienia: komórka dopełniona do liczby
   kolumn nie dostaje wyrównania (zmiana z 2026-11-21, po porównaniu na dokumentach rzeczywistych).
   Każdą różnicę wobec parsera odniesienia opisuje `../differences.md`.
4. **Nagłówkiem tabeli jest ostatni wiersz akapitu** poprzedzający wiersz ograniczników; wcześniejsze wiersze
   zostają akapitem. Specyfikacja tego nie rozstrzyga; tak samo robi parser odniesienia.
5. **Filtr znaczników działa tylko z `RawHTMLPass`**: przy pozostałych politykach surowy HTML i tak nie trafia do
   wyniku jako znaczniki.
6. **Rozszerzone autolinki podlegają liście schematów** (ADR 0006): `ftp://` przy ustawieniach domyślnych zostaje
   tekstem.

## Konsekwencje

+ kryteria K1 i K2 mierzy się niezależnie: z `Disable: GFM` 652/652, przy ustawieniach domyślnych 24/24;
− przy ustawieniach domyślnych 10 z 652 przykładów CommonMark daje inny wynik niż w specyfikacji (zawierają
  składnię rozszerzeń, głównie adresy w zwykłym tekście); odbiorca przepinany z parsera odniesienia, który
  rozszerzeń nie włączał, musi ustawić `Disable: GFM`;
− odbiorca przepinany z `goldmark` zobaczyłby różnice w plikach wzorcowych tam, gdzie polegał na zachowaniu
  szerszym niż specyfikacja; dla jednej tyldy i zapisu wyrównania są opcje zgodności (ADR 0008);
− tabela o wielu kolumnach i wielu krótkich wierszach dawałaby wynik o rozmiarze kolumny × wiersze; ogranicza
  to limit dopełniania (ADR 0008).

# 0006 Wynik bezpieczny z konstrukcji

Status: przyjęta (punkt 3: decyzja właściciela; punkty 2 i 4 to wybory wykonawcy, do potwierdzenia) · 2026-11-21

## Kontekst

ADR 0001 ustala zasadę: domyślnie surowy HTML ze źródła jest escapowany, a adresy przechodzą przez listę
dozwolonych schematów; osobny sanityzator nie jest potrzebny. Etap 3 musiał rozstrzygnąć szczegóły.

## Decyzja

1. **Surowy HTML** ma trzy polityki (`Config.RawHTML`): `RawHTMLEscape` (domyślna) zapisuje go jako tekst,
   `RawHTMLOmit` pomija, `RawHTMLPass` przepuszcza bez zmian.
2. **Blok HTML jako tekst zachowuje się jak akapit**: dostaje `<p>`, a w liście zwartej jest samym tekstem.
   Czytelnik widzi, co autor wpisał; struktura dokumentu zostaje poprawna.
3. **Lista domyślna schematów to `http`, `https`, `mailto`, `tel`** (potwierdzona przez właściciela
   2026-11-21; `tel` dopisany na jego decyzję, `ftp` i `data` świadomie poza listą). Adres bez schematu (względny, `#fragment`,
   `//host`) jest dozwolony zawsze. To lista dozwolonych, nie zakazanych: nowy niebezpieczny schemat nie wymaga
   zmiany biblioteki. `data:` nie jest dozwolony domyślnie także dla obrazów; odbiorca może go dopisać.
4. **Adres o niedozwolonym schemacie znika razem ze znacznikiem**: odnośnik zostaje samą swoją treścią, obraz samym
   opisem. Pusty `href=""` wskazywałby bieżącą stronę i udawał działający odnośnik; `<img>` bez `src` jest
   niepoprawny. W drzewie dokumentu adres zostaje, więc walidator odbiorcy może taki odnośnik zgłosić.
5. **Schemat rozpoznaje się na adresie po rozwinięciu** znaków ucieczki i odwołań znakowych, bez rozróżniania
   wielkości liter: litera, potem litery, cyfry i `+ . -` aż do pierwszego dwukropka. Każdy inny bajt przed
   dwukropkiem oznacza adres względny, a w wyniku jest kodowany procentowo (`SAFE-004`), więc przeglądarka nie
   złoży z niego schematu.
6. **Dowód jest testem niezależnym od renderera** (`checkSafe`): wynik bez `RawHTMLPass` zawiera wyłącznie
   znaczniki i atrybuty z zamkniętej listy, a adresy mają postać z punktu 5. Test działa na przykładach
   specyfikacji, korpusie, liście ataków i w fuzzingu.

## Konsekwencje

+ kryterium K6 ma test (`TestSafeByDefault`) i stały sprawdzian w `FuzzRender`;
+ ustawienia niebezpieczne (`RawHTMLPass`, `AnyScheme`) są jawne i nazwane;
− tryb bezpieczny różni się od `goldmark`, który stosuje listę schematów zakazanych i zostawia `href=""`:
  różnice są w `../differences.md`; odnośniki `ftp:`, `irc:`, `sms:` wymagają dopisania schematu przez odbiorcę;
− adres w postaci `host:port/ścieżka` bez `//` jest odczytywany jako schemat `host` i odrzucany, tak jak
  odczytałaby go przeglądarka.

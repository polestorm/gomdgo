# Pytania otwarte

Numery są stałe; pytania rozstrzygnięte znikają z listy, a decyzja jest w ADR.

- pytanie 1 (przykłady specyfikacji w repozytorium) → rozstrzygnięte, `0002-spec-examples-in-repo.md`;
- pytanie 2 (licencja) → rozstrzygnięte, `0013-licence.md`;
- pytanie 3 (publiczne drzewo dokumentu) → rozstrzygnięte, `0010-document-walk.md`;
- pytanie 4 (wzorzec dla przypisów i bloków uwag) → rozstrzygnięte, `0011-alerts-and-footnotes.md`;
- pytanie 5 (pozycje w źródle) → rozstrzygnięte, `0003-source-positions.md`;
- pytanie 6 (niepoprawny UTF-8 na wejściu) → rozstrzygnięte, `0004-invalid-utf8.md`;
- pytanie 7 (sygnatura `Render`) → rozstrzygnięte, `0005-api-shape.md`;
- pytanie 9 (rozmiar wyniku dla tabel) → rozstrzygnięte, `0008-compat-and-table-limit.md`;
- pytanie 10 (zgodność z plikami wzorcowymi odbiorcy przy rozszerzeniach) → rozstrzygnięte, tamże.

8. **Wybory robocze do potwierdzenia.** Z `0006-safe-output.md`: (b) odnośnik o niedozwolonym schemacie zostaje
   samą treścią zamiast `href=""`; (c) blok HTML jako tekst w `<p>`. Z `0007-extensions.md`: (e) ścisłe czytanie
   specyfikacji GFM (dwie tyldy, odstęp po `[ ]`), z opcjami zgodności z ADR 0008. Z `0009-extension-points.md`: (f) kształt funkcji punktów
   rozszerzeń (blok kodu zwraca gotowy HTML, odwołanie zwraca adres, tekst i tytuł). Z `0011-alerts-and-footnotes.md`:
   (g) postać HTML bloków uwag i przypisów (klasy `alert alert-…`, tytuły tylko od odbiorcy, identyfikatory `fn:N`). Punkty (a), lista schematów
   z `tel`, i (d), rozszerzenia domyślnie włączone, właściciel rozstrzygnął 2026-11-21.

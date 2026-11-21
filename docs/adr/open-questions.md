# Pytania otwarte

Numery są stałe; pytania rozstrzygnięte znikają z listy, a decyzja jest w ADR.

- pytanie 1 (przykłady specyfikacji w repozytorium) → rozstrzygnięte, `0002-spec-examples-in-repo.md`;
- pytanie 5 (pozycje w źródle) → rozstrzygnięte, `0003-source-positions.md`;
- pytanie 6 (niepoprawny UTF-8 na wejściu) → rozstrzygnięte, `0004-invalid-utf8.md`;
- pytanie 7 (sygnatura `Render`) → rozstrzygnięte, `0005-api-shape.md`;
- pytanie 9 (rozmiar wyniku dla tabel) → rozstrzygnięte, `0008-compat-and-table-limit.md`;
- pytanie 10 (zgodność z plikami wzorcowymi odbiorcy przy rozszerzeniach) → rozstrzygnięte, tamże.

2. **Licencja `gomdgo`.** Repozytorium jest prywatne. Czy licencja jak w `gopdfgo` (Apache 2.0), czy zamknięta
   do czasu decyzji o udostępnieniu? Przy każdej z nich pliki w `testdata/spec/` zostają na CC BY-SA 4.0 (ADR 0002).
3. **Publiczne drzewo dokumentu.** Czy wystawiamy typy węzłów jako stabilne API (walidatory odbiorców tego potrzebują),
   czy tylko funkcję odwiedzającą z kodem rodzaju węzła i pozycją w źródle? Do rozstrzygnięcia przed etapem 5.
4. **Przypisy i bloki uwag** nie mają specyfikacji. Czy przyjmujemy zachowanie GitHub opisane w jego dokumentacji
   użytkowej, czy własny opis w `requirements/` jako jedyne źródło? Do rozstrzygnięcia przed etapem 5.
8. **Wybory robocze do potwierdzenia.** Z `0006-safe-output.md`: (b) odnośnik o niedozwolonym schemacie zostaje
   samą treścią zamiast `href=""`; (c) blok HTML jako tekst w `<p>`. Z `0007-extensions.md`: (e) ścisłe czytanie
   specyfikacji GFM (dwie tyldy, odstęp po `[ ]`), z opcjami zgodności z ADR 0008. Punkty (a), lista schematów
   z `tel`, i (d), rozszerzenia domyślnie włączone, właściciel rozstrzygnął 2026-11-21.

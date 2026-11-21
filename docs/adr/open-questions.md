# Pytania otwarte

Numery są stałe; pytania rozstrzygnięte znikają z listy, a decyzja jest w ADR.

- pytanie 1 (przykłady specyfikacji w repozytorium) → rozstrzygnięte, `0002-spec-examples-in-repo.md`;
- pytanie 5 (pozycje w źródle) → rozstrzygnięte, `0003-source-positions.md`.

2. **Licencja `gomdgo`.** Repozytorium jest prywatne. Czy licencja jak w `gopdfgo` (Apache 2.0), czy zamknięta
   do czasu decyzji o udostępnieniu? Przy każdej z nich pliki w `testdata/spec/` zostają na CC BY-SA 4.0 (ADR 0002).
3. **Publiczne drzewo dokumentu.** Czy wystawiamy typy węzłów jako stabilne API (walidatory odbiorców tego potrzebują),
   czy tylko funkcję odwiedzającą z kodem rodzaju węzła i pozycją w źródle? Do rozstrzygnięcia przed etapem 5.
4. **Przypisy i bloki uwag** nie mają specyfikacji. Czy przyjmujemy zachowanie GitHub opisane w jego dokumentacji
   użytkowej, czy własny opis w `requirements/` jako jedyne źródło? Do rozstrzygnięcia przed etapem 5.
6. **Niepoprawny UTF-8 na wejściu.** Specyfikacja mówi tylko o U+0000 (zamiana na U+FFFD). Czy niepoprawne bajty
   zamieniamy na U+FFFD, czy przepuszczamy bez zmian? Dziś `DET-001` wymaga jedynie braku paniki i poprawnego
   UTF-8 w wyniku dla poprawnego wejścia. Do rozstrzygnięcia w etapie 1.
7. **Sygnatura `Render`.** Zaślepka etapu 0 to `Render([]byte) []byte`. Opcje (polityka HTML, schematy, punkty
   rozszerzeń) i błędy funkcji wywołującego wymagają decyzji o kształcie API przed etapem 3.

# Pytania otwarte

1. **Przykłady specyfikacji w repozytorium.** Specyfikacja CommonMark jest na licencji CC BY-SA 4.0. Czy trzymamy
   wyciąg przykładów (`spec.json`) w `testdata/spec/` z przypisaniem autorstwa, czy pobieramy go skryptem przy
   pierwszym uruchomieniu testów (wersja przypięta sumą kontrolną)? Tak samo dla GFM.
2. **Licencja `gomdgo`.** Repozytorium jest prywatne. Czy licencja jak w `gopdfgo` (Apache 2.0), czy zamknięta
   do czasu decyzji o udostępnieniu?
3. **Publiczne drzewo dokumentu.** Czy wystawiamy typy węzłów jako stabilne API (walidatory odbiorców tego potrzebują),
   czy tylko funkcję odwiedzającą z kodem rodzaju węzła i pozycją w źródle?
4. **Przypisy i bloki uwag** nie mają specyfikacji. Czy przyjmujemy zachowanie GitHub opisane w jego dokumentacji
   użytkowej, czy własny opis w `requirements/` jako jedyne źródło?
5. **Pozycje w źródle.** Czy każdy węzeł niesie numer wiersza i kolumny (komunikaty walidatorów odbiorcy), mimo
   kosztu pamięci?

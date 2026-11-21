# 0001 Zakres i kryteria zastąpienia parsera odniesienia

Status: przyjęta (robocza) · 2026-11-21

## Kontekst

Produkty, które potrzebują Markdown → HTML, używają dziś `goldmark` (MIT, tylko biblioteka standardowa, pełny
CommonMark). `gomdgo` ma go zastąpić dopiero wtedy, gdy będzie co najmniej tak samo poprawny, a przy tym mniejszy,
bezpieczny z konstrukcji i w całości nasz (rodowód: `../provenance.md`). Do tego czasu `goldmark` jest parserem
w użyciu i odniesieniem w testach różnicowych.

## Decyzja

1. **Zakres składni.**
   - CommonMark **0.31.2**, w całości: także bloki HTML, odsyłacze referencyjne, nagłówki setext, kod przez wcięcie,
     kontynuacja leniwa. Odbiorca może zawęzić składnię swoim walidatorem na drzewie dokumentu; parser rozpoznaje
     wszystko, żeby ten sam tekst nigdy nie znaczył u nas czegoś innego niż w innych narzędziach.
   - GFM **0.29-gfm**, rozszerzenia: tabele, przekreślenie, listy zadań, rozszerzone autolinki, filtr niedozwolonych
     znaczników.
   - Rozszerzenia bez specyfikacji (przypisy, bloki uwag `> [!NOTE]`) według publicznej dokumentacji użytkowej GitHub;
     każde opisane wymaganiami i włączane osobno.
2. **Punkty rozszerzeń zamiast surowego HTML.** Blok kodu z zarejestrowanym językiem (np. `mermaid`) renderuje funkcja
   wywołującego; tak samo odwołania w wierszu (np. `[[NC-2026-017]]`). Drzewo dokumentu jest dostępne dla walidatorów.
3. **Bezpieczeństwo.** Domyślnie surowy HTML ze źródła jest escapowany; adresy odnośników i obrazów przechodzą przez
   listę dozwolonych schematów. Przepuszczenie HTML to jawna opcja.
4. **Determinizm i czas liniowy.** To samo wejście i opcje dają ten sam wynik co do bajta; czas rośnie liniowo także
   dla danych złośliwych.

## Kryteria zastąpienia `goldmark`

Wszystkie muszą być spełnione jednocześnie, zmierzone i zapisane z datą w `../testing.md`:

| # | Kryterium | Miara |
|---|---|---|
| K1 | zgodność z CommonMark | 652 z 652 przykładów specyfikacji 0.31.2 |
| K2 | zgodność z GFM | wszystkie przykłady rozdziałów rozszerzeń GFM 0.29-gfm |
| K3 | podobieństwo do `goldmark` | korpus różnicowy (przykłady specyfikacji, korpus własny, dokumenty odbiorców, wejścia z fuzzingu): HTML po normalizacji identyczny w co najmniej 99,5% przypadków; **każda** różnica opisana w `../differences.md` z uzasadnieniem (gdy `goldmark` odbiega od specyfikacji, wygrywa specyfikacja) |
| K4 | wymagania | każde wymaganie z `../../requirements/` ma test; testy przechodzą |
| K5 | odporność | fuzzing bez paniki i z poprawnym HTML co najmniej 24 h łącznie; zestaw przypadków patologicznych w czasie liniowym |
| K6 | bezpieczeństwo | testy: brak surowego HTML i niedozwolonych schematów w wyniku przy ustawieniach domyślnych |
| K7 | wydajność | nie wolniej niż 2× `goldmark` na korpusie, pomiar `go test -bench` |
| K8 | odbiorca | pliki wzorcowe odbiorcy przechodzą bez zmian po przepięciu z `goldmark` na `gomdgo` |
| K9 | dowód | pakiet dowodów do walidacji u odbiorcy: wymagania → testy → wynik ostatniego przebiegu |

Przepięcie odbiorcy to jego decyzja, podejmowana na podstawie tych wyników.

## Konsekwencje

+ odbiorcy nie czekają: używają `goldmark` od dziś i zmieniają tylko implementację za własnym interfejsem;
+ kryteria są mierzalne i sprawdzane automatycznie, a nie przez ocenę „wydaje się gotowe”;
− pełny CommonMark to większy zakres niż sama odmiana odbiorcy (szacunek w `../roadmap.md`);
− K3 wymaga utrzymywania modułu porównawczego z `goldmark` poza biblioteką (`tools/compare`).

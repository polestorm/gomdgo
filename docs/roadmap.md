# Plan

Szacunki z 2026-11-21, w godzinach pracy z asystentem. Kryteria zakończenia całości: ADR 0001, kryteria K1–K9.

| Etap | Zakres | Szacunek | Stan |
|---|---|---|---|
| 0 | infrastruktura testów: wczytywanie przykładów specyfikacji (pytanie 1), test pokrycia wymagań, szkielet fuzzingu, moduł `tools/compare` z `goldmark` i normalizacją HTML | 4–6 h | gotowy 2026-11-21 |
| 1 | struktura bloków: wiersze i tabulatory, cytaty i listy (z kontynuacją leniwą), nagłówki ATX i setext, linia pozioma, kod przez wcięcie i ogrodzony, akapity, definicje odsyłaczy, bloki HTML (7 rodzajów) | 15–25 h | gotowy 2026-11-21 |
| 2 | treść w wierszu: znaki ucieczki, encje, kod w wierszu, algorytm ograniczników (wyróżnienia), odnośniki i obrazy (w wierszu i referencyjne), autolinki, HTML w wierszu, łamanie wierszy | 15–25 h | gotowy 2026-11-21 |
| 3 | renderer HTML i bezpieczeństwo: escapowanie, lista dozwolonych schematów, polityka surowego HTML (escapuj, pomiń, przepuść), identyfikatory nagłówków | 4–6 h | gotowy 2026-11-21 |
| 4 | GFM: tabele, przekreślenie, listy zadań, rozszerzone autolinki, filtr znaczników | 8–12 h | gotowy 2026-11-21 |
| 5 | punkty rozszerzeń: renderer bloku według języka, odwołania w wierszu, bloki uwag, przypisy, publiczne drzewo dokumentu dla walidatorów | 8–12 h | gotowy 2026-11-21 |
| 6 | utwardzanie: przypadki patologiczne, fuzzing, korpus różnicowy, `differences.md`, testy wydajności | 6–10 h | w toku: dokumenty rzeczywiste, `BenchmarkCorpus` (K7) i generator pakietu dowodów (K9) gotowe 2026-11-21; fuzzing pośredni 9 h 40 min bez błędu; 24 h na kodzie biblioteki z `9c8577b` bez błędu (2026-11-23); 2026-11-24 zmniejszenie pamięci (węzeł drzewa), licznik K5 od nowa, przebieg 24 h w toku (ADR 0012) |
| 7 | odbiór: pomiar K1–K9, pakiet dowodów, próba przepięcia u odbiorcy na jego plikach wzorcowych | 3–5 h | — |
| | **razem** | **65–100 h** | |

Parser odniesienia w `tools/compare`: `goldmark` v1.8.7 (decyzja z 2026-11-21).

Rozmiar docelowy: 5–6 tys. linii kodu bez testów (dla porównania `goldmark` v1.5.3: ok. 15 tys.).

Kolejność jest zależnością: etap 2 potrzebuje drzewa bloków z etapu 1, etap 4 rozszerza oba. Etapy 3 i 5 można
zaczynać, gdy jest prosty akapit. Liczba zaliczonych przykładów specyfikacji jest datowaną miarą postępu po każdym
etapie.

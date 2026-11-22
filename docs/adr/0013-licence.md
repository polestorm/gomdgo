# 0013 Licencja: Apache 2.0

Status: przyjęta · 2026-11-22 (decyzja właściciela; zamyka pytanie otwarte 2)

## Kontekst

Repozytorium jest prywatne, ale może zostać upublicznione. Biblioteki rodziny (`gopdfgo`, `gomermaidgo`) są
na licencji Apache 2.0. `gomdgo` nie zawiera cudzego kodu (`docs/provenance.md`); jedynym cudzym materiałem
są teksty specyfikacji w `testdata/spec/` na licencji CC BY-SA 4.0 (ADR 0002).

## Decyzja

1. Kod i dokumenty `gomdgo` są na licencji Apache License 2.0. Tekst licencji jest w pliku `LICENSE`, a nota
   o prawach autorskich w pliku `NOTICE`.
2. Pliki w `testdata/spec/` zostają na CC BY-SA 4.0 i są wyłączone z licencji kodu; mówi o tym `NOTICE`,
   `README.md` i `testdata/spec/README.md`.
3. Decyzja o upublicznieniu repozytorium jest osobna i nie zapadła; do tego czasu repozytorium pozostaje prywatne.
4. `README.md` zawiera porównanie z parserem odniesienia (wielkość, szybkość, zależności, cechy) z datą i sposobem
   pomiaru; przy każdej zmianie liczb w `docs/testing.md` trzeba je tam uaktualnić.

## Konsekwencje

+ ta sama licencja w całej rodzinie bibliotek, z jawnym udzieleniem licencji patentowej;
+ zerowa lista zależności oznacza brak cudzych not licencyjnych do przenoszenia;
− dane testowe na innej licencji wymagają osobnego opisu przy każdej dystrybucji źródeł.

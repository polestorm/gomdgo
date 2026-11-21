# Wymagania

Każde wymaganie ma identyfikator `PREFIKS-nnn`. Testy i pliki korpusu odwołują się do nich przez `REQ: PREFIKS-nnn`;
test `TestRequirementsCovered` sprawdza pokrycie w obie strony.

| Prefiks | Plik | Zakres |
|---|---|---|
| `BLK` | `blocks.md` | struktura bloków (CommonMark, rozdziały 4–5) |
| `INL` | `inlines.md` | treść w wierszu (CommonMark, rozdział 6) |
| `GFM` | `gfm.md` | rozszerzenia GFM |
| `EXT` | `extensions.md` | punkty rozszerzeń, bloki uwag, przypisy |
| `SAFE` | `safety.md` | escapowanie, schematy adresów, polityka surowego HTML |
| `DET` | `determinism.md` | powtarzalność, czas liniowy, odporność |

Wymaganie dla przykładu specyfikacji wskazuje rozdział, a nie pojedynczy przykład; przykłady są testami tego wymagania.
Pliki powstają razem z etapami w `../docs/roadmap.md`.

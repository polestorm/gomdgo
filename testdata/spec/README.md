# Teksty specyfikacji (dane testowe)

Pliki w tym katalogu są **danymi testowymi, nie kodem**. Nie są objęte licencją `gomdgo`; pozostają na licencji
swoich autorów. Decyzja o trzymaniu ich w repozytorium: `docs/adr/0002-spec-examples-in-repo.md`.

| Plik | Dzieło | Autor | Źródło | Licencja | SHA-256 |
|---|---|---|---|---|---|
| `commonmark-0.31.2.txt` | CommonMark Spec 0.31.2 (2024-01-28) | John MacFarlane | `spec.txt`, repozytorium `commonmark/commonmark-spec`, tag `0.31.2` | [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/) | `257c41ad946f7a1414a499aca402a1aa8fdac3678532266611348c1cf54f4b80` |
| `gfm-0.29.txt` | GitHub Flavored Markdown Spec 0.29 (2019-04-06) | GitHub, na podstawie CommonMark Spec Johna MacFarlane'a | `test/spec.txt`, repozytorium `github/cmark-gfm`, tag `0.29.0.gfm.0` | [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/) | `7cea1221ffba48559d8748c8510d3c5bda40487a13667b80e77c14a1505b9821` |

Pliki są niezmienione względem źródła (pobrane 2026-11-21). Nie poprawiamy ich ręcznie: zmiana wersji specyfikacji
to nowy plik i nowa decyzja.

Tag `0.29.0.gfm.0` odpowiada opublikowanej wersji 0.29-gfm. Późniejsze tagi `0.29.0.gfm.*` zmieniają także część
wspólną z CommonMark (m.in. zagnieżdżone `<strong>`, komentarze HTML) i dodają przykłady rozszerzonych autolinków;
dla części wspólnej wzorcem jest u nas CommonMark 0.31.2, a z pliku GFM mierzymy tylko rozdziały rozszerzeń.

## Pliki pochodne

`passing-commonmark.txt` i `passing-gfm.txt` to listy numerów przykładów, które `gomdgo` zalicza. Są nasze,
przepisuje je `go test -update` (opis w `docs/testing.md`).

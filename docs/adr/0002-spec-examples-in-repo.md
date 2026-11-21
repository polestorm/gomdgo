# 0002 Teksty specyfikacji w repozytorium

Status: przyjęta · 2026-11-21 (decyzja właściciela; zamyka pytanie otwarte 1)

## Kontekst

Kryteria K1 i K2 mierzymy na przykładach ze specyfikacji CommonMark 0.31.2 i GFM 0.29-gfm. Obie są na licencji
CC BY-SA 4.0. Do wyboru było: trzymać je w repozytorium albo pobierać skryptem przy pierwszym uruchomieniu testów.

## Decyzja

1. Trzymamy w `testdata/spec/` **niezmienione pliki `spec.txt`** obu specyfikacji, a nie pochodny wyciąg JSON:
   to jedno źródło i jeden format dla obu, a zarazem tekst, z którego piszemy parser. Przykłady wyciąga
   `internal/spectest`.
2. `testdata/spec/README.md` podaje dla każdego pliku dzieło, autora, źródło (repozytorium i tag), licencję
   i sumę SHA-256. Pliki są danymi testowymi, nie kodem, i nie są objęte licencją `gomdgo`.
3. Dla GFM bierzemy tag `0.29.0.gfm.0` (opublikowana wersja 0.29-gfm) i mierzymy na nim tylko rozdziały rozszerzeń;
   dla części wspólnej wzorcem jest CommonMark 0.31.2.
4. Plików nie poprawiamy ręcznie. Nowa wersja specyfikacji to nowy plik i wpis w tej decyzji.

## Konsekwencje

+ `go test ./...` działa bez sieci i jest powtarzalne;
+ pochodzenie każdego przykładu da się sprawdzić sumą kontrolną;
− repozytorium zawiera ok. 400 kB cudzego tekstu na licencji CC BY-SA 4.0; przy wyborze licencji `gomdgo`
  (pytanie otwarte 2) i przy ewentualnym udostępnieniu trzeba zachować przypisanie autorstwa i oznaczyć te pliki
  jako wyłączone z licencji kodu.

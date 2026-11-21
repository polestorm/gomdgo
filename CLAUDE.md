# gomdgo — instrukcje dla Claude Code

Projekt: parser i renderer Markdown → HTML w czystym Go. Cele: zgodność z CommonMark i wybranymi rozszerzeniami GFM,
zero zależności poza biblioteką standardową, wynik bezpieczny z konstrukcji (bez osobnego sanityzatora), powtarzalny
co do bajta, czas liniowy także dla danych złośliwych, mały kod, który da się przeczytać i zwalidować. Projekt jest
agnostyczny wobec odbiorców.

Zacznij od `README.md`, potem `docs/roadmap.md`, `docs/adr/0001-scope-and-acceptance.md` i `docs/provenance.md`.

Zasady:
- **czysty rodowód** (`docs/provenance.md`): piszemy ze specyfikacji CommonMark i GFM. Nie czytaj, nie kopiuj i nie
  przenoś kodu innych parserów Markdown (goldmark, blackfriday, gomarkdown, cmark, commonmark.js, markdown-it, pulldown-cmark).
  Wolno uruchamiać je jako czarną skrzynkę, żeby porównać wynik;
- **zero zależności**: `go.mod` w katalogu głównym nie ma sekcji `require`; porównanie z parserem odniesienia żyje
  w osobnym module `tools/compare` i nie jest importowane przez bibliotekę;
- wzorcem są specyfikacje, nie zachowanie parsera odniesienia: gdy goldmark różni się od specyfikacji, wygrywa
  specyfikacja, a różnicę wpisujemy do `docs/differences.md`;
- nazwy plików, pakietów i identyfikatorów po angielsku; komentarze i dokumentacja po polsku; `README.md` po angielsku;
  komunikaty błędów biblioteki po angielsku, ze stałym kodem do tłumaczenia przez wywołującego;
- każdy obsługiwany element składni ma wymaganie w `requirements/` i test, który je wskazuje;
- każda decyzja = plik w `docs/adr/`; pytania bez decyzji w `docs/adr/open-questions.md`;
- liczby (zaliczone przykłady, zgodność z odniesieniem) są datowane i podpisane sposobem pomiaru;
- opisy commitów po angielsku; przed commitem muszą przechodzić `go vet ./...` i `go test ./...`.

# Wymagania: API i postać wyniku

Decyzja o kształcie API: `docs/adr/0005-api-shape.md`.

| Id | Wymaganie |
|---|---|
| OUT-001 | API ma dwa kroki: `Converter.Parse` zwraca dokument i nie zawodzi dla żadnego wejścia, `Document.HTML` zapisuje go jako HTML i zwraca błąd tylko z funkcji wywołującego podłączonej w punkcie rozszerzenia. `Render(src)` to skrót z ustawieniami domyślnymi (CommonMark z rozszerzeniami GFM, wynik bezpieczny), bez błędu, równy `New(Config{}).Render(src)`. Konwerter i dokument są niezmienne: można ich używać z wielu gorutyn, a kolejne wywołania `HTML` dają ten sam wynik. |
| OUT-002 | Z `Config.HeadingIDs` każdy nagłówek dostaje atrybut `id` utworzony z jego tekstu bez znaczników: litery i cyfry (także spoza ASCII) małymi literami, odstępy jako łączniki, `-` i `_` zachowane, pozostałe znaki pominięte; białe znaki na brzegach tekstu są pomijane; pusty wynik daje `section`. Identyfikator jest unikalny w dokumencie: powtórzony dostaje kolejny numer po łączniku. Domyślnie nagłówki nie mają `id`. |

# 0017 Wersja v0.1.1 i zasada: wydanych znaczników nie przesuwamy

Status: przyjęta · 2026-11-25 (polecenie właściciela)

## Kontekst

Znacznik v0.1.0 powstał 2026-11-24, a potem tego samego dnia historia została przepisana (ADR 0016, punkt 5)
i znacznik wskazał nowy commit; po nim doszła jeszcze zmiana najniższej wersji Go na 1.22 (punkt 8).
Skutek stwierdzony 2026-11-25 przy zakładaniu `gopeekgo`: pod tą samą nazwą wersji istnieją dwie treści modułu.

| Skąd | Suma modułu v0.1.0 | Dyrektywa `go` |
|---|---|---|
| pamięć podręczna modułów na komputerze właściciela (pobrana przed przesunięciem) | `h1:cpwoQcxVXFU6Vnc3ulBpZEYzKKqRaeQ9n1Y3pQMQk9o=` | 1.26.0 |
| GitHub, pobranie do czystej pamięci podręcznej | `h1:PCs0HC1uUdQCUOQ3p5h70S4ITA3R2rqh+AhEptPS0gY=` | 1.22 |

Odbiorcy (`gopressgo`, `gopeekgo`) zapisali w `go.sum` starą sumę: budowały się lokalnie, ale na czystej
maszynie pobranie kończyłoby się błędem sumy kontrolnej. Go zakłada, że treść wydanej wersji nigdy się nie zmienia.

## Decyzja

1. Wydajemy v0.1.1 z bieżącego stanu gałęzi `main`. Kod biblioteki nie różni się od v0.1.0 z GitHuba poza
   komentarzami dokumentacyjnymi publicznego API (po angielsku); licznik K5 biegnie dalej (ADR 0012).
2. Odbiorcy przechodzą na v0.1.1. Wersji v0.1.0 nie używamy; znacznik zostaje, żeby nie przesuwać go po raz trzeci.
3. Zasada na przyszłość: wydanego znacznika wersji nie przesuwamy i nie usuwamy. Błąd w wydaniu naprawia
   następna wersja. Przepisanie historii po wydaniu wymaga nowego numeru wersji.

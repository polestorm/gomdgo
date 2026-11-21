# 0005 Kształt publicznego API

Status: przyjęta · 2026-11-21 (decyzja właściciela; zamyka pytanie otwarte 7)

## Kontekst

API musi pomieścić ustawienia (polityka surowego HTML, schematy adresów, identyfikatory nagłówków, rozszerzenia),
punkty rozszerzeń, które jako jedyne mogą zwrócić błąd (funkcja renderująca blok kodu według języka, odwołania
w wierszu), oraz dostęp do drzewa dla walidatorów odbiorcy. Rozważane były trzy kształty: jedna funkcja z opcjami
funkcyjnymi, struktura ustawień z konwerterem wielokrotnego użytku, oraz dwa kroki (parsowanie i renderowanie osobno).

## Decyzja

Dwa kroki, z konwerterem wielokrotnego użytku i skrótem dla ustawień domyślnych:

```go
conv := gomdgo.New(gomdgo.Config{...}) // ustawienia sprawdzane raz
doc := conv.Parse(src)                 // nigdy nie zawodzi
// tu walidator odbiorcy ogląda dokument
html, err := doc.HTML()                // błąd tylko z funkcji wywołującego

html, err := conv.Render(src)          // skrót: Parse + HTML
html := gomdgo.Render(src)             // skrót: ustawienia domyślne, bez błędu
```

1. `Config` jest zwykłą strukturą; wartość zerowa to ustawienia domyślne: bezpieczne, z rozszerzeniami GFM
   (ADR 0007). Pola opisują **co** ma być
   w wyniku, a nie jak działa parser.
2. `Converter` po utworzeniu jest niezmienny (kopiuje listy z `Config`) i bezpieczny dla wielu gorutyn.
3. `Parse` nie zwraca błędu: każdy ciąg bajtów jest jakimś dokumentem Markdown (ADR 0004 dla niepoprawnego UTF-8).
4. `Document` jest niezmienny; `HTML` można wołać wielokrotnie i równolegle, zawsze z tym samym wynikiem.
5. Błąd z punktu rozszerzenia **przerywa** renderowanie: wynik jest albo kompletny, albo go nie ma. Walidator może
   sprawdzić bloki wcześniej, na dokumencie.

## Konsekwencje

+ jedno parsowanie na dokument także wtedy, gdy odbiorca waliduje przed renderowaniem;
+ `error` pojawia się tylko tam, gdzie może powstać; najprostsze wywołanie zostaje jednowierszowe;
− typ `Document` jest częścią publicznego API. Ile drzewa wystawia (typy węzłów czy funkcja odwiedzająca),
  rozstrzyga pytanie 3, przed etapem 5; do tego czasu `Document` ma tylko metodę `HTML`;
− tymczasowe `RenderWith` i `Options` z etapu 1 znikają bez okresu przejściowego (nie miały użytkowników).

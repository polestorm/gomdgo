# Wymagania: powtarzalność i odporność

| Id | Wymaganie |
|---|---|
| DET-001 | Dowolne wejście, także niepoprawny UTF-8 i znaki sterujące, nie powoduje paniki, a wynik jest poprawnym fragmentem HTML: znaczniki są domknięte w kolejności otwarcia, znaki `<` i `&` występują tylko jako początek znacznika i odwołania znakowego. Wynik jest zawsze poprawnym UTF-8 i nie zawiera znaku U+0000. |
| DET-002 | To samo wejście daje ten sam wynik co do bajta przy każdym wywołaniu. |
| DET-003 | Czas przetwarzania rośnie liniowo z długością wejścia także dla wejść dobranych złośliwie. Zestaw przypadków patologicznych rośnie z etapami; obejmuje strukturę bloków (głębokie zagnieżdżenia cytatów i list, kontynuację leniwą w głębokim zagnieżdżeniu, długie serie definicji odsyłaczy, niezamknięte bloki) i treść w wierszu (serie ograniczników wyróżnień, niedomknięte nawiasy kwadratowe i okrągłe, serie znaków kodu różnej długości, niedomknięte komentarze i wartości atrybutów, głębokie zagnieżdżenie wyróżnień, odnośników i obrazów). |
| DET-004 | Niepoprawny UTF-8 na wejściu: każda nieprzerwana seria bajtów, które nie tworzą poprawnie zakodowanego znaku, jest zastępowana jednym znakiem U+FFFD przed parsowaniem, tak jak znak U+0000 (ADR 0004). Poprawne wejście nie jest kopiowane. Pozycje węzłów odnoszą się do źródła po zamianie; numery wierszy są takie same jak w oryginale. |

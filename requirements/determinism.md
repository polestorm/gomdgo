# Wymagania: powtarzalność i odporność

| Id | Wymaganie |
|---|---|
| DET-001 | Dowolne wejście, także niepoprawny UTF-8 i znaki sterujące, nie powoduje paniki, a wynik jest poprawnym fragmentem HTML: znaczniki są domknięte w kolejności otwarcia, znaki `<` i `&` występują tylko jako początek znacznika i odwołania znakowego. Dla wejścia w poprawnym UTF-8 wynik jest poprawnym UTF-8. |
| DET-002 | To samo wejście daje ten sam wynik co do bajta przy każdym wywołaniu. |
| DET-003 | Czas przetwarzania rośnie liniowo z długością wejścia także dla wejść dobranych złośliwie. Zestaw przypadków patologicznych rośnie z etapami; po etapie 1 obejmuje strukturę bloków: głębokie zagnieżdżenia cytatów i list, kontynuację leniwą w głębokim zagnieżdżeniu, długie serie definicji odsyłaczy i niezamknięte bloki. |

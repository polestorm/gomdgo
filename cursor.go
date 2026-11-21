package gomdgo

// cursor to pozycja w wierszu z rachunkiem kolumn. Tabulator nie jest zamieniany na spacje w treści, ale
// przy wyznaczaniu struktury bloków zachowuje się jak dopełnienie do następnej kolumny podzielnej przez 4.
// Gdy struktura zużywa tylko część tabulatora, reszta zostaje jako spacje wirtualne (pad).
type cursor struct {
	s   []byte // wiersz bez znaku końca
	i   int    // pozycja pierwszego niezużytego bajtu
	col int    // kolumna, w której zaczyna się s[i]
	pad int    // spacje wirtualne przed s[i], pozostałe z częściowo zużytego tabulatora
}

// indent zwraca szerokość wiodących białych znaków w kolumnach i pozycję pierwszego bajtu za nimi.
func (c *cursor) indent() (cols, next int) {
	cols, col := c.pad, c.col
	for next = c.i; next < len(c.s); next++ {
		switch c.s[next] {
		case ' ':
			cols++
			col++
		case '\t':
			w := 4 - col%4
			cols += w
			col += w
		default:
			return cols, next
		}
	}
	return cols, next
}

// blank mówi, czy reszta wiersza to same spacje i tabulatory.
func (c *cursor) blank() bool {
	_, next := c.indent()
	return next == len(c.s)
}

// skipCols zużywa najwyżej n kolumn białych znaków.
func (c *cursor) skipCols(n int) {
	for n > 0 {
		if c.pad > 0 {
			k := min(n, c.pad)
			c.pad -= k
			n -= k
			continue
		}
		if c.i >= len(c.s) {
			return
		}
		switch c.s[c.i] {
		case ' ':
			c.i++
			c.col++
			n--
		case '\t':
			w := 4 - c.col%4
			c.i++
			c.col += w
			if w > n {
				c.pad = w - n
			}
			n -= w
		default:
			return
		}
	}
}

// skipSpaces zużywa wszystkie wiodące białe znaki.
func (c *cursor) skipSpaces() {
	cols, next := c.indent()
	c.col += cols - c.pad
	c.pad = 0
	c.i = next
}

// advance zużywa n bajtów niebędących tabulatorami.
func (c *cursor) advance(n int) {
	c.i += n
	c.col += n
}

// appendRest dopisuje resztę wiersza, ze spacjami wirtualnymi na początku.
func (c *cursor) appendRest(dst []byte) []byte {
	for k := 0; k < c.pad; k++ {
		dst = append(dst, ' ')
	}
	return append(dst, c.s[c.i:]...)
}

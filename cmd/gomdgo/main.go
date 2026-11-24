// Polecenie gomdgo zamienia dokument Markdown na HTML biblioteką gomdgo.
//
//	gomdgo [flagi] [plik.md]
//
// Bez nazwy pliku (albo z nazwą "-") czyta standardowe wejście; wynik trafia na standardowe wyjście albo
// do pliku wskazanego flagą -o. Ustawienia domyślne są ustawieniami biblioteki: CommonMark z rozszerzeniami
// GFM, surowy HTML zapisany jako tekst, adresy tylko o dozwolonych schematach. Polecenie nie uruchamia innych
// programów i nie zna diagramów ani PDF (docs/adr/0014-command-line.md).
package main

import (
	"os"

	"github.com/polestorm/gomdgo/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, cli.Options{}))
}

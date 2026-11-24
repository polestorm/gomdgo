// Polecenie gomdgo-mermaid to polecenie gomdgo z jednym dodatkiem: blok kodu o języku "mermaid" jest rysowany
// biblioteką gomermaidgo i trafia do wyniku jako SVG. Flagi, wejście i wyjście są takie same jak w gomdgo.
//
//	gomdgo-mermaid [flagi] [plik.md]
//
// Żyje w osobnym module, bo zależy od gomermaidgo; biblioteka gomdgo nie ma zależności
// (docs/adr/0015-combined-binary.md). Diagram powstaje w tym samym procesie, przez punkt rozszerzenia
// Config.CodeBlocks, więc surowy HTML z dokumentu pozostaje zapisany jako tekst. Błąd w diagramie przerywa
// pracę z kodem wyjścia 1 i podaje wiersz dokumentu.
package main

import (
	"os"

	"github.com/polestorm/gomdgo/internal/cli"
	gomdgo "github.com/polestorm/gomdgo/markdown"
	"github.com/polestorm/gomermaidgo/markdown"
	"github.com/polestorm/gomermaidgo/mermaid"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, options()))
}

func options() cli.Options {
	return cli.Options{
		Name:       "gomdgo-mermaid",
		CodeBlocks: map[string]gomdgo.CodeBlockFunc{"mermaid": diagram},
		Head:       "<style>" + markdown.CSS + "</style>\n",
	}
}

// diagram rysuje jeden blok mermaid. Opakowanie <figure class="diagram"> jest takie samo jak
// w gomermaidgo/markdown, więc pasuje do niego ten sam arkusz stylów.
func diagram(b gomdgo.CodeBlock) ([]byte, error) {
	svg, err := mermaid.Render(string(b.Code))
	if err != nil {
		return nil, err
	}
	out := append([]byte("<figure class=\"diagram\">\n"), svg...)
	if n := len(out); n > 0 && out[n-1] != '\n' {
		out = append(out, '\n')
	}
	return append(out, "</figure>\n"...), nil
}

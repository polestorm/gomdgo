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
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"strings"

	"github.com/polestorm/gomdgo"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run wykonuje polecenie i zwraca kod wyjścia: 0 przy powodzeniu, 1 przy błędzie odczytu lub zapisu,
// 2 przy błędnych argumentach.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gomdgo", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		output     = fs.String("o", "", "write the result to this `file` instead of standard output")
		rawHTML    = fs.String("raw-html", "escape", "raw HTML `policy`: escape, omit or pass")
		schemes    = fs.String("schemes", "", "comma-separated `list` of allowed link schemes (default http,https,mailto,tel)")
		anyScheme  = fs.Bool("any-scheme", false, "allow every link scheme")
		commonMark = fs.Bool("commonmark", false, "plain CommonMark: turn the GFM extensions off")
		alerts     = fs.Bool("alerts", false, "enable alerts (> [!NOTE])")
		footnotes  = fs.Bool("footnotes", false, "enable footnotes")
		headingIDs = fs.Bool("heading-ids", false, "add id attributes to headings")
		tilde      = fs.Bool("single-tilde", false, "strike through ~text~ as well as ~~text~~")
		alignStyle = fs.Bool("table-align-style", false, "write table alignment as a style attribute")
		standalone = fs.Bool("standalone", false, "write a complete HTML document instead of a fragment")
		title      = fs.String("title", "", "document `title` for -standalone")
	)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: gomdgo [flags] [file.md]")
		fmt.Fprintln(stderr, "Converts Markdown to HTML. Reads standard input when no file is given.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	fail := func(code int, format string, a ...any) int {
		fmt.Fprintf(stderr, "gomdgo: "+format+"\n", a...)
		return code
	}
	if fs.NArg() > 1 {
		return fail(2, "expected at most one input file, got %d", fs.NArg())
	}

	cfg := gomdgo.Config{
		AnyScheme:       *anyScheme,
		HeadingIDs:      *headingIDs,
		SingleTilde:     *tilde,
		TableAlignStyle: *alignStyle,
	}
	switch *rawHTML {
	case "escape":
		cfg.RawHTML = gomdgo.RawHTMLEscape
	case "omit":
		cfg.RawHTML = gomdgo.RawHTMLOmit
	case "pass":
		cfg.RawHTML = gomdgo.RawHTMLPass
	default:
		return fail(2, "unknown -raw-html policy %q (escape, omit, pass)", *rawHTML)
	}
	schemesSet := false
	fs.Visit(func(f *flag.Flag) { schemesSet = schemesSet || f.Name == "schemes" })
	if schemesSet {
		// Pusta lista jest poprawna: dopuszcza tylko adresy bez schematu.
		cfg.Schemes = []string{}
		for _, s := range strings.Split(*schemes, ",") {
			if s = strings.TrimSpace(s); s != "" {
				cfg.Schemes = append(cfg.Schemes, s)
			}
		}
	}
	if *commonMark {
		cfg.Disable = gomdgo.GFM
	}
	if *alerts {
		cfg.Enable |= gomdgo.Alerts
	}
	if *footnotes {
		cfg.Enable |= gomdgo.Footnotes
	}

	var src []byte
	var err error
	if name := fs.Arg(0); name == "" || name == "-" {
		src, err = io.ReadAll(stdin)
	} else {
		src, err = os.ReadFile(name)
	}
	if err != nil {
		return fail(1, "%v", err)
	}

	body, err := gomdgo.New(cfg).Render(src)
	if err != nil {
		return fail(1, "%v", err)
	}
	out := body
	if *standalone {
		out = []byte("<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n<title>" +
			html.EscapeString(*title) + "</title>\n</head>\n<body>\n")
		out = append(out, body...)
		out = append(out, "</body>\n</html>\n"...)
	}

	if *output == "" {
		_, err = stdout.Write(out)
	} else {
		err = os.WriteFile(*output, out, 0o644)
	}
	if err != nil {
		return fail(1, "%v", err)
	}
	return 0
}

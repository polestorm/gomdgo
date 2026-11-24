// Package cli zawiera wspólną część poleceń wiersza poleceń: flagi odpowiadające polom markdown.Config, odczyt
// wejścia i zapis wyniku. Polecenie cmd/gomdgo woła Run bez rozszerzeń; polecenie łączące gomdgo z inną
// biblioteką (cmd/gomdgo-mermaid, osobny moduł) przekazuje w Options funkcje dla bloków kodu
// (docs/adr/0014-command-line.md, docs/adr/0015-combined-binary.md).
package cli

import (
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"strings"

	"github.com/polestorm/gomdgo/markdown"
)

// Options opisuje, czym polecenie różni się od zwykłego gomdgo.
type Options struct {
	Name       string                            // nazwa polecenia w komunikatach; pusta oznacza "gomdgo"
	CodeBlocks map[string]markdown.CodeBlockFunc // funkcje dla bloków kodu według języka (Config.CodeBlocks)
	Head       string                            // HTML dopisywany do <head> przy -standalone, np. arkusz stylów
}

// Run wykonuje polecenie i zwraca kod wyjścia: 0 przy powodzeniu, 1 przy błędzie odczytu lub zapisu,
// 2 przy błędnych argumentach.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer, opts Options) int {
	name := opts.Name
	if name == "" {
		name = "gomdgo"
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
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
		fmt.Fprintf(stderr, "usage: %s [flags] [file.md]\n", name)
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
		fmt.Fprintf(stderr, "%s: %s\n", name, fmt.Sprintf(format, a...))
		return code
	}
	if fs.NArg() > 1 {
		return fail(2, "expected at most one input file, got %d", fs.NArg())
	}

	cfg := markdown.Config{
		AnyScheme:       *anyScheme,
		HeadingIDs:      *headingIDs,
		SingleTilde:     *tilde,
		TableAlignStyle: *alignStyle,
		CodeBlocks:      opts.CodeBlocks,
	}
	switch *rawHTML {
	case "escape":
		cfg.RawHTML = markdown.RawHTMLEscape
	case "omit":
		cfg.RawHTML = markdown.RawHTMLOmit
	case "pass":
		cfg.RawHTML = markdown.RawHTMLPass
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
		cfg.Disable = markdown.GFM
	}
	if *alerts {
		cfg.Enable |= markdown.Alerts
	}
	if *footnotes {
		cfg.Enable |= markdown.Footnotes
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

	body, err := markdown.New(cfg).Render(src)
	if err != nil {
		return fail(1, "%v", err)
	}
	out := body
	if *standalone {
		out = []byte("<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n<title>" +
			html.EscapeString(*title) + "</title>\n" + opts.Head + "</head>\n<body>\n")
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

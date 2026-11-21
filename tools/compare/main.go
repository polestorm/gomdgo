// Polecenie compare mierzy kryterium K3: ten sam korpus przechodzi przez gomdgo i przez parser odniesienia
// (goldmark, uruchamiany przez publiczne API jako czarna skrzynka), a wyniki są porównywane po normalizacji HTML.
//
// Żyje w osobnym module, bo zależy od goldmark; biblioteka go nie importuje.
//
//	go run . -spec ../../testdata/spec/commonmark-0.31.2.txt -spec ../../testdata/spec/gfm-0.29.txt \
//	    -out ../../out/compare ../../testdata/corpus/*.md
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/polestorm/gomdgo"
	"github.com/polestorm/gomdgo/internal/htmlnorm"
	"github.com/polestorm/gomdgo/internal/spectest"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

// input to jeden dokument korpusu różnicowego.
type input struct {
	name     string
	markdown string
	gfm      bool   // parser odniesienia z rozszerzeniami GFM
	want     string // oczekiwany HTML ze specyfikacji; pusty dla plików korpusu
	spec     bool
}

// tally liczy zgodności w jednej grupie wejść.
type tally struct {
	total, same       int // gomdgo i goldmark po normalizacji
	specTotal         int
	oursSpec, refSpec int // zgodność każdego z parserów ze specyfikacją, po normalizacji
}

func main() {
	var specs multiFlag
	flag.Var(&specs, "spec", "plik specyfikacji z przykładami (można podać kilka razy)")
	outDir := flag.String("out", "", "katalog na raport różnic (poza repozytorium, np. ../../out/compare)")
	flag.Parse()

	// Przepuszczanie surowego HTML jest włączone, bo przykłady specyfikacji go oczekują. Porównanie w trybie
	// domyślnym gomdgo (HTML escapowany) dojdzie razem z polityką surowego HTML w etapie 3.
	plain := goldmark.New(goldmark.WithRendererOptions(html.WithUnsafe()))
	gfm := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(html.WithUnsafe()))

	var inputs []input
	var groups []string
	for _, path := range specs {
		examples, err := spectest.Load(path)
		if err != nil {
			fatal(err)
		}
		groups = append(groups, filepath.Base(path))
		// Ze specyfikacji GFM bierzemy tylko rozdziały rozszerzeń (tak jak TestSpecGFM): jej część wspólna
		// to starsza wersja CommonMark, dla której wzorcem jest plik CommonMark.
		extensionsOnly := false
		for _, e := range examples {
			extensionsOnly = extensionsOnly || e.IsExtension()
		}
		for _, e := range examples {
			if extensionsOnly && !e.IsExtension() {
				continue
			}
			inputs = append(inputs, input{
				name:     fmt.Sprintf("%s#%d", filepath.Base(path), e.Number),
				markdown: e.Markdown,
				gfm:      e.IsExtension(),
				want:     e.HTML,
				spec:     true,
			})
		}
	}
	if flag.NArg() > 0 {
		groups = append(groups, "corpus")
	}
	for _, path := range flag.Args() {
		data, err := os.ReadFile(path)
		if err != nil {
			fatal(err)
		}
		inputs = append(inputs, input{name: "corpus#" + filepath.Base(path), markdown: string(data), gfm: true})
	}
	if len(inputs) == 0 {
		fatal(fmt.Errorf("no inputs: give -spec files or Markdown files"))
	}

	counts := map[string]*tally{}
	var report strings.Builder
	for _, in := range inputs {
		group := in.name[:strings.IndexByte(in.name, '#')]
		c := counts[group]
		if c == nil {
			c = &tally{}
			counts[group] = c
		}
		md := plain
		if in.gfm {
			md = gfm
		}
		var buf bytes.Buffer
		if err := md.Convert([]byte(in.markdown), &buf); err != nil {
			fatal(fmt.Errorf("%s: reference parser: %w", in.name, err))
		}
		ours := htmlnorm.Normalize(string(gomdgo.Render([]byte(in.markdown))))
		ref := htmlnorm.Normalize(buf.String())
		c.total++
		if ours == ref {
			c.same++
		} else {
			fmt.Fprintf(&report, "=== %s\n--- markdown\n%s--- gomdgo\n%s\n--- goldmark\n%s\n\n", in.name, in.markdown, ours, ref)
		}
		if in.spec {
			want := htmlnorm.Normalize(in.want)
			c.specTotal++
			if ours == want {
				c.oursSpec++
			}
			if ref == want {
				c.refSpec++
			} else {
				fmt.Fprintf(&report, "=== %s: goldmark differs from the specification\n--- markdown\n%s--- spec\n%s\n--- goldmark\n%s\n\n",
					in.name, in.markdown, want, ref)
			}
		}
	}

	fmt.Printf("reference: goldmark %s (raw HTML passed through)\n\n", referenceVersion())
	fmt.Printf("%-24s %8s %22s %18s %18s\n", "group", "inputs", "gomdgo = goldmark", "gomdgo = spec", "goldmark = spec")
	var all tally
	for _, g := range groups {
		c := counts[g]
		if c == nil {
			continue
		}
		printRow(g, c)
		all.total += c.total
		all.same += c.same
		all.specTotal += c.specTotal
		all.oursSpec += c.oursSpec
		all.refSpec += c.refSpec
	}
	printRow("total", &all)

	if *outDir != "" {
		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			fatal(err)
		}
		path := filepath.Join(*outDir, "differences.txt")
		if err := os.WriteFile(path, []byte(report.String()), 0o644); err != nil {
			fatal(err)
		}
		fmt.Printf("\ndifferences written to %s\n", path)
	}
}

func printRow(name string, c *tally) {
	spec := func(n int) string {
		if c.specTotal == 0 {
			return "—"
		}
		return fmt.Sprintf("%d / %d", n, c.specTotal)
	}
	fmt.Printf("%-24s %8d %13d (%5.1f%%) %18s %18s\n", name, c.total, c.same,
		100*float64(c.same)/float64(c.total), spec(c.oursSpec), spec(c.refSpec))
}

func referenceVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/yuin/goldmark" {
				return dep.Version
			}
		}
	}
	return "unknown"
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "compare:", err)
	os.Exit(1)
}

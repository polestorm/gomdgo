package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gmd "github.com/gomarkdown/markdown"
	gmdhtml "github.com/gomarkdown/markdown/html"
	gmdparser "github.com/gomarkdown/markdown/parser"
	"github.com/polestorm/gomdgo/internal/htmlnorm"
	"github.com/polestorm/gomdgo/internal/spectest"
	"github.com/polestorm/gomdgo/markdown"
	"github.com/russross/blackfriday/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// Plik porównuje gomdgo z trzema bibliotekami Markdown w Go: goldmark (parser odniesienia), gomarkdown/markdown
// i blackfriday v2. Wszystkie są uruchamiane przez publiczne API jako czarne skrzynki (docs/provenance.md).
// To zestawienie do README, a nie kryterium: K3 i K7 mierzy się tylko wobec goldmark.

// rival to jedna biblioteka w dwóch ustawieniach: bez rozszerzeń i z rozszerzeniami typu GFM.
type rival struct {
	name       string
	plain, gfm func(src []byte) []byte
}

func rivals() []rival {
	pass := markdown.Config{RawHTML: markdown.RawHTMLPass, AnyScheme: true}
	plainCfg := pass
	plainCfg.Disable = markdown.GFM
	ours, oursPlain := markdown.New(pass), markdown.New(plainCfg)
	ref := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(html.WithUnsafe()))
	refPlain := goldmark.New(goldmark.WithRendererOptions(html.WithUnsafe()))

	oursRun := func(c *markdown.Converter) func([]byte) []byte {
		return func(src []byte) []byte {
			out, err := c.Render(src)
			if err != nil {
				panic(err)
			}
			return out
		}
	}
	refRun := func(md goldmark.Markdown) func([]byte) []byte {
		return func(src []byte) []byte {
			var buf bytes.Buffer
			if err := md.Convert(src, &buf); err != nil {
				panic(err)
			}
			return buf.Bytes()
		}
	}
	// gomarkdown: parser jest jednorazowy, więc powstaje dla każdego dokumentu.
	gmdRun := func(ext gmdparser.Extensions) func([]byte) []byte {
		return func(src []byte) []byte {
			r := gmdhtml.NewRenderer(gmdhtml.RendererOptions{})
			return gmd.ToHTML(src, gmdparser.NewWithExtensions(ext), r)
		}
	}
	bfRun := func(ext blackfriday.Extensions) func([]byte) []byte {
		return func(src []byte) []byte {
			return blackfriday.Run(src, blackfriday.WithExtensions(ext))
		}
	}
	return []rival{
		{"gomdgo", oursRun(oursPlain), oursRun(ours)},
		{"goldmark", refRun(refPlain), refRun(ref)},
		{"gomarkdown", gmdRun(gmdparser.NoExtensions), gmdRun(gmdparser.CommonExtensions)},
		{"blackfriday", bfRun(blackfriday.NoExtensions), bfRun(blackfriday.CommonExtensions)},
	}
}

// TestRivalsSpec liczy, ile przykładów CommonMark 0.31.2 każda biblioteka zalicza po normalizacji HTML,
// w ustawieniu bez rozszerzeń i z rozszerzeniami; wynik jest tylko logowany (go test -v -run TestRivalsSpec .).
func TestRivalsSpec(t *testing.T) {
	examples, err := spectest.Load("../../testdata/spec/commonmark-0.31.2.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rivals() {
		plain, gfm, panics := 0, 0, 0
		for _, e := range examples {
			want := htmlnorm.Normalize(e.HTML)
			run := func(f func([]byte) []byte) (ok bool) {
				defer func() {
					if recover() != nil {
						panics++
					}
				}()
				return htmlnorm.Normalize(string(f([]byte(e.Markdown)))) == want
			}
			if run(r.plain) {
				plain++
			}
			if run(r.gfm) {
				gfm++
			}
		}
		t.Logf("%-12s bez rozszerzeń %d/%d, z rozszerzeniami %d/%d, paniki %d",
			r.name, plain, len(examples), gfm, len(examples), panics)
	}
}

// BenchmarkRivals mierzy czas, pamięć i rozmiar wyniku czterech bibliotek na zestawach z BenchmarkCorpus
// (spec bez rozszerzeń; corpus i large z rozszerzeniami). Zestaw large jest tu inny niż w BenchmarkCorpus:
// pliki korpusu bez 003 i 004 sklejone 60 razy, ok. 260 kB.
//
//	go test -run xxx -bench BenchmarkRivals -benchtime 2s -count 5 .
func BenchmarkRivals(b *testing.B) {
	examples, err := spectest.Load("../../testdata/spec/commonmark-0.31.2.txt")
	if err != nil {
		b.Fatal(err)
	}
	var spec, corpus [][]byte
	for _, e := range examples {
		spec = append(spec, []byte(e.Markdown))
	}
	var large []byte
	paths, _ := filepath.Glob("../../testdata/corpus/*.md")
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			b.Fatal(err)
		}
		corpus = append(corpus, data)
		// Bez plików 003 i 004: oba mają niezamknięte ogrodzenie kodu. Plik 003 zamieniłby resztę dokumentu
		// w kod u wszystkich, a plik 004 (ogrodzenie w cytacie) robi to w gomarkdown, który wbrew CommonMark
		// szuka zamknięcia poza cytatem; połowa sklejonego dokumentu wychodziła wtedy jako kod (docs/testing.md).
		if name := filepath.Base(p); !strings.HasPrefix(name, "003-") && !strings.HasPrefix(name, "004-") {
			large = append(append(large, data...), '\n')
		}
	}
	large = bytes.Repeat(large, 60)
	var extra [][]byte
	for _, p := range strings.Fields(os.Getenv("GOMDGO_BENCH_FILES")) {
		data, err := os.ReadFile(p)
		if err != nil {
			b.Fatal(err)
		}
		extra = append(extra, data)
	}

	sets := []struct {
		name string
		docs [][]byte
		gfm  bool
	}{{"spec", spec, false}, {"corpus", corpus, true}, {"large", [][]byte{large}, true}, {"extra", extra, true}}
	for _, set := range sets {
		if len(set.docs) == 0 {
			continue
		}
		size := 0
		for _, d := range set.docs {
			size += len(d)
		}
		for _, r := range rivals() {
			f := r.plain
			if set.gfm {
				f = r.gfm
			}
			b.Run(set.name+"/"+r.name, func(b *testing.B) {
				b.SetBytes(int64(size))
				b.ReportAllocs()
				out := 0
				for i := 0; i < b.N; i++ {
					out = 0
					for _, d := range set.docs {
						out += len(f(d))
					}
				}
				b.ReportMetric(float64(out), "out-bytes")
			})
		}
	}
}

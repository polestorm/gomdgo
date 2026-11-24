package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/polestorm/gomdgo/internal/spectest"
	"github.com/polestorm/gomdgo/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// BenchmarkCorpus mierzy kryterium K7: gomdgo nie wolniej niż 2× parser odniesienia. Oba parsery mają
// te same ustawienia co w porównaniu różnicowym (surowy HTML przepuszczany, bez filtrowania adresów).
//
//	go test -run xxx -bench BenchmarkCorpus -benchtime 2s -count 5 .
//
// Zestawy: spec (652 przykłady CommonMark, każdy jako osobny dokument, bez rozszerzeń), corpus (pliki
// korpusu własnego osobno, z rozszerzeniami GFM), large (te same pliki sklejone 50 razy w jeden dokument;
// bez pliku 003, który kończy się niezamkniętym blokiem kodu i zamieniłby resztę w kod). Dodatkowe pliki
// można podać w zmiennej GOMDGO_BENCH_FILES (ścieżki rozdzielone znakiem nowego wiersza): zestaw extra.
// Miara out-bytes to łączny rozmiar wyniku HTML dla zestawu.
func BenchmarkCorpus(b *testing.B) {
	examples, err := spectest.Load("../../testdata/spec/commonmark-0.31.2.txt")
	if err != nil {
		b.Fatal(err)
	}
	var spec, corpus, extra [][]byte
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
		if !strings.HasPrefix(filepath.Base(p), "003-") {
			large = append(append(large, data...), '\n')
		}
	}
	large = bytes.Repeat(large, 50)
	for _, p := range strings.Fields(os.Getenv("GOMDGO_BENCH_FILES")) {
		data, err := os.ReadFile(p)
		if err != nil {
			b.Fatal(err)
		}
		extra = append(extra, data)
	}

	pass := markdown.Config{RawHTML: markdown.RawHTMLPass, AnyScheme: true}
	plain := pass
	plain.Disable = markdown.GFM
	ours, oursPlain := markdown.New(pass), markdown.New(plain)
	ref := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(html.WithUnsafe()))
	refPlain := goldmark.New(goldmark.WithRendererOptions(html.WithUnsafe()))

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
		conv, md := oursPlain, refPlain
		if set.gfm {
			conv, md = ours, ref
		}
		b.Run(set.name+"/gomdgo", func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ReportAllocs()
			out := 0
			for i := 0; i < b.N; i++ {
				out = 0
				for _, d := range set.docs {
					res, err := conv.Render(d)
					if err != nil {
						b.Fatal(err)
					}
					out += len(res)
				}
			}
			b.ReportMetric(float64(out), "out-bytes")
		})
		b.Run(set.name+"/goldmark", func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ReportAllocs()
			out := 0
			for i := 0; i < b.N; i++ {
				out = 0
				for _, d := range set.docs {
					var buf bytes.Buffer
					if err := md.Convert(d, &buf); err != nil {
						b.Fatal(err)
					}
					out += buf.Len()
				}
			}
			b.ReportMetric(float64(out), "out-bytes")
		})
	}
}

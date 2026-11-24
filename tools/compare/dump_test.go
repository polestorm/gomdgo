package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDumpRivals zapisuje wyniki czterech bibliotek dla plików korpusu i dla dokumentu sklejonego (zmienna
// GOMDGO_DUMP wskazuje katalog); służy do ręcznej analizy różnic.
func TestDumpRivals(t *testing.T) {
	dir := os.Getenv("GOMDGO_DUMP")
	if dir == "" {
		t.Skip("GOMDGO_DUMP not set")
	}
	docs := map[string][]byte{}
	var one []byte
	paths, _ := filepath.Glob("../../testdata/corpus/*.md")
	for _, p := range paths {
		data, _ := os.ReadFile(p)
		docs[strings.TrimSuffix(filepath.Base(p), ".md")] = data
		if !strings.HasPrefix(filepath.Base(p), "003-") {
			one = append(append(one, data...), '\n')
		}
	}
	docs["one"] = one
	docs["large"] = bytes.Repeat(one, 50)
	for name, src := range docs {
		os.WriteFile(filepath.Join(dir, name+".md"), src, 0o644)
		for _, r := range rivals() {
			out := r.gfm(src)
			os.WriteFile(filepath.Join(dir, name+"."+r.name+".html"), out, 0o644)
			t.Logf("%-40s %-12s in %7d out %7d", name, r.name, len(src), len(out))
		}
	}
}

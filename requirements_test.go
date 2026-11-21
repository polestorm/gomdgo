package gomdgo

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var reqRef = regexp.MustCompile(`\b[A-Z]{3,5}-\d{3}\b`)

// TestRequirementsCovered mierzy kryterium K4: każde wymaganie z requirements/ ma test albo plik korpusu,
// a każde odwołanie „REQ:” wskazuje istniejące wymaganie.
func TestRequirementsCovered(t *testing.T) {
	read := func(path string) string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	defined := map[string]bool{}
	docs, _ := filepath.Glob("requirements/*.md")
	for _, doc := range docs {
		for _, line := range strings.Split(read(doc), "\n") {
			if id := reqRef.FindString(line); id != "" && strings.HasPrefix(line, "| "+id) {
				if defined[id] {
					t.Errorf("requirement %s is defined twice", id)
				}
				defined[id] = true
			}
		}
	}
	if len(defined) == 0 {
		t.Fatal("no requirements found")
	}
	corpus, _ := filepath.Glob("testdata/corpus/*.md")
	sources := append([]string{}, corpus...)
	for _, pattern := range []string{"*_test.go", "internal/*/*_test.go"} {
		m, _ := filepath.Glob(pattern)
		sources = append(sources, m...)
	}
	used := map[string]bool{}
	for _, f := range sources {
		for _, line := range strings.Split(read(f), "\n") {
			if !strings.Contains(line, "REQ:") {
				continue
			}
			for _, id := range reqRef.FindAllString(line, -1) {
				used[id] = true
				if !defined[id] {
					t.Errorf("%s refers to unknown requirement %s", f, id)
				}
			}
		}
	}
	for id := range defined {
		if !used[id] {
			t.Errorf("requirement %s has no corpus file or test", id)
		}
	}
	for _, f := range corpus {
		if !strings.Contains(read(f), "REQ:") {
			t.Errorf("%s has no REQ line", f)
		}
	}
}

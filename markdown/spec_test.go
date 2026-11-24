package markdown

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/polestorm/gomdgo/internal/spectest"
)

var update = flag.Bool("update", false, "przepisz listy zaliczonych przykładów (testdata/spec/passing-*.txt)")

const (
	commonMarkSpec = "../testdata/spec/commonmark-0.31.2.txt"
	gfmSpec        = "../testdata/spec/gfm-0.29.txt"
)

// specConverter odpowiada założeniom przykładów specyfikacji CommonMark: bez rozszerzeń, surowy HTML przechodzi
// bez zmian, a schemat adresu nie jest sprawdzany.
var specConverter = New(Config{RawHTML: RawHTMLPass, AnyScheme: true, Disable: GFM})

// gfmSpecConverter: to samo z rozszerzeniami GFM, dla przykładów z rozdziałów rozszerzeń.
var gfmSpecConverter = New(Config{RawHTML: RawHTMLPass, AnyScheme: true})

func renderSpec(markdown string) []byte {
	return renderWith(specConverter, markdown)
}

func renderWith(conv *Converter, markdown string) []byte {
	out, err := conv.Render([]byte(markdown))
	if err != nil {
		panic(err)
	}
	return out
}

func loadSpec(t testing.TB, path string) []spectest.Example {
	t.Helper()
	ex, err := spectest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return ex
}

// TestSpecCommonMark mierzy kryterium K1: wszystkie przykłady CommonMark 0.31.2.
func TestSpecCommonMark(t *testing.T) {
	runSpec(t, specConverter, loadSpec(t, commonMarkSpec), "../testdata/spec/passing-commonmark.txt")
}

// TestSpecGFM mierzy kryterium K2: przykłady z rozdziałów rozszerzeń GFM 0.29-gfm. Część wspólną z CommonMark
// mierzy TestSpecCommonMark na nowszej wersji specyfikacji.
func TestSpecGFM(t *testing.T) {
	var ext []spectest.Example
	for _, e := range loadSpec(t, gfmSpec) {
		if e.IsExtension() {
			ext = append(ext, e)
		}
	}
	runSpec(t, gfmSpecConverter, ext, "../testdata/spec/passing-gfm.txt")
}

// runSpec renderuje przykłady i porównuje wynik co do bajta. Przykład, który jeszcze nie przechodzi, nie psuje
// testu; psuje go przykład zapisany na liście zaliczonych, który przestał przechodzić. Dzięki temu liczba
// zaliczonych nie może spaść niezauważenie: spadek wymaga zmiany pliku listy, widocznej w przeglądzie.
func runSpec(t *testing.T, conv *Converter, examples []spectest.Example, passingFile string) {
	recorded := readPassing(t, passingFile)
	var passed []int
	type count struct{ ok, all int }
	sections := map[string]*count{}
	var order []string
	for _, e := range examples {
		c := sections[e.Section]
		if c == nil {
			c = &count{}
			sections[e.Section] = c
			order = append(order, e.Section)
		}
		c.all++
		got := string(renderWith(conv, e.Markdown))
		if got == e.HTML {
			c.ok++
			passed = append(passed, e.Number)
			continue
		}
		if recorded[e.Number] && !*update {
			t.Errorf("example %d (%s, spec line %d) regressed\nmarkdown: %q\nwant:     %q\ngot:      %q",
				e.Number, e.Section, e.StartLine, e.Markdown, e.HTML, got)
		}
	}
	for _, s := range order {
		t.Logf("%-45s %3d / %3d", s, sections[s].ok, sections[s].all)
	}
	t.Logf("passed %d of %d", len(passed), len(examples))

	if *update {
		writePassing(t, passingFile, passed)
		return
	}
	for _, n := range passed {
		if !recorded[n] {
			t.Errorf("%d examples pass but %d are recorded in %s; run go test -update and review the diff",
				len(passed), len(recorded), passingFile)
			break
		}
	}
}

func readPassing(t *testing.T, path string) map[int]bool {
	t.Helper()
	set := map[int]bool{}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return set
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Fields(string(data)) {
		n, err := strconv.Atoi(line)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		set[n] = true
	}
	return set
}

func writePassing(t *testing.T, path string, passed []int) {
	t.Helper()
	sort.Ints(passed)
	var b strings.Builder
	for _, n := range passed {
		fmt.Fprintln(&b, n)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

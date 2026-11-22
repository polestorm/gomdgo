package gomdgo

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// sectionRequirements przypisuje rozdziały specyfikacji do wymagań; przykłady rozdziału są testami wymagania.
var sectionRequirements = map[string]string{
	"Tabs":                       "BLK-002", // REQ: BLK-002
	"Precedence":                 "BLK-003", // REQ: BLK-003
	"Thematic breaks":            "BLK-004", // REQ: BLK-004
	"ATX headings":               "BLK-005", // REQ: BLK-005
	"Setext headings":            "BLK-006", // REQ: BLK-006
	"Indented code blocks":       "BLK-007", // REQ: BLK-007
	"Fenced code blocks":         "BLK-008", // REQ: BLK-008
	"HTML blocks":                "BLK-009", // REQ: BLK-009
	"Link reference definitions": "BLK-010", // REQ: BLK-010
	"Paragraphs":                 "BLK-011", // REQ: BLK-011
	"Blank lines":                "BLK-012", // REQ: BLK-012
	"Block quotes":               "BLK-013", // REQ: BLK-013
	"List items":                 "BLK-014", // REQ: BLK-014
	"Lists":                      "BLK-015", // REQ: BLK-015

	"Backslash escapes":                       "INL-001", // REQ: INL-001
	"Entity and numeric character references": "INL-002", // REQ: INL-002
	"Inlines":                      "INL-003", // REQ: INL-003
	"Code spans":                   "INL-003",
	"Emphasis and strong emphasis": "INL-004", // REQ: INL-004
	"Links":                        "INL-005", // REQ: INL-005
	"Images":                       "INL-006", // REQ: INL-006
	"Autolinks":                    "INL-007", // REQ: INL-007
	"Raw HTML":                     "INL-008", // REQ: INL-008
	"Hard line breaks":             "INL-009", // REQ: INL-009
	"Soft line breaks":             "INL-010", // REQ: INL-010
	"Textual content":              "INL-011", // REQ: INL-011
}

// TestSpecSectionsHaveRequirements pilnuje, żeby rozdziały specyfikacji z tabeli istniały pod tą nazwą.
func TestSpecSectionsHaveRequirements(t *testing.T) {
	sections := map[string]bool{}
	for _, e := range loadSpec(t, commonMarkSpec) {
		sections[e.Section] = true
	}
	for s := range sectionRequirements {
		if !sections[s] {
			t.Errorf("section %q is not in the specification", s)
		}
	}
	for s := range sections {
		if sectionRequirements[s] == "" {
			t.Errorf("section %q has no requirement", s)
		}
	}
}

// REQ: BLK-001
func TestLineEndings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a\nb", "<p>a\nb</p>\n"},
		{"a\r\nb\r\n", "<p>a\nb</p>\n"},
		{"a\rb\r", "<p>a\nb</p>\n"},
		{"a\r\n\r\nb", "<p>a</p>\n<p>b</p>\n"},
		{"a\r\rb", "<p>a</p>\n<p>b</p>\n"},
		{"# a\r- b\r\n- c", "<h1>a</h1>\n<ul>\n<li>b</li>\n<li>c</li>\n</ul>\n"},
		{"a\x00b", "<p>a�b</p>\n"},
		{"    \x00\n", "<pre><code>�\n</code></pre>\n"},
		{"", ""},
		{"\n\n", ""},
	}
	for _, c := range cases {
		if got := string(Render([]byte(c.in))); got != c.want {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
}

// REQ: DET-004
func TestInvalidUTF8(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a\xffb", "<p>a\uFFFDb</p>\n"},
		{"a\xff\xfe\xfdb", "<p>a\uFFFDb</p>\n"}, // seria niepoprawnych bajtów to jeden znak
		{"za\xc5", "<p>za\uFFFD</p>\n"},         // urwana sekwencja na końcu
		// bajty nie mogą złożyć się w znak ponad znacznikiem ani w opisie obrazu
		{"\xc5*\x82*", "<p>\uFFFD<em>\uFFFD</em></p>\n"},
		{"![\xc5*\x82*](u)", "<p><img src=\"u\" alt=\"\uFFFD\uFFFD\" /></p>\n"},
		{"[a](/\xff)", "<p><a href=\"/%EF%BF%BD\">a</a></p>\n"},
		{"[a](/\x00)", "<p><a href=\"/%EF%BF%BD\">a</a></p>\n"},
		{"`\xff`", "<p><code>\uFFFD</code></p>\n"},
		{"```\xff\n\xfe\n```", "<pre><code class=\"language-\uFFFD\">\uFFFD\n</code></pre>\n"},
		{"\xed\xa0\x80", "<p>\uFFFD</p>\n"}, // surogat zakodowany jak w CESU-8
		{"zażółć", "<p>zażółć</p>\n"},
	}
	for _, c := range cases {
		for _, conv := range []*Converter{defaultConverter, specConverter} {
			if got, _ := conv.Render([]byte(c.in)); string(got) != c.want {
				t.Errorf("%q (%+v): got %q, want %q", c.in, conv.cfg, got, c.want)
			}
		}
	}
	// Pozycje odnoszą się do źródła po zamianie: wiersze bez zmian, kolumny za zamienionym bajtem przesunięte.
	doc := parse(sanitize([]byte("\xff *a*\n\n# b\xff")), 0)
	want := "document 1:1\n  paragraph 1:1\n    text 1:1 \"\uFFFD \"\n    emph 1:5\n      text 1:6 \"a\"\n" +
		"  heading 3:1\n    text 3:3 \"b\uFFFD\"\n"
	if got := dumpTree(doc, true); got != want {
		t.Errorf("tree:\n%s\nwant:\n%s", got, want)
	}
	valid := []byte("zażółć gęślą jaźń\n")
	if got := sanitize(valid); &got[0] != &valid[0] {
		t.Error("valid source must not be copied")
	}
}

// REQ: BLK-010
func TestReferenceDefinitions(t *testing.T) {
	type ref struct{ label, dest, title string }
	cases := []struct {
		in   string
		want []ref
		rest string // HTML tego, co nie jest definicją
	}{
		{"[foo]: /url \"title\"\n", []ref{{"foo", "/url", "title"}}, ""},
		{"   [foo]: \n      /url  \n           'the title'  \n", []ref{{"foo", "/url", "the title"}}, ""},
		{"[Foo bar]:\n<my url>\n'title'\n", []ref{{"foo bar", "my url", "title"}}, ""},
		{"[foo]: /url '\ntitle\nline1\n'\n", []ref{{"foo", "/url", "\ntitle\nline1\n"}}, ""},
		{"[foo]:\n/url\n", []ref{{"foo", "/url", ""}}, ""},
		{"[foo]: <>\n", []ref{{"foo", "", ""}}, ""},
		{"[foo]:\n", nil, "<p>[foo]:</p>\n"},
		{"[foo]: /url \"title\" ok\n", nil, "<p>[foo]: /url &quot;title&quot; ok</p>\n"},
		{"[foo]: /url\n\"title\" ok\n", []ref{{"foo", "/url", ""}}, "<p>&quot;title&quot; ok</p>\n"},
		{"[foo]: first\n[foo]: second\n", []ref{{"foo", "first", ""}}, ""},
		{"[ΑΓΩ]: /φου\n[Zażółć  GĘŚLĄ\n jaźń]: /pl\n", []ref{{"αγω", "/φου", ""}, {"zażółć gęślą jaźń", "/pl", ""}}, ""},
		{"[ẞ]: /url\n", []ref{{"ss", "/url", ""}}, ""},
		{"Foo\n[bar]: /baz\n", nil, "<p>Foo\n[bar]: /baz</p>\n"},
		{"[foo]: /url\nbar\n===\n", []ref{{"foo", "/url", ""}}, "<h1>bar</h1>\n"},
		{"[foo]: /url\n===\n", []ref{{"foo", "/url", ""}}, "<p>===</p>\n"},
		{"> [foo]: /url\n", []ref{{"foo", "/url", ""}}, "<blockquote>\n</blockquote>\n"},
		{"[a]: (b)c\n", []ref{{"a", "(b)c", ""}}, ""},
		{"[a]: b(c\n", nil, "<p>[a]: b(c</p>\n"},
		{"[]: /url\n", nil, "<p>[]: /url</p>\n"},
		{"[a[b]: /url\n", nil, "<p>[a[b]: /url</p>\n"},
		{"[a\\]b]: /url (t (x)\n", nil, "<p>[a]b]: /url (t (x)</p>\n"},
		{"[" + strings.Repeat("a", 999) + "]: /url\n", []ref{{strings.Repeat("a", 999), "/url", ""}}, ""},
	}
	for _, c := range cases {
		doc := parse([]byte(c.in), 0)
		if len(doc.refs) != len(c.want) {
			t.Errorf("%q: got %d definitions, want %d", c.in, len(doc.refs), len(c.want))
		}
		for _, w := range c.want {
			got := doc.refs[w.label]
			if got == nil {
				t.Errorf("%q: no definition for %q", c.in, w.label)
			} else if got.dest != w.dest || got.title != w.title {
				t.Errorf("%q: got destination %q title %q, want %q %q", c.in, got.dest, got.title, w.dest, w.title)
			}
		}
		if got := string(Render([]byte(c.in))); got != c.rest {
			t.Errorf("%q: got %q, want %q", c.in, got, c.rest)
		}
	}
	if doc := parse([]byte("["+strings.Repeat("a", 1000)+"]: /url\n"), 0); len(doc.refs) != 0 {
		t.Error("a label of 1000 characters must not be accepted")
	}
}

// dumpTree zapisuje drzewo jako wiersze „rodzaj wiersz:kolumna”, z wcięciem według głębokości; inlines
// włącza węzły treści w wierszu.
func dumpTree(doc *document, inlines bool) string {
	var b strings.Builder
	var walk func(n *node, depth int)
	walk = func(n *node, depth int) {
		if n.kind >= kindText && n.kind < kindTable && !inlines {
			return
		}
		line, col := doc.position(n.pos)
		fmt.Fprintf(&b, "%s%s %d:%d", strings.Repeat("  ", depth), n.kind, line, col)
		if n.kind == kindText || n.kind == kindCode || n.kind == kindHTMLInline {
			fmt.Fprintf(&b, " %q", n.literal)
		}
		if n.kind == kindLink || n.kind == kindImage {
			fmt.Fprintf(&b, " %q %q", n.info, n.title)
		}
		b.WriteByte('\n')
		for c := n.first; c != nil; c = c.next {
			walk(c, depth+1)
		}
	}
	walk(doc.root, 0)
	return b.String()
}

// REQ: BLK-016
func TestBlockPositions(t *testing.T) {
	src := "# Tytuł\r\n" +
		"\n" +
		"  Zażółć\n" +
		"gęślą\n" +
		"> - jaźń\n" +
		">   ```go\n" +
		">   x\n" +
		">   ```\n" +
		"\n" +
		" 1. a\n" +
		"\n" +
		"    b\n" +
		"***\n" +
		"\tkod\n" +
		"[x]: /y\n" +
		"Setext\n" +
		"---\n" +
		"<div>\n"
	want := `document 1:1
  heading 1:1
  paragraph 3:3
  block_quote 5:1
    list 5:3
      item 5:3
        paragraph 5:5
        code_block 6:5
  list 10:2
    item 10:2
      paragraph 10:5
      paragraph 12:5
  thematic_break 13:1
  code_block 14:2
  heading 16:1
  html_block 18:1
`
	doc := parse([]byte(src), 0)
	if got := dumpTree(doc, false); got != want {
		t.Errorf("tree:\n%s\nwant:\n%s", got, want)
	}
	if line, col := doc.position(len(src)); line != 18 || col != 7 {
		t.Errorf("end of source: got %d:%d, want 18:7", line, col)
	}
}

// REQ: BLK-002
func TestCursorTabs(t *testing.T) {
	c := &cursor{s: []byte(" \t\tx")}
	if cols, next := c.indent(); cols != 8 || next != 3 {
		t.Fatalf("indent: got %d columns up to byte %d, want 8 and 3", cols, next)
	}
	c.skipCols(2) // spacja i jedna z trzech kolumn tabulatora
	if c.i != 2 || c.pad != 2 {
		t.Fatalf("after 2 columns: byte %d, pad %d; want 2 and 2", c.i, c.pad)
	}
	if got := string(c.appendRest(nil)); got != "  \tx" {
		t.Fatalf("rest: got %q", got)
	}
	c.skipCols(4) // dwie spacje wirtualne i dwie kolumny następnego tabulatora
	if got := string(c.appendRest(nil)); got != "  x" {
		t.Fatalf("rest: got %q", got)
	}
	c.skipCols(10) // nie zużywa treści
	if got := string(c.appendRest(nil)); got != "x" || c.col != 8 {
		t.Fatalf("rest: got %q at column %d", got, c.col)
	}
}

// corpusConverter: korpus własny jest renderowany przy ustawieniach domyślnych (z rozszerzeniami GFM).
var corpusConverter = defaultConverter

// extrasConverter: z blokami uwag i przypisami.
var extrasConverter = New(Config{Enable: Alerts | Footnotes, AlertTitles: map[string]string{"note": "Uwaga", "warning": "Ostrzeżenie"}})

// TestCorpus porównuje wynik dla plików korpusu własnego z plikami wzorcowymi; go test -update je przepisuje.
func TestCorpus(t *testing.T) {
	files, _ := filepath.Glob("testdata/corpus/*.md")
	extras, _ := filepath.Glob("testdata/corpus-extras/*.md")
	if len(files) == 0 || len(extras) == 0 {
		t.Fatal("no corpus files")
	}
	for _, path := range append(files, extras...) {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		conv := corpusConverter
		if strings.Contains(path, "corpus-extras") {
			conv = extrasConverter // rozszerzenia spoza GFM; ten katalog nie wchodzi do porównania z odniesieniem
		}
		got, _ := conv.Render(src)
		golden := strings.TrimSuffix(path, ".md") + ".html"
		if *update {
			if err := os.WriteFile(golden, got, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("%s: output differs from %s\ngot:\n%s", path, golden, got)
		}
	}
}

// pathologicalConverter: wszystkie rozszerzenia i odwołania w wierszu, żeby pomiar obejmował wszystkie
// ścieżki parsera.
var pathologicalConverter = New(Config{Enable: Alerts | Footnotes,
	References: func(Reference) (ReferenceLink, error) { return ReferenceLink{}, nil }})

// REQ: DET-003
func TestPathological(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test")
	}
	cases := []struct {
		name string
		gen  func(n int) string
	}{
		{"nested quotes on one line", func(n int) string { return strings.Repeat("> ", n) + "a\n" }},
		{"nested items on one line", func(n int) string { return strings.Repeat("* ", n) + "a\n" }},
		{"nested ordered items on one line", func(n int) string { return strings.Repeat("1. ", n) + "a\n" }},
		{"lazy lines under deep quotes", func(n int) string {
			return strings.Repeat("> ", n/2) + "a\n" + strings.Repeat("b\n", n/2)
		}},
		{"lazy lines under deep items", func(n int) string {
			return strings.Repeat("- ", n/2) + "a\n" + strings.Repeat("b\n", n/2)
		}},
		{"many reference definitions", func(n int) string { return strings.Repeat("[a]: /b\n", n) }},
		{"unclosed labels", func(n int) string { return strings.Repeat("[a\n", n) }},
		{"long unclosed label", func(n int) string { return "[" + strings.Repeat("a ", n) }},
		{"underlines after definitions", func(n int) string { return "[a]: /b\n" + strings.Repeat("===\n", n) }},
		{"unclosed fence", func(n int) string { return "```\n" + strings.Repeat("a\n", n) }},
		{"unclosed comment", func(n int) string { return "<!--\n" + strings.Repeat("a\n", n) }},
		{"loose list", func(n int) string { return strings.Repeat("- a\n\n", n) }},
		{"almost thematic breaks", func(n int) string { return strings.Repeat("- ", n) + "x\n" }},
		{"open tag with many attributes", func(n int) string { return "<a" + strings.Repeat(" b=c", n) + "\n" }},
		{"tabs", func(n int) string { return strings.Repeat("\t", n) + "a\n" }},
	}
	cases = append(cases, pathologicalInlines...)
	cases = append(cases, pathologicalGFM...)
	// Rozmiary różnią się 16 razy: wzrost liniowy to czynnik 16, kwadratowy 256. Próg 96 zostawia zapas
	// na odśmiecanie pamięci, które przy małym rozmiarze jeszcze nie pracuje (pomiar w docs/testing.md).
	const small, factor, slack = 10_000, 16, 6
	// Odśmiecanie jest na czas pomiaru wyłączone i wykonywane między pomiarami: inaczej jego koszt, zależny
	// od tego, co zostawiły wcześniejsze przypadki, przykrywa mierzony algorytm.
	// Mierzymy czas procesora, nie zegar: na obciążonej maszynie proces bywa wywłaszczany w trakcie pomiaru
	// i czas zegarowy rośnie kilkukrotnie bez związku z algorytmem.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	measure := func(src []byte, runs int) time.Duration {
		best := time.Duration(1 << 62)
		for i := 0; i < runs; i++ {
			runtime.GC()
			start := cpuTime()
			pathologicalConverter.Render(src)
			best = min(best, cpuTime()-start)
		}
		return best
	}
	for _, c := range cases {
		var base, big time.Duration
		// Jedna powtórka na wypadek pomiaru zaburzonego mimo wszystko (np. przez pracę środowiska Go w tle).
		for attempt := 0; attempt < 2; attempt++ {
			base = max(measure([]byte(c.gen(small)), 3), 100*time.Microsecond)
			big = measure([]byte(c.gen(small*factor)), 3)
			if big <= base*factor*slack {
				break
			}
		}
		if big > base*factor*slack {
			t.Errorf("%s: %v for size %d, %v for size %d", c.name, base, small, big, small*factor)
		}
		if big > 5*time.Second {
			t.Errorf("%s: %v for size %d is too slow", c.name, big, small*factor)
		}
	}
}

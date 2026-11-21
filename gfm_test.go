package gomdgo

import (
	"strings"
	"testing"
)

// REQ: GFM-001 GFM-002 GFM-003 GFM-004 GFM-005 (przykłady rozdziałów rozszerzeń: TestSpecGFM)
var gfmSections = []string{
	"Tables (extension)", "Task list items (extension)", "Strikethrough (extension)", "Autolinks (extension)",
	"Disallowed Raw HTML (extension)",
}

func TestGFMSectionsPresent(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range loadSpec(t, gfmSpec) {
		if e.IsExtension() {
			seen[e.Section] = true
		}
	}
	for _, s := range gfmSections {
		if !seen[s] {
			t.Errorf("section %q is not in the specification", s)
		}
	}
	if len(seen) != len(gfmSections) {
		t.Errorf("extension sections in the specification: %v", seen)
	}
}

func checkCases(t *testing.T, conv *Converter, cases [][2]string) {
	t.Helper()
	for _, c := range cases {
		if got, _ := conv.Render([]byte(c[0])); string(got) != c[1] {
			t.Errorf("%q:\ngot  %q\nwant %q", c[0], got, c[1])
		}
	}
}

func table(head string, rows ...string) string {
	out := "<table>\n<thead>\n<tr>\n" + head + "</tr>\n</thead>\n"
	if len(rows) > 0 {
		out += "<tbody>\n"
		for _, r := range rows {
			out += "<tr>\n" + r + "</tr>\n"
		}
		out += "</tbody>\n"
	}
	return out + "</table>\n"
}

// REQ: GFM-001
func TestTablesBeyondSpec(t *testing.T) {
	checkCases(t, corpusConverter, [][2]string{
		{"Wstęp\n| Miasto | Liczba |\n|:--|--:|\n| Łódź | 1 |", "<p>Wstęp</p>\n" +
			table("<th align=\"left\">Miasto</th>\n<th align=\"right\">Liczba</th>\n", "<td align=\"left\">Łódź</td>\n<td align=\"right\">1</td>\n")},
		{"a|b\n-|-\n*c*|`d|e`", table("<th>a</th>\n<th>b</th>\n", "<td><em>c</em></td>\n<td>`d</td>\n")},
		{"|a|\n|-|\n|`x\\|y` \\| z|", table("<th>a</th>\n", "<td><code>x|y</code> | z</td>\n")},
		{"|a|b|\n|-|-|\n||\n|\n|c", table("<th>a</th>\n<th>b</th>\n", "<td></td>\n<td></td>\n", "<td></td>\n<td></td>\n", "<td>c</td>\n<td></td>\n")},
		{"a\n---", "<h2>a</h2>\n"}, // bez znaku | to nagłówek setext
		{"a | b\n---", "<h2>a | b</h2>\n"},
		{"|a|\n---", table("<th>a</th>\n")},
		{"a | b\n- | -", table("<th>a</th>\n<th>b</th>\n")},
		{"a | b\n-- | -- | --", "<p>a | b\n-- | -- | --</p>\n"},
		{"|a|\n|-|\n# b", table("<th>a</th>\n") + "<h1>b</h1>\n"},
		{"|a|\n|-|\n- b", table("<th>a</th>\n") + "<ul>\n<li>b</li>\n</ul>\n"},
		{"|a|\n|-|\n    b", table("<th>a</th>\n", "<td>b</td>\n")},
		{"> |a|\n> |-|\nb", "<blockquote>\n" + table("<th>a</th>\n") + "</blockquote>\n<p>b</p>\n"},
		{"- |a|\n  |-|\n  |b|\n- c", "<ul>\n<li>\n" + table("<th>a</th>\n", "<td>b</td>\n") + "</li>\n<li>c</li>\n</ul>\n"},
		{"|a|\n|-|\n\n|b|", table("<th>a</th>\n") + "<p>|b|</p>\n"},
		{"[r]: /u\n|[r]|\n|-|", table("<th><a href=\"/u\">r</a></th>\n")},
		{"| a |\n| :-: |\n| ~~b~~ www.x.pl |", table("<th align=\"center\">a</th>\n",
			"<td align=\"center\"><del>b</del> <a href=\"http://www.x.pl\">www.x.pl</a></td>\n")},
	})
}

// REQ: GFM-002
func TestTaskListsBeyondSpec(t *testing.T) {
	box, checked := `<input disabled="" type="checkbox">`, `<input checked="" disabled="" type="checkbox">`
	checkCases(t, corpusConverter, [][2]string{
		{"- [ ] zrób\n- [X] zrobione\n- [x]bez spacji\n- [ ]", "<ul>\n<li>" + box + " zrób</li>\n<li>" + checked +
			" zrobione</li>\n<li>[x]bez spacji</li>\n<li>[ ]</li>\n</ul>\n"},
		{"1. [x] a\n\n2. [ ] b", "<ol>\n<li>\n<p>" + checked + " a</p>\n</li>\n<li>\n<p>" + box + " b</p>\n</li>\n</ol>\n"},
		{"- [ ]\n  dalej", "<ul>\n<li>" + box + "\ndalej</li>\n</ul>\n"},
		{"- [x] *a*\n\n  [ ] b", "<ul>\n<li>\n<p>" + checked + " <em>a</em></p>\n<p>[ ] b</p>\n</li>\n</ul>\n"},
		{"[ ] poza listą", "<p>[ ] poza listą</p>\n"},
		{"- # [ ] nagłówek", "<ul>\n<li>\n<h1>[ ] nagłówek</h1>\n</li>\n</ul>\n"},
		{"- [x]: /u\n\n[x]", "<ul>\n<li></li>\n</ul>\n<p><a href=\"/u\">x</a></p>\n"},
	})
}

// REQ: GFM-003
func TestStrikethroughBeyondSpec(t *testing.T) {
	checkCases(t, corpusConverter, [][2]string{
		{"~~zażółć~~ gęślą", "<p><del>zażółć</del> gęślą</p>\n"},
		{"~jedna~ ~~~trzy~~~ ~~ spacje ~~", "<p>~jedna~ ~~~trzy~~~ ~~ spacje ~~</p>\n"},
		{"**~~a~~** ~~*b*~~ ~~a **b~~ c**", "<p><strong><del>a</del></strong> <del><em>b</em></del> <del>a **b</del> c**</p>\n"},
		{"~~a~~b~~c~~", "<p><del>a</del>b<del>c</del></p>\n"},
		{"~~[a](/b)~~ `~~c~~`", "<p><del><a href=\"/b\">a</a></del> <code>~~c~~</code></p>\n"},
		{"~~~\nkod\n~~~", "<pre><code>kod\n</code></pre>\n"},
	})
}

// REQ: GFM-007
func TestCompatibilityOptions(t *testing.T) {
	compat := New(Config{SingleTilde: true, TableAlignStyle: true})
	// Oczekiwania pochodzą z parsera odniesienia uruchomionego jako czarna skrzynka (goldmark v1.8.7).
	checkCases(t, compat, [][2]string{
		{"~a~", "<p><del>a</del></p>\n"},
		{"~~a~~", "<p><del>a</del></p>\n"},
		{"~a~~", "<p><del>a</del>~</p>\n"},
		{"~~a~", "<p>~<del>a</del></p>\n"},
		{"x ~~~a~~~", "<p>x ~~~a~~~</p>\n"},
		{"~a ~~b~~ c~", "<p><del>a <del>b</del> c</del></p>\n"},
		{"~~a ~b~ c~~", "<p><del>a <del>b</del> c</del></p>\n"},
		{"a~b~c", "<p>a<del>b</del>c</p>\n"},
		{"~ a ~", "<p>~ a ~</p>\n"},
		{"~a~b~", "<p><del>a</del>b~</p>\n"},
		{"~a~ ~~b~~ ~~~c~~~", "<p><del>a</del> <del>b</del> ~~~c~~~</p>\n"},
		{"|a|b|c|\n|:-|:-:|-:|\n|d|e|f|", "<table>\n<thead>\n<tr>\n<th style=\"text-align:left\">a</th>\n" +
			"<th style=\"text-align:center\">b</th>\n<th style=\"text-align:right\">c</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n" +
			"<td style=\"text-align:left\">d</td>\n<td style=\"text-align:center\">e</td>\n<td style=\"text-align:right\">f</td>\n</tr>\n</tbody>\n</table>\n"},
	})
	// Opcja jednej tyldy nie włącza przekreślenia sama z siebie.
	if got, _ := New(Config{SingleTilde: true, Disable: Strikethrough}).Render([]byte("~a~ ~~b~~")); string(got) != "<p>~a~ ~~b~~</p>\n" {
		t.Errorf("SingleTilde without Strikethrough: %q", got)
	}
}

// REQ: GFM-008
func TestPaddedCellLimit(t *testing.T) {
	const cols, rows = 1000, 150 // 150 wierszy po 999 brakujących komórek: limit kończy się w 101. wierszu
	src := strings.Repeat("|a", cols) + "\n" + strings.Repeat("|-", cols) + "\n" + strings.Repeat("|b\n", rows)
	out, _ := corpusConverter.Render([]byte(src))
	if got := strings.Count(string(out), "<td></td>"); got != maxPaddedCells {
		t.Errorf("padded cells: got %d, want %d", got, maxPaddedCells)
	}
	if got := strings.Count(string(out), "<td>b</td>"); got != rows {
		t.Errorf("rows: got %d, want %d", got, rows)
	}
	if !strings.HasSuffix(string(out), "<tr>\n<td>b</td>\n</tr>\n</tbody>\n</table>\n") {
		t.Errorf("rows after the limit must keep their own cells only: %q", out[len(out)-80:])
	}
	if err := checkSafe(out); err != nil {
		t.Error(err)
	}
	// Limit dotyczy dokumentu, nie tabeli: druga tabela nie dostaje nowej puli.
	out, _ = corpusConverter.Render([]byte(src + "\n|a|b|\n|-|-|\n|c|\n"))
	if got := strings.Count(string(out), "<td></td>"); got != maxPaddedCells {
		t.Errorf("padded cells in two tables: got %d, want %d", got, maxPaddedCells)
	}
}

// REQ: GFM-004
func TestAutolinksBeyondSpec(t *testing.T) {
	a := func(href, text string) string { return `<a href="` + href + `">` + text + `</a>` }
	checkCases(t, corpusConverter, [][2]string{
		{"Zobacz https://example.com/ścieżka?q=1, potem www.example.com.", "<p>Zobacz " +
			a("https://example.com/%C5%9Bcie%C5%BCka?q=1", "https://example.com/ścieżka?q=1") + ", potem " +
			a("http://www.example.com", "www.example.com") + ".</p>\n"},
		{"*www.example.com* (https://example.com/a_b_c) ~~www.x.pl~~", "<p><em>" + a("http://www.example.com", "www.example.com") +
			"</em> (" + a("https://example.com/a_b_c", "https://example.com/a_b_c") + ") <del>" + a("http://www.x.pl", "www.x.pl") + "</del></p>\n"},
		{"xwww.example.com ahttp://example.com http://localhost www. www.a_b.pl www.a_b.c.pl", "<p>xwww.example.com ahttp://example.com http://localhost www. www.a_b.pl " +
			a("http://www.a_b.c.pl", "www.a_b.c.pl") + "</p>\n"},
		{"[www.example.com](/x) [a www.example.com b](/y) ![www.example.com](/z)", "<p><a href=\"/x\">www.example.com</a> <a href=\"/y\">a www.example.com b</a> <img src=\"/z\" alt=\"www.example.com\" /></p>\n"},
		{"`www.example.com` <www.example.com> <https://example.com>", "<p><code>www.example.com</code> &lt;www.example.com&gt; " + a("https://example.com", "https://example.com") + "</p>\n"},
		{"jan.kowalski+test@example.com, jan@example i jan@example.com.", "<p>" + a("mailto:jan.kowalski+test@example.com", "jan.kowalski+test@example.com") +
			", jan@example i " + a("mailto:jan@example.com", "jan@example.com") + ".</p>\n"},
		{"plik_z_danymi@example.com i a_b_@example.com", "<p>" + a("mailto:plik_z_danymi@example.com", "plik_z_danymi@example.com") + " i " +
			a("mailto:a_b_@example.com", "a_b_@example.com") + "</p>\n"},
		{"ftp://example.com/plik i https://example.com/a)b) (https://example.com/a(b))", "<p>ftp://example.com/plik i " +
			a("https://example.com/a)b", "https://example.com/a)b") + ") (" + a("https://example.com/a(b)", "https://example.com/a(b)") + ")</p>\n"},
		{"https://example.com/a&amp; https://example.com/a&amp;b https://example.com/?a=1&b=2;", "<p>" + a("https://example.com/a&amp;", "https://example.com/a&amp;") + " " +
			a("https://example.com/a&amp;b", "https://example.com/a&amp;b") + " " + a("https://example.com/?a=1&amp;b=2;", "https://example.com/?a=1&amp;b=2;") + "</p>\n"},
		{"wiersz\nwww.example.com\nhttps://example.com", "<p>wiersz\n" + a("http://www.example.com", "www.example.com") + "\n" +
			a("https://example.com", "https://example.com") + "</p>\n"},
	})
}

// REQ: GFM-005
func TestTagFilterBeyondSpec(t *testing.T) {
	conv := New(Config{RawHTML: RawHTMLPass})
	checkCases(t, conv, [][2]string{
		{"<script>alert(1)</script>\n", "&lt;script>alert(1)&lt;/script>\n"},
		{"a <SCRIPT src=x> <Style> </iframe > <b> <titles> <xmp/> <plaintext", "<p>a &lt;SCRIPT src=x> &lt;Style> &lt;/iframe > <b> <titles> &lt;xmp/> &lt;plaintext</p>\n"},
		{"<div>\n<textarea>x</textarea><noembed><noframes>\n</div>", "<div>\n&lt;textarea>x&lt;/textarea>&lt;noembed>&lt;noframes>\n</div>\n"},
	})
	checkCases(t, New(Config{RawHTML: RawHTMLPass, Disable: TagFilter}), [][2]string{
		{"<script>alert(1)</script>\n", "<script>alert(1)</script>\n"},
	})
}

// REQ: GFM-006
func TestExtensionsAreSeparate(t *testing.T) {
	src := "| a |\n| - |\n\n- [ ] b ~~c~~ www.example.com\n"
	plain := "<p>| a |\n| - |</p>\n<ul>\n<li>[ ] b ~~c~~ www.example.com</li>\n</ul>\n"
	full := table("<th>a</th>\n") + "<ul>\n<li><input disabled=\"\" type=\"checkbox\"> b <del>c</del> " +
		"<a href=\"http://www.example.com\">www.example.com</a></li>\n</ul>\n"
	// Domyślnie włączone są wszystkie rozszerzenia GFM; Disable: GFM daje czysty CommonMark.
	if got := string(Render([]byte(src))); got != full {
		t.Errorf("defaults: got %q, want %q", got, full)
	}
	if got, _ := New(Config{Disable: GFM}).Render([]byte(src)); string(got) != plain {
		t.Errorf("all extensions disabled: got %q, want %q", got, plain)
	}
	// Każde rozszerzenie włączone osobno zmienia tylko swoją składnię, a wyłączone osobno tylko ją zostawia.
	changes := map[Extension][2]string{
		Tables:        {"<p>| a |\n| - |</p>\n", table("<th>a</th>\n")},
		TaskLists:     {"[ ] b", `<input disabled="" type="checkbox"> b`},
		Strikethrough: {"~~c~~", "<del>c</del>"},
		Autolinks:     {"www.example.com", `<a href="http://www.example.com">www.example.com</a>`},
		TagFilter:     {"", ""},
	}
	for ext, c := range changes {
		only := strings.Replace(plain, c[0], c[1], 1)
		if got, _ := New(Config{Disable: GFM &^ ext}).Render([]byte(src)); string(got) != only {
			t.Errorf("only extension %d: got %q, want %q", ext, got, only)
		}
		without := strings.Replace(full, c[1], c[0], 1)
		if got, _ := New(Config{Disable: ext}).Render([]byte(src)); string(got) != without {
			t.Errorf("without extension %d: got %q, want %q", ext, got, without)
		}
	}
	// Z rozszerzeniami wynik przykładów CommonMark różni się tylko tam, gdzie przykład zawiera składnię
	// rozszerzenia; z Disable: GFM musi być zgodny ze specyfikacją (TestSpecCommonMark).
	changed := 0
	for _, e := range loadSpec(t, commonMarkSpec) {
		if string(renderWith(gfmSpecConverter, e.Markdown)) != e.HTML {
			changed++
		}
	}
	if changed > 20 {
		t.Errorf("%d CommonMark examples change with GFM extensions; expected only a few", changed)
	}
	t.Logf("CommonMark examples whose output changes with GFM extensions: %d", changed)
}

// REQ: INL-012
func TestGFMPositions(t *testing.T) {
	src := "| a | `b\\|c` *d* |\n|---|:-:|\n| ~~e~~ | www.x.pl |\n\n- [x] f\n"
	want := `document 1:1
  table 1:1
    table_row 1:1
      table_cell 1:3
        text 1:3 "a"
      table_cell 1:7
        code 1:7 "b|c"
        text 1:13 " "
        emph 1:14
          text 1:15 "d"
    table_row 3:1
      table_cell 3:3
        strike 3:3
          text 3:5 "e"
      table_cell 3:11
        link 3:11 "http://www.x.pl" ""
          text 3:11 "www.x.pl"
  list 5:1
    item 5:1
      paragraph 5:3
        task_checkbox 5:3
        text 5:6 " f"
`
	if got := dumpTree(parse([]byte(src), GFM), true); got != want {
		t.Errorf("tree:\n%s\nwant:\n%s", got, want)
	}
}

// pathologicalGFM to przypadki złośliwe dla rozszerzeń GFM.
// REQ: DET-003
var pathologicalGFM = []struct {
	name string
	gen  func(n int) string
}{
	{"table with many rows", func(n int) string { return "|a|b|\n|-|-|\n" + strings.Repeat("|c|d|\n", n) }},
	{"table row with many cells", func(n int) string { return "|a|b|\n|-|-|\n" + strings.Repeat("|c", n) }},
	{"table row with escaped pipes", func(n int) string { return "|a|\n|-|\n|" + strings.Repeat("*x*\\|", n) }},
	{"delimiter rows under a paragraph", func(n int) string { return strings.Repeat("|a|b|\n|-|\n", n) }},
	{"almost delimiter rows", func(n int) string { return "a\n" + strings.Repeat("|-", n) + "x\n" }},
	{"long delimiter row", func(n int) string { return strings.Repeat("|a", n) + "\n" + strings.Repeat("|-", n) + "\n" }},
	{"tilde pairs", func(n int) string { return strings.Repeat("~~a ", n) }},
	{"tilde runs", func(n int) string { return strings.Repeat("~", n) + "a" + strings.Repeat("~~~ ", n) }},
	{"strikethrough and emphasis", func(n int) string { return strings.Repeat("~~a *b ", n/2) + strings.Repeat("c~~ d* ", n/2) }},
	{"task items", func(n int) string { return strings.Repeat("- [ ] a\n", n) }},
	{"www without domain", func(n int) string { return strings.Repeat("(www._", n) }},
	{"www with long tail", func(n int) string { return strings.Repeat("(www.a.b", n) }},
	{"url with open parens", func(n int) string { return "www.a.b/" + strings.Repeat("(", n) + strings.Repeat(")", n) + ")" }},
	{"url with close parens", func(n int) string { return "www.a.b/" + strings.Repeat(")", n) }},
	{"url with entities", func(n int) string { return "www.a.b/" + strings.Repeat("&a;", n) }},
	{"url with trailing punctuation", func(n int) string { return "www.a.b/" + strings.Repeat("?!.,", n) }},
	{"scheme separators", func(n int) string { return strings.Repeat("http://", n) + strings.Repeat("://", n) }},
	{"at signs", func(n int) string { return strings.Repeat("@", n) + strings.Repeat("a@", n) + strings.Repeat("a@b", n) }},
	{"long local part", func(n int) string { return strings.Repeat("a", n) + "@" + strings.Repeat("b", n) + "@c" }},
	{"many text nodes to merge", func(n int) string { return strings.Repeat("a\\*b_c&amp;", n) + "@d.e" }},
	{"links around autolinks", func(n int) string { return strings.Repeat("[www.a.b](c) ", n) }},
	{"filtered tags", func(n int) string { return strings.Repeat("<script><title", n) }},
	{"wide table with short rows", func(n int) string {
		return strings.Repeat("|a", n/4) + "\n" + strings.Repeat("|-", n/4) + "\n" + strings.Repeat("|b\n", n/6)
	}},
}

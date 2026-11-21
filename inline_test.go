package gomdgo

import (
	"strings"
	"testing"
)

// REQ: INL-012
func TestInlinePositions(t *testing.T) {
	src := "Zażółć *gęślą* `jaźń`\n" +
		"  i **mocno _bardzo_** [tu](/a \"t\")  \n" +
		"koniec\\\n" +
		"<b> &amp; \\* ![o *p*][r] <https://x.pl>\n" +
		"\n" +
		"> - # Tytuł [r]\n" +
		"\n" +
		"[r]: /ref 'Tytuł'\n"
	want := `document 1:1
  paragraph 1:1
    text 1:1 "Zażółć "
    emph 1:12
      text 1:13 "gęślą"
    text 1:22 " "
    code 1:23 "jaźń"
    soft_break 1:31
    text 2:3 "i "
    strong 2:5
      text 2:7 "mocno "
      emph 2:13
        text 2:14 "bardzo"
    text 2:23 " "
    link 2:24 "/a" "t"
      text 2:25 "tu"
    hard_break 2:36
    text 3:1 "koniec"
    hard_break 3:7
    html_inline 4:1 "<b>"
    text 4:4 " "
    text 4:5 "&"
    text 4:10 " "
    text 4:12 "* "
    image 4:14 "/ref" "Tytuł"
      text 4:16 "o "
      emph 4:18
        text 4:19 "p"
    text 4:25 " "
    link 4:26 "https://x.pl" ""
      text 4:27 "https://x.pl"
  block_quote 6:1
    list 6:3
      item 6:3
        heading 6:5
          text 6:7 "Tytuł "
          link 6:14 "/ref" "Tytuł"
            text 6:15 "r"
  ref_defs 8:1
`
	doc := parse([]byte(src), 0)
	if got := dumpTree(doc, true); got != want {
		t.Errorf("tree:\n%s\nwant:\n%s", got, want)
	}
}

// REQ: INL-004
func TestEmphasisBeyondSpec(t *testing.T) {
	cases := []struct{ in, want string }{
		{"*zażółć*gęślą", "<p><em>zażółć</em>gęślą</p>\n"},
		{"zażółć_gęślą_jaźń", "<p>zażółć_gęślą_jaźń</p>\n"},
		{"„*cytat*”", "<p>„<em>cytat</em>”</p>\n"},
		{"*„cytat”*ą", "<p>*„cytat”*ą</p>\n"},
		{"**a*", "<p>*<em>a</em></p>\n"},
		{"*a**", "<p><em>a</em>*</p>\n"},
		{"***a** b*", "<p><em><strong>a</strong> b</em></p>\n"},
		{"*a **b***", "<p><em>a <strong>b</strong></em></p>\n"},
		{"******a******", "<p><strong><strong><strong>a</strong></strong></strong></p>\n"},
		{"*a\u00a0*", "<p>*a\u00a0*</p>\n"},
		{"a * b * c", "<p>a * b * c</p>\n"},
	}
	for _, c := range cases {
		if got := string(Render([]byte(c.in))); got != c.want {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
}

// REQ: INL-005
func TestLinksBeyondSpec(t *testing.T) {
	cases := []struct{ in, want string }{
		{"[a](/ścieżka?q=ż&x=1 \"Tytuł „x”\")", "<p><a href=\"/%C5%9Bcie%C5%BCka?q=%C5%BC&amp;x=1\" title=\"Tytuł „x”\">a</a></p>\n"},
		{"[a](<b c>)", "<p><a href=\"b%20c\">a</a></p>\n"},
		{"[a](b\\\"c)", "<p><a href=\"b%22c\">a</a></p>\n"},
		{"[a](%41%zz)", "<p><a href=\"%41%zz\">a</a></p>\n"},
		{"[a](" + strings.Repeat("(", 32) + "x" + strings.Repeat(")", 32) + ")",
			"<p><a href=\"" + strings.Repeat("(", 32) + "x" + strings.Repeat(")", 32) + "\">a</a></p>\n"},
		{"[a](" + strings.Repeat("(", 33) + "x" + strings.Repeat(")", 33) + ")",
			"<p>[a](" + strings.Repeat("(", 33) + "x" + strings.Repeat(")", 33) + ")</p>\n"},
		{"[Zażółć GĘŚLĄ]\n\n[zażółć  gęślą]: /pl", "<p><a href=\"/pl\">Zażółć GĘŚLĄ</a></p>\n"},
		{"[a [b](/c) d](/e)", "<p>[a <a href=\"/c\">b</a> d](/e)</p>\n"},
		{"![a [b](/c) d](/e)", "<p><img src=\"/e\" alt=\"a b d\" /></p>\n"},
		{"[![a](/i)](/l)", "<p><a href=\"/l\"><img src=\"/i\" alt=\"a\" /></a></p>\n"},
		{"[a]: /x\n\n[a] [a][] [a][a] [b][a] [a][b]", "<p><a href=\"/x\">a</a> <a href=\"/x\">a</a> <a href=\"/x\">a</a> <a href=\"/x\">b</a> [a][b]</p>\n"},
	}
	for _, c := range cases {
		if got := string(Render([]byte(c.in))); got != c.want {
			t.Errorf("%q:\ngot  %q\nwant %q", c.in, got, c.want)
		}
	}
}

// pathologicalInlines to przypadki złośliwe dla treści w wierszu; n to przybliżony rozmiar wejścia.
// REQ: DET-003
var pathologicalInlines = []struct {
	name string
	gen  func(n int) string
}{
	{"emphasis openers", func(n int) string { return strings.Repeat("*a ", n) }},
	{"emphasis closers", func(n int) string { return strings.Repeat("a* ", n) }},
	{"mixed emphasis openers", func(n int) string { return strings.Repeat("*a _b **c __d ", n/4) }},
	{"openers then closers", func(n int) string { return strings.Repeat("*a ", n/2) + strings.Repeat("b* ", n/2) }},
	{"mismatched closers", func(n int) string { return strings.Repeat("*a ", n/2) + strings.Repeat("b_ ", n/2) }},
	{"runs of three and one", func(n int) string { return strings.Repeat("a***b c*d ", n/2) }},
	{"intraword runs", func(n int) string { return strings.Repeat("a**b", n) }},
	{"long delimiter run", func(n int) string { return strings.Repeat("*", n) + "a" + strings.Repeat("*", n) }},
	{"nested emphasis", func(n int) string { return strings.Repeat("*a **b ", n/2) + strings.Repeat("c** d* ", n/2) }},
	{"open brackets", func(n int) string { return strings.Repeat("[", n) }},
	{"close brackets", func(n int) string { return strings.Repeat("]", n) }},
	{"nested brackets", func(n int) string { return strings.Repeat("[", n) + "a" + strings.Repeat("]", n) }},
	{"open brackets then links", func(n int) string { return strings.Repeat("[ ", n/2) + strings.Repeat("[a](b) ", n/2) }},
	{"open images then links", func(n int) string { return strings.Repeat("![ ", n/2) + strings.Repeat("[a](b) ", n/2) }},
	{"nested images", func(n int) string { return strings.Repeat("![", n) + "a" + strings.Repeat("](b)", n) }},
	{"unclosed link tails", func(n int) string { return strings.Repeat("[a](b", n) }},
	{"unclosed link parens", func(n int) string { return strings.Repeat("[a](", n) }},
	{"unclosed angle destinations", func(n int) string { return strings.Repeat("[a](<b", n) }},
	{"unclosed titles", func(n int) string { return strings.Repeat("[a](b \"c [a](b 'c [a](b (c ", n/3) }},
	{"link with deep parens", func(n int) string { return "[a](" + strings.Repeat("(", n) + ")" }},
	{"undefined references", func(n int) string { return strings.Repeat("[a][b] ", n) }},
	{"long undefined labels", func(n int) string { return strings.Repeat("["+strings.Repeat("a", 990)+"] ", n/990+1) }},
	{"defined references", func(n int) string { return "[a]: /b\n\n" + strings.Repeat("[a] ", n) }},
	{"brackets around emphasis", func(n int) string { return strings.Repeat("[*a ", n/2) + strings.Repeat("b*] ", n/2) }},
	{"backtick runs of growing length", func(n int) string {
		var b strings.Builder
		for i := 1; b.Len() < n; i++ {
			b.WriteString(strings.Repeat("`", i) + "a ")
		}
		return b.String()
	}},
	{"single backticks then double", func(n int) string { return strings.Repeat("``a ", n/2) + strings.Repeat("`b` ", n/2) }},
	{"unclosed comments", func(n int) string { return strings.Repeat("<!-- a ", n) }},
	{"unclosed instructions", func(n int) string { return strings.Repeat("<? a ", n) }},
	{"unclosed declarations", func(n int) string { return strings.Repeat("<!A a ", n) }},
	{"unclosed CDATA", func(n int) string { return strings.Repeat("<![CDATA[ a ", n) }},
	{"unclosed attribute values", func(n int) string { return strings.Repeat("<a b=\"c ", n/2) + strings.Repeat("<a b='c ", n/2) }},
	{"open tags without end", func(n int) string { return strings.Repeat("<a b c d e ", n) }},
	{"long attribute list", func(n int) string { return "<a" + strings.Repeat(" b=c", n) }},
	{"autolink starts", func(n int) string { return strings.Repeat("<ab:c ", n) + strings.Repeat("<a@b ", n) }},
	{"long autolink without end", func(n int) string { return "<ab:" + strings.Repeat("c", n) + " <" + strings.Repeat("d", n) }},
	{"ampersands", func(n int) string { return strings.Repeat("&", n) + strings.Repeat("&#12; &amp ", n/8) }},
	{"backslashes", func(n int) string { return strings.Repeat("\\", n) + strings.Repeat("\\a\\*", n/4) }},
	{"hard breaks", func(n int) string { return strings.Repeat("a  \n", n) }},
	{"trailing spaces", func(n int) string { return "a" + strings.Repeat(" ", n) + "\nb" + strings.Repeat(" ", n) }},
	{"many lines of inlines", func(n int) string { return strings.Repeat("*a* `b` [c](d)\n", n/4) }},
	{"emphasis across many lines", func(n int) string { return "*" + strings.Repeat("a\n", n) + "*" }},
	{"setext heading with many lines", func(n int) string { return strings.Repeat("*a* b\n", n) + "===\n" }},
}

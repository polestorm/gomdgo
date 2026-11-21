package gomdgo

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// REQ: EXT-005
func TestAlerts(t *testing.T) {
	conv := New(Config{Enable: Alerts, AlertTitles: map[string]string{"note": "Uwaga", "tip": "A & <B>"}})
	checkCases(t, conv, [][2]string{
		{"> [!NOTE]\n> Zażółć.", "<blockquote class=\"alert alert-note\">\n<p class=\"alert-title\">Uwaga</p>\n<p>Zażółć.</p>\n</blockquote>\n"},
		{"> [!tip]  \n> a\n>\n> b", "<blockquote class=\"alert alert-tip\">\n<p class=\"alert-title\">A &amp; &lt;B&gt;</p>\n<p>a</p>\n<p>b</p>\n</blockquote>\n"},
		{"> [!Important]\n> a", "<blockquote class=\"alert alert-important\">\n<p>a</p>\n</blockquote>\n"},
		{"> [!WARNING]\n> # h\n> - x", "<blockquote class=\"alert alert-warning\">\n<h1>h</h1>\n<ul>\n<li>x</li>\n</ul>\n</blockquote>\n"},
		{"> [!CAUTION]\nleniwie", "<blockquote class=\"alert alert-caution\">\n<p>leniwie</p>\n</blockquote>\n"},
		{"> [!NOTE]", "<blockquote class=\"alert alert-note\">\n<p class=\"alert-title\">Uwaga</p>\n</blockquote>\n"},
		{"> [!NOTE] a", "<blockquote>\n<p>[!NOTE] a</p>\n</blockquote>\n"},
		{"> [!INFO]\n> a", "<blockquote>\n<p>[!INFO]\na</p>\n</blockquote>\n"},
		{"> a\n> [!NOTE]", "<blockquote>\n<p>a\n[!NOTE]</p>\n</blockquote>\n"},
		{"> # h\n>\n> [!NOTE]", "<blockquote>\n<h1>h</h1>\n<p>[!NOTE]</p>\n</blockquote>\n"},
		{"> [!NOTE\n> a", "<blockquote>\n<p>[!NOTE\na</p>\n</blockquote>\n"},
		{">     [!NOTE]", "<blockquote>\n<pre><code>[!NOTE]\n</code></pre>\n</blockquote>\n"},
		{"- > [!NOTE]\n  > a\n\n> > [!TIP]\n> > b", "<ul>\n<li>\n<blockquote class=\"alert alert-note\">\n<p class=\"alert-title\">Uwaga</p>\n<p>a</p>\n</blockquote>\n</li>\n</ul>\n" +
			"<blockquote>\n<blockquote class=\"alert alert-tip\">\n<p class=\"alert-title\">A &amp; &lt;B&gt;</p>\n<p>b</p>\n</blockquote>\n</blockquote>\n"},
		{"[!NOTE]\na", "<p>[!NOTE]\na</p>\n"},
	})
	if got := string(Render([]byte("> [!NOTE]\n> a"))); got != "<blockquote>\n<p>[!NOTE]\na</p>\n</blockquote>\n" {
		t.Errorf("without the extension: %q", got)
	}
}

// REQ: EXT-006
func TestFootnotes(t *testing.T) {
	ref := func(n int, k ...int) string {
		id := fmt.Sprint(n)
		if len(k) > 0 {
			id += fmt.Sprint(":", k[0])
		}
		return fmt.Sprintf(`<sup><a href="#fn:%d" id="fnref:%s">%d</a></sup>`, n, id, n)
	}
	back := func(n int, k ...int) string {
		id := fmt.Sprint(n)
		if len(k) > 0 {
			id += fmt.Sprint(":", k[0])
		}
		return `<a href="#fnref:` + id + `" class="footnote-backref">↩</a>`
	}
	section := func(items ...string) string {
		return "<section class=\"footnotes\">\n<ol>\n" + strings.Join(items, "") + "</ol>\n</section>\n"
	}
	conv := New(Config{Enable: Footnotes, HeadingIDs: true})
	checkCases(t, conv, [][2]string{
		{"a[^1]\n\n[^1]: Zażółć *gęślą*.", "<p>a" + ref(1) + "</p>\n" + section("<li id=\"fn:1\">\n<p>Zażółć <em>gęślą</em>. "+back(1)+"</p>\n</li>\n")},
		// numeracja według pierwszego odwołania, nie według kolejności definicji ani etykiety
		{"[^b]: B\n[^a]: A\n\nx[^a] y[^b] z[^A]", "<p>x" + ref(1) + " y" + ref(2) + " z" + ref(1, 2) + "</p>\n" +
			section("<li id=\"fn:1\">\n<p>A "+back(1)+" "+back(1, 2)+"</p>\n</li>\n", "<li id=\"fn:2\">\n<p>B "+back(2)+"</p>\n</li>\n")},
		{"a[^x]\n\n[^x]:\n    pierwszy\n\n    drugi\nleniwie\n\n        kod\n\nkoniec", "<p>a" + ref(1) + "</p>\n<p>koniec</p>\n" +
			section("<li id=\"fn:1\">\n<p>pierwszy</p>\n<p>drugi\nleniwie</p>\n<pre><code>kod\n</code></pre>\n"+back(1)+"\n</li>\n")},
		{"a[^1]\n\n[^1]: - x\n    - y", "<p>a" + ref(1) + "</p>\n" + section("<li id=\"fn:1\">\n<ul>\n<li>x</li>\n<li>y</li>\n</ul>\n"+back(1)+"\n</li>\n")},
		{"a[^1]\n\n[^1]:", "<p>a" + ref(1) + "</p>\n" + section("<li id=\"fn:1\">\n"+back(1)+"\n</li>\n")},
		// przypis w przypisie dostaje numer przy zapisie treści pierwszego
		{"a[^1]\n\n[^2]: dwa\n[^1]: jeden[^2]", "<p>a" + ref(1) + "</p>\n" +
			section("<li id=\"fn:1\">\n<p>jeden"+ref(2)+" "+back(1)+"</p>\n</li>\n", "<li id=\"fn:2\">\n<p>dwa "+back(2)+"</p>\n</li>\n")},
		{"a[^1]\n\n[^1]: sam do siebie[^1]", "<p>a" + ref(1) + "</p>\n" + section("<li id=\"fn:1\">\n<p>sam do siebie"+ref(1, 2)+" "+back(1)+" "+back(1, 2)+"</p>\n</li>\n")},
		{"a[^brak] [^] [^a b] [^1]\n\n[^nieużywany]: x", "<p>a[^brak] [^] [^a b] [^1]</p>\n"},
		{"akapit\n[^1]: przerywa akapit\n\nb[^1]", "<p>akapit</p>\n<p>b" + ref(1) + "</p>\n" + section("<li id=\"fn:1\">\n<p>przerywa akapit "+back(1)+"</p>\n</li>\n")},
		{"[^1]: pierwsza\n[^1]: druga\n\na[^1]", "<p>a" + ref(1) + "</p>\n" + section("<li id=\"fn:1\">\n<p>pierwsza "+back(1)+"</p>\n</li>\n")},
		{"> [^1]: w cytacie\n\n- a[^1] `[^1]`\n\n# fn:1 [^1]", "<blockquote>\n</blockquote>\n<ul>\n<li>a" + ref(1) + " <code>[^1]</code></li>\n</ul>\n<h1 id=\"fn1\">fn:1 " + ref(1, 2) + "</h1>\n" +
			section("<li id=\"fn:1\">\n<p>w cytacie "+back(1)+" "+back(1, 2)+"</p>\n</li>\n")},
		{"    [^1]: kod\n\n[^1](/u)", "<pre><code>[^1]: kod\n</code></pre>\n<p><a href=\"/u\">^1</a></p>\n"},
	})
	// Bez rozszerzenia [^1]: jest definicją odsyłacza CommonMark.
	if got := string(Render([]byte("a[^1]\n\n[^1]: /url"))); got != "<p>a<a href=\"/url\">^1</a></p>\n" {
		t.Errorf("without the extension: %q", got)
	}
}

// REQ: EXT-007
func TestWalk(t *testing.T) {
	src := "# Tytuł *a*\n\n" +
		"Tekst [odn](javascript:x \"t\") ![obraz](/i.png) `kod` <b> [[cel]] www.x.pl[^1]\\\n~~s~~\n\n" +
		"> [!TIP]\n> rada\n\n" +
		"3. [x] zadanie\n\n" +
		"```go opis\nx := 1\n```\n\n" +
		"| a | b |\n|:--|--:|\n| c |\n\n" +
		"<div>\n\n---\n\n[r]: /u\n\n[^1]: przypis\n"
	conv := New(Config{Enable: Alerts | Footnotes, References: func(Reference) (ReferenceLink, error) { return ReferenceLink{}, nil }})
	doc := conv.Parse([]byte(src))
	var b strings.Builder
	depth := 0
	doc.Walk(func(n Node, entering bool) bool {
		if !entering {
			depth--
			return true
		}
		line, col := n.Position()
		fmt.Fprintf(&b, "%s%s %d:%d", strings.Repeat(" ", depth), n.Kind(), line, col)
		switch n.Kind() {
		case KindText, KindCode, KindCodeBlock, KindHTMLBlock, KindHTMLInline, KindReference, KindHeading, KindFootnote, KindFootnoteRef:
			fmt.Fprintf(&b, " %q", n.Text())
		case KindLink, KindImage:
			fmt.Fprintf(&b, " %q %q %q", n.Destination(), n.Title(), n.Text())
		case KindList:
			fmt.Fprintf(&b, " ordered=%v start=%d tight=%v", n.Ordered(), n.Start(), n.Tight())
		case KindTaskCheckbox:
			fmt.Fprintf(&b, " checked=%v", n.Checked())
		case KindTableRow:
			fmt.Fprintf(&b, " header=%v", n.Header())
		case KindTableCell:
			fmt.Fprintf(&b, " column=%d align=%q", n.Column(), n.Align())
		case KindBlockQuote:
			fmt.Fprintf(&b, " alert=%q", n.Alert())
		}
		if n.Kind() == KindHeading {
			fmt.Fprintf(&b, " level=%d", n.Level())
		}
		if n.Kind() == KindCodeBlock {
			fmt.Fprintf(&b, " lang=%q info=%q", n.Language(), n.Info())
		}
		b.WriteByte('\n')
		depth++
		return n.Kind() != KindImage // dzieci obrazu pomijamy; wyjście i tak jest zgłaszane
	})
	want := `document 1:1
 heading 1:1 "Tytuł a" level=1
  text 1:3 "Tytuł "
  emph 1:10
   text 1:11 "a"
 paragraph 3:1
  text 3:1 "Tekst "
  link 3:7 "javascript:x" "t" "odn"
   text 3:8 "odn"
  text 3:30 " "
  image 3:31 "/i.png" "" "obraz"
  text 3:47 " "
  code 3:48 "kod"
  text 3:53 " "
  html_inline 3:54 "<b>"
  text 3:57 " "
  reference 3:58 "cel"
  text 3:65 " "
  link 3:66 "http://www.x.pl" "" "www.x.pl"
   text 3:66 "www.x.pl"
  footnote_ref 3:74 "1"
  hard_break 3:78
  strike 4:1
   text 4:3 "s"
 block_quote 6:1 alert="tip"
  paragraph 7:3
   text 7:3 "rada"
 list 9:1 ordered=true start=3 tight=true
  item 9:1
   paragraph 9:4
    task_checkbox 9:4 checked=true
    text 9:7 " zadanie"
 code_block 11:1 "x := 1\n" lang="go" info="go opis"
 table 15:1
  table_row 15:1 header=true
   table_cell 15:3 column=0 align="left"
    text 15:3 "a"
   table_cell 15:7 column=1 align="right"
    text 15:7 "b"
  table_row 17:1 header=false
   table_cell 17:3 column=0 align="left"
    text 17:3 "c"
   table_cell 17:6 column=1 align="right"
 html_block 19:1 "<div>\n"
 thematic_break 21:1
 footnote 25:1 "1"
  paragraph 25:7
   text 25:7 "przypis"
`
	if b.String() != want {
		t.Errorf("walk:\n%s\nwant:\n%s", b.String(), want)
	}
	if depth != 0 {
		t.Errorf("enter and leave calls are not paired: depth %d", depth)
	}

	// Odwiedzanie i renderowanie mogą biec równolegle; dokument się nie zmienia.
	before, _ := doc.HTML()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); doc.Walk(func(Node, bool) bool { return true }) }()
		go func() {
			defer wg.Done()
			if out, _ := doc.HTML(); string(out) != string(before) {
				t.Error("output changed")
			}
		}()
	}
	wg.Wait()
}

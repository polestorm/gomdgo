package gomdgo

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Znaczniki i atrybuty, które renderer ma prawo wytworzyć przy ustawieniach bez przepuszczania HTML.
var safeTags = map[string][]string{
	"p": nil, "h1": {"id"}, "h2": {"id"}, "h3": {"id"}, "h4": {"id"}, "h5": {"id"}, "h6": {"id"},
	"ul": nil, "ol": {"start"}, "li": nil, "blockquote": nil, "pre": nil, "code": {"class"}, "hr": nil, "br": nil,
	"em": nil, "strong": nil, "a": {"href", "title"}, "img": {"src", "alt", "title"},
}

// checkSafe sprawdza wynik niezależnie od renderera: poprawny fragment, wyłącznie znaczniki i atrybuty
// z zamkniętej listy, a adresy zakodowane i ze schematem z listy domyślnej albo bez schematu.
func checkSafe(out []byte) error {
	tags, err := scanFragment(out)
	if err != nil {
		return err
	}
	for _, tag := range tags {
		allowed, ok := safeTags[tag.name]
		if !ok {
			return fmt.Errorf("element <%s> is not on the list", tag.name)
		}
		for _, attr := range tag.attrs {
			name, value := attr[0], html.UnescapeString(attr[1])
			if !strings.Contains(" "+strings.Join(allowed, " ")+" ", " "+name+" ") {
				return fmt.Errorf("attribute %s on <%s> is not on the list", name, tag.name)
			}
			switch name {
			case "href", "src":
				if err := checkSafeURL(value); err != nil {
					return fmt.Errorf("%s=%q: %v", name, value, err)
				}
			case "class":
				if !strings.HasPrefix(value, "language-") {
					return fmt.Errorf("class %q", value)
				}
			case "start":
				if strings.Trim(value, "0123456789") != "" {
					return fmt.Errorf("start %q", value)
				}
			}
		}
	}
	return nil
}

func checkSafeURL(url string) error {
	for _, c := range []byte(url) {
		if c <= ' ' || c >= 0x7f || strings.IndexByte("\"<>\\^`{|}", c) >= 0 {
			return fmt.Errorf("byte %#x is not percent-encoded", c)
		}
	}
	// Przeglądarka uznaje za schemat wszystko do pierwszego dwukropka, o ile wcześniej nie ma / ? ani #.
	head, _, found := strings.Cut(url, ":")
	if !found || strings.ContainsAny(head, "/?#") {
		return nil
	}
	switch strings.ToLower(head) {
	case "http", "https", "mailto":
		return nil
	}
	if head == "" || strings.ContainsAny(head, "%@!$&'()*,;=~_") || head[0] >= '0' && head[0] <= '9' ||
		head[0] == '+' || head[0] == '-' || head[0] == '.' {
		return nil // to nie jest poprawny schemat, więc adres jest względny
	}
	return fmt.Errorf("scheme %q is not allowed", head)
}

// attacks to wejścia, które mają przemycić wykonywalną treść albo niedozwolony adres.
var attacks = []string{
	"<script>alert(1)</script>",
	"<img src=x onerror=alert(1)>",
	"a <b onclick=\"x()\">b</b> <!-- c --> <?php ?> <![CDATA[x]]> <!DOCTYPE html>",
	"<div>\n*a*\n</div>\n\n<style>p{}</style>\n\n<iframe src=\"//x\"></iframe>",
	"[a](javascript:alert(1))", "[a](JaVaScRiPt:alert(1))", "[a](<javascript:alert(1)>)",
	"[a](java\tscript:alert(1))", "[a](\\javascript:alert(1))", "[a](javascript\\:alert(1))",
	"[a](&#106;avascript:alert(1))", "[a](javascript&colon;alert(1))", "[a](javascript&#58;alert(1))",
	"[a](&#x6A;&#x61;vascript&#x3A;alert(1))", "[a](%6Aavascript:alert(1))", "[a]( javascript:alert(1))",
	"[a](\x01javascript:alert(1))", "[a](<\x01javascript:alert(1)>)", "[a](<  javascript:alert(1)>)",
	"[a](vbscript:x)", "[a](data:text/html,<script>alert(1)</script>)", "[a](file:///etc/passwd)",
	"![a](javascript:alert(1))", "![a](data:image/svg+xml,<svg onload=alert(1)>)",
	"<javascript:alert(1)>", "<JAVASCRIPT:alert(1)>", "<data:text/html,x>", "<vbscript:x>",
	"[a]: javascript:alert(1)\n\n[a] [b][a] ![c][a]",
	"[a](x \"t\\\" onmouseover=\\\"alert(1)\")", "[a](x 't\" onmouseover=\"alert(1)')",
	"![a\" onerror=\"alert(1)](x)", "![a](x\" onerror=\"alert(1))", "[a](x\"><script>alert(1)</script>)",
	"```\"><script>\ncode\n```", "``` a\" onclick=\"x\n```", "~~~ &quot;&gt;&lt;script&gt;\n~~~",
	"# <script>x</script>", "> <script>x</script>", "- <script>x</script>", "    <script>x</script>",
	"&lt;script&gt;alert(1)&lt;/script&gt;", "&#60;script&#62;", "\\<script>",
	"<a href=\"javascript:alert(1)\">a</a>", "<svg/onload=alert(1)>", "<math><mi xlink:href=\"data:x\">",
	"1234567890. a\n\n999999999) b",
	"[a](//example.com/x)", "[a](/x:y)", "[a](x/y:z)", "[a](?x:y)", "[a](#x:y)", "[a](1a:b)", "[a](a%20b:c)",
}

// REQ: SAFE-001 SAFE-005
func TestSafeByDefault(t *testing.T) {
	inputs := append([]string{}, attacks...)
	for _, path := range []string{commonMarkSpec, gfmSpec} {
		for _, e := range loadSpec(t, path) {
			inputs = append(inputs, e.Markdown)
		}
	}
	files, _ := filepath.Glob("testdata/corpus/*.md")
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, string(b))
	}
	for _, conv := range []*Converter{defaultConverter, New(Config{RawHTML: RawHTMLOmit, HeadingIDs: true})} {
		for _, in := range inputs {
			out, err := conv.Render([]byte(in))
			if err != nil {
				t.Fatal(err)
			}
			if err := checkSafe(out); err != nil {
				t.Errorf("%+v: %v\nsource: %q\noutput: %q", conv.cfg, err, in, out)
			}
			if lower := bytes.ToLower(out); bytes.Contains(lower, []byte("<script")) {
				t.Errorf("script element in output\nsource: %q\noutput: %q", in, out)
			}
		}
	}
}

// REQ: SAFE-002
func TestRawHTMLPolicy(t *testing.T) {
	src := "a <b>c</b> <!-- d -->\n\n<div class=\"x\">\ne\n</div>\n\nf\n"
	cases := []struct {
		policy RawHTML
		want   string
	}{
		{RawHTMLEscape, "<p>a &lt;b&gt;c&lt;/b&gt; &lt;!-- d --&gt;</p>\n<p>&lt;div class=&quot;x&quot;&gt;\ne\n&lt;/div&gt;</p>\n<p>f</p>\n"},
		{RawHTMLOmit, "<p>a c </p>\n<p>f</p>\n"},
		{RawHTMLPass, "<p>a <b>c</b> <!-- d --></p>\n<div class=\"x\">\ne\n</div>\n<p>f</p>\n"},
		{RawHTML(99), "<p>a &lt;b&gt;c&lt;/b&gt; &lt;!-- d --&gt;</p>\n<p>&lt;div class=&quot;x&quot;&gt;\ne\n&lt;/div&gt;</p>\n<p>f</p>\n"},
	}
	for _, c := range cases {
		got, _ := New(Config{RawHTML: c.policy}).Render([]byte(src))
		if string(got) != c.want {
			t.Errorf("policy %d:\ngot  %q\nwant %q", c.policy, got, c.want)
		}
	}
	if got := string(Render([]byte("- a <b>\n- c"))); got != "<ul>\n<li>a &lt;b&gt;</li>\n<li>c</li>\n</ul>\n" {
		t.Errorf("escaped inline HTML in a tight list: %q", got)
	}
	if got, _ := New(Config{RawHTML: RawHTMLOmit}).Render([]byte("- <div>\n- c")); string(got) != "<ul>\n<li></li>\n<li>c</li>\n</ul>\n" {
		t.Errorf("omitted HTML block in a tight list: %q", got)
	}
	if got := string(Render([]byte("- a\n  <div>\n- c\n\n> <div>"))); got !=
		"<ul>\n<li>a\n&lt;div&gt;</li>\n<li>c</li>\n</ul>\n<blockquote>\n<p>&lt;div&gt;</p>\n</blockquote>\n" {
		t.Errorf("escaped HTML block after a paragraph in a tight list: %q", got)
	}
	if got := string(Render([]byte("- <div>\n- c"))); got != "<ul>\n<li>&lt;div&gt;</li>\n<li>c</li>\n</ul>\n" {
		t.Errorf("escaped HTML block in a tight list: %q", got)
	}
}

// REQ: SAFE-003 SAFE-004
func TestSchemes(t *testing.T) {
	link := func(href string) string { return "<p><a href=\"" + href + "\">a</a></p>\n" }
	const dropped = "<p>a</p>\n"
	cases := []struct{ in, want string }{
		{"[a](https://example.com/ż)", link("https://example.com/%C5%BC")},
		{"[a](HTTP://EXAMPLE.COM)", link("HTTP://EXAMPLE.COM")},
		{"[a](mailto:jan@example.com)", link("mailto:jan@example.com")},
		{"[a](/ścieżka)", link("/%C5%9Bcie%C5%BCka")},
		{"[a](plik.md#sekcja)", link("plik.md#sekcja")},
		{"[a](#x:y)", link("#x:y")},
		{"[a](./x:y)", link("./x:y")},
		{"[a](//example.com)", link("//example.com")},
		{"[a](1a:b)", link("1a:b")},
		{"[a](a%20b:c)", link("a%20b:c")},
		{"[a](<a b:c>)", link("a%20b:c")},
		{"[a](:b)", link(":b")},
		{"[a](javascript:alert(1))", dropped},
		{"[a](JAVASCRIPT:alert(1))", dropped},
		{"[a](javascript&colon;alert(1))", dropped},
		{"[a](&#106;avascript:alert(1))", dropped},
		{"[a](javascript\\:alert(1))", dropped},
		{"[a](ftp://example.com)", dropped},
		{"[a](tel:+48123456789)", dropped},
		{"[a](data:text/plain,x)", dropped},
		{"[*a*](x-y.z+w:b \"t\")", "<p><em>a</em></p>\n"},
		{"<ftp://example.com>", "<p>ftp://example.com</p>\n"},
		{"![o *p*](data:image/png;base64,AAAA \"t\")", "<p>o p</p>\n"},
		{"![o](https://example.com/o.png)", "<p><img src=\"https://example.com/o.png\" alt=\"o\" /></p>\n"},
		{"[a]: ftp://x\n\n[a] i [b][a]", "<p>a i b</p>\n"},
	}
	for _, c := range cases {
		if got := string(Render([]byte(c.in))); got != c.want {
			t.Errorf("%q:\ngot  %q\nwant %q", c.in, got, c.want)
		}
	}

	custom := New(Config{Schemes: []string{"FTP", "tel:"}})
	for in, want := range map[string]string{
		"[a](ftp://x)":   link("ftp://x"),
		"[a](Tel:+48)":   link("Tel:+48"),
		"[a](https://x)": dropped,
		"[a](/x)":        link("/x"),
	} {
		if got, _ := custom.Render([]byte(in)); string(got) != want {
			t.Errorf("custom schemes, %q: got %q, want %q", in, got, want)
		}
	}
	none := New(Config{Schemes: []string{}})
	if got, _ := none.Render([]byte("[a](https://x) [a](/x)")); string(got) != "<p>a <a href=\"/x\">a</a></p>\n" {
		t.Errorf("empty scheme list: got %q", got)
	}
	any := New(Config{AnyScheme: true})
	if got, _ := any.Render([]byte("[a](javascript:x)")); string(got) != link("javascript:x") {
		t.Errorf("any scheme: got %q", got)
	}
	// Lista przekazana do New nie jest współdzielona z wywołującym.
	list := []string{"ftp"}
	conv := New(Config{Schemes: list})
	list[0] = "javascript"
	if got, _ := conv.Render([]byte("[a](javascript:x) [a](ftp:x)")); string(got) != "<p>a <a href=\"ftp:x\">a</a></p>\n" {
		t.Errorf("scheme list must be copied: got %q", got)
	}
}

// REQ: OUT-002
func TestHeadingIDs(t *testing.T) {
	src := "# Zażółć gęślą jaźń!\n\n## Wstęp\n\n## Wstęp\n\n### *Kod* `a_b` i [odnośnik](/x)\n\n# ?!\n\n# ?!\n\n" +
		"## wstęp-1\n\n## Wstęp\n\nRozdział 2.1\n---\n\n# <b>x</b> \"y\"\n"
	want := "<h1 id=\"zażółć-gęślą-jaźń\">Zażółć gęślą jaźń!</h1>\n" +
		"<h2 id=\"wstęp\">Wstęp</h2>\n<h2 id=\"wstęp-1\">Wstęp</h2>\n" +
		"<h3 id=\"kod-a_b-i-odnośnik\"><em>Kod</em> <code>a_b</code> i <a href=\"/x\">odnośnik</a></h3>\n" +
		"<h1 id=\"section\">?!</h1>\n<h1 id=\"section-1\">?!</h1>\n" +
		"<h2 id=\"wstęp-1-1\">wstęp-1</h2>\n<h2 id=\"wstęp-2\">Wstęp</h2>\n" +
		"<h2 id=\"rozdział-21\">Rozdział 2.1</h2>\n" +
		"<h1 id=\"x-y\">&lt;b&gt;x&lt;/b&gt; &quot;y&quot;</h1>\n"
	got, _ := New(Config{HeadingIDs: true}).Render([]byte(src))
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	if got := string(Render([]byte("# a"))); got != "<h1>a</h1>\n" {
		t.Errorf("heading ids must be off by default: %q", got)
	}
}

// REQ: OUT-001
func TestConverterAPI(t *testing.T) {
	src := []byte("# Tytuł\n\n[a](/b) <i>c</i>\n\n# Tytuł\n")
	conv := New(Config{HeadingIDs: true})
	doc := conv.Parse(src)
	first, err := doc.HTML()
	if err != nil {
		t.Fatal(err)
	}
	// Ten sam dokument renderowany równolegle daje ten sam wynik; numeracja identyfikatorów nie przecieka
	// między przebiegami.
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				again, _ := doc.HTML()
				other, _ := conv.Render(src)
				if !bytes.Equal(first, again) || !bytes.Equal(first, other) {
					t.Errorf("output differs between calls:\n%s\n%s\n%s", first, again, other)
					return
				}
			}
		}()
	}
	wg.Wait()
	want := "<h1 id=\"tytuł\">Tytuł</h1>\n<p><a href=\"/b\">a</a> &lt;i&gt;c&lt;/i&gt;</p>\n<h1 id=\"tytuł-1\">Tytuł</h1>\n"
	if string(first) != want {
		t.Errorf("got %q, want %q", first, want)
	}
	if def, viaConv := Render(src), New(Config{}); true {
		out, _ := viaConv.Render(src)
		if !bytes.Equal(def, out) {
			t.Errorf("Render must equal New(Config{}).Render: %q, %q", def, out)
		}
	}
}

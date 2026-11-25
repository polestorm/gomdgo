package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// REQ: EXT-001
func TestCodeBlockRendererInput(t *testing.T) {
	var seen []CodeBlock
	conv := New(Config{CodeBlocks: map[string]CodeBlockFunc{
		"mermaid": func(b CodeBlock) ([]byte, error) {
			seen = append(seen, b)
			if len(b.Code) > 0 {
				b.Code[0] = '!' // funkcja dostaje kopię, więc dokumentu to nie zmienia
			}
			return []byte("<svg></svg>"), nil
		},
	}})
	doc := conv.Parse([]byte("```mermaid tytuł=\"A&amp;B\"\ngraf TD\n  A-->B\n```\n\n> - ~~~mermaid\n>   C\n\n```mermaid\n```\n"))
	for i := 0; i < 2; i++ {
		if _, err := doc.HTML(); err != nil {
			t.Fatal(err)
		}
	}
	want := []CodeBlock{
		{Lang: "mermaid", Info: "mermaid tytuł=\"A&B\"", Code: []byte("graf TD\n  A-->B\n"), Line: 1},
		{Lang: "mermaid", Info: "mermaid", Code: []byte("C\n"), Line: 6},
		{Lang: "mermaid", Info: "mermaid", Code: nil, Line: 9},
	}
	if len(seen) != 2*len(want) {
		t.Fatalf("got %d calls, want %d", len(seen), 2*len(want))
	}
	for i, got := range seen[len(want):] { // drugi przebieg: treść musi być nietknięta przez pierwszy
		w := want[i]
		code := append([]byte(nil), got.Code...)
		if len(code) > 0 {
			code[0] = w.Code[0]
		}
		if got.Lang != w.Lang || got.Info != w.Info || got.Line != w.Line || !bytes.Equal(code, w.Code) || (len(got.Code) > 0 && got.Code[0] != '!') {
			t.Errorf("call %d: got %+v (code %q), want %+v (code %q)", i, got, got.Code, w, w.Code)
		}
	}
}

// REQ: EXT-001 EXT-004
func TestCodeBlockRendererOutput(t *testing.T) {
	calls := 0
	registry := map[string]CodeBlockFunc{
		"d": func(b CodeBlock) ([]byte, error) {
			calls++
			return []byte(fmt.Sprintf("<figure data-line=\"%d\">%s</figure>", b.Line, bytes.TrimSpace(b.Code))), nil
		},
	}
	conv := New(Config{CodeBlocks: registry})
	delete(registry, "d") // rejestr jest kopiowany
	cases := [][2]string{
		{"```d\nA\n```", "<figure data-line=\"1\">A</figure>\n"},
		{"a\n```d\nA\n```\nb", "<p>a</p>\n<figure data-line=\"2\">A</figure>\n<p>b</p>\n"},
		{"- x\n  ```d\n  A\n  ```\n- y", "<ul>\n<li>x\n<figure data-line=\"2\">A</figure>\n</li>\n<li>y</li>\n</ul>\n"},
		{"> ~~~d opis\n> A", "<blockquote>\n<figure data-line=\"1\">A</figure>\n</blockquote>\n"},
		{"    d\n\n```D\nA\n```\n\n```dd\nA\n```", "<pre><code>d\n</code></pre>\n<figure data-line=\"3\">A</figure>\n<pre><code class=\"language-dd\">A\n</code></pre>\n"},
		{"``` &#100; opis\nB\n```", "<figure data-line=\"1\">B</figure>\n"},
	}
	checkCases(t, conv, cases)
	doc := conv.Parse([]byte("```d\nA\n```"))
	before := calls
	first, _ := doc.HTML()
	second, _ := doc.HTML()
	if !bytes.Equal(first, second) || calls != before+2 {
		t.Errorf("repeated rendering: %q, %q, %d calls", first, second, calls-before)
	}
	if got := string(Render([]byte("```d\nA\n```"))); got != "<pre><code class=\"language-d\">A\n</code></pre>\n" {
		t.Errorf("without a renderer: %q", got)
	}
}

// REQ: EXT-001
func TestCodeBlockLanguageCase(t *testing.T) {
	var langs []string
	record := func(tag string) CodeBlockFunc {
		return func(b CodeBlock) ([]byte, error) {
			langs = append(langs, tag+":"+b.Lang)
			return []byte("<" + tag + ">"), nil
		}
	}
	src := []byte("```Go\nA\n```\n\n```GO\nB\n```\n\n```go\nC\n```\n\n```ŻÓŁW\nD\n```\n")
	// Klucze różniące się tylko wielkością liter: wygrywa pierwszy w porządku bajtowym („Go” < „go”), niezależnie
	// od kolejności przechodzenia mapy; sprawdzamy wiele konwerterów.
	for i := 0; i < 20; i++ {
		langs = langs[:0]
		conv := New(Config{CodeBlocks: map[string]CodeBlockFunc{"go": record("lower"), "Go": record("upper"), "żółw": record("turtle")}})
		out, err := conv.Parse(src).HTML()
		if err != nil {
			t.Fatal(err)
		}
		want := "<upper>\n<upper>\n<upper>\n<turtle>\n"
		if string(out) != want || strings.Join(langs, " ") != "upper:Go upper:GO upper:go turtle:ŻÓŁW" {
			t.Fatalf("got %q, calls %v", out, langs)
		}
	}
}

// REQ: EXT-002
func TestReferences(t *testing.T) {
	var seen []Reference
	conv := New(Config{References: func(r Reference) (ReferenceLink, error) {
		seen = append(seen, r)
		switch {
		case strings.HasPrefix(r.Target, "NC-"):
			return ReferenceLink{Href: "/nc/" + r.Target, Title: "Niezgodność \"" + r.Target + "\""}, nil
		case r.Target == "tekst":
			return ReferenceLink{Href: "/ścieżka?a=1&b=2", Text: "<b>Zażółć</b> & gęślą"}, nil
		case r.Target == "js":
			return ReferenceLink{Href: "javascript:alert(1)", Text: "klik"}, nil
		case r.Target == "tel":
			return ReferenceLink{Href: "tel:+48123"}, nil
		}
		return ReferenceLink{}, nil
	}})
	checkCases(t, conv, [][2]string{
		{"Patrz [[NC-2026-017]].", "<p>Patrz <a href=\"/nc/NC-2026-017\" title=\"Niezgodność &quot;NC-2026-017&quot;\">NC-2026-017</a>.</p>\n"},
		{"[[ tekst ]] i [[js]] i [[tel]]", "<p><a href=\"/%C5%9Bcie%C5%BCka?a=1&amp;b=2\">&lt;b&gt;Zażółć&lt;/b&gt; &amp; gęślą</a> i klik i <a href=\"tel:+48123\">tel</a></p>\n"},
		{"[[nieznane]] [[<x>]]", "<p>nieznane &lt;x&gt;</p>\n"},
		{"[[]] [[ ]] [[a\nb]] [[a]b]] [[a[b]] [a] `[[NC-1]]`", "<p>[[]] [[ ]] [[a\nb]] [[a]b]] [[a[b]] [a] <code>[[NC-1]]</code></p>\n"},
		{"*[[NC-1]]* ~~[[NC-2]]~~ # nie nagłówek", "<p><em><a href=\"/nc/NC-1\" title=\"Niezgodność &quot;NC-1&quot;\">NC-1</a></em> <del><a href=\"/nc/NC-2\" title=\"Niezgodność &quot;NC-2&quot;\">NC-2</a></del> # nie nagłówek</p>\n"},
		{"| a |\n|---|\n| [[NC-3]] |", table("<th>a</th>\n", "<td><a href=\"/nc/NC-3\" title=\"Niezgodność &quot;NC-3&quot;\">NC-3</a></td>\n")},
		{"[[NC-4]](/x) [a [[NC-5]] b](/y)", "<p><a href=\"/nc/NC-4\" title=\"Niezgodność &quot;NC-4&quot;\">NC-4</a>(/x) <a href=\"/y\">a <a href=\"/nc/NC-5\" title=\"Niezgodność &quot;NC-5&quot;\">NC-5</a> b</a></p>\n"},
	})
	seen = nil
	if _, err := conv.Render([]byte("a\n\n> x [[NC-9]]\n")); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 || seen[0] != (Reference{Target: "NC-9", Line: 3, Column: 5}) {
		t.Errorf("reference passed to the resolver: %+v", seen)
	}
	// Bez funkcji zapis [[cel]] jest zwykłym tekstem CommonMark, a z etykietą także odsyłaczem.
	if got := string(Render([]byte("[[NC-1]]\n\n[nc-1]: /u"))); got != "<p>[<a href=\"/u\">NC-1</a>]</p>\n" {
		t.Errorf("without a resolver: %q", got)
	}
	// Opis nagłówka i obrazu zawiera cel odwołania.
	ids := New(Config{HeadingIDs: true, References: func(Reference) (ReferenceLink, error) { return ReferenceLink{}, nil }})
	if got, _ := ids.Render([]byte("# Raport [[NC-1]]\n\n![a [[b]]](/i)")); string(got) != "<h1 id=\"raport-nc-1\">Raport NC-1</h1>\n<p><img src=\"/i\" alt=\"a b\" /></p>\n" {
		t.Errorf("plain text of references: %q", got)
	}
}

// REQ: EXT-003
func TestExtensionErrors(t *testing.T) {
	boom := errors.New("boom")
	var calls []string
	conv := New(Config{
		CodeBlocks: map[string]CodeBlockFunc{"d": func(b CodeBlock) ([]byte, error) {
			calls = append(calls, "code:"+strings.TrimSpace(string(b.Code)))
			if bytes.Contains(b.Code, []byte("zły")) {
				return []byte("częściowy wynik"), boom
			}
			return []byte("ok"), nil
		}},
		References: func(r Reference) (ReferenceLink, error) {
			calls = append(calls, "ref:"+r.Target)
			if r.Target == "zły" {
				return ReferenceLink{Href: "/x"}, boom
			}
			return ReferenceLink{}, nil
		},
	})
	out, err := conv.Render([]byte("[[a]]\n\n```d\ndobry\n```\n\n> ```d\n> zły\n> ```\n\n[[b]]\n\n```d\npóźniej\n```\n"))
	var e *Error
	if out != nil || !errors.As(err, &e) || !errors.Is(err, boom) {
		t.Fatalf("got output %q and error %v", out, err)
	}
	if e.Code != ErrCodeBlock || e.Line != 7 || e.Column != 3 || e.Detail != "d" {
		t.Errorf("error fields: %+v", e)
	}
	if want := `gomdgo: line 7, column 3: code block renderer for language "d" failed: boom`; err.Error() != want {
		t.Errorf("message: got %q, want %q", err.Error(), want)
	}
	if got := strings.Join(calls, " "); got != "ref:a code:dobry code:zły" {
		t.Errorf("calls: %s", got)
	}

	out, err = conv.Render([]byte("tekst\n*x [[zły]]* [[c]]\n"))
	if out != nil || !errors.As(err, &e) || e.Code != ErrReference || e.Line != 2 || e.Column != 4 || e.Detail != "zły" {
		t.Errorf("reference error: output %q, error %+v", out, err)
	}
	if want := `gomdgo: line 2, column 4: resolver for reference "zły" failed: boom`; err.Error() != want {
		t.Errorf("message: got %q, want %q", err.Error(), want)
	}
	if out, err := conv.Render([]byte("[[a]]\n\n```d\ndobry\n```")); err != nil || string(out) != "<p>a</p>\nok\n" {
		t.Errorf("after an error the converter still works: %q, %v", out, err)
	}
}

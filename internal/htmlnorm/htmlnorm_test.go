package htmlnorm

import "testing"

func TestNormalizeEqual(t *testing.T) {
	pairs := [][2]string{
		{"<p>a</p>\n", "<p>a</p>"},
		{"<ul>\n<li>a</li>\n</ul>\n", "<ul><li>a</li></ul>"},
		{"<p>a<br />\nb</p>", "<p>a<br>b</p>"},
		{"<hr />", "<hr>"},
		{`<a title="t" href="/u">x</a>`, `<A HREF='/u' title="t">x</a>`},
		{`<input disabled="" type="checkbox">`, `<input type="checkbox" disabled />`},
		{"<p>&quot;a&quot; &#39;b&#39; &ouml;</p>", "<p>\"a\" 'b' ö</p>"},
		{"<p>a  \n b</p>", "<p>a b</p>"},
		{"<p><em>a</em> <strong>b</strong></p>", "<p><em>a</em>\n<strong>b</strong></p>"},
		{`<a href="/u?a=1&amp;b=2">x</a>`, `<a href="/u?a=1&b=2">x</a>`},
		{"<p>a &lt; b &amp; c</p>", "<p>a &#60; b &#38; c</p>"},
		{"<li>\n<p>a</p>\n</li>", "<li><p>a</p></li>"},
	}
	for _, p := range pairs {
		if a, b := Normalize(p[0]), Normalize(p[1]); a != b {
			t.Errorf("expected equal:\n%q -> %q\n%q -> %q", p[0], a, p[1], b)
		}
	}
}

func TestNormalizeDifferent(t *testing.T) {
	pairs := [][2]string{
		{"<p>a</p>", "<p>b</p>"},
		{"<p><em>a</em> <strong>b</strong></p>", "<p><em>a</em><strong>b</strong></p>"},
		{"<p>a b</p>", "<p>ab</p>"},
		{"<pre><code>a\n  b\n</code></pre>", "<pre><code>a\n b\n</code></pre>"},
		{"<pre><code>a\n</code></pre>", "<pre><code>a</code></pre>"},
		{`<a href="/a%20b">x</a>`, `<a href="/a b">x</a>`},
		{"<p>&lt;b&gt;</p>", "<p><b></p>"},
		{"<p>a</p>", "<h1>a</h1>"},
		{"<!-- a -->", "<!-- b -->"},
		{"<p>a <em>b</em></p>", "<p>a<em>b</em></p>"},
	}
	for _, p := range pairs {
		if a, b := Normalize(p[0]), Normalize(p[1]); a == b {
			t.Errorf("expected different, both give %q:\n%q\n%q", a, p[0], p[1])
		}
	}
}

func TestNormalizeIdempotent(t *testing.T) {
	for _, s := range []string{
		"<ul>\n<li>a <em>b</em>\n</li>\n</ul>\n", "a < b", "<p", "<pre>\n x\n</pre>", `<a href='"'>x</a>`,
	} {
		once := Normalize(s)
		if twice := Normalize(once); once != twice {
			t.Errorf("%q: %q then %q", s, once, twice)
		}
	}
}

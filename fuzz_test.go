package gomdgo

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

var fuzzStrict = New(Config{RawHTML: RawHTMLOmit, Schemes: []string{}, HeadingIDs: true,
	SingleTilde: true, TableAlignStyle: true})

// fuzzCommonMark: czysty CommonMark przy ustawieniach bezpiecznych.
var fuzzCommonMark = New(Config{Disable: GFM})

// REQ: DET-001 SAFE-005
func FuzzRender(f *testing.F) {
	for _, path := range []string{commonMarkSpec, gfmSpec} {
		for _, e := range loadSpec(f, path) {
			f.Add([]byte(e.Markdown))
		}
	}
	f.Add([]byte{})
	f.Add([]byte("\x00\xff<a href=\"x\">&amp;\r\n"))
	f.Fuzz(func(t *testing.T, src []byte) {
		out := Render(src)
		if err := checkSafe(out); err != nil {
			t.Fatalf("unsafe or malformed output: %v\nsource: %q\noutput: %q", err, src, out)
		}
		plain, _ := fuzzCommonMark.Render(src)
		if err := checkSafe(plain); err != nil {
			t.Fatalf("unsafe or malformed output (extensions disabled): %v\nsource: %q\noutput: %q", err, src, plain)
		}
		if raw := renderWith(gfmSpecConverter, string(src)); !utf8.Valid(raw) || bytes.IndexByte(raw, 0) >= 0 {
			t.Fatalf("raw HTML output with GFM is not valid UTF-8 or contains U+0000\nsource: %q\noutput: %q", src, raw)
		}
		strict, _ := fuzzStrict.Render(src)
		if err := checkSafe(strict); err != nil {
			t.Fatalf("unsafe or malformed output (raw HTML omitted, heading ids): %v\nsource: %q\noutput: %q", err, src, strict)
		}
		if !utf8.Valid(out) || bytes.IndexByte(out, 0) >= 0 {
			t.Fatalf("output is not valid UTF-8 or contains U+0000\nsource: %q\noutput: %q", src, out)
		}
		// Z przepuszczaniem surowego HTML wynik nie musi być poprawnym fragmentem, ale kodowanie musi.
		if raw := renderSpec(string(src)); !utf8.Valid(raw) || bytes.IndexByte(raw, 0) >= 0 {
			t.Fatalf("raw HTML output is not valid UTF-8 or contains U+0000\nsource: %q\noutput: %q", src, raw)
		}
		if again := Render(src); !bytes.Equal(out, again) {
			t.Fatalf("output differs between calls\nsource: %q", src)
		}
	})
}

// REQ: DET-002
func TestDeterministic(t *testing.T) {
	for _, path := range []string{commonMarkSpec, gfmSpec} {
		for _, e := range loadSpec(t, path) {
			first := Render([]byte(e.Markdown))
			for i := 0; i < 2; i++ {
				if again := Render([]byte(e.Markdown)); !bytes.Equal(first, again) {
					t.Errorf("%s example %d: output differs between calls", path, e.Number)
				}
			}
		}
	}
}

// REQ: DET-001
func TestCheckFragment(t *testing.T) {
	good := []string{
		"", "<p>a &amp; b &#35; &#x1F; &copy;</p>\n", "<ul>\n<li>a<br />b</li>\n</ul>\n",
		`<p><a href="/x?a=1&amp;b=2" title="t">x</a><img src="i.png" alt=""></p>`, "<hr />\n",
	}
	for _, s := range good {
		if err := checkFragment([]byte(s)); err != nil {
			t.Errorf("%q: unexpected error: %v", s, err)
		}
	}
	bad := []string{
		"<p>a", "a</p>", "<p><em>a</p></em>", "a < b", "a & b", "<p", `<a href="x>y</a>`, "<p>&;</p>", "<a href=x>y</a>",
	}
	for _, s := range bad {
		if err := checkFragment([]byte(s)); err == nil {
			t.Errorf("%q: expected an error", s)
		}
	}
}

var voidElements = map[string]bool{"br": true, "hr": true, "img": true, "input": true}

// checkFragment sprawdza, czy wynik renderera jest poprawnym fragmentem HTML w sensie DET-001. To sprawdzian
// naszego własnego wyniku, a nie parser HTML: przyjmuje tylko zapis, który renderer ma prawo wytworzyć
// (atrybuty w cudzysłowach, znaczniki domknięte jawnie, elementy puste z listy).
func checkFragment(out []byte) error {
	_, err := scanFragment(out)
	return err
}

// tagInfo to znacznik otwierający znaleziony w wyniku, z atrybutami w kolejności zapisu.
type tagInfo struct {
	name  string
	attrs [][2]string
}

// scanFragment sprawdza fragment jak checkFragment i zwraca jego znaczniki otwierające.
func scanFragment(out []byte) ([]tagInfo, error) {
	var tags []tagInfo
	var stack []string
	s := string(out)
	for i := 0; i < len(s); {
		switch s[i] {
		case '&':
			end := strings.IndexByte(s[i:], ';')
			if end < 2 || end > 34 || !isReference(s[i+1:i+end]) {
				return nil, fmt.Errorf("offset %d: bare ampersand", i)
			}
			i += end + 1
		case '<':
			end, err := checkTag(s, i, &stack, &tags)
			if err != nil {
				return nil, err
			}
			i = end
		default:
			i++
		}
	}
	if len(stack) > 0 {
		return nil, fmt.Errorf("unclosed elements: %v", stack)
	}
	return tags, nil
}

func isReference(name string) bool {
	if rest, ok := strings.CutPrefix(name, "#"); ok {
		digits := "0123456789"
		if r, hex := strings.CutPrefix(strings.ToLower(rest), "x"); hex {
			rest, digits = r, "0123456789abcdef"
		}
		return rest != "" && strings.Trim(strings.ToLower(rest), digits) == ""
	}
	return isName(name)
}

func isName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		letter := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if !letter && (i == 0 || !(c >= '0' && c <= '9' || c == '-')) {
			return false
		}
	}
	return true
}

// checkTag sprawdza znacznik zaczynający się w s[i] i zwraca pozycję za nim.
func checkTag(s string, i int, stack *[]string, tags *[]tagInfo) (int, error) {
	bad := func(msg string) (int, error) { return 0, fmt.Errorf("offset %d: %s", i, msg) }
	j := i + 1
	closing := j < len(s) && s[j] == '/'
	if closing {
		j++
	}
	start := j
	for j < len(s) && s[j] != ' ' && s[j] != '>' && s[j] != '/' {
		j++
	}
	name := s[start:j]
	if !isName(name) {
		return bad("bare less-than sign")
	}
	if closing {
		if j >= len(s) || s[j] != '>' {
			return bad("malformed closing tag")
		}
		st := *stack
		if len(st) == 0 || st[len(st)-1] != name {
			return bad("closing tag </" + name + "> does not match open elements")
		}
		*stack = st[:len(st)-1]
		return j + 1, nil
	}
	info := tagInfo{name: name}
	defer func() { *tags = append(*tags, info) }()
	for {
		if j >= len(s) {
			return bad("unterminated tag")
		}
		if s[j] == '>' {
			j++
			break
		}
		if strings.HasPrefix(s[j:], " />") || strings.HasPrefix(s[j:], "/>") {
			if !voidElements[name] {
				return bad("self-closing non-void element " + name)
			}
			return j + strings.IndexByte(s[j:], '>') + 1, nil
		}
		if s[j] != ' ' {
			return bad("malformed tag")
		}
		j++
		start = j
		for j < len(s) && s[j] != '=' && s[j] != ' ' && s[j] != '>' {
			j++
		}
		if !isName(s[start:j]) || !strings.HasPrefix(s[j:], `="`) {
			return bad("attribute without a quoted value")
		}
		attrName := s[start:j]
		end := strings.IndexByte(s[j+2:], '"')
		if end < 0 {
			return bad("unterminated attribute value")
		}
		if strings.ContainsAny(s[j+2:j+2+end], "<>") {
			return bad("unescaped angle bracket in attribute value")
		}
		info.attrs = append(info.attrs, [2]string{attrName, s[j+2 : j+2+end]})
		j += 2 + end + 1
	}
	if !voidElements[name] {
		*stack = append(*stack, name)
	}
	return j, nil
}

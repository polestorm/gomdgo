package spectest

import (
	"strings"
	"testing"
)

func TestLoadCommonMark(t *testing.T) {
	ex, err := Load("../../testdata/spec/commonmark-0.31.2.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(ex) != 652 {
		t.Fatalf("got %d examples, want 652", len(ex))
	}
	first := ex[0]
	if first.Section != "Tabs" || first.Markdown != "\tfoo\tbaz\t\tbim\n" ||
		first.HTML != "<pre><code>foo\tbaz\t\tbim\n</code></pre>\n" {
		t.Errorf("unexpected first example: %+v", first)
	}
	last := ex[len(ex)-1]
	if last.Section != "Textual content" || last.Number != 652 {
		t.Errorf("unexpected last example: %+v", last)
	}
	for _, e := range ex {
		if strings.Contains(e.Markdown, "→") || strings.Contains(e.HTML, "→") {
			t.Errorf("example %d still contains the tab marker", e.Number)
		}
		if e.IsExtension() || e.Extension != "" {
			t.Errorf("example %d: CommonMark has no extension examples", e.Number)
		}
	}
}

func TestLoadGFM(t *testing.T) {
	ex, err := Load("../../testdata/spec/gfm-0.29.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(ex) != 673 {
		t.Fatalf("got %d examples, want 673", len(ex))
	}
	bySection := map[string]int{}
	for _, e := range ex {
		if e.IsExtension() {
			bySection[e.Section]++
		}
	}
	want := map[string]int{
		"Tables (extension)":              8,
		"Task list items (extension)":     2,
		"Strikethrough (extension)":       2,
		"Autolinks (extension)":           11,
		"Disallowed Raw HTML (extension)": 1,
	}
	for s, n := range want {
		if bySection[s] != n {
			t.Errorf("section %q: got %d examples, want %d", s, bySection[s], n)
		}
	}
	if len(bySection) != len(want) {
		t.Errorf("extension sections: got %v", bySection)
	}
}

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/polestorm/gomdgo/internal/cli"
)

const doc = "# T\n\n<script>x</script>\n\n```mermaid\nflowchart LR\n  A[Start] --> B[End]\n```\n\n```go\nx := 1\n```\n"

// TestDiagram sprawdza, że blok mermaid wychodzi jako SVG w <figure class="diagram">, inne bloki kodu zostają
// blokami kodu, a surowy HTML dokumentu nadal jest zapisany jako tekst.
func TestDiagram(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := cli.Run([]string{"-standalone"}, strings.NewReader(doc), &out, &errOut, options()); code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, errOut.String())
	}
	got := out.String()
	for _, want := range []string{
		"<figure class=\"diagram\">\n<svg ", "</svg>\n</figure>\n",
		"&lt;script&gt;x&lt;/script&gt;", "<pre><code class=\"language-go\">x := 1\n</code></pre>",
		"<style>figure.diagram{",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "<script>") || strings.Contains(got, "language-mermaid") {
		t.Errorf("unexpected raw script or untouched mermaid block:\n%s", got)
	}
}

// TestDiagramError sprawdza, że błąd w diagramie kończy polecenie kodem 1 i wskazuje wiersz dokumentu.
func TestDiagramError(t *testing.T) {
	var out, errOut bytes.Buffer
	src := "text\n\n```mermaid\nnonsense here\n```\n"
	if code := cli.Run(nil, strings.NewReader(src), &out, &errOut, options()); code != 1 {
		t.Fatalf("exit code %d, want 1", code)
	}
	if out.Len() != 0 || !strings.Contains(errOut.String(), "gomdgo-mermaid: ") || !strings.Contains(errOut.String(), "line 3") {
		t.Errorf("stdout %q, stderr %q", out.String(), errOut.String())
	}
}

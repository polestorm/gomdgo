# gomdgo

Markdown → HTML in pure Go, written from the specifications, with no dependencies outside the standard library.

**Status: CommonMark and the GFM extensions complete, safe by default; extension points next.** All 652 examples
of the CommonMark 0.31.2 specification and all 24 examples of the GFM 0.29 extension sections pass byte for byte.
Raw HTML is escaped and link destinations are checked against a scheme allow-list unless the caller opts out.
The GFM extensions are on by default; `Config.Disable` turns them off one by one (`gomdgo.GFM` for plain
CommonMark). Code blocks
can be rendered by caller-supplied functions per language and `[[target]]` references resolved by a caller-supplied
function. Alerts (`> [!NOTE]`) and footnotes are opt-in through `Config.Enable`. `Document.Walk` gives validators a
read-only view of the document tree with source positions.

## Usage

```go
html := gomdgo.Render(src) // CommonMark + GFM, safe defaults, never fails

conv := gomdgo.New(gomdgo.Config{RawHTML: gomdgo.RawHTMLOmit, HeadingIDs: true})
doc := conv.Parse(src)   // never fails
doc.Walk(func(n gomdgo.Node, entering bool) bool { // a validator inspects the document here
    if entering && n.Kind() == gomdgo.KindImage {
        line, col := n.Position()
        log.Printf("%d:%d: image %s", line, col, n.Destination())
    }
    return true
})
out, err := doc.HTML()   // an error can only come from a caller-supplied extension

conv = gomdgo.New(gomdgo.Config{
    CodeBlocks: map[string]gomdgo.CodeBlockFunc{
        "mermaid": func(b gomdgo.CodeBlock) ([]byte, error) { return renderDiagram(b.Code) },
    },
    References: func(r gomdgo.Reference) (gomdgo.ReferenceLink, error) {
        return gomdgo.ReferenceLink{Href: "/records/" + r.Target}, nil
    },
})
``` See [docs/roadmap.md](docs/roadmap.md) for the stages and
[docs/adr/0001-scope-and-acceptance.md](docs/adr/0001-scope-and-acceptance.md) for the scope and the acceptance
criteria.

## Goals

- **Conformance to the specifications**: CommonMark (all specification examples) and the GitHub Flavored Markdown
  extensions that documents need (tables, strikethrough, task lists, extended autolinks).
- **Safe output by construction**: raw HTML from the source is escaped unless the caller allows it, and link
  destinations are checked against an allow-list of schemes. No separate sanitizer is required.
- **Extension points instead of raw HTML**: a fenced code block with a registered language (for example `mermaid`)
  is rendered by a caller-supplied function; the same holds for inline references and mathematics.
- **Deterministic output** and linear running time on adversarial input.
- **Zero dependencies**: `go.mod` has no `require` section. The comparison with the reference parser lives in a
  separate module under `tools/`.

## Reference

Until gomdgo meets the acceptance criteria, [goldmark](https://github.com/yuin/goldmark) is the parser in use and
the reference for differential tests. gomdgo is written from the specifications, not from goldmark's source code.

## Documentation

The rest of the documentation is in Polish: [roadmap](docs/roadmap.md), [testing](docs/testing.md),
[provenance](docs/provenance.md), [decisions](docs/adr/).

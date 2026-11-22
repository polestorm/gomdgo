# gomdgo

Markdown → HTML in pure Go, written from the specifications, with no dependencies outside the standard library.

**Status: feature complete, hardening in progress.** All 652 examples of the CommonMark 0.31.2 specification and
all 24 examples of the GFM 0.29 extension sections pass byte for byte. Raw HTML is escaped and link destinations
are checked against a scheme allow-list unless the caller opts out. What remains before a first release is the
24-hour fuzzing run on the release candidate and the check against the golden files of the first consumer
(criteria K5 and K8 in [docs/adr/0001-scope-and-acceptance.md](docs/adr/0001-scope-and-acceptance.md)).

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
```

## What it does

- **CommonMark 0.31.2**, complete.
- **GFM extensions**, on by default: tables, strikethrough, task lists, extended autolinks, the tag filter.
  `Config.Disable` turns them off one by one (`gomdgo.GFM` for plain CommonMark).
- **Opt-in extras** through `Config.Enable`: GitHub alerts (`> [!NOTE]`) and footnotes.
- **Safe output by construction**: raw HTML from the source is escaped by default (or omitted, or passed through,
  as the caller chooses), and link destinations must use an allowed scheme (`http`, `https`, `mailto`, `tel` by
  default). No separate sanitizer is required. Invalid UTF-8 and U+0000 become U+FFFD before parsing.
- **Extension points instead of raw HTML**: a fenced code block with a registered language is rendered by a
  caller-supplied function, and `[[target]]` references are resolved by a caller-supplied function. Errors carry
  a stable code, line and column.
- **A read-only document tree** for validators: `Document.Walk`, with the source line and column of every node.
- **Deterministic output**, byte for byte, and **linear running time** on adversarial input (92 pathological
  cases are timed in the test suite).
- **Heading identifiers** on request.

What it deliberately does not do: definition lists, typographic replacements, attribute syntax (`{#id .class}`),
hard line breaks as an option, XHTML output, caller-defined syntax, and modification of the document tree.

## Comparison with goldmark

[goldmark](https://github.com/yuin/goldmark) is the reference for the differential tests. gomdgo is written from
the specifications, not from goldmark's source code ([docs/provenance.md](docs/provenance.md)); goldmark is run
as a black box only. The numbers below compare gomdgo at commit `3acafeb` with goldmark v1.8.7, Go 1.26.6,
on an Apple M-series machine, measured on 2026-11-21 and 2026-11-22. They are reproduced by `tools/compare`
([docs/testing.md](docs/testing.md)).

### Size

| | gomdgo | goldmark | Ratio |
|---|---:|---:|---:|
| Library code without tests | 3,561 lines in 14 files | 13,293 lines in 53 files | 0.27 |
| Packages | 1 | 11 | |
| Added to a binary (a "stdin → HTML" program with GFM, `-ldflags="-s -w"`, macOS arm64) | 466 kB | 1,289 kB | 0.36 |

### Speed

`BenchmarkCorpus` in `tools/compare`, median of three runs. A fuzzing job was running in the background during
the measurement, so the ratio is meaningful and the absolute times are overstated.

| Input | gomdgo | goldmark | Time ratio | Memory, gomdgo / goldmark |
|---|---:|---:|---:|---:|
| 652 CommonMark examples, one by one | 1.20 ms | 2.85 ms | 0.42 | 1.5 MB / 4.4 MB |
| 7 files of the project corpus, one by one | 0.23 ms | 0.39 ms | 0.58 | 0.29 MB / 0.26 MB |
| one 258 kB document | 13.8 ms | 19.3 ms | 0.72 | 13.5 MB / 10.5 MB |
| 123 real documents, 948 kB | 20.1 ms | 30.9 ms | 0.65 | 20.6 MB / 17.8 MB |

gomdgo is 1.4 to 2.4 times faster and uses 16–29% more memory on large documents.

### Dependencies

Neither library has dependencies outside the Go standard library.

### Features

| Feature | gomdgo | goldmark |
|---|---|---|
| CommonMark 0.31.2 examples | 652 / 652 | 652 / 652 |
| GFM tables, strikethrough, task lists, autolinks | yes, on by default | yes, as an extension to enable |
| GFM tag filter (`<title>`, `<script>`, …) | yes | no |
| Footnotes | yes, opt-in | yes, as an extension |
| GitHub alerts (`> [!NOTE]`) | yes, opt-in | not in the package; third-party extensions |
| `[[target]]` references | yes, resolved by the caller's function | not in the package |
| Custom rendering of a code block by language | yes, one function per language | yes, through a custom node renderer |
| Definition lists, typographer, attributes, CJK options | no | yes |
| Hard line breaks, XHTML output | no | yes, as options |
| Heading identifiers | yes | yes |
| Raw HTML by default | escaped (or omitted, or passed through) | omitted, with a comment in the output |
| Link destinations | allow-list of schemes | block-list of schemes |
| Document tree | read-only, through `Walk` | public and mutable, with transformers |
| Caller-defined syntax | no, two extension points only | yes, custom parsers and renderers |
| Source positions | line and column of every node | byte segments of lines |
| Errors with a stable code, line and column | yes | not in this form |

### Agreement of the output

After normalization of the HTML, gomdgo and goldmark give identical output for 681 of 683 inputs (676
specification examples and 7 corpus files) and for 122 of 123 real documents. Every difference is explained in
[docs/differences.md](docs/differences.md); where goldmark differs from a specification, gomdgo follows the
specification.

### What the numbers do not show

- **Maturity.** goldmark has been in wide use since 2019. gomdgo is new and has no production users yet.
- **Ecosystem.** goldmark has third-party extensions (syntax highlighting, emoji, metadata and others). gomdgo
  has none, by design.
- **Robustness.** gomdgo has been fuzzed for 9 h 40 min (about 3.6 billion executions) without a failure, with
  a check that the output is well-formed and safe. No such measurement was made for goldmark, so this is not
  a claim that goldmark is weaker here.

## Goals

- **Conformance to the specifications**: CommonMark (all specification examples) and the GitHub Flavored Markdown
  extensions that documents need.
- **Safe output by construction**, with no separate sanitizer.
- **Extension points instead of raw HTML.**
- **Deterministic output** and linear running time on adversarial input.
- **Zero dependencies**: `go.mod` has no `require` section. The comparison with the reference parser lives in a
  separate module under `tools/`.
- **Small code** that can be read and validated: every supported syntax element has a requirement in
  `requirements/` and a test that names it.

## Documentation

The rest of the documentation is in Polish: [roadmap](docs/roadmap.md), [testing](docs/testing.md),
[provenance](docs/provenance.md), [differences from the reference](docs/differences.md),
[decisions](docs/adr/), [requirements](requirements/).

## Licence

Apache License 2.0 (`LICENSE`), Copyright 2026 Michał Głębowski (`NOTICE`). gomdgo contains no code by other
authors. The specification texts in `testdata/spec/` are test data under the Creative Commons
Attribution-ShareAlike 4.0 licence and are not covered by the licence of gomdgo; their authors and sources are
listed in [testdata/spec/README.md](testdata/spec/README.md).

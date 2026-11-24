# gomdgo

Markdown → HTML in pure Go, written from the specifications, with no dependencies outside the standard library.

**Status: feature complete, hardening in progress.** All 652 examples of the CommonMark 0.31.2 specification and
all 24 examples of the GFM 0.29 extension sections pass byte for byte. Raw HTML is escaped and link destinations
are checked against a scheme allow-list unless the caller opts out. The current library code has passed a
24-hour fuzzing run without a failure. What remains before a first release is the check against the golden files
of the first consumer (criterion K8 in [docs/adr/0001-scope-and-acceptance.md](docs/adr/0001-scope-and-acceptance.md)).

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

## Command line

```sh
go install github.com/polestorm/gomdgo/cmd/gomdgo@latest

gomdgo README.md > README.html           # a fragment, safe defaults
gomdgo -standalone -title "Read me" -heading-ids -o README.html README.md
cat notes.md | gomdgo -alerts -footnotes
```

The flags mirror `Config`: `-raw-html escape|omit|pass`, `-schemes`, `-any-scheme`, `-commonmark`, `-alerts`,
`-footnotes`, `-heading-ids`, `-single-tilde`, `-table-align-style`. The command does not run other programs and
has no extension points; a `mermaid` code block comes out as an ordinary code block. The stripped binary is
about 2.0 MB on macOS arm64.

### With diagrams: `gomdgo-mermaid`

The same command with one addition: a fenced code block with the language `mermaid` is drawn by
[gomermaidgo](https://github.com/polestorm/gomermaidgo) and written as inline SVG. It lives in its own module
inside this repository, so the library itself stays free of dependencies. Build it from a clone:

```sh
cd cmd/gomdgo-mermaid
go build -ldflags="-s -w" .              # about 6.3 MB on macOS arm64
./gomdgo-mermaid -standalone -o doc.html doc.md
```

The diagram is rendered in the same process, through `Config.CodeBlocks`, so raw HTML in the document is still
escaped. A diagram that cannot be rendered stops the command with exit code 1 and the line of the block. Go has
no optional dependencies, which is why this is a second command and not a build tag
([docs/adr/0015-combined-binary.md](docs/adr/0015-combined-binary.md)). To wire in another renderer, write a
`main.go` of your own against the public API, as in the usage example above.

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
on an Apple M-series machine, measured on 2026-11-22. They are reproduced by `tools/compare`
([docs/testing.md](docs/testing.md)).

The "gomdgo advantage" column is `(goldmark − gomdgo) / goldmark × 100%` for measures where less is better.
A positive value means gomdgo does better, a negative value means it does worse.

### Size

| Measure | gomdgo | goldmark | gomdgo advantage |
|---|---:|---:|---:|
| Lines of library code, without tests | 3,561 | 13,293 | +73.2% |
| Bytes of library code, without tests | 98,765 | 461,450 | +78.6% |
| Source files | 14 | 53 | +73.6% |
| Added to a binary (a "stdin → HTML" program with GFM, `-ldflags="-s -w"`, macOS arm64) | 466 kB | 1,289 kB | +63.8% |

### Speed, memory and output size

`BenchmarkCorpus` in `tools/compare`, median of five runs, measured on 2026-11-22.

| Input | Measure | gomdgo | goldmark | gomdgo advantage |
|---|---|---:|---:|---:|
| 652 CommonMark examples, one by one | time | 0.52 ms | 1.24 ms | **+58.0%** |
| | memory | 1.53 MB | 4.42 MB | +65.4% |
| | allocations | 12,742 | 18,450 | +30.9% |
| | output size | 27,572 B | 27,444 B | −0.5% |
| 7 files of the project corpus, one by one | time | 0.091 ms | 0.180 ms | **+49.6%** |
| | memory | 0.290 MB | 0.265 MB | −9.2% |
| | allocations | 1,550 | 1,753 | +11.6% |
| | output size | 8,725 B | 8,888 B | +1.8% |
| one 258 kB document | time | 4.98 ms | 8.44 ms | **+40.9%** |
| | memory | 13.45 MB | 10.52 MB | −27.9% |
| | allocations | 65,898 | 74,952 | +12.1% |
| | output size | 395,400 B | 403,550 B | +2.0% |
| 123 real documents, about 1 MB | time | 9.36 ms | 15.91 ms | **+41.2%** |
| | memory | 21.58 MB | 18.58 MB | −16.1% |
| | allocations | 78,223 | 101,696 | +23.1% |
| | output size | 1,202,562 B | 1,204,674 B | +0.2% |

gomdgo takes 41–58% less time. Across three runs on the same day the time advantage stayed within 58–61%,
48–51%, 39–43% and 41–42% for the four inputs. Its weak point is memory on large documents, where it uses
16–28% more. The output is practically the same size.

### Other Go libraries

Two more Markdown libraries are in common use in Go: [gomarkdown/markdown](https://github.com/gomarkdown/markdown)
and its ancestor [blackfriday](https://github.com/russross/blackfriday) v2. Measured on 2026-11-24 with
`TestRivalsSpec` and `BenchmarkRivals` in `tools/compare` (goldmark v1.8.7, gomarkdown of 2026-11-23,
blackfriday v2.1.0; median of five runs).

| Measure | gomdgo | goldmark | gomarkdown | blackfriday |
|---|---:|---:|---:|---:|
| CommonMark 0.31.2 examples passed | **652 / 652** | 652 / 652 | 380 / 652 | 332 / 652 |
| `<script>` in the source, by default | escaped | omitted | passed through | passed through |
| `javascript:` link, by default | removed | removed | passed through | passed through |
| Lines of code compiled into a program | **3,561** | 12,850 | 10,719 | 7,911 |
| Added to a binary | **466 kB** | 1,289 kB | 1,161 kB | 847 kB |
| Dependencies in `go.mod` | none | none | 2 | none |
| Time, 652 CommonMark examples | **0.49 ms** | 1.31 ms | 1.51 ms | 1.07 ms |
| Time, 7 corpus files | **0.095 ms** | 0.188 ms | 0.151 ms | 0.109 ms |
| Time, one 260 kB document | **4.11 ms** | 7.59 ms | 5.89 ms | 4.73 ms |
| Memory, one 260 kB document | 13.05 MB | 10.23 MB | **6.96 MB** | 13.50 MB |

gomdgo advantage in time:

| Input | over goldmark | over gomarkdown | over blackfriday |
|---|---:|---:|---:|
| 652 CommonMark examples | +62.8% | +67.7% | +54.4% |
| 7 corpus files | +49.6% | +37.4% | +13.6% |
| one 260 kB document | +45.8% | +30.2% | +13.1% |

gomdgo advantage in memory on the 260 kB document: −27.5% against goldmark, −87.5% against gomarkdown, +3.3%
against blackfriday.

Notes:

- For gomarkdown and blackfriday the table gives the better of two settings (with their common extensions); with
  no extensions they pass 353 and 307 examples. They implement an older Markdown dialect, not CommonMark, so
  their output differs in list nesting, task lists and autolinks.
- The two dependencies of gomarkdown are not compiled into a program that uses only its parser and HTML renderer.
- The line count here covers only the packages such a program pulls in, which is why goldmark shows 12,850 lines
  and not the 13,293 of the whole library in the table above.
- The 260 kB document is not the one of the previous section. That one contains an unclosed code fence inside
  a block quote, followed by a line outside the quote. CommonMark ends the code block with the quote; gomarkdown
  searches for the closing fence beyond the quote and, when it finds one anywhere later, turns everything in
  between into code. On that document it rendered half of the content as code and looked 20% faster than gomdgo.
  The details and the minimal input are in [docs/testing.md](docs/testing.md).

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
- **Robustness.** the current gomdgo code has been fuzzed for 24 hours in one run (1.34 billion executions, after
  9 h 40 min of earlier runs) without a failure, with
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

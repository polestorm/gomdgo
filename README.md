# gomdgo

Converts Markdown (CommonMark with the GitHub Flavored Markdown extensions) to HTML in pure Go, written from the
specifications, with no dependencies outside the Go standard library. The output is safe by default, and the
same input always gives the same HTML, byte for byte.

```go
import "github.com/polestorm/gomdgo/markdown"

html := markdown.Render(src)
```

## Why gomdgo

Compared with three other Markdown libraries in Go, [goldmark](https://github.com/yuin/goldmark),
[gomarkdown](https://github.com/gomarkdown/markdown) and [blackfriday](https://github.com/russross/blackfriday)
(details in [Comparison](#comparison)):

- **Follows the specifications**: all 652 examples of CommonMark 0.31.2 and all 24 examples of the GFM 0.29
  extensions pass byte for byte.
- **Safe without a sanitizer**: raw HTML from the source is written as text and link destinations are checked
  against an allow-list of schemes, unless the caller opts out.
- **The fastest of the four**: 24–69% less time than the others on the three test inputs.
- **The smallest**: 3,586 lines of library code, 466 kB added to a program, no dependencies.
- **Made to be built on**: a function per code block language (diagrams, highlighting), a function for
  `[[target]]` references, errors with a stable code, line and column, and a read-only document tree with source
  positions for validators.
- **Linear time on hostile input**: 92 pathological cases are timed in the test suite, and the library is fuzzed.

Where it is weaker: gomarkdown uses 16% less memory on a large document. gomdgo is new and has no ecosystem of
extensions; definition lists, typographic replacements, attribute syntax and math are deliberately left out.

## Contents

- [Why gomdgo](#why-gomdgo) · [Comparison](#comparison) · [Status](#status) · [Documentation](#documentation) ·
  [How the comparison was measured](#how-the-comparison-was-measured) · [License](#license)
- [Usage](docs/usage.md): options, extension points, errors, the document tree, command line, diagrams

## Comparison

goldmark[^ref] is today's usual choice in Go and, like gomdgo, implements CommonMark. gomarkdown and its ancestor
blackfriday implement an older Markdown dialect.

| | gomdgo | goldmark 1.8.7 | gomarkdown[^gmd] | blackfriday 2.1.0 |
|---|:---:|:---:|:---:|:---:|
| CommonMark 0.31.2 examples[^spec] | 652 of 652 🥇 | 652 of 652 🥇 | 380 of 652 | 332 of 652 |
| Dependencies[^deps] | 0 🥇 | 0 🥇 | 2 | 0 🥇 |
| Lines of code[^size] | 3,586 🥇 | 12,850 | 10,719 | 7,911 |
| Added to a program | 466 kB 🥇 | 1,289 kB | 1,161 kB | 847 kB |
| Time, 652 small documents[^bench] | 0.46 ms 🥇 | 1.30 ms | 1.49 ms | 1.06 ms |
| Time, 7 documents | 0.082 ms 🥇 | 0.188 ms | 0.150 ms | 0.107 ms |
| Time, one 260 kB document | 3.34 ms 🥇 | 7.47 ms | 5.81 ms | 4.49 ms |
| Memory, 652 small documents | 1.18 MB 🥇 | 4.43 MB | 5.31 MB | 5.05 MB |
| Memory, 7 documents | 0.195 MB 🥇 | 0.273 MB | 0.207 MB | 0.345 MB |
| Memory, one 260 kB document | 8.05 MB | 10.25 MB | 6.96 MB 🥇 | 13.50 MB |
| `<script>` in the source, by default[^safe] | written as text | left out | passed through | passed through |
| `javascript:` link, by default | removed | removed | passed through | passed through |
| Tables, strikethrough | yes, on by default | yes, an extension | yes | yes |
| Task lists, bare `www.` links[^feat] | yes, on by default | yes, an extension | no | no |
| Footnotes | yes, opt-in | yes, an extension | yes | yes |
| GitHub alerts (`> [!NOTE]`) | yes, opt-in | no | no | no |
| Heading identifiers | yes | yes | yes | yes |
| Definition lists, typographic replacements | no | yes | yes | yes |
| Attribute syntax (`{#id .class}`) | no | yes | not in our probe | no |
| Math | no | no | yes | no |
| Code block rendered by the caller, per language | yes, one function per language | through a custom renderer | through a render hook | through a custom renderer |
| Caller-defined syntax | no | yes | yes | no |
| Document tree | read-only, with line and column | public and mutable | public and mutable | public and mutable |
| Errors with a code, line and column | yes | no | no | no |
| Deterministic output | yes, documented and tested | not documented | not documented | not documented |
| License | Apache 2.0 | MIT | BSD 2-Clause | BSD 2-Clause |
| GitHub popularity[^pop] | new | 5,114 stars | 1,741 stars | 5,598 stars |

[^ref]: goldmark is the reference parser this library is compared with throughout its tests. It is only run as
    a program; gomdgo is written from the specifications, not from the source code of other parsers
    ([docs/provenance.md](docs/provenance.md)).
[^gmd]: gomarkdown has no releases; the version measured is the commit of 2026-11-23.
[^spec]: Examples whose output equals the specification's after HTML normalization. For gomarkdown and
    blackfriday the better of two settings is shown (with their common extensions); with no extensions they pass
    353 and 307.
[^deps]: Modules required in `go.mod`. The two of gomarkdown are not compiled into a program that uses only its
    parser and HTML renderer.
[^size]: Lines of the packages that a "standard input to HTML" program pulls in, without tests. "Added to
    a program" is the size of that program minus the same program without a parser (`-ldflags="-s -w"`, macOS
    arm64).
[^bench]: The 652 CommonMark examples one by one, the 7 files of this project's corpus one by one, and one
    260 kB document. Median of five runs of `BenchmarkRivals`; memory is the number of bytes allocated per run.
    Measured on 2026-11-24 on one Apple-silicon laptop. 🥇 marks the best value in a row.
[^safe]: Output of each library with its default settings for `<script>x</script>` and
    `[x](javascript:alert(1))`.
[^feat]: The feature rows are the result of running each library with its extensions enabled on a small input
    for each feature; "no" means the feature did not appear in the output. The rows on rendering hooks,
    caller-defined syntax and the document tree come from each library's documentation and were not tested.
[^pop]: Stars on 2026-11-24. The last release of blackfriday is from 2020.

## Status

Version 0.1.0: feature complete, hardening in progress. The API may still change before 1.0.

- CommonMark 0.31.2: 652 of 652 examples; GFM 0.29 extensions: 24 of 24.
- Agreement with goldmark after HTML normalization: 681 of 683 inputs (specification examples and the project
  corpus) and 122 of 123 real documents; every difference is explained in
  [docs/differences.md](docs/differences.md). Where goldmark differs from a specification, gomdgo follows the
  specification.
- Fuzzing: 33 h 40 min in total (about 5 billion executions) without a failure, with a check that the output is
  well-formed and safe. 24 hours of it ran on the code before the last change to the library (a reduction of
  memory use on 2026-11-24); the 24-hour run on the current code is in progress.
- Not yet done: the check against the documents of a first consumer.

## Documentation

- [`docs/usage.md`](docs/usage.md): how to use the library and the commands.
- In Polish: [`docs/provenance.md`](docs/provenance.md) (sources and clean-room rules),
  [`docs/testing.md`](docs/testing.md) (tests, measurements, comparison with other libraries),
  [`docs/differences.md`](docs/differences.md) (differences from the reference parser),
  [`requirements/`](requirements/) (numbered requirements), [`docs/adr/`](docs/adr/) (decisions),
  [`docs/roadmap.md`](docs/roadmap.md).

## How the comparison was measured

All four libraries are run through their public API on the same inputs, by `TestRivalsSpec` and `BenchmarkRivals`
in `tools/compare` (a separate module, so the library itself has no dependencies). On the specification examples
each library runs without extensions; on the other inputs with its GFM or "common" extensions. gomdgo and goldmark
are set to pass raw HTML through, so that all four do the same work.

The 260 kB document is the project corpus joined 60 times, without two files that contain an unclosed code fence.
One of them has the fence inside a block quote, followed by a line outside the quote. CommonMark ends the code
block with the quote; gomarkdown searches for the closing fence beyond the quote and, when it finds one anywhere
later, turns everything in between into code. On a document with that file it rendered half of the content as
code, which made it look faster than it is. The minimal input is in [docs/testing.md](docs/testing.md).

gomarkdown and blackfriday are not CommonMark parsers, so their output differs from the other two in list
nesting, task lists and autolinks; the times compare the work each library does on the same text, not identical
results.

## License

Apache License 2.0, see `LICENSE` and `NOTICE`. gomdgo contains no code by other authors. The specification texts
in `testdata/spec/` are test data under the Creative Commons Attribution-ShareAlike 4.0 licence and are not
covered by the licence of gomdgo; their authors and sources are listed in
[testdata/spec/README.md](testdata/spec/README.md).

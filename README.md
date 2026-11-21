# gomdgo

Markdown → HTML in pure Go, written from the specifications, with no dependencies outside the standard library.

**Status: planning.** No parser code yet. See [docs/roadmap.md](docs/roadmap.md) for the stages and
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

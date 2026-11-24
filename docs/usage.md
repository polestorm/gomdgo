# Using gomdgo

## Library

```go
import "github.com/polestorm/gomdgo/markdown"

html := markdown.Render(src) // CommonMark + GFM, safe defaults, never fails
```

Settings go into `markdown.Config`; parsing and writing are separate steps, so a validator can inspect the
document in between:

```go
conv := markdown.New(markdown.Config{RawHTML: markdown.RawHTMLOmit, HeadingIDs: true})
doc := conv.Parse(src) // never fails
doc.Walk(func(n markdown.Node, entering bool) bool {
    if entering && n.Kind() == markdown.KindImage {
        line, col := n.Position()
        log.Printf("%d:%d: image %s", line, col, n.Destination())
    }
    return true
})
out, err := doc.HTML() // an error can only come from a caller-supplied extension
```

### What is on by default

- **CommonMark 0.31.2**, complete.
- **GFM extensions**: tables, strikethrough (`~~text~~`), task lists, extended autolinks and the tag filter.
  `Config.Disable` turns them off one by one; `markdown.GFM` turns all of them off for plain CommonMark.
- **Safe output**: raw HTML from the source is written as text, and link destinations must use an allowed scheme
  (`http`, `https`, `mailto`, `tel`). Invalid UTF-8 and U+0000 become U+FFFD before parsing.

### Options

| Field of `Config` | Effect |
|---|---|
| `RawHTML` | `RawHTMLEscape` (default), `RawHTMLOmit` or `RawHTMLPass` |
| `Schemes`, `AnyScheme` | the allow-list of link schemes, or no check at all |
| `Disable` | GFM extensions to turn off |
| `Enable` | opt-in extras: `Alerts` (`> [!NOTE]`) and `Footnotes` |
| `HeadingIDs` | `id` attributes on headings |
| `SingleTilde` | `~text~` is struck through as well |
| `TableAlignStyle` | table alignment as a `style` attribute instead of `align` |
| `CodeBlocks` | a function per code block language, see below |
| `References` | a function that resolves `[[target]]` references, see below |

### Extension points

Instead of raw HTML, the caller plugs in functions. A fenced code block with a registered language is rendered by
the caller's function; `[[target]]` references are resolved by the caller's function.

```go
conv := markdown.New(markdown.Config{
    CodeBlocks: map[string]markdown.CodeBlockFunc{
        "mermaid": func(b markdown.CodeBlock) ([]byte, error) { return renderDiagram(b.Code) },
    },
    References: func(r markdown.Reference) (markdown.ReferenceLink, error) {
        return markdown.ReferenceLink{Href: "/records/" + r.Target}, nil
    },
})
out, err := conv.Render(src)
```

When such a function fails, `err` is a `*markdown.Error` with a stable `Code` (`code-block-renderer-failed`,
`reference-resolver-failed`), the `Line` and `Column` of the block or reference in the source, and the original
error. No partial output is returned.

### What it deliberately does not do

Definition lists, typographic replacements, attribute syntax (`{#id .class}`), math, hard line breaks as an
option, XHTML output, caller-defined syntax, and modification of the document tree.

## Command line

```sh
go install github.com/polestorm/gomdgo/cmd/gomdgo@latest

gomdgo README.md > README.html           # a fragment, safe defaults
gomdgo -standalone -title "Read me" -heading-ids -o README.html README.md
cat notes.md | gomdgo -alerts -footnotes
```

The flags mirror `Config`: `-raw-html escape|omit|pass`, `-schemes`, `-any-scheme`, `-commonmark`, `-alerts`,
`-footnotes`, `-heading-ids`, `-single-tilde`, `-table-align-style`. `-standalone` writes a complete HTML
document. The command does not run other programs; a `mermaid` code block comes out as an ordinary code block.
Exit codes: 0 success, 1 read, write or rendering error, 2 wrong arguments.

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
([adr/0015-combined-binary.md](adr/0015-combined-binary.md), in Polish). To wire in another renderer, write
a `main.go` of your own against the library, as in the extension point example above.

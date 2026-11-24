// Package markdown converts Markdown (CommonMark with the GitHub Flavored Markdown extensions) to HTML.
//
// The simplest use is Render(src): CommonMark with the GFM extensions and output that is safe by default.
// Settings go into New(Config); Converter.Parse and Document.HTML separate parsing from writing, so that
// a validator can inspect the document in between with Document.Walk.
//
// Raw HTML from the source is written as text and link destinations are checked against an allow-list of
// schemes unless Config says otherwise. Instead of raw HTML, the caller plugs in functions: one per code
// block language (Config.CodeBlocks) and one for [[target]] references (Config.References). The same input
// and settings always give the same output, byte for byte. The specifications followed are CommonMark
// 0.31.2 and GFM 0.29; usage examples are in docs/usage.md of the repository.
package markdown

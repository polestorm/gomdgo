// Package spectest wczytuje przykłady z tekstu specyfikacji CommonMark i GFM (plik spec.txt).
//
// Format przykładu opisuje sama specyfikacja (rozdział „About this document”): blok ogrodzony 32 odwrotnymi
// apostrofami ze słowem example, źródło Markdown, wiersz z samą kropką, oczekiwany HTML. Znak → oznacza tabulator.
package spectest

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Example to jeden przykład specyfikacji.
type Example struct {
	Number    int    // numer kolejny w pliku, od 1 (taki sam jak w opublikowanej specyfikacji)
	Section   string // tytuł ostatniego nagłówka przed przykładem
	Extension string // słowo po „example” (np. table, autolink, disabled); puste dla zwykłych przykładów
	StartLine int    // wiersz otwierającego ogrodzenia w pliku specyfikacji
	Markdown  string
	HTML      string
}

// IsExtension mówi, czy przykład należy do rozdziału rozszerzenia GFM.
func (e Example) IsExtension() bool {
	return strings.HasSuffix(e.Section, "(extension)")
}

const (
	fence     = "````````````````````````````````"
	openFence = fence + " example"
	endMarker = "<!-- END TESTS -->"
)

// Load wczytuje wszystkie przykłady z pliku specyfikacji.
func Load(path string) ([]Example, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	const (
		outside = iota
		inMarkdown
		inHTML
	)
	var (
		examples []Example
		cur      Example
		md, html strings.Builder
		section  string
		state    = outside
		lineNo   int
	)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		switch state {
		case outside:
			switch {
			case line == endMarker:
				return examples, nil
			case strings.HasPrefix(line, openFence):
				cur = Example{
					Number:    len(examples) + 1,
					Section:   section,
					Extension: strings.TrimSpace(strings.TrimPrefix(line, openFence)),
					StartLine: lineNo,
				}
				md.Reset()
				html.Reset()
				state = inMarkdown
			case strings.HasPrefix(line, "#"):
				section = strings.TrimSpace(strings.TrimLeft(line, "#"))
			}
		case inMarkdown:
			if line == "." {
				state = inHTML
				continue
			}
			md.WriteString(line)
			md.WriteByte('\n')
		case inHTML:
			if line == fence {
				cur.Markdown = strings.ReplaceAll(md.String(), "→", "\t")
				cur.HTML = strings.ReplaceAll(html.String(), "→", "\t")
				examples = append(examples, cur)
				state = outside
				continue
			}
			html.WriteString(line)
			html.WriteByte('\n')
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if state != outside {
		return nil, fmt.Errorf("%s: example opened at line %d is not closed", path, cur.StartLine)
	}
	return examples, nil
}

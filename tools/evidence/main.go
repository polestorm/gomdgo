// Polecenie evidence składa pakiet dowodów (kryterium K9): dla każdego wymagania z requirements/ wskazuje
// testy, które się na nie powołują, i wynik ich ostatniego przebiegu. Uruchamia go test, więc wynik opisuje
// stan drzewa roboczego w chwili wywołania.
//
//	go run ./tools/evidence -o out/evidence.md
//
// Należy do modułu głównego i korzysta tylko z biblioteki standardowej.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

var (
	reqRef   = regexp.MustCompile(`\b[A-Z]{3,5}-\d{3}\b`)
	funcDecl = regexp.MustCompile(`^func ((?:Test|Fuzz)\w+)\(`)
	varDecl  = regexp.MustCompile(`^var (\w+)`)
	passed   = regexp.MustCompile(`passed (\d+) of (\d+)`)
)

// Tabele danych testowych, na które powołują się wymagania, i testy, które z nich korzystają.
var tableTests = map[string]string{
	"sectionRequirements": "TestSpecCommonMark",
	"gfmSections":         "TestSpecGFM",
	"pathologicalInlines": "TestPathological",
	"pathologicalGFM":     "TestPathological",
}

type requirement struct {
	id, text, file string
}

type event struct {
	Action, Package, Test, Output string
}

func main() {
	out := flag.String("o", "out/evidence.md", "plik wynikowy")
	flag.Parse()

	reqs := readRequirements()
	evidence := map[string][]string{} // wymaganie -> testy i pliki korpusu
	add := func(id, what string) {
		for _, e := range evidence[id] {
			if e == what {
				return
			}
		}
		evidence[id] = append(evidence[id], what)
	}
	tests, _ := filepath.Glob("*_test.go")
	internal, _ := filepath.Glob("internal/*/*_test.go")
	for _, path := range append(tests, internal...) {
		lines := strings.Split(read(path), "\n")
		for i, line := range lines {
			if !strings.Contains(line, "REQ:") {
				continue
			}
			target := ""
			// Komentarz na końcu wiersza tabeli należy do tabeli, w której stoi; komentarz w osobnym
			// wierszu do najbliższej deklaracji poniżej.
			from, step := i, 1
			if !strings.HasPrefix(strings.TrimSpace(line), "//") {
				step = -1
			}
			for j := from; j >= 0 && j < len(lines) && target == ""; j += step {
				if m := funcDecl.FindStringSubmatch(lines[j]); m != nil {
					target = m[1]
				} else if m := varDecl.FindStringSubmatch(lines[j]); m != nil {
					target = tableTests[m[1]]
					if target == "" {
						target = "?" + m[1]
					}
				}
			}
			for _, id := range reqRef.FindAllString(line[strings.Index(line, "REQ:"):], -1) {
				add(id, target)
			}
		}
	}
	for _, pattern := range []string{"testdata/corpus/*.md", "testdata/corpus-extras/*.md"} {
		files, _ := filepath.Glob(pattern)
		for _, path := range files {
			first, _, _ := strings.Cut(read(path), "\n")
			for _, id := range reqRef.FindAllString(first, -1) {
				add(id, "TestCorpus")
				add(id, path)
			}
		}
	}

	results, specCounts, failed := runTests()

	var b strings.Builder
	fmt.Fprintf(&b, "# Pakiet dowodów gomdgo\n\n")
	fmt.Fprintf(&b, "Wygenerowany %s przez `go run ./tools/evidence`; commit `%s`%s; %s %s/%s.\n\n",
		time.Now().Format("2006-01-02 15:04"), git("rev-parse", "--short", "HEAD"), dirty(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&b, "Wynik `go test -count=1 ./...`: ")
	if failed == 0 {
		fmt.Fprintf(&b, "**wszystkie testy przeszły** (%d testów).\n\n", len(results))
	} else {
		fmt.Fprintf(&b, "**%d testów nie przeszło** (z %d).\n\n", failed, len(results))
	}
	fmt.Fprintf(&b, "| Kryterium | Wynik |\n|---|---|\n")
	fmt.Fprintf(&b, "| K1, przykłady CommonMark 0.31.2 | %s |\n", specCounts["TestSpecCommonMark"])
	fmt.Fprintf(&b, "| K2, przykłady rozszerzeń GFM 0.29 | %s |\n", specCounts["TestSpecGFM"])
	covered := 0
	for _, r := range reqs {
		if len(evidence[r.id]) > 0 {
			covered++
		}
	}
	fmt.Fprintf(&b, "| K4, wymagania z testem | %d z %d |\n\n", covered, len(reqs))
	fmt.Fprintf(&b, "Kryteria K3 i K7 mierzy `tools/compare` (osobny moduł), K5 fuzzing; ich wyniki są w `docs/testing.md`.\n\n")

	file := ""
	for _, r := range reqs {
		if r.file != file {
			file = r.file
			fmt.Fprintf(&b, "## %s\n\n| Wymaganie | Treść | Dowód | Wynik |\n|---|---|---|---|\n", file)
		}
		var names []string
		status := "przeszedł"
		if len(evidence[r.id]) == 0 {
			status = "**brak testu**"
		}
		for _, e := range evidence[r.id] {
			names = append(names, "`"+e+"`")
			if res, isTest := results[e]; isTest && res != "pass" {
				status = "**nie przeszedł: " + e + "**"
			} else if !isTest && !strings.Contains(e, "/") {
				status = "**nie znaleziono testu " + e + "**"
			}
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", r.id, r.text, strings.Join(names, ", "), status)
		if file != "" && r == lastOf(reqs, file) {
			b.WriteString("\n")
		}
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("%d requirements, %d with evidence, %d tests, %d failed; written to %s\n", len(reqs), covered, len(results), failed, *out)
	if failed > 0 || covered < len(reqs) {
		os.Exit(1)
	}
}

func lastOf(reqs []requirement, file string) requirement {
	var last requirement
	for _, r := range reqs {
		if r.file == file {
			last = r
		}
	}
	return last
}

func readRequirements() []requirement {
	var reqs []requirement
	files, _ := filepath.Glob("requirements/*.md")
	sort.Strings(files)
	for _, path := range files {
		for _, line := range strings.Split(read(path), "\n") {
			id := reqRef.FindString(line)
			if id == "" || !strings.HasPrefix(line, "| "+id+" |") {
				continue
			}
			text := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "| "+id+" |")), "|")
			reqs = append(reqs, requirement{id: id, text: strings.TrimSpace(text), file: path})
		}
	}
	if len(reqs) == 0 {
		fatal(fmt.Errorf("no requirements found; run from the repository root"))
	}
	return reqs
}

// runTests uruchamia testy modułu i zwraca wynik każdego testu najwyższego poziomu, liczby zaliczonych
// przykładów specyfikacji oraz liczbę testów, które nie przeszły.
func runTests() (results map[string]string, specCounts map[string]string, failed int) {
	cmd := exec.Command("go", "test", "-count=1", "-json", "./...")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run() // niezerowy kod wyjścia oznacza nieudane testy; widać je w zdarzeniach
	results, specCounts = map[string]string{}, map[string]string{}
	sc := bufio.NewScanner(&stdout)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<26)
	for sc.Scan() {
		var e event
		if json.Unmarshal(sc.Bytes(), &e) != nil || e.Test == "" || strings.Contains(e.Test, "/") {
			continue
		}
		switch e.Action {
		case "pass", "fail", "skip":
			results[e.Test] = e.Action
			if e.Action == "fail" {
				failed++
			}
		case "output":
			if m := passed.FindStringSubmatch(e.Output); m != nil {
				specCounts[e.Test] = m[1] + " z " + m[2]
			}
		}
	}
	if len(results) == 0 {
		fatal(fmt.Errorf("go test produced no results"))
	}
	return results, specCounts, failed
}

func read(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	return string(data)
}

func git(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "nieznany"
	}
	return strings.TrimSpace(string(out))
}

func dirty() string {
	if git("status", "--porcelain") != "" {
		return " ze zmianami w drzewie roboczym"
	}
	return ""
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "evidence:", err)
	os.Exit(1)
}

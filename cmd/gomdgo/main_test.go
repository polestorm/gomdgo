package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	const src = "# Title\n\n<b>raw</b> ~one~ ~~two~~ [x](javascript:alert(1))\n\n| a |\n|:-:|\n| b |\n\n> [!NOTE]\n> text[^1]\n\n[^1]: note\n"
	cases := []struct {
		name     string
		args     []string
		contains []string
		absent   []string
	}{
		{"defaults", nil,
			[]string{"<h1>Title</h1>", "&lt;b&gt;raw&lt;/b&gt;", "~one~", "<del>two</del>", "<table>", `align="center"`},
			[]string{"javascript:", "<b>", "class=\"alert", "footnotes"}},
		{"raw html omitted", []string{"-raw-html", "omit"}, nil, []string{"raw</b>", "&lt;b&gt;"}},
		{"raw html passed", []string{"-raw-html", "pass"}, []string{"<b>raw</b>"}, nil},
		{"any scheme", []string{"-any-scheme"}, []string{`href="javascript:alert(1)"`}, nil},
		{"schemes", []string{"-schemes", "javascript"}, []string{`href="javascript:alert(1)"`}, nil},
		{"commonmark", []string{"-commonmark"}, []string{"~~two~~"}, []string{"<table>", "<del>"}},
		{"alerts and footnotes", []string{"-alerts", "-footnotes"},
			[]string{`<blockquote class="alert alert-note">`, `<section class="footnotes">`}, nil},
		{"heading ids", []string{"-heading-ids"}, []string{`<h1 id="title">`}, nil},
		{"single tilde", []string{"-single-tilde"}, []string{"<del>one</del>"}, nil},
		{"table align style", []string{"-table-align-style"}, []string{`style="text-align:center"`}, []string{`align="center"`}},
		{"standalone", []string{"-standalone", "-title", "A & B"},
			[]string{"<!DOCTYPE html>", `<meta charset="utf-8">`, "<title>A &amp; B</title>", "</body>\n</html>\n"}, nil},
	}
	for _, c := range cases {
		var out, errOut bytes.Buffer
		if code := run(c.args, strings.NewReader(src), &out, &errOut); code != 0 {
			t.Errorf("%s: exit code %d, stderr %q", c.name, code, errOut.String())
			continue
		}
		for _, want := range c.contains {
			if !strings.Contains(out.String(), want) {
				t.Errorf("%s: output lacks %q:\n%s", c.name, want, out.String())
			}
		}
		for _, bad := range c.absent {
			if strings.Contains(out.String(), bad) {
				t.Errorf("%s: output contains %q:\n%s", c.name, bad, out.String())
			}
		}
	}
}

func TestRunFiles(t *testing.T) {
	dir := t.TempDir()
	in, out := filepath.Join(dir, "in.md"), filepath.Join(dir, "out.html")
	if err := os.WriteFile(in, []byte("*a*\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-o", out, in}, strings.NewReader("ignored"), &stdout, &stderr); code != 0 {
		t.Fatalf("exit code %d, stderr %q", code, stderr.String())
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "<p><em>a</em></p>\n" || stdout.Len() != 0 {
		t.Errorf("file %q, stdout %q", got, stdout.String())
	}
}

func TestRunErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
		code int
		msg  string
	}{
		{"unknown flag", []string{"-nope"}, 2, "flag provided but not defined"},
		{"unknown policy", []string{"-raw-html", "keep"}, 2, "unknown -raw-html policy"},
		{"two files", []string{"a.md", "b.md"}, 2, "at most one input file"},
		{"missing file", []string{filepath.Join(t.TempDir(), "none.md")}, 1, "no such file"},
		{"help", []string{"-h"}, 0, "usage: gomdgo"},
	}
	for _, c := range cases {
		var out, errOut bytes.Buffer
		code := run(c.args, strings.NewReader(""), &out, &errOut)
		if code != c.code || !strings.Contains(errOut.String(), c.msg) || out.Len() != 0 {
			t.Errorf("%s: code %d, stdout %q, stderr %q", c.name, code, out.String(), errOut.String())
		}
	}
}

// Command fakecheck is the check program of the hand test for the folder
// check (.marksafe.yaml, testing/handtest/README.md, section O). It follows
// the format in docs/FOLDER-CHECK.md and reacts to marker words in the
// Markdown files of the folder:
//
//	FEHLER   error at that line        WARNUNG  warning at that line
//	ABSTURZ  crash (exit 3, stderr)    LANGSAM  sleeps longer than the timeout
//	KEIN-VORLAGENORDNER  exit 2 with a message
//
// A word counts only at the start of a line, so sop.md can name them.
//
// It also compares the front matter version of x.md and x.<lang>.md, as a
// finding without a line in another file.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type finding struct {
	File     string `json:"file"`
	Line     int    `json:"line,omitempty"`
	Col      int    `json:"col,omitempty"`
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type output struct {
	OK       bool      `json:"ok"`
	Findings []finding `json:"findings"`
	Error    string    `json:"error,omitempty"`
}

var translation = regexp.MustCompile(`^(.*)\.([a-z]{2})\.md$`)

func main() {
	dir := "."
	for _, a := range os.Args[1:] {
		if !strings.HasPrefix(a, "-") {
			dir = a
		}
	}
	fmt.Println("fakecheck: prüfe", dir) // a log line before the result
	findings, versions, err := scan(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	findings = append(findings, compareVersions(versions)...)
	sort.SliceStable(findings, func(i, j int) bool { return findings[i].File < findings[j].File })
	out := output{OK: len(findings) == 0, Findings: findings}
	data, _ := json.Marshal(out)
	fmt.Println(string(data))
	if len(findings) > 0 {
		os.Exit(1)
	}
}

// scan reads every .md file below dir; versions maps file → front matter version.
func scan(dir string) ([]finding, map[string]string, error) {
	findings := []finding{}
	versions := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for n := 1; sc.Scan(); n++ {
			findings = append(findings, markers(rel, n, sc.Text())...)
			if v, ok := strings.CutPrefix(sc.Text(), "version:"); ok {
				versions[rel] = strings.TrimSpace(v)
			}
		}
		return sc.Err()
	})
	return findings, versions, err
}

func markers(file string, line int, text string) []finding {
	trimmed := strings.TrimSpace(text)
	col := len(text) - len(strings.TrimLeft(text, " \t")) + 1
	switch {
	case strings.HasPrefix(trimmed, "ABSTURZ"):
		fmt.Fprintln(os.Stderr, "panic: Absturz zum Test in", file)
		os.Exit(3)
	case strings.HasPrefix(trimmed, "LANGSAM"):
		time.Sleep(10 * time.Minute)
	case strings.HasPrefix(trimmed, "KEIN-VORLAGENORDNER"):
		data, _ := json.Marshal(output{Error: "Das ist kein Vorlagen-Ordner (Testfall)."})
		fmt.Println(string(data))
		os.Exit(2)
	}
	switch {
	case strings.HasPrefix(trimmed, "FEHLER"):
		return []finding{{file, line, col, "error", "demo-fehler", "Testfehler: Die Zeile beginnt mit FEHLER."}}
	case strings.HasPrefix(trimmed, "WARNUNG"):
		return []finding{{file, line, col, "warning", "demo-warnung", "Testwarnung: Die Zeile beginnt mit WARNUNG."}}
	}
	return nil
}

// compareVersions: a translation must have the version of its base file.
func compareVersions(versions map[string]string) []finding {
	var out []finding
	for file, v := range versions {
		m := translation.FindStringSubmatch(file)
		if m == nil {
			continue
		}
		if base, ok := versions[m[1]+".md"]; ok && base != v {
			out = append(out, finding{
				File: file, Severity: "error", Code: "translation-version",
				Message: fmt.Sprintf("Die Übersetzung hat Version %s, die Grundfassung %s.", v, base),
			})
		}
	}
	return out
}

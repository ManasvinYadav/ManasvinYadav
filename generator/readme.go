package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	projectsStart = "<!-- projects:start -->"
	projectsEnd   = "<!-- projects:end -->"
)

// projectsBlock is the HTML between the projects markers in the README: one
// linked, theme-aware card per project, two per row.
func projectsBlock(cfg Config, repos []Repo) string {
	if len(repos) == 0 {
		return "<p align=\"center\"><sub>Nothing public yet — watch this space.</sub></p>"
	}
	var b strings.Builder
	b.WriteString(`<p align="center">`)
	for i, r := range repos {
		alt := r.Name
		if d := strings.TrimSpace(r.Description); d != "" {
			alt += ": " + d
		}
		fmt.Fprintf(&b, "\n  %s", picture(cfg, fmt.Sprintf("project-%d", i+1), alt, r.URL, "49%"))
	}
	b.WriteString("\n</p>")
	return b.String()
}

// picture links a dark/light pair of cards from the output branch.
func picture(cfg Config, name, alt, href, width string) string {
	return fmt.Sprintf(`<a href="%s"><picture>`+
		`<source media="(prefers-color-scheme: dark)" srcset="%s">`+
		`<source media="(prefers-color-scheme: light)" srcset="%s">`+
		`<img alt="%s" src="%s" width="%s">`+
		`</picture></a>`,
		esc(href), cfg.rawURL(name+"-dark.svg"), cfg.rawURL(name+"-light.svg"),
		esc(alt), cfg.rawURL(name+"-light.svg"), width)
}

// replaceBlock swaps the text between start and end markers, keeping the
// markers themselves. It reports whether the markers were found.
func replaceBlock(doc, start, end, body string) (string, bool) {
	i := strings.Index(doc, start)
	if i < 0 {
		return doc, false
	}
	j := strings.Index(doc[i+len(start):], end)
	if j < 0 {
		return doc, false
	}
	j += i + len(start)
	return doc[:i+len(start)] + "\n" + body + "\n" + doc[j:], true
}

func updateReadme(path string, cfg Config, repos []Repo) (changed bool, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	doc, ok := replaceBlock(string(raw), projectsStart, projectsEnd, projectsBlock(cfg, repos))
	if !ok {
		return false, fmt.Errorf("%s: missing %s … %s markers", path, projectsStart, projectsEnd)
	}
	if doc == string(raw) {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(doc), 0o644)
}

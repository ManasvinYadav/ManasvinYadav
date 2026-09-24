package main

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func day(s string, n int) Day {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return Day{d, n}
}

func TestStreaks(t *testing.T) {
	days := []Day{
		day("2026-09-01", 2), day("2026-09-02", 1), day("2026-09-03", 4), // 3-day run
		day("2026-09-04", 0),
		day("2026-09-05", 1), day("2026-09-06", 1), // run into "yesterday"
		day("2026-09-07", 0), // today, nothing yet
	}
	today := time.Date(2026, 9, 7, 15, 0, 0, 0, time.UTC)
	cur, long := streaks(days, today)
	if cur != 2 || long != 3 {
		t.Fatalf("streaks = %d, %d; want 2, 3", cur, long)
	}

	// A gap before yesterday ends the current streak.
	cur, _ = streaks(days[:5], today)
	if cur != 0 {
		t.Fatalf("current streak after a gap = %d; want 0", cur)
	}

	// Days after "today" (time-zone skew) are ignored.
	cur, long = streaks(append(days, day("2026-09-08", 9)), today)
	if cur != 2 || long != 3 {
		t.Fatalf("streaks with a future day = %d, %d; want 2, 3", cur, long)
	}
}

func TestWeekly(t *testing.T) {
	// 2026-09-05 is a Saturday, 2026-09-06 a Sunday.
	weeks := weekly([]Day{day("2026-09-04", 1), day("2026-09-05", 2), day("2026-09-06", 3), day("2026-09-07", 4)})
	if len(weeks) != 2 || weeks[0].Count != 3 || weeks[1].Count != 7 {
		t.Fatalf("weekly = %+v", weeks)
	}
	if weeks[1].Start.Weekday() != time.Sunday {
		t.Fatalf("weeks should start on Sunday, got %s", weeks[1].Start.Weekday())
	}
}

func TestMergeDays(t *testing.T) {
	got := mergeDays([]Day{day("2026-01-02", 1), day("2026-01-01", 3)}, []Day{day("2026-01-02", 1)})
	if len(got) != 2 || got[0].Count != 3 || got[1].Count != 1 || !got[0].Date.Before(got[1].Date) {
		t.Fatalf("mergeDays = %+v", got)
	}
}

func TestCompact(t *testing.T) {
	for n, want := range map[int]string{0: "0", 999: "999", 1284: "1,284", 9999: "9,999", 12900: "12.9K", 20000: "20K", 4_200_000: "4.2M"} {
		if got := compact(n); got != want {
			t.Errorf("compact(%d) = %q; want %q", n, got, want)
		}
	}
}

func TestNiceStep(t *testing.T) {
	for v, want := range map[float64]float64{0: 1, 0.5: 1, 1.2: 2, 27: 30, 50: 50, 51: 60, 130: 150, 700: 800} {
		if got := niceStep(v); got != want {
			t.Errorf("niceStep(%v) = %v; want %v", v, got, want)
		}
	}
}

func TestWrap(t *testing.T) {
	got := wrap("the quick brown fox jumps over the lazy dog", 10, 2)
	if len(got) != 2 || got[0] != "the quick" || !strings.HasSuffix(got[1], "…") || len([]rune(got[1])) > 10 {
		t.Fatalf("wrap = %q", got)
	}
	if got := wrap("short", 10, 2); len(got) != 1 || got[0] != "short" {
		t.Fatalf("wrap(short) = %q", got)
	}
	if got := wrap("", 10, 2); got != nil {
		t.Fatalf("wrap(empty) = %q", got)
	}
	if got := wrap("supercalifragilistic", 8, 2); got[0] != "superca…" {
		t.Fatalf("wrap(long word) = %q", got)
	}
}

func TestLegible(t *testing.T) {
	for _, bg := range []string{"#0d1117", "#ffffff"} {
		for _, fg := range []string{"#f1e05a", "#555555", "#00add8", "#000080"} {
			if c := contrast(legible(fg, bg, 3), bg); c < 3 {
				t.Errorf("legible(%s on %s) contrast %.2f < 3", fg, bg, c)
			}
		}
	}
}

func TestPickProjects(t *testing.T) {
	cfg := Config{Login: "me"}
	cfg.Projects.Max = 2
	owned := []Repo{
		{Name: "me", Stars: 100}, // the profile repo itself
		{Name: "old", Stars: 50, Archived: true},
		{Name: "b", Stars: 5, PushedAt: time.Unix(1, 0)},
		{Name: "c", Stars: 5, PushedAt: time.Unix(2, 0)},
		{Name: "a", Stars: 1},
	}
	got := pickProjects(cfg, nil, owned)
	if len(got) != 2 || got[0].Name != "c" || got[1].Name != "b" {
		t.Fatalf("fallback projects = %+v", got)
	}
	got = pickProjects(cfg, []Repo{{Name: "old", Archived: true}}, owned)
	if len(got) != 1 || got[0].Name != "old" {
		t.Fatalf("pinned projects should win, even archived: %+v", got)
	}
}

func TestReplaceBlock(t *testing.T) {
	doc := "intro\n" + projectsStart + "\nold\n" + projectsEnd + "\noutro\n"
	got, ok := replaceBlock(doc, projectsStart, projectsEnd, "new")
	want := "intro\n" + projectsStart + "\nnew\n" + projectsEnd + "\noutro\n"
	if !ok || got != want {
		t.Fatalf("replaceBlock = %q, %v", got, ok)
	}
	if _, ok := replaceBlock("no markers", projectsStart, projectsEnd, "x"); ok {
		t.Fatal("replaceBlock found markers that are not there")
	}
}

// Every card must be well-formed XML for rich and empty profiles alike, or
// GitHub shows a broken image.
func TestCardsAreWellFormed(t *testing.T) {
	cfg, err := loadConfig("profile.json")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	empty := &Profile{Login: cfg.Login, CreatedAt: now.AddDate(0, -1, 0), Now: now}
	for d := now.AddDate(-1, 0, 0); !d.After(now); d = d.AddDate(0, 0, 1) {
		empty.Year = append(empty.Year, Day{truncDay(d), 0})
	}
	empty.AllTime = empty.Year

	for _, p := range []*Profile{demoProfile(cfg, now), empty} {
		for _, th := range themes {
			docs := map[string]string{
				"hero": renderHero(cfg, th), "stack": renderStack(cfg, th), "contact": renderContact(cfg, th),
				"overview": renderOverview(p, th), "languages": renderLanguages(p, th), "highlights": renderHighlights(p, th),
			}
			for i, r := range p.Projects {
				docs["project"+string(rune('1'+i))] = renderProject(r, cfg.Login, th)
			}
			for name, doc := range docs {
				dec := xml.NewDecoder(strings.NewReader(doc))
				for {
					_, err := dec.Token()
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						t.Fatalf("%s (%s): %v", name, th.Name, err)
					}
				}
			}
		}
	}
}

func TestProjectsBlockEscapes(t *testing.T) {
	cfg := Config{Repository: "me/me", OutputBranch: "output"}
	block := projectsBlock(cfg, []Repo{{Name: "x", Description: `"quoted" <b>`, URL: "https://github.com/me/x"}})
	if strings.Contains(block, `"quoted"`) || strings.Contains(block, "<b>") {
		t.Fatalf("description not escaped: %s", block)
	}
	if !strings.Contains(block, "https://raw.githubusercontent.com/me/me/output/project-1-dark.svg") {
		t.Fatalf("missing card URL: %s", block)
	}
}

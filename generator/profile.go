package main

import (
	"sort"
	"time"
)

// Profile is everything the data-driven cards need, however it was gathered.
type Profile struct {
	Login     string
	CreatedAt time.Time
	Now       time.Time

	Year    []Day // the last year, oldest first — drives the activity chart
	AllTime []Day // every day since the account was created, oldest first

	Commits      int // commit contributions in the last year
	PullRequests int // all time
	Issues       int // all time
	Stars        int // stars across the user's public repositories
	PublicRepos  int

	Languages []Language // largest first
	Projects  []Repo
}

type Day struct {
	Date  time.Time
	Count int
}

type Language struct {
	Name  string
	Color string
	Bytes int64
}

type Repo struct {
	Owner, Name   string
	Description   string
	URL           string
	Language      string
	LanguageColor string
	Stars, Forks  int
	Archived      bool
	PushedAt      time.Time
}

func sum(days []Day) int {
	n := 0
	for _, d := range days {
		n += d.Count
	}
	return n
}

func activeDays(days []Day) int {
	n := 0
	for _, d := range days {
		if d.Count > 0 {
			n++
		}
	}
	return n
}

// streaks returns the current and longest run of consecutive days with at
// least one contribution. A quiet today does not break the current streak —
// the day is not over yet.
func streaks(days []Day, today time.Time) (current, longest int) {
	today = truncDay(today)
	run := 0
	var prev time.Time
	for _, d := range days {
		if d.Date.After(today) {
			break
		}
		if d.Count > 0 {
			if run > 0 && d.Date.Sub(prev) == 24*time.Hour {
				run++
			} else {
				run = 1
			}
			prev = d.Date
			longest = max(longest, run)
		}
	}
	// The run is current only if it reaches today or yesterday.
	if run > 0 && !prev.Before(today.AddDate(0, 0, -1)) {
		current = run
	}
	return current, longest
}

func truncDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Week is seven days of contributions, starting on a Sunday like GitHub's grid.
type Week struct {
	Start time.Time
	Count int
}

func weekly(days []Day) []Week {
	var weeks []Week
	for _, d := range days {
		start := d.Date.AddDate(0, 0, -int(d.Date.Weekday()))
		if len(weeks) == 0 || !weeks[len(weeks)-1].Start.Equal(start) {
			weeks = append(weeks, Week{Start: start})
		}
		weeks[len(weeks)-1].Count += d.Count
	}
	return weeks
}

// mergeDays folds calendars together, keeping one entry per date, sorted.
func mergeDays(sets ...[]Day) []Day {
	byDate := map[time.Time]int{}
	for _, set := range sets {
		for _, d := range set {
			byDate[d.Date] = max(byDate[d.Date], d.Count)
		}
	}
	out := make([]Day, 0, len(byDate))
	for date, n := range byDate {
		out = append(out, Day{date, n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.Before(out[j].Date) })
	return out
}

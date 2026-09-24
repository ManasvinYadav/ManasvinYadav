package main

import (
	"math"
	"math/rand/v2"
	"time"
)

// demoProfile fabricates plausible data so the cards can be previewed
// without a token. It is never published: the workflow always runs live.
func demoProfile(cfg Config, now time.Time) *Profile {
	rng := rand.New(rand.NewPCG(7, 42))
	created := time.Date(2021, 5, 16, 0, 0, 0, 0, time.UTC)
	today := truncDay(now)

	var all []Day
	for d := created; !d.After(today); d = d.AddDate(0, 0, 1) {
		// Busier on weekdays and during a few "project sprints".
		rate := 0.9 + 1.6*math.Max(0, math.Sin(float64(d.YearDay())/58))
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			rate *= 0.55
		}
		n := 0
		if rng.Float64() < 0.62 {
			n = int(rng.ExpFloat64() * rate * 2.2)
		}
		all = append(all, Day{d, n})
	}
	yearStart := today.AddDate(-1, 0, 0)
	yearStart = yearStart.AddDate(0, 0, -int(yearStart.Weekday()))
	var year []Day
	for _, d := range all {
		if !d.Date.Before(yearStart) {
			year = append(year, d)
		}
	}

	return &Profile{
		Login: cfg.Login, CreatedAt: created, Now: now,
		Year: year, AllTime: all,
		Commits: sum(year) * 7 / 10, PullRequests: 38, Issues: 17, Stars: 64, PublicRepos: 8,
		Languages: []Language{
			{"Go", "#00ADD8", 412_000},
			{"TypeScript", "#3178c6", 268_000},
			{"JavaScript", "#f1e05a", 96_000},
			{"CSS", "#663399", 41_000},
			{"HTML", "#e34c26", 30_000},
			{"Shell", "#89e051", 6_400},
			{"Dockerfile", "#384d54", 1_900},
			{"Makefile", "#427819", 900},
			{"Lua", "#000080", 400},
		},
		Projects: []Repo{
			{Owner: cfg.Login, Name: "sample-api", Description: "A tiny, fast HTTP service in Go with graceful shutdown, structured logs and zero dependencies.", URL: "https://github.com/" + cfg.Login, Language: "Go", LanguageColor: "#00ADD8", Stars: 23, Forks: 4, PushedAt: now.AddDate(0, 0, -3)},
			{Owner: cfg.Login, Name: "sample-dashboard", Description: "Realtime dashboard built with TypeScript and web sockets.", URL: "https://github.com/" + cfg.Login, Language: "TypeScript", LanguageColor: "#3178c6", Stars: 17, Forks: 2, PushedAt: now.AddDate(0, -1, 0)},
			{Owner: cfg.Login, Name: "sample-dotfiles", Description: "", URL: "https://github.com/" + cfg.Login, Language: "Shell", LanguageColor: "#89e051", Stars: 9, Forks: 1, PushedAt: now.AddDate(0, -2, 0)},
			{Owner: cfg.Login, Name: "sample-cli-with-a-really-long-repository-name", Description: "Command-line tool that does one thing well, with a description long enough that it needs to wrap onto a second line and then be cut short.", URL: "https://github.com/" + cfg.Login, Language: "JavaScript", LanguageColor: "#f1e05a", Stars: 1204, Forks: 0, Archived: true, PushedAt: now.AddDate(-1, 0, 0)},
		},
	}
}

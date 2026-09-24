package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const graphqlURL = "https://api.github.com/graphql"

const repoFragment = `
fragment repo on Repository {
  name description url stargazerCount forkCount isArchived isPrivate pushedAt
  owner { login }
  primaryLanguage { name color }
}`

const userQuery = `query($login: String!) {
  user(login: $login) {
    createdAt
    pullRequests { totalCount }
    issues { totalCount }
    publicRepos: repositories(privacy: PUBLIC, ownerAffiliations: OWNER) { totalCount }
    pinnedItems(first: 6, types: REPOSITORY) { nodes { ... on Repository { ...repo } } }
    contributionsCollection { totalCommitContributions }
  }
}` + repoFragment

const reposQuery = `query($login: String!, $cursor: String) {
  user(login: $login) {
    repositories(first: 100, after: $cursor, privacy: PUBLIC, ownerAffiliations: OWNER, isFork: false,
                 orderBy: {field: STARGAZERS, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes {
        ...repo
        languages(first: 12, orderBy: {field: SIZE, direction: DESC}) { edges { size node { name color } } }
      }
    }
  }
}` + repoFragment

type gqlRepo struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	URL            string    `json:"url"`
	StargazerCount int       `json:"stargazerCount"`
	ForkCount      int       `json:"forkCount"`
	IsArchived     bool      `json:"isArchived"`
	IsPrivate      bool      `json:"isPrivate"`
	PushedAt       time.Time `json:"pushedAt"`
	Owner          struct {
		Login string `json:"login"`
	} `json:"owner"`
	PrimaryLanguage *struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"primaryLanguage"`
	Languages struct {
		Edges []struct {
			Size int64 `json:"size"`
			Node struct {
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"languages"`
}

func (r gqlRepo) repo() Repo {
	out := Repo{
		Owner: r.Owner.Login, Name: r.Name, Description: r.Description, URL: r.URL,
		Stars: r.StargazerCount, Forks: r.ForkCount, Archived: r.IsArchived, PushedAt: r.PushedAt,
	}
	if r.PrimaryLanguage != nil {
		out.Language, out.LanguageColor = r.PrimaryLanguage.Name, r.PrimaryLanguage.Color
	}
	return out
}

type gqlCalendar struct {
	ContributionCalendar struct {
		Weeks []struct {
			ContributionDays []struct {
				Date              string `json:"date"`
				ContributionCount int    `json:"contributionCount"`
			} `json:"contributionDays"`
		} `json:"weeks"`
	} `json:"contributionCalendar"`
}

func (c gqlCalendar) days() ([]Day, error) {
	var out []Day
	for _, w := range c.ContributionCalendar.Weeks {
		for _, d := range w.ContributionDays {
			date, err := time.Parse(time.DateOnly, d.Date)
			if err != nil {
				return nil, err
			}
			out = append(out, Day{date, d.ContributionCount})
		}
	}
	return out, nil
}

// fetchProfile gathers everything from the GitHub GraphQL API. Only public
// repositories are ever read, so a personal token cannot leak private names.
func fetchProfile(ctx context.Context, token string, cfg Config, now time.Time) (*Profile, error) {
	c := &client{token: token, http: &http.Client{Timeout: 30 * time.Second}}
	vars := map[string]any{"login": cfg.Login}

	var u struct {
		User *struct {
			CreatedAt    time.Time                `json:"createdAt"`
			PullRequests struct{ TotalCount int } `json:"pullRequests"`
			Issues       struct{ TotalCount int } `json:"issues"`
			PublicRepos  struct{ TotalCount int } `json:"publicRepos"`
			PinnedItems  struct {
				Nodes []gqlRepo `json:"nodes"`
			} `json:"pinnedItems"`
			ContributionsCollection struct {
				TotalCommitContributions int `json:"totalCommitContributions"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}
	if err := c.query(ctx, userQuery, vars, &u); err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	if u.User == nil {
		return nil, fmt.Errorf("user %q not found", cfg.Login)
	}
	p := &Profile{
		Login:        cfg.Login,
		CreatedAt:    u.User.CreatedAt,
		Now:          now,
		Commits:      u.User.ContributionsCollection.TotalCommitContributions,
		PullRequests: u.User.PullRequests.TotalCount,
		Issues:       u.User.Issues.TotalCount,
		PublicRepos:  u.User.PublicRepos.TotalCount,
	}

	var owned []gqlRepo
	for cursor := (*string)(nil); ; {
		var r struct {
			User struct {
				Repositories struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []gqlRepo `json:"nodes"`
				} `json:"repositories"`
			} `json:"user"`
		}
		if err := c.query(ctx, reposQuery, map[string]any{"login": cfg.Login, "cursor": cursor}, &r); err != nil {
			return nil, fmt.Errorf("repositories: %w", err)
		}
		owned = append(owned, r.User.Repositories.Nodes...)
		if !r.User.Repositories.PageInfo.HasNextPage {
			break
		}
		cursor = &r.User.Repositories.PageInfo.EndCursor
	}

	langs := map[string]*Language{}
	for _, r := range owned {
		p.Stars += r.StargazerCount
		if strings.EqualFold(r.Name, cfg.Login) {
			continue // the profile repo's own generator is not the user's work
		}
		for _, e := range r.Languages.Edges {
			if containsFold(cfg.HideLanguages, e.Node.Name) {
				continue
			}
			l := langs[e.Node.Name]
			if l == nil {
				l = &Language{Name: e.Node.Name, Color: e.Node.Color}
				langs[e.Node.Name] = l
			}
			l.Bytes += e.Size
		}
	}
	for _, l := range langs {
		p.Languages = append(p.Languages, *l)
	}
	sort.Slice(p.Languages, func(i, j int) bool {
		if p.Languages[i].Bytes != p.Languages[j].Bytes {
			return p.Languages[i].Bytes > p.Languages[j].Bytes
		}
		return p.Languages[i].Name < p.Languages[j].Name
	})

	var pinned, candidates []Repo
	for _, r := range u.User.PinnedItems.Nodes {
		if !r.IsPrivate && r.Name != "" {
			pinned = append(pinned, r.repo())
		}
	}
	for _, r := range owned {
		candidates = append(candidates, r.repo())
	}
	p.Projects = pickProjects(cfg, pinned, candidates)

	year, all, err := fetchCalendars(ctx, c, cfg.Login, p.CreatedAt, now)
	if err != nil {
		return nil, fmt.Errorf("contributions: %w", err)
	}
	p.Year, p.AllTime = year, all
	return p, nil
}

// pickProjects prefers the repositories pinned on the profile; without pins
// it falls back to the most-starred, most recently pushed originals.
func pickProjects(cfg Config, pinned, owned []Repo) []Repo {
	usable := func(r Repo) bool {
		return !containsFold(cfg.Projects.Hide, r.Name) && !strings.EqualFold(r.Name, cfg.Login)
	}
	var out []Repo
	for _, r := range pinned {
		if usable(r) {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		sorted := append([]Repo(nil), owned...)
		sort.SliceStable(sorted, func(i, j int) bool {
			if sorted[i].Stars != sorted[j].Stars {
				return sorted[i].Stars > sorted[j].Stars
			}
			return sorted[i].PushedAt.After(sorted[j].PushedAt)
		})
		for _, r := range sorted {
			if usable(r) && !r.Archived {
				out = append(out, r)
			}
		}
	}
	if len(out) > cfg.Projects.Max {
		out = out[:cfg.Projects.Max]
	}
	return out
}

// fetchCalendars loads the rolling last-year calendar plus one calendar per
// calendar year since the account was created, in a single aliased query.
func fetchCalendars(ctx context.Context, c *client, login string, created, now time.Time) (year, all []Day, err error) {
	var q strings.Builder
	q.WriteString("query($login: String!) { user(login: $login) {\n")
	const fields = "{ contributionCalendar { weeks { contributionDays { date contributionCount } } } }"
	q.WriteString("  recent: contributionsCollection " + fields + "\n")
	for y := created.Year(); y <= now.Year(); y++ {
		from := time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
		if from.Before(created) {
			from = created
		}
		to := time.Date(y, 12, 31, 23, 59, 59, 0, time.UTC)
		if to.After(now) {
			to = now
		}
		fmt.Fprintf(&q, "  y%d: contributionsCollection(from: %q, to: %q) %s\n",
			y, from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339), fields)
	}
	q.WriteString("} }")

	var r struct {
		User map[string]gqlCalendar `json:"user"`
	}
	if err := c.query(ctx, q.String(), map[string]any{"login": login}, &r); err != nil {
		return nil, nil, err
	}
	var years [][]Day
	for alias, cal := range r.User {
		days, err := cal.days()
		if err != nil {
			return nil, nil, err
		}
		if alias == "recent" {
			year = days
		} else {
			years = append(years, days)
		}
	}
	return mergeDays(year), mergeDays(years...), nil
}

type client struct {
	token string
	http  *http.Client
}

// query runs one GraphQL request, retrying transient server errors.
func (c *client) query(ctx context.Context, q string, vars map[string]any, out any) error {
	payload, err := json.Marshal(map[string]any{"query": q, "variables": vars})
	if err != nil {
		return err
	}
	var lastErr error
	for attempt := range 4 {
		if attempt > 0 {
			select {
			case <-time.After(time.Duration(1<<attempt) * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlURL, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "profile-card-generator")
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("%s", resp.Status)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s: %s", resp.Status, truncate(strings.TrimSpace(string(body)), 300))
		}
		var env struct {
			Data   json.RawMessage `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(body, &env); err != nil {
			return err
		}
		if len(env.Errors) > 0 {
			msgs := make([]string, len(env.Errors))
			for i, e := range env.Errors {
				msgs[i] = e.Message
			}
			return fmt.Errorf("%s", strings.Join(msgs, "; "))
		}
		return json.Unmarshal(env.Data, out)
	}
	return lastErr
}

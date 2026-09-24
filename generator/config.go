package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config is the hand-edited part of the profile: who it is for and what the
// static artwork says. Everything else comes from the GitHub API.
type Config struct {
	Login        string `json:"login"`
	Name         string `json:"name"`
	Repository   string `json:"repository"`
	OutputBranch string `json:"outputBranch"`

	Hero struct {
		Prompt   string `json:"prompt"`
		Tagline  string `json:"tagline"`
		Chips    []Chip `json:"chips"`
		Terminal struct {
			Title   string   `json:"title"`
			Command string   `json:"command"`
			Output  []string `json:"output"`
		} `json:"terminal"`
	} `json:"hero"`

	Stack []struct {
		Group string   `json:"group"`
		Items []string `json:"items"`
	} `json:"stack"`

	Contact struct {
		Label string `json:"label"`
	} `json:"contact"`

	Projects struct {
		Max  int      `json:"max"`
		Hide []string `json:"hide"`
	} `json:"projects"`

	HideLanguages []string `json:"hideLanguages"`
}

// Chip is a pill in the hero: a coloured dot or an octicon, then a label.
type Chip struct {
	Label string `json:"label"`
	Color string `json:"color,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("%s: %w", path, err)
	}
	if cfg.Login == "" || cfg.Repository == "" {
		return cfg, errors.New("profile.json: login and repository are required")
	}
	if cfg.Name == "" {
		cfg.Name = cfg.Login
	}
	if cfg.OutputBranch == "" {
		cfg.OutputBranch = "output"
	}
	if cfg.Projects.Max <= 0 {
		cfg.Projects.Max = 4
	}
	for _, group := range cfg.Stack {
		for _, slug := range group.Items {
			if _, ok := brands[slug]; !ok {
				return cfg, fmt.Errorf("profile.json: unknown stack item %q", slug)
			}
		}
	}
	for _, chip := range cfg.Hero.Chips {
		if chip.Icon != "" {
			if _, ok := octicons[chip.Icon]; !ok {
				return cfg, fmt.Errorf("profile.json: unknown chip icon %q", chip.Icon)
			}
		}
	}
	return cfg, nil
}

// rawURL is where a file on the output branch is served from.
func (c Config) rawURL(file string) string {
	return "https://raw.githubusercontent.com/" + c.Repository + "/" + c.OutputBranch + "/" + file
}

func containsFold(list []string, s string) bool {
	for _, v := range list {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}

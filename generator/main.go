// Command generator draws the SVG artwork for the profile README.
//
// Static artwork (hero, toolbox, contact button) only depends on profile.json
// and is committed to assets/. Data cards are rendered from the GitHub API
// into -out, which the workflow publishes to the output branch.
//
//	go run . -demo                 preview every card with sample data
//	GITHUB_TOKEN=… go run .        render live cards and refresh the README
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var (
		configPath = flag.String("config", "profile.json", "profile configuration")
		assetsDir  = flag.String("assets", "../assets", "where static artwork is written")
		outDir     = flag.String("out", "../dist", "where data-driven cards are written")
		readmePath = flag.String("readme", "../README.md", "README whose projects section is refreshed (empty to skip)")
		demo       = flag.Bool("demo", false, "use sample data instead of the GitHub API; never touches the README")
		staticOnly = flag.Bool("static", false, "only render the static artwork")
	)
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("generator: ")

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range themes {
		must(write(*assetsDir, "hero-"+t.Name+".svg", renderHero(cfg, t)))
		must(write(*assetsDir, "stack-"+t.Name+".svg", renderStack(cfg, t)))
		must(write(*assetsDir, "contact-"+t.Name+".svg", renderContact(cfg, t)))
	}
	if *staticOnly {
		return
	}

	now := time.Now().UTC()
	var p *Profile
	if *demo {
		p = demoProfile(cfg, now)
	} else {
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			log.Fatal("GITHUB_TOKEN is not set (run with -demo to preview using sample data)")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		if p, err = fetchProfile(ctx, token, cfg, now); err != nil {
			log.Fatal(err)
		}
	}

	for _, t := range themes {
		must(write(*outDir, "overview-"+t.Name+".svg", renderOverview(p, t)))
		must(write(*outDir, "languages-"+t.Name+".svg", renderLanguages(p, t)))
		must(write(*outDir, "highlights-"+t.Name+".svg", renderHighlights(p, t)))
		for i, r := range p.Projects {
			must(write(*outDir, fmt.Sprintf("project-%d-%s.svg", i+1, t.Name), renderProject(r, cfg.Login, t)))
		}
	}
	log.Printf("rendered cards for %s: %d contributions in the last year, %d languages, %d projects",
		p.Login, sum(p.Year), len(p.Languages), len(p.Projects))

	if *readmePath != "" && !*demo {
		changed, err := updateReadme(*readmePath, cfg, p.Projects)
		if err != nil {
			log.Fatal(err)
		}
		if changed {
			log.Printf("updated the projects section of %s", *readmePath)
		}
	}
}

func write(dir, name, content string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

# Profile card generator

A small, dependency-free Go program that draws every card on the profile
README. No third-party stats services: if GitHub is up, the profile renders.

## How it fits together

| What | Where it comes from | Where it lives |
| --- | --- | --- |
| Hero, toolbox, "Say hello" button | `profile.json` | `assets/` on `main` |
| Activity, languages, highlights, project cards | GitHub GraphQL API | `output` branch |
| Contribution snake | [Platane/snk](https://github.com/Platane/snk) | `output` branch |
| Projects section of the README | pinned repos (or top repos) | between the `projects` markers |

`.github/workflows/profile.yml` runs on every push to `main`, daily at 03:17 UTC,
and on demand. It tests and runs the generator, renders the snake, force-pushes
the cards to the `output` branch as a single commit (so history never grows), and
commits any README or `assets/` changes back to `main`.

Every card is drawn twice, dark and light, and the README picks one with
`<picture>` and `prefers-color-scheme`. Fonts are subset and embedded, so
text looks and measures the same everywhere, and animations switch off
for readers who prefer reduced motion.

## Common changes

- **Tagline, chips, terminal lines, toolbox:** edit `profile.json` and push. The
  workflow redraws `assets/`.
- **Featured projects:** pin repositories on your GitHub profile. Without pins,
  the most-starred public repositories are shown. Hide a repository with
  `projects.hide`. Hide a language (e.g. `"Jupyter Notebook"`) with
  `hideLanguages`.
- **Private contributions:** add a `PROFILE_TOKEN` repository secret (a classic
  token with the `read:user` scope). Only public repositories are ever listed,
  so private repository names never appear.
- **Refresh now:** Actions → Profile → Run workflow.

## Local preview

```sh
cd generator
go run . -demo -assets /tmp/preview -out /tmp/preview   # sample data, no token needed
GITHUB_TOKEN=$(gh auth token) go run . -out ../dist      # live data
go test ./...
```

`-demo` never touches the README or the committed assets, so sample data can't
end up on the profile by accident.

## Notes

- GitHub pauses scheduled workflows in repositories with no activity for 60
  days. If the cards stop updating, re-enable the workflow from the Actions tab.
- Stats count public activity. A streak counts consecutive days with at least
  one contribution (UTC). A quiet today doesn't break it until the day is over.

## Credits

Fonts: [JetBrains Mono](https://github.com/JetBrains/JetBrainsMono) and
[Space Grotesk](https://github.com/floriankarsten/space-grotesk), both under the
SIL Open Font License (see `fonts/`). Icons:
[Octicons](https://github.com/primer/octicons) (MIT, see `LICENSE-octicons`) and
[Simple Icons](https://simpleicons.org) (CC0). Brand logos belong to their
owners.

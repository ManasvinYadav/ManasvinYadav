<!--
  Most of this page is drawn by code:
    • assets/ — hero, toolbox and contact button, rendered from generator/profile.json
    • stats and project cards — redrawn daily from live GitHub data by
      .github/workflows/profile.yml and served from the `output` branch
  See generator/README.md for how it all fits together.
-->

<a href="https://github.com/ManasvinYadav">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/hero-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="assets/hero-light.svg">
    <img alt="Manasvin Yadav — Student · part-time larper · Go & TypeScript" src="assets/hero-light.svg" width="100%">
  </picture>
</a>

<p align="center">
  <a href="#about"><samp>about</samp></a> &nbsp;·&nbsp;
  <a href="#toolbox"><samp>toolbox</samp></a> &nbsp;·&nbsp;
  <a href="#activity"><samp>activity</samp></a> &nbsp;·&nbsp;
  <a href="#projects"><samp>projects</samp></a> &nbsp;·&nbsp;
  <a href="#contact"><samp>contact</samp></a>
</p>

### <samp>~/about</samp>

```go
package main

import "fmt"

type Developer struct {
    Name, Role, Based string
    Focus             []string
    Philosophy        string
}

func main() {
    me := Developer{
        Name:       "Manasvin Yadav",
        Role:       "Student · part-time larper",
        Based:      "India",
        Focus:      []string{"Go backends", "full-stack web", "modern tooling"},
        Philosophy: "clean code, high performance, robust engineering",
    }
    fmt.Printf("Hi, I'm %s. Thanks for stopping by!\n", me.Name)
}
```

### <samp>~/toolbox</samp>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/stack-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="assets/stack-light.svg">
  <img alt="Languages: Go, TypeScript, JavaScript, HTML, CSS. Tooling: Linux, Docker, Git." src="assets/stack-light.svg" width="100%">
</picture>

### <samp>~/activity</samp>

<a href="https://github.com/ManasvinYadav">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/overview-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/overview-light.svg">
    <img alt="Contributions in the last year, streaks and a weekly activity chart" src="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/overview-light.svg" width="100%">
  </picture>
</a>

<p align="center">
  <a href="https://github.com/ManasvinYadav?tab=repositories"><picture><source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/languages-dark.svg"><source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/languages-light.svg"><img alt="Most used languages by code size" src="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/languages-light.svg" width="49%"></picture></a>
  <a href="https://github.com/ManasvinYadav?tab=repositories"><picture><source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/highlights-dark.svg"><source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/highlights-light.svg"><img alt="Highlights: commits, pull requests, issues, stars and repositories" src="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/highlights-light.svg" width="49%"></picture></a>
</p>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/snake-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/snake-light.svg">
  <img alt="A snake eating the contribution graph" src="https://raw.githubusercontent.com/ManasvinYadav/ManasvinYadav/output/snake-light.svg" width="100%">
</picture>

### <samp>~/projects</samp>

<!-- projects:start -->
<p align="center"><sub>Project cards appear here after the first workflow run.</sub></p>
<!-- projects:end -->

### <samp>~/contact</samp>

<p align="center">
  <a href="mailto:manasvinyadav@icloud.com">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="assets/contact-dark.svg">
      <source media="(prefers-color-scheme: light)" srcset="assets/contact-light.svg">
      <img alt="Say hello by email" src="assets/contact-light.svg" height="48">
    </picture>
  </a>
</p>

<p align="center">
  <sub>Cards are redrawn every day from live GitHub data by a small Go program in <a href="generator"><code>generator/</code></a>.</sub>
  <br><br>
  <img alt="Profile views" src="https://komarev.com/ghpvc/?username=ManasvinYadav&label=profile%20views&color=00b37e&style=flat-square">
</p>

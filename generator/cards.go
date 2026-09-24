package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// renderOverview is the wide activity card: a headline number, four stat
// tiles and a weekly contributions chart for the last year.
func renderOverview(p *Profile, t Theme) string {
	const W, H = 840.0, 300.0
	var b strings.Builder
	b.WriteString(card(t, W, H))

	// Header.
	b.WriteString(octicon("mark-github", 28, 24, 16, t.Ink))
	fmt.Fprintf(&b, `<text class="s" x="52" y="37" font-size="15" font-weight="600" fill="%s">%s</text>`, t.Ink, esc(p.Login))
	fmt.Fprintf(&b, `<text class="m" x="%s" y="37" font-size="12" fill="%s">· on GitHub since %d</text>`,
		num(52+sansWidth(p.Login, 15, 600)+8), t.Muted, p.CreatedAt.Year())
	fmt.Fprintf(&b, `<text class="m" x="812" y="37" font-size="11" fill="%s" text-anchor="end">Updated %s</text>`,
		t.Muted, p.Now.Format("2 Jan 2006"))

	// Headline figure.
	total := sum(p.Year)
	fmt.Fprintf(&b, `<text class="s" x="26" y="112" font-size="48" font-weight="600" letter-spacing="-1" fill="%s">%s</text>`, t.Ink, compact(total))
	fmt.Fprintf(&b, `<text class="m" x="28" y="136" font-size="12" fill="%s">contributions in the last year</text>`, t.Ink2)

	// Stat tiles.
	current, longest := streaks(p.AllTime, p.Now)
	best := 0
	for _, d := range p.Year {
		best = max(best, d.Count)
	}
	active := activeDays(p.Year)
	tiles := []struct {
		label string
		value int
		unit  string
	}{
		{"Current streak", current, plural(current, "day", "days")},
		{"Longest streak", longest, plural(longest, "day", "days")},
		{"Active days", active, plural(active, "day", "days")},
		{"Best day", best, plural(best, "contribution", "contributions")},
	}
	for i, tile := range tiles {
		x := 322 + float64(i)*124
		fmt.Fprintf(&b, `<path d="M%s 80V138" stroke="%s"/>`, num(x-18), t.Hairline)
		v := compact(tile.value)
		fmt.Fprintf(&b, `<text class="s" x="%s" y="106" font-size="26" font-weight="600" fill="%s">%s</text>`, num(x), t.Ink, v)
		unit := tile.unit
		if len(unit) > 6 {
			unit = "" // long units only fit in the label
		}
		if unit != "" {
			fmt.Fprintf(&b, `<text class="m" x="%s" y="106" font-size="11.5" fill="%s">%s</text>`,
				num(x+sansWidth(v, 26, 600)+6), t.Muted, esc(unit))
		}
		fmt.Fprintf(&b, `<text class="m" x="%s" y="130" font-size="11.5" fill="%s">%s</text>`, num(x), t.Ink2, esc(tile.label))
	}

	b.WriteString(activityChart(p, t))

	css := `.draw{animation:draw 1.8s cubic-bezier(.3,.6,.2,1) .2s both}@keyframes draw{from{stroke-dashoffset:1}}` +
		`.fade{animation:fade 1s ease .7s both}@keyframes fade{from{opacity:0}}` +
		`.pop{animation:fade .4s ease 1.8s both}`
	title := fmt.Sprintf("%s contributions in the last year; current streak %d %s, longest streak %d %s",
		thousands(total), current, plural(current, "day", "days"), longest, plural(longest, "day", "days"))
	return document(int(W), int(H), title, css, b.String(), mono400, sans600)
}

// activityChart plots weekly contribution totals as a single smoothed series.
func activityChart(p *Profile, t Theme) string {
	const (
		left, right = 64.0, 812.0
		top, base   = 180.0, 256.0
	)
	weeks := weekly(p.Year)
	var b strings.Builder
	if len(weeks) < 2 {
		return ""
	}
	peak := 0
	for i, w := range weeks {
		if w.Count > weeks[peak].Count {
			peak = i
		}
	}
	step := niceStep(float64(weeks[peak].Count) / 2)
	yMax := 2 * step

	// Gridlines and y ticks: 0, step, 2·step.
	for i := 0; i <= 2; i++ {
		y := base - float64(i)/2*(base-top)
		fmt.Fprintf(&b, `<path d="M%s %sH%s" stroke="%s"/>`, num(left), num(y+.5), num(right), t.Hairline)
		fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="10.5" fill="%s" text-anchor="end">%s</text>`,
			num(left-10), num(y+4), t.Muted, compact(int(float64(i)*step)))
	}

	pts := make([]point, len(weeks))
	for i, w := range weeks {
		pts[i] = point{
			left + float64(i)*(right-left)/float64(len(weeks)-1),
			base - float64(w.Count)/yMax*(base-top),
		}
	}

	// Month labels under the weeks in which a month begins.
	lastX := -100.0
	for i, w := range weeks {
		end := w.Start.AddDate(0, 0, 6)
		if end.Day() > 7 || pts[i].x-lastX < 40 || pts[i].x > right-12 {
			continue
		}
		fmt.Fprintf(&b, `<text class="m" x="%s" y="280" font-size="10.5" fill="%s" text-anchor="middle">%s</text>`,
			num(pts[i].x), t.Muted, end.Format("Jan"))
		lastX = pts[i].x
	}

	line := smoothPath(pts)
	accent := legible(t.Accent, t.BgBottom, 3)
	fmt.Fprintf(&b, `<defs><linearGradient id="wash" x1="0" y1="0" x2="0" y2="1">`+
		`<stop offset="0" stop-color="%s" stop-opacity=".2"/><stop offset="1" stop-color="%s" stop-opacity="0"/></linearGradient></defs>`,
		accent, accent)
	fmt.Fprintf(&b, `<path class="fade" d="%s L%s %s L%s %s Z" fill="url(#wash)"/>`, line, num(right), num(base), num(left), num(base))
	fmt.Fprintf(&b, `<path class="draw" d="%s" fill="none" stroke="%s" stroke-width="2" stroke-linejoin="round" stroke-linecap="round" pathLength="1" stroke-dasharray="1 1"/>`,
		line, accent)

	if weeks[peak].Count == 0 {
		fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="12" fill="%s" text-anchor="middle">No public contributions in the last year — yet.</text>`,
			num((left+right)/2), num(top+24), t.Ink2)
		return b.String()
	}

	// Mark and label the peak week only.
	pk := pts[peak]
	label := fmt.Sprintf("Peak week · %s", thousands(weeks[peak].Count))
	lx := math.Min(math.Max(pk.x, left+monoWidth(label, 11)/2), right-monoWidth(label, 11)/2)
	fmt.Fprintf(&b, `<g class="pop"><circle cx="%s" cy="%s" r="4.5" fill="%s" stroke="%s" stroke-width="2"/>`+
		`<text class="m" x="%s" y="%s" font-size="11" fill="%s" text-anchor="middle">%s</text></g>`,
		num(pk.x), num(pk.y), accent, t.BgBottom, num(lx), num(pk.y-12), t.Ink2, esc(label))
	return b.String()
}

// niceStep rounds v up to a whole "round" number — 1, 1.5, 2, 2.5, 3, 4, 5,
// 6 or 8 times a power of ten — so the axis hugs the data without odd ticks.
func niceStep(v float64) float64 {
	if v <= 1 {
		return 1
	}
	mag := math.Pow(10, math.Floor(math.Log10(v)))
	for _, m := range []float64{1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10} {
		if s := m * mag; s >= v && s == math.Trunc(s) {
			return s
		}
	}
	return 10 * mag
}

type point struct{ x, y float64 }

// smoothPath draws a monotone cubic curve through pts (Fritsch–Carlson), so
// the line never overshoots the data or dips below zero.
func smoothPath(pts []point) string {
	n := len(pts)
	d := make([]float64, n-1)
	for i := range d {
		d[i] = (pts[i+1].y - pts[i].y) / (pts[i+1].x - pts[i].x)
	}
	m := make([]float64, n)
	m[0], m[n-1] = d[0], d[n-2]
	for i := 1; i < n-1; i++ {
		if d[i-1]*d[i] > 0 {
			m[i] = (d[i-1] + d[i]) / 2
		}
	}
	for i := range d {
		if d[i] == 0 {
			m[i], m[i+1] = 0, 0
			continue
		}
		a, c := m[i]/d[i], m[i+1]/d[i]
		if s := a*a + c*c; s > 9 {
			tau := 3 / math.Sqrt(s)
			m[i], m[i+1] = tau*a*d[i], tau*c*d[i]
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "M%s %s", num(pts[0].x), num(pts[0].y))
	for i := 0; i < n-1; i++ {
		h := pts[i+1].x - pts[i].x
		fmt.Fprintf(&b, " C%s %s %s %s %s %s",
			num(pts[i].x+h/3), num(pts[i].y+m[i]*h/3),
			num(pts[i+1].x-h/3), num(pts[i+1].y-m[i+1]*h/3),
			num(pts[i+1].x), num(pts[i+1].y))
	}
	return b.String()
}

// renderLanguages is a part-to-whole bar of the languages across the user's
// public repositories, with a legend that carries names and shares.
func renderLanguages(p *Profile, t Theme) string {
	const (
		W, H    = 412.0, 216.0
		x0, x1  = 24.0, 388.0
		barY    = 60.0
		barH    = 10.0
		gap     = 2.0
		maxRows = 8
	)
	var b strings.Builder
	b.WriteString(card(t, W, H))
	fmt.Fprintf(&b, `<text class="s" x="24" y="40" font-size="16" font-weight="600" fill="%s">Languages</text>`, t.Ink)
	fmt.Fprintf(&b, `<text class="m" x="388" y="40" font-size="11" fill="%s" text-anchor="end">by code size</text>`, t.Muted)

	langs := p.Languages
	var total int64
	for _, l := range langs {
		total += l.Bytes
	}
	if total == 0 {
		fmt.Fprintf(&b, `<text class="m" x="206" y="120" font-size="12" fill="%s" text-anchor="middle">No public code to measure yet.</text>`, t.Ink2)
		return document(int(W), int(H), "Languages: none yet", "", b.String(), mono400, sans600)
	}
	if len(langs) > maxRows {
		other := Language{Name: "Other", Color: t.Muted}
		for _, l := range langs[maxRows-1:] {
			other.Bytes += l.Bytes
		}
		langs = append(append([]Language(nil), langs[:maxRows-1]...), other)
	}

	// Bar: segments separated by 2px gaps, clipped to a rounded track.
	avail := x1 - x0 - gap*float64(len(langs)-1)
	fmt.Fprintf(&b, `<clipPath id="bar"><rect x="%s" y="%s" width="%s" height="%s" rx="%s"/></clipPath><g clip-path="url(#bar)">`,
		num(x0), num(barY), num(x1-x0), num(barH), num(barH/2))
	x := x0
	for i, l := range langs {
		w := math.Max(2, float64(l.Bytes)/float64(total)*avail)
		if i == len(langs)-1 {
			w = math.Max(2, x1-x) // absorb rounding so the bar ends flush
		}
		fmt.Fprintf(&b, `<rect class="grow" style="animation-delay:%.2fs" x="%s" y="%s" width="%s" height="%s" fill="%s"/>`,
			0.1+float64(i)*0.07, num(x), num(barY), num(w), num(barH), langColor(l, t))
		x += w + gap
	}
	b.WriteString(`</g>`)

	// Legend: two columns, name on the left, share on the right.
	var names []string
	for i, l := range langs {
		col, row := i/4, i%4
		cx := x0 + float64(col)*190
		y := 104 + float64(row)*28
		share := float64(l.Bytes) / float64(total) * 100
		pct := strconv.FormatFloat(share, 'f', 1, 64) + "%"
		if share < 0.1 {
			pct = "<0.1%"
		}
		name := truncate(l.Name, int((174-18-monoWidth(pct, 12.5)-10)/(12.5*0.6)))
		fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="5" fill="%s"/>`, num(cx+5), num(y-4.5), langColor(l, t))
		fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="12.5" fill="%s">%s</text>`, num(cx+18), num(y), t.Ink, esc(name))
		fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="12.5" fill="%s" text-anchor="end">%s</text>`, num(cx+174), num(y), t.Ink2, pct)
		names = append(names, l.Name+" "+pct)
	}
	css := `.grow{animation:grow .7s cubic-bezier(.2,.7,.2,1) both;transform-box:fill-box}` +
		`@keyframes grow{from{transform:scaleX(0)}}`
	return document(int(W), int(H), "Languages by code size: "+strings.Join(names, ", "), css, b.String(), mono400, sans600)
}

func langColor(l Language, t Theme) string {
	c := l.Color
	if c == "" {
		c = t.Muted
	}
	return legible(c, t.BgBottom, 2)
}

// renderHighlights lists the headline counts, one per row.
func renderHighlights(p *Profile, t Theme) string {
	const W, H = 412.0, 216.0
	var b strings.Builder
	b.WriteString(card(t, W, H))
	fmt.Fprintf(&b, `<text class="s" x="24" y="40" font-size="16" font-weight="600" fill="%s">Highlights</text>`, t.Ink)
	fmt.Fprintf(&b, `<text class="m" x="388" y="40" font-size="11" fill="%s" text-anchor="end">public activity</text>`, t.Muted)

	rows := []struct {
		icon, label string
		value       int
	}{
		{"git-commit", "Commits in the last year", p.Commits},
		{"git-pull-request", "Pull requests opened", p.PullRequests},
		{"issue-opened", "Issues opened", p.Issues},
		{"star", "Stars earned", p.Stars},
		{"repo", "Public repositories", p.PublicRepos},
		{"pulse", fmt.Sprintf("Contributions since %d", p.CreatedAt.Year()), sum(p.AllTime)},
	}
	icon := legible(t.Accent, t.BgBottom, 3)
	var parts []string
	for i, r := range rows {
		y := 76 + float64(i)*23
		if i > 0 {
			fmt.Fprintf(&b, `<path d="M24 %sH388" stroke="%s"/>`, num(y-15.5), t.Hairline)
		}
		b.WriteString(octicon(r.icon, 24, y-11.5, 14, icon))
		fmt.Fprintf(&b, `<text class="m" x="48" y="%s" font-size="12.5" fill="%s">%s</text>`, num(y), t.Ink2, esc(r.label))
		fmt.Fprintf(&b, `<text class="s" x="388" y="%s" font-size="15" font-weight="600" fill="%s" text-anchor="end">%s</text>`,
			num(y+.5), t.Ink, compact(r.value))
		parts = append(parts, fmt.Sprintf("%s: %s", r.label, thousands(r.value)))
	}
	return document(int(W), int(H), "Highlights — "+strings.Join(parts, "; "), "", b.String(), mono400, sans600)
}

// renderProject is one repository card for the projects grid.
func renderProject(r Repo, login string, t Theme) string {
	const W, H = 412.0, 150.0
	var b strings.Builder
	b.WriteString(card(t, W, H))
	fmt.Fprintf(&b, `<defs><linearGradient id="edge" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="%s"/><stop offset=".45" stop-color="%s" stop-opacity="0"/></linearGradient></defs>`,
		legible(t.Accent, t.BgTop, 2), t.Accent2)
	b.WriteString(`<path d="M17 1.5H240" stroke="url(#edge)" stroke-width="2" stroke-linecap="round"/>`)

	name := r.Name
	if !strings.EqualFold(r.Owner, login) && r.Owner != "" {
		name = r.Owner + "/" + r.Name
	}
	maxName := 340.0
	if r.Archived {
		maxName -= 78
	}
	for sansWidth(name, 16, 600) > maxName {
		name = truncate(name, len([]rune(name))-1)
	}
	b.WriteString(octicon("repo", 24, 25, 16, t.Ink2))
	fmt.Fprintf(&b, `<text class="s" x="48" y="38" font-size="16" font-weight="600" fill="%s">%s</text>`, t.Ink, esc(name))
	if r.Archived {
		px := 48 + sansWidth(name, 16, 600) + 10
		fmt.Fprintf(&b, `<rect x="%s" y="24" width="64" height="19" rx="9.5" fill="none" stroke="%s"/>`+
			`<text class="m" x="%s" y="37.5" font-size="10.5" fill="%s" text-anchor="middle">archived</text>`,
			num(px), t.Border, num(px+32), t.Muted)
	}

	desc := strings.TrimSpace(r.Description)
	descColor := t.Ink2
	if desc == "" {
		desc, descColor = "No description yet.", t.Muted
	}
	for i, line := range wrap(desc, 50, 2) {
		fmt.Fprintf(&b, `<text class="m" x="24" y="%s" font-size="12" fill="%s">%s</text>`, num(68+float64(i)*19), descColor, esc(line))
	}

	// Footer: language · stars · forks, and when it was last pushed.
	x := 24.0
	const fy = 126.0
	if r.Language != "" {
		fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="5" fill="%s"/>`, num(x+5), num(fy-4.5), langColor(Language{Color: r.LanguageColor}, t))
		fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="12" fill="%s">%s</text>`, num(x+16), num(fy), t.Ink2, esc(truncate(r.Language, 16)))
		x += 16 + monoWidth(truncate(r.Language, 16), 12) + 18
	}
	for _, s := range []struct {
		icon string
		n    int
	}{{"star", r.Stars}, {"repo-forked", r.Forks}} {
		b.WriteString(octicon(s.icon, x, fy-11.5, 14, t.Muted))
		v := compact(s.n)
		fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="12" fill="%s">%s</text>`, num(x+19), num(fy), t.Ink2, v)
		x += 19 + monoWidth(v, 12) + 18
	}
	if !r.PushedAt.IsZero() {
		fmt.Fprintf(&b, `<text class="m" x="388" y="%s" font-size="11" fill="%s" text-anchor="end">Updated %s</text>`,
			num(fy), t.Muted, r.PushedAt.Format("Jan 2006"))
	}
	title := r.Name
	if strings.TrimSpace(r.Description) != "" {
		title += " — " + r.Description
	}
	return document(int(W), int(H), title, "", b.String(), mono400, sans600)
}

package main

import (
	"fmt"
	"strings"
)

// brand describes how a Simple Icons logo is drawn on each theme.
type brand struct {
	title       string
	dark, light string // logo colour per theme
	plate       string // colour shown through the logo's cut-outs, if any
}

var brands = map[string]brand{
	"go":         {"Go", "#00add8", "#007d9c", ""},
	"typescript": {"TypeScript", "#3178c6", "#3178c6", "#ffffff"},
	"javascript": {"JavaScript", "#f7df1e", "#f7df1e", "#1f2328"},
	"html5":      {"HTML", "#e34f26", "#e34f26", ""},
	"css":        {"CSS", "#a67ce0", "#663399", ""},
	"linux":      {"Linux", "#fcc624", "#1f2328", ""},
	"docker":     {"Docker", "#2496ed", "#1d63ed", ""},
	"git":        {"Git", "#f03c2e", "#f03c2e", ""},
}

// renderStack draws the toolbox: one row of logo tiles, grouped and labelled.
func renderStack(cfg Config, t Theme) string {
	const (
		W, H    = 840.0, 158.0
		tile    = 84.0
		gap     = 12.0
		groupGp = 40.0
		top     = 50.0
	)
	total := 0.0
	for i, g := range cfg.Stack {
		total += float64(len(g.Items))*tile + float64(len(g.Items)-1)*gap
		if i > 0 {
			total += groupGp
		}
	}
	var b strings.Builder
	b.WriteString(card(t, W, H))
	b.WriteString(`<defs>`)
	for slug, br := range stackBrands(cfg) {
		c := br.dark
		if !t.Dark {
			c = br.light
		}
		fmt.Fprintf(&b, `<radialGradient id="glow-%s"><stop offset="0" stop-color="%s" stop-opacity="%.2f"/><stop offset="1" stop-color="%s" stop-opacity="0"/></radialGradient>`,
			slug, c, map[bool]float64{true: .22, false: .14}[t.Dark], c)
	}
	b.WriteString(`</defs>`)

	x := (W - total) / 2
	n := 0
	for gi, g := range cfg.Stack {
		if gi > 0 {
			// A hairline between groups, centred in the gap.
			fmt.Fprintf(&b, `<path d="M%s %sV%s" stroke="%s"/>`, num(x-groupGp/2), num(top+14), num(top+tile-14), t.Hairline)
		}
		fmt.Fprintf(&b, `<text class="m" x="%s" y="34" font-size="11" letter-spacing="1.6" fill="%s">%s</text>`,
			num(x+2), t.Muted, esc(strings.ToUpper(g.Group)))
		for _, slug := range g.Items {
			br := brands[slug]
			c := br.dark
			if !t.Dark {
				c = br.light
			}
			fmt.Fprintf(&b, `<g class="rise" style="animation-delay:%.2fs">`, 0.05+float64(n)*0.06)
			fmt.Fprintf(&b, `<rect x="%s" y="%s" width="%s" height="%s" rx="16" fill="%s" stroke="%s"/>`,
				num(x+.5), num(top+.5), num(tile-1), num(tile-1), t.Tile, t.Border)
			fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="34" fill="url(#glow-%s)"/>`, num(x+tile/2), num(top+33), slug)
			const icon = 30.0
			ix, iy := x+(tile-icon)/2, top+16
			if br.plate != "" {
				// Square logos (JS, TS) are letters cut out of a tile; fill the
				// cut-outs so they read correctly on any background.
				fmt.Fprintf(&b, `<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s"/>`,
					num(ix+icon*0.025), num(iy+icon*0.025), num(icon*0.95), num(icon*0.95), num(icon*0.03), br.plate)
			}
			fmt.Fprintf(&b, `<path transform="translate(%s %s) scale(%s)" d="%s" fill="%s"/>`,
				num(ix), num(iy), num(icon/24), brandPaths[slug], c)
			fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="11.5" fill="%s" text-anchor="middle">%s</text></g>`,
				num(x+tile/2), num(top+tile-13), t.Ink2, esc(br.title))
			x += tile + gap
			n++
		}
		x += groupGp - gap
	}

	css := `.rise{animation:rise .5s cubic-bezier(.2,.7,.2,1) both}` +
		`@keyframes rise{from{opacity:0;transform:translateY(6px)}to{opacity:1;transform:none}}`
	var names []string
	for _, g := range cfg.Stack {
		for _, slug := range g.Items {
			names = append(names, brands[slug].title)
		}
	}
	return document(int(W), int(H), "Toolbox: "+strings.Join(names, ", "), css, b.String(), mono400)
}

// stackBrands yields the stack's brands in config order, each once.
func stackBrands(cfg Config) func(func(string, brand) bool) {
	return func(yield func(string, brand) bool) {
		seen := map[string]bool{}
		for _, g := range cfg.Stack {
			for _, slug := range g.Items {
				if seen[slug] {
					continue
				}
				seen[slug] = true
				if !yield(slug, brands[slug]) {
					return
				}
			}
		}
	}
}

// renderContact draws the call-to-action button that links to email.
func renderContact(cfg Config, t Theme) string {
	label := cfg.Contact.Label
	if label == "" {
		label = "Say hello"
	}
	const H = 48.0
	tw := sansWidth(label, 16, 600)
	W := 22 + 18 + 12 + tw + 14 + 16 + 22
	var b strings.Builder
	fmt.Fprintf(&b, `<defs><linearGradient id="btn" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient></defs>`,
		t.Accent, t.Accent2)
	fill, ink := "url(#btn)", "#04130d"
	if !t.Dark {
		fill, ink = t.Ink, "#ffffff"
	}
	fmt.Fprintf(&b, `<rect x="1" y="1" width="%s" height="%s" rx="%s" fill="%s"/>`, num(W-2), num(H-2), num((H-2)/2), fill)
	b.WriteString(octicon("mail", 22, 15, 18, ink))
	fmt.Fprintf(&b, `<text class="s" x="52" y="29.5" font-size="16" font-weight="600" fill="%s">%s</text>`, ink, esc(label))
	b.WriteString(`<g class="nudge">`)
	b.WriteString(octicon("arrow-right", 52+tw+12, 16, 16, ink))
	b.WriteString(`</g>`)
	css := `.nudge{animation:nudge 2.4s ease-in-out infinite}@keyframes nudge{0%,60%,100%{transform:none}30%{transform:translateX(3px)}}`
	return document(int(W+.5), int(H), label, css, b.String(), sans600)
}

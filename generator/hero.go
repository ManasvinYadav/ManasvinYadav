package main

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// renderHero draws the banner at the top of the README: name, tagline and
// chips on the left, a terminal that "runs" the profile on the right.
func renderHero(cfg Config, t Theme) string {
	const W, H = 840, 280
	var b strings.Builder

	bgFrom, bgTo := "#0c1118", "#101822"
	shadow := `<feDropShadow dx="0" dy="14" stdDeviation="16" flood-color="#000" flood-opacity=".55"/>`
	if !t.Dark {
		bgFrom, bgTo = "#ffffff", "#f2f5f8"
		shadow = `<feDropShadow dx="0" dy="12" stdDeviation="14" flood-color="#1f2328" flood-opacity=".22"/>`
	}
	accentText := legible(t.Accent, bgFrom, 4.5)

	fmt.Fprintf(&b, `<defs>`+
		`<linearGradient id="hbg" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`+
		`<clipPath id="frame"><rect width="%d" height="%d" rx="20"/></clipPath>`+
		`<radialGradient id="g1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s" stop-opacity="0"/></radialGradient>`+
		`<radialGradient id="g2"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s" stop-opacity="0"/></radialGradient>`+
		`<radialGradient id="g3"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s" stop-opacity="0"/></radialGradient>`+
		`<pattern id="dots" width="22" height="22" patternUnits="userSpaceOnUse"><circle cx="1.5" cy="1.5" r="1.1" fill="%s" fill-opacity="%.2f"/></pattern>`+
		`<radialGradient id="fadeg" cx=".72" cy=".35" r=".75"><stop offset="0" stop-color="#fff"/><stop offset="1" stop-color="#fff" stop-opacity="0"/></radialGradient>`+
		`<mask id="fade"><rect width="%d" height="%d" fill="url(#fadeg)"/></mask>`+
		`<linearGradient id="name" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="%s"/><stop offset=".52" stop-color="%s"/><stop offset=".8" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`+
		`<filter id="shadow" x="-30%%" y="-30%%" width="160%%" height="170%%">%s</filter>`+
		`</defs>`,
		bgFrom, bgTo, W, H,
		t.Accent, t.Accent, t.Accent2, t.Accent2, t.Accent3, t.Accent3,
		t.Ink, t.Dots, W, H,
		t.Ink, t.Ink, legible(t.Accent, bgFrom, 3), legible(t.Accent2, bgFrom, 3),
		shadow)

	// Backdrop: gradient, drifting aurora, a dot grid that fades out to the left.
	fmt.Fprintf(&b, `<rect width="%d" height="%d" rx="20" fill="url(#hbg)"/>`, W, H)
	fmt.Fprintf(&b, `<g clip-path="url(#frame)"><g fill-opacity="%.2f">`+
		`<circle class="a1" cx="620" cy="10" r="260" fill="url(#g1)"/>`+
		`<circle class="a2" cx="820" cy="260" r="220" fill="url(#g2)"/>`+
		`<circle class="a3" cx="360" cy="310" r="240" fill="url(#g3)"/></g>`+
		`<rect width="%d" height="%d" fill="url(#dots)" mask="url(#fade)"/></g>`,
		t.Glow, W, H)
	fmt.Fprintf(&b, `<rect x=".5" y=".5" width="%d" height="%d" rx="19.5" fill="none" stroke="%s"/>`, W-1, H-1, t.Border)

	// Identity.
	const left = 44.0
	fmt.Fprintf(&b, `<text class="m" x="%s" y="74" font-size="13"><tspan fill="%s">%s</tspan><tspan fill="%s"> ❯ whoami</tspan></text>`,
		num(left), accentText, esc(cfg.Hero.Prompt), t.Ink2)

	nameSize := 52.0
	if w := sansWidth(cfg.Name, nameSize, 700) - 1.2*float64(utf8.RuneCountInString(cfg.Name)); w > 420 {
		nameSize *= 420 / w
	}
	fmt.Fprintf(&b, `<text class="s" x="%s" y="134" font-size="%s" font-weight="700" letter-spacing="-1.2" fill="url(#name)">%s</text>`,
		num(left-2), num(nameSize), esc(cfg.Name))
	fmt.Fprintf(&b, `<text class="m" x="%s" y="170" font-size="14" fill="%s">%s</text>`,
		num(left), t.Ink2, esc(truncate(cfg.Hero.Tagline, 50)))

	// Chips.
	x := left
	for i, chip := range cfg.Hero.Chips {
		tw := monoWidth(chip.Label, 12)
		w := 24 + tw + 13
		if chip.Icon != "" {
			w = 28 + tw + 13
		}
		if x+w > 468 {
			break
		}
		fmt.Fprintf(&b, `<g class="rise" style="animation-delay:%.2fs">`, 0.15+float64(i)*0.08)
		fmt.Fprintf(&b, `<rect x="%s" y="196" width="%s" height="30" rx="15" fill="%s" fill-opacity=".8" stroke="%s"/>`,
			num(x), num(w), t.Tile, t.Border)
		textX := x + 24
		if chip.Icon != "" {
			b.WriteString(octicon(chip.Icon, x+11, 204, 13, t.Ink2))
			textX = x + 28
		} else {
			fmt.Fprintf(&b, `<circle cx="%s" cy="211" r="4" fill="%s"/>`, num(x+14), legible(chip.Color, t.Tile, 2.2))
		}
		fmt.Fprintf(&b, `<text class="m" x="%s" y="215.2" font-size="12" fill="%s">%s</text></g>`, num(textX), t.Ink, esc(chip.Label))
		x += w + 8
	}

	b.WriteString(terminal(cfg, t))

	css := `.a1{animation:d1 15s ease-in-out infinite alternate}` +
		`.a2{animation:d2 19s ease-in-out infinite alternate}` +
		`.a3{animation:d3 23s ease-in-out infinite alternate}` +
		`@keyframes d1{to{transform:translate(-70px,34px)}}` +
		`@keyframes d2{to{transform:translate(-46px,-30px)}}` +
		`@keyframes d3{to{transform:translate(80px,-24px)}}` +
		`.rise{animation:rise .5s cubic-bezier(.2,.7,.2,1) both}` +
		`@keyframes rise{from{opacity:0;transform:translateY(5px)}to{opacity:1;transform:none}}` +
		`.blink{animation:blink 1.1s steps(1) infinite}` +
		`@keyframes blink{50%{opacity:0}}` +
		`@media (prefers-reduced-motion:reduce){.type{display:none}}`

	title := cfg.Name + " — " + cfg.Hero.Tagline
	return document(W, H, title, css, b.String(), mono400, mono700, sans700)
}

// terminal draws the window on the right of the hero. The command is typed
// out by sliding a cover off it, then the output lines rise in one by one.
func terminal(cfg Config, t Theme) string {
	const (
		x, y, w, h = 492.0, 46.0, 308.0, 188.0
		size       = 12.5
		ch         = size * 0.6
		pad        = 16.0
		lineH      = 22.0
	)
	border := "#252e3b"
	if !t.Dark {
		border = "#1c232d"
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<clipPath id="term"><rect x="%s" y="%s" width="%s" height="%s" rx="12"/></clipPath>`, num(x), num(y), num(w), num(h))
	fmt.Fprintf(&b, `<rect x="%s" y="%s" width="%s" height="%s" rx="12" fill="%s" filter="url(#shadow)"/>`, num(x), num(y), num(w), num(h), termBg)
	fmt.Fprintf(&b, `<g clip-path="url(#term)"><rect x="%s" y="%s" width="%s" height="30" fill="%s"/>`, num(x), num(y), num(w), termBar)
	fmt.Fprintf(&b, `<path d="M%s %sH%s" stroke="#1b232d"/>`, num(x), num(y+30.5), num(x+w))
	for i, c := range []string{"#ff5f57", "#febc2e", "#28c840"} {
		fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="5.5" fill="%s"/>`, num(x+18+float64(i)*16), num(y+15), c)
	}
	fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="11" fill="%s" text-anchor="middle">%s</text>`,
		num(x+w/2), num(y+19.5), termDim, esc(cfg.Hero.Terminal.Title))

	maxChars := int(math.Floor((w - 2*pad) / ch))
	baseline := y + 58
	textX := x + pad

	// Line 1: the command, typed.
	cmd := truncate(cfg.Hero.Terminal.Command, maxChars-2)
	cmdX := textX + 2*ch
	cmdW := monoWidth(cmd, size)
	steps := utf8.RuneCountInString(cmd)
	fmt.Fprintf(&b, `<text class="m" x="%s" y="%s" font-size="%s" fill="%s">❯ <tspan fill="%s">%s</tspan></text>`,
		num(textX), num(baseline), num(size), termAccent, termText, esc(cmd))
	fmt.Fprintf(&b, `<g class="type" style="animation:type %.2fs steps(%d,end) .5s both">`+
		`<rect x="%s" y="%s" width="%s" height="18" fill="%s"/>`+
		`<rect class="caret" x="%s" y="%s" width="%s" height="15" fill="%s"/></g>`,
		0.07*float64(steps), steps,
		num(cmdX), num(baseline-13), num(cmdW+ch+2), termBg,
		num(cmdX), num(baseline-11.5), num(ch), termAccent)
	typed := 0.5 + 0.07*float64(steps)
	b.WriteString(`</g>`)

	// Output lines: symbol · key · value, with the value dimmed.
	delay := typed + 0.3
	for i, line := range cfg.Hero.Terminal.Output {
		if i >= 4 {
			break
		}
		line = truncate(line, maxChars-2)
		sym, rest, _ := strings.Cut(line, " ")
		symColor := termAccent
		if sym == "→" {
			symColor = termCyan
		}
		key, val := rest, ""
		if i := strings.Index(rest, "  "); i > 0 {
			key, val = rest[:i], rest[i:]
		}
		fmt.Fprintf(&b, `<g class="rise" style="animation-delay:%.2fs"><text class="m" x="%s" y="%s" font-size="%s" xml:space="preserve">`+
			`<tspan fill="%s">%s</tspan> <tspan fill="%s">%s</tspan><tspan fill="%s">%s</tspan></text></g>`,
			delay, num(textX+2*ch), num(baseline+lineH*float64(i+1)), num(size),
			symColor, esc(sym), termText, esc(key), termDim, esc(val))
		delay += 0.3
	}

	// Final prompt with a blinking cursor.
	last := baseline + lineH*float64(min(len(cfg.Hero.Terminal.Output), 4)+1)
	fmt.Fprintf(&b, `<g class="rise" style="animation-delay:%.2fs"><text class="m" x="%s" y="%s" font-size="%s" fill="%s">❯</text>`+
		`<rect class="blink" x="%s" y="%s" width="%s" height="15" fill="%s"/></g>`,
		delay+0.1, num(textX), num(last), num(size), termAccent,
		num(textX+2*ch), num(last-11.5), num(ch), termAccent)

	fmt.Fprintf(&b, `<rect x="%s" y="%s" width="%s" height="%s" rx="11.5" fill="none" stroke="%s"/>`,
		num(x+.5), num(y+.5), num(w-1), num(h-1), border)

	// The cover slides right by the command's width; the caret rides its left
	// edge and disappears once typing is done.
	fmt.Fprintf(&b, `<style>@keyframes type{to{transform:translateX(%spx)}}`+
		`.caret{animation:blink .5s steps(1) infinite,gone .01s linear %.2fs forwards}`+
		`@keyframes gone{to{opacity:0}}</style>`, num(cmdW), typed)
	return b.String()
}

package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Theme is one colour scheme. Every card is rendered once per theme and the
// README picks the right file with <picture> + prefers-color-scheme.
type Theme struct {
	Name string
	Dark bool

	BgTop, BgBottom string // card background gradient
	Border          string
	Tile            string // raised surfaces inside a card

	Ink      string // primary text
	Ink2     string // secondary text
	Muted    string // captions, axis labels
	Hairline string // gridlines, dividers
	Track    string // empty part of a bar

	Accent  string // mint — the signature colour
	Accent2 string // cyan
	Accent3 string // violet, only ever used as a soft glow

	Glow float64 // opacity of the hero aurora
	Dots float64 // opacity of the hero dot grid
}

var themes = []Theme{
	{
		Name: "dark", Dark: true,
		BgTop: "#121922", BgBottom: "#0d1117", Border: "#232c37", Tile: "#151c25",
		Ink: "#e6edf3", Ink2: "#a3adb8", Muted: "#7d8793", Hairline: "#1f2731", Track: "#1c242e",
		Accent: "#00ffa3", Accent2: "#00e5ff", Accent3: "#a78bfa",
		Glow: 0.26, Dots: 0.07,
	},
	{
		Name: "light", Dark: false,
		BgTop: "#ffffff", BgBottom: "#f6f8fa", Border: "#d8dee4", Tile: "#f6f8fa",
		Ink: "#1f2328", Ink2: "#4c5561", Muted: "#6b7480", Hairline: "#e8ecf0", Track: "#eaeef2",
		Accent: "#00955f", Accent2: "#0886a8", Accent3: "#7c3aed",
		Glow: 0.16, Dots: 0.08,
	},
}

// The hero terminal is dark in both themes, like a code block on a landing page.
const (
	termBg     = "#0b0f15"
	termBar    = "#10161e"
	termText   = "#d7dee6"
	termDim    = "#7d8793"
	termAccent = "#00ffa3"
	termCyan   = "#00e5ff"
)

type rgb struct{ r, g, b float64 }

func parseHex(s string) rgb {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil || len(s) != 6 {
		return rgb{0.5, 0.5, 0.5}
	}
	return rgb{float64(v>>16&0xff) / 255, float64(v>>8&0xff) / 255, float64(v&0xff) / 255}
}

func (c rgb) hex() string {
	to := func(f float64) int { return int(math.Round(math.Max(0, math.Min(1, f)) * 255)) }
	return fmt.Sprintf("#%02x%02x%02x", to(c.r), to(c.g), to(c.b))
}

func (c rgb) luminance() float64 {
	lin := func(f float64) float64 {
		if f <= 0.03928 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.r) + 0.7152*lin(c.g) + 0.0722*lin(c.b)
}

func contrast(a, b string) float64 {
	la, lb := parseHex(a).luminance(), parseHex(b).luminance()
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func mix(a, b string, t float64) string {
	x, y := parseHex(a), parseHex(b)
	return rgb{x.r + (y.r-x.r)*t, x.g + (y.g-x.g)*t, x.b + (y.b-x.b)*t}.hex()
}

// legible nudges fg toward white (on dark backgrounds) or black (on light
// ones) until it reaches the WCAG contrast ratio min against bg. Hue is kept,
// so a language or brand colour stays recognisable.
func legible(fg, bg string, min float64) string {
	target := "#000000"
	if parseHex(bg).luminance() < 0.5 {
		target = "#ffffff"
	}
	out := fg
	for t := 0.0; t <= 1 && contrast(out, bg) < min; t += 0.04 {
		out = mix(fg, target, t)
	}
	return out
}

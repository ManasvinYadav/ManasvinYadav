package main

import (
	"embed"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

//go:embed fonts/*.woff2
var fontFiles embed.FS

// Fonts are subset to Latin + a handful of symbols and inlined as data URIs:
// an SVG shown through <img> may not fetch anything, so this is the only way
// to get the same typography (and the same text widths) on every OS.
type face struct {
	family string
	weight int
	file   string
}

var (
	mono400 = face{"JBM", 400, "fonts/jetbrains-mono-400.woff2"}
	mono700 = face{"JBM", 700, "fonts/jetbrains-mono-700.woff2"}
	sans600 = face{"SG", 600, "fonts/space-grotesk-600.woff2"}
	sans700 = face{"SG", 700, "fonts/space-grotesk-700.woff2"}
)

const baseCSS = `.m{font-family:'JBM',ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}` +
	`.s{font-family:'SG','Segoe UI',system-ui,-apple-system,Helvetica,Arial,sans-serif}` +
	`@media (prefers-reduced-motion:reduce){*{animation:none!important}}`

func fontCSS(faces ...face) string {
	var b strings.Builder
	for _, f := range faces {
		data, err := fontFiles.ReadFile(f.file)
		if err != nil {
			panic(err) // embedded at build time; cannot be missing
		}
		fmt.Fprintf(&b, "@font-face{font-family:'%s';font-weight:%d;src:url(data:font/woff2;base64,%s) format('woff2')}",
			f.family, f.weight, base64.StdEncoding.EncodeToString(data))
	}
	return b.String()
}

// document wraps body in a standalone, accessible SVG.
func document(w, h int, title, css, body string, faces ...face) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-labelledby="title">`+
		`<title id="title">%s</title><style>%s%s%s</style>%s</svg>`+"\n",
		w, h, w, h, esc(title), fontCSS(faces...), baseCSS, css, body)
}

// card draws the shared rounded panel every card sits on.
func card(t Theme, w, h float64) string {
	return fmt.Sprintf(`<defs><linearGradient id="bg" x1="0" y1="0" x2="0" y2="1">`+
		`<stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient></defs>`+
		`<rect x=".5" y=".5" width="%s" height="%s" rx="16" fill="url(#bg)" stroke="%s"/>`,
		t.BgTop, t.BgBottom, num(w-1), num(h-1), t.Border)
}

// octicon places a 16×16 octicon with its top-left corner at (x, y).
func octicon(name string, x, y, size float64, fill string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<g transform="translate(%s %s) scale(%s)" fill="%s">`, num(x), num(y), num(size/16), fill)
	for _, d := range octicons[name] {
		fmt.Fprintf(&b, `<path d="%s"/>`, d)
	}
	b.WriteString(`</g>`)
	return b.String()
}

var xmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")

func esc(s string) string { return xmlEscaper.Replace(s) }

// num prints a coordinate with at most two decimals and no trailing zeros.
func num(f float64) string {
	return strconv.FormatFloat(float64(int64(f*100+sign(f)*0.5))/100, 'f', -1, 64)
}

func sign(f float64) float64 {
	if f < 0 {
		return -1
	}
	return 1
}

// monoWidth is exact for JetBrains Mono (every glyph is 0.6em); characters
// outside the subset fall back to a system font and are assumed to be wide.
func monoWidth(s string, size float64) float64 {
	w := 0.0
	for _, r := range s {
		if r < 0x2e80 {
			w += 0.6
		} else {
			w += 1.0
		}
	}
	return w * size
}

// sansWidth measures Space Grotesk text from its advance-width table.
func sansWidth(s string, size float64, weight int) float64 {
	table := spaceGrotesk600
	if weight >= 700 {
		table = spaceGrotesk700
	}
	units := 0
	for _, r := range s {
		if adv, ok := table[r]; ok {
			units += adv
		} else {
			units += 600
		}
	}
	return float64(units) / 1000 * size
}

// truncate shortens s to at most max runes, ending in an ellipsis if cut.
func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return strings.TrimRight(string(r[:max-1]), " ,.;:-") + "…"
}

// wrap breaks s into at most lines lines of at most width runes each,
// word by word, with an ellipsis if the text does not fit.
func wrap(s string, width, lines int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	if lines <= 1 {
		return []string{truncate(strings.Join(words, " "), width)}
	}
	var out []string
	cur := words[0]
	for i := 1; i < len(words); i++ {
		if utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(words[i]) <= width {
			cur += " " + words[i]
			continue
		}
		out = append(out, truncate(cur, width))
		if len(out) == lines-1 {
			// Whatever is left becomes the last line.
			return append(out, truncate(strings.Join(words[i:], " "), width))
		}
		cur = words[i]
	}
	return append(out, truncate(cur, width))
}

// compact formats a count the way a stat tile should: 1,284 · 12.9K · 4.2M.
func compact(n int) string {
	switch {
	case n >= 1_000_000:
		return trimZero(fmt.Sprintf("%.1f", float64(n)/1e6)) + "M"
	case n >= 10_000:
		return trimZero(fmt.Sprintf("%.1f", float64(n)/1e3)) + "K"
	default:
		return thousands(n)
	}
}

func trimZero(s string) string { return strings.TrimSuffix(s, ".0") }

func thousands(n int) string {
	s := strconv.Itoa(n)
	if n < 0 {
		return "-" + thousands(-n)
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

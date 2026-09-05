// Package tui implements cloudlume's Bubble Tea dashboard, styled to match
// the serverlume family look: rounded cards on a dark background, a
// gradient wordmark, status pills, and gradient gauges — see the
// serverlume TUI style guide for the source conventions.
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	colorful "github.com/lucasb-eyer/go-colorful"

	"github.com/steamedeo/cloudlume/internal/model"
)

// cloudlume's identity is a sky-blue + violet accent pair, distinct from
// serverlume's pink+lavender but built from the same recipe: one primary,
// one secondary, blended in LUV space for gradients.
const (
	hexSky    = "#5fd7ff" // primary accent
	hexViolet = "#8b7cff" // secondary accent

	hexBgPage   = "#0b0e14"
	hexBgCard   = "#121722"
	hexBgCardHi = "#1a2130"
	hexBorder   = "#2a3245"
	hexBorderHi = "#5fd7ff"

	hexText = "#eef2fb"
	hexDim  = "#7c88a6"

	hexGreen = "#4ade80"
	hexAmber = "#fbbf24"
	hexRed   = "#fb7185"
)

var (
	colSky    = lipgloss.Color(hexSky)
	colViolet = lipgloss.Color(hexViolet)

	colBgPage   = lipgloss.Color(hexBgPage)
	colBgCard   = lipgloss.Color(hexBgCard)
	colBgCardHi = lipgloss.Color(hexBgCardHi)
	colBorder   = lipgloss.Color(hexBorder)
	colBorderHi = lipgloss.Color(hexBorderHi)

	colText = lipgloss.Color(hexText)
	colDim  = lipgloss.Color(hexDim)

	colGreen = lipgloss.Color(hexGreen)
	colAmber = lipgloss.Color(hexAmber)
	colRed   = lipgloss.Color(hexRed)
)

// healthColor maps a model.Health to its status color.
func healthColor(h model.Health) lipgloss.Color {
	switch h {
	case model.HealthOK:
		return colGreen
	case model.HealthWarn:
		return colAmber
	case model.HealthDown:
		return colRed
	default:
		return colDim
	}
}

var (
	pageStyle = lipgloss.NewStyle().Background(colBgPage).Foreground(colText)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder).
			Background(colBgCard).
			Padding(0, 1)

	cardStyleFocused = cardStyle.BorderForeground(colBorderHi)
)

// cardBorderSize is how many extra columns/rows a cardStyle box occupies
// on screen beyond the width/height passed to Style.Width()/Height() —
// lipgloss adds border characters on top of that content-box size (1 char
// each on the left/right and top/bottom, since RoundedBorder draws all
// four sides). Anything laying two cards side by side against a known
// terminal width/height must add this back in, or the combined line ends
// up wider than the terminal and gets wrapped by the outer page style.
const cardBorderSize = 2

// cardPaddingX is the total horizontal padding cardStyle applies inside
// its border (Padding(0, 1) — none top/bottom, 1 each side). Style.Width()
// covers the padded content box, so text placed inside must subtract this,
// not the border, to get the space actually available for characters.
const cardPaddingX = 2

var (
	dimStyle  = lipgloss.NewStyle().Foreground(colDim).Background(colBgCard)
	textStyle = lipgloss.NewStyle().Foreground(colText).Background(colBgCard)
	boldStyle = lipgloss.NewStyle().Foreground(colText).Bold(true).Background(colBgCard)

	footerDimStyle = lipgloss.NewStyle().Foreground(colDim).Background(colBgPage)
)

// blend interpolates two hex colors in perceptual (LUV) space, which is
// what keeps cloudlume's gradients smooth instead of muddy like a naive
// linear RGB lerp.
func blend(hexA, hexB string, t float64) colorful.Color {
	a, _ := colorful.Hex(hexA)
	b, _ := colorful.Hex(hexB)
	return a.BlendLuv(b, t)
}

// gradientText renders each character of s blended from hexA to hexB
// across its length, bolded — used for wordmarks and rule characters.
func gradientText(s string, hexA, hexB string, bold bool) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	var b strings.Builder
	for i, r := range runes {
		t := 0.0
		if len(runes) > 1 {
			t = float64(i) / float64(len(runes)-1)
		}
		c := blend(hexA, hexB, t)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex())).Bold(bold)
		b.WriteString(style.Render(string(r)))
	}
	return b.String()
}

// gradientRule renders a full-width horizontal rule blended from hexA to
// hexB.
func gradientRule(width int, hexA, hexB string) string {
	return gradientText(strings.Repeat("─", width), hexA, hexB, false)
}

// pill renders a padded, background-filled label — the page background
// color is used as the foreground so the pill "punches through" rather
// than reading as light-on-dark outline text.
func pill(label string, bg lipgloss.Color, bold bool) string {
	return lipgloss.NewStyle().
		Background(bg).
		Foreground(colBgPage).
		Bold(bold).
		Padding(0, 1).
		Render(label)
}

// healthPill renders a short status label pilled in its health color.
func healthPill(label string, h model.Health) string {
	return pill(strings.ToUpper(label), healthColor(h), true)
}

// gaugeColors picks the two gradient endpoint colors for a gauge based on
// how "hot" its fill ratio is, so color communicates severity rather than
// only bar length.
func gaugeColors(ratio float64) (string, string) {
	switch {
	case ratio < 0.6:
		return hexGreen, hexViolet
	case ratio < 0.85:
		return hexViolet, hexAmber
	default:
		return hexSky, hexRed
	}
}

// progressBar renders a horizontal gradient bar of width cells, filled to
// ratio (0..1) with the sky->violet accent gradient, and the unfilled
// remainder shown as a dim dashed track — used for the startup loading
// screen, where "more filled" always means "closer to done" rather than
// signaling severity (contrast gaugeColors, used for utilization gauges).
func progressBar(width int, ratio float64) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(float64(width) * ratio)

	var b strings.Builder
	for i := 0; i < filled; i++ {
		t := 0.0
		if filled > 1 {
			t = float64(i) / float64(filled-1)
		}
		c := blend(hexSky, hexViolet, t)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex())).Render("█"))
	}
	if width-filled > 0 {
		track := lipgloss.NewStyle().Foreground(colDim).Render(strings.Repeat("╌", width-filled))
		b.WriteString(track)
	}
	return b.String()
}

// padLine pads a rendered (possibly ANSI-styled) line to width with a
// backgrounded blank fill, so mixed content never lets a partial redraw
// show through as flicker.
func padLine(s string, width int, bg lipgloss.Color) string {
	visible := lipgloss.Width(s)
	if visible >= width {
		return s
	}
	filler := lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", width-visible))
	return s + filler
}

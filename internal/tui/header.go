package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/steamedeo/cloudlume/internal/model"
)

func (m Model) renderHeader() string {
	wordmark := gradientText("cloudlume", hexSky, hexViolet, true)
	tagline := lipgloss.NewStyle().Foreground(colDim).Background(colBgPage).Render(" · local cloud resources, live")

	ok, warn, down := 0, 0, 0
	for _, st := range m.statuses {
		switch st.Health {
		case model.HealthOK:
			ok++
		case model.HealthWarn:
			warn++
		case model.HealthDown:
			down++
		}
	}

	var pills []string
	if ok > 0 {
		pills = append(pills, pill(fmt.Sprintf("%d UP", ok), colGreen, true))
	}
	if warn > 0 {
		pills = append(pills, pill(fmt.Sprintf("%d SYNCING", warn), colAmber, true))
	}
	if down > 0 {
		pills = append(pills, pill(fmt.Sprintf("%d ERROR", down), colRed, true))
	}
	if m.paused {
		pills = append(pills, pill("PAUSED", colViolet, true))
	}
	status := strings.Join(pills, " ")

	clock := lipgloss.NewStyle().Foreground(colText).Background(colBgPage).Render(time.Now().Format("15:04:05"))

	left := wordmark + tagline
	right := status
	if right != "" {
		right += "  "
	}
	right += clock

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	// If there's no room for both sides, drop the right side (status
	// pills/clock are decorative) rather than let the line run wider than
	// the terminal — an overlong line gets soft-wrapped by the outer page
	// style, which corrupts the whole layout below it, not just this row.
	var line string
	if leftW+1+rightW > m.width {
		line = padLine(left, m.width, colBgPage)
	} else {
		gap := m.width - leftW - rightW
		line = left + lipgloss.NewStyle().Background(colBgPage).Render(strings.Repeat(" ", gap)) + right
	}
	line = padLine(line, m.width, colBgPage)

	rule := gradientRule(m.width, hexSky, hexViolet)

	return line + "\n" + rule
}

func (m Model) renderFooter() string {
	legend := "↑/↓ select   ←/→ panel   enter details   / filter   tab/1-5 category   p pause   r refresh   q quit"
	if m.filtering {
		legend = "type to search   enter confirm   esc cancel"
	}
	left := footerDimStyle.Render(legend)
	right := footerDimStyle.Render(fmt.Sprintf("cloudlume %s", appVersion))

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	// Same overflow guard as the header: drop the version tag rather than
	// let the line exceed the terminal width and get soft-wrapped.
	if leftW+1+rightW > m.width {
		return padLine(left, m.width, colBgPage)
	}
	gap := m.width - leftW - rightW
	line := left + lipgloss.NewStyle().Background(colBgPage).Render(strings.Repeat(" ", gap)) + right
	return padLine(line, m.width, colBgPage)
}

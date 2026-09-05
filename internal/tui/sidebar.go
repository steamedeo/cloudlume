package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/steamedeo/cloudlume/internal/model"
)

const sidebarWidth = 28

func (m Model) renderSidebar(height int) string {
	style := cardStyle
	if m.focus == focusSidebar {
		style = cardStyleFocused
	}

	var rows []string
	rows = append(rows, boldStyle.Render("Accounts"), "")

	totalResources := 0
	for _, rs := range m.byAccount {
		totalResources += len(rs)
	}
	rows = append(rows, m.sidebarRow(sidebarAll, model.HealthOK, totalResources, m.sidebarIndex == 0))

	for i, name := range m.sortedAccountNames() {
		st := m.statuses[name]
		health := model.HealthUnknown
		count := 0
		if st != nil {
			health = st.Health
			count = st.ResourceCount
		}
		rows = append(rows, m.sidebarRow(name, health, count, m.sidebarIndex == i+1))
	}

	// Never emit more lines than the card has room for — an overflowing
	// sidebar would wrap/spill past its box and break the side-by-side
	// join with the main panel.
	if len(rows) > height {
		rows = rows[:height]
	}

	content := strings.Join(rows, "\n")
	return style.Width(sidebarWidth).Height(height).Render(content)
}

func (m Model) sidebarRow(name string, health model.Health, count int, selected bool) string {
	// One indicator column + one space, dot + space, then name/pad/count.
	inner := sidebarWidth - cardPaddingX // usable width inside the card's padding
	const indicatorWidth = 2             // "▎ "

	countRaw := fmt.Sprintf("%d", count)
	nameWidth := inner - indicatorWidth - 2 - len(countRaw) - 1
	if nameWidth < 1 {
		nameWidth = 1
	}
	name = truncate(name, nameWidth)

	bg := colBgCard
	nameColor := colText
	if selected {
		bg = colBgCardHi
		nameColor = colSky
	}

	indicator := lipgloss.NewStyle().Foreground(bg).Background(bg).Render("▎")
	if selected {
		indicator = lipgloss.NewStyle().Foreground(colSky).Background(bg).Render("▎")
	}

	dot := lipgloss.NewStyle().Foreground(healthColor(health)).Background(bg).Render("●")
	nameRendered := lipgloss.NewStyle().Foreground(nameColor).Background(bg).Bold(selected).Render(name)
	label := fmt.Sprintf("%s %s %s", indicator, dot, nameRendered)

	countStr := lipgloss.NewStyle().Foreground(colDim).Background(bg).Render(countRaw)

	pad := inner - lipgloss.Width(label) - lipgloss.Width(countStr)
	if pad < 1 {
		pad = 1
	}
	padding := lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", pad))

	return label + padding + countStr
}

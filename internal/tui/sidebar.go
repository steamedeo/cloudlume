package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/steamedeo/cloudlume/internal/model"
)

const sidebarWidth = 28

func (m Model) renderSidebar(height int) string {
	focused := true // sidebar navigation is always live via left/right
	style := cardStyle
	if focused && m.sidebarIndex >= 0 {
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

	content := strings.Join(rows, "\n")
	return style.Width(sidebarWidth).Height(height).Render(content)
}

func (m Model) sidebarRow(name string, health model.Health, count int, selected bool) string {
	dot := lipgloss.NewStyle().Foreground(healthColor(health)).Render("●")
	label := fmt.Sprintf("%s %s", dot, name)

	countStr := dimStyle.Render(fmt.Sprintf("%d", count))

	inner := sidebarWidth - 4 // card padding/border
	pad := inner - lipgloss.Width(label) - lipgloss.Width(countStr)
	if pad < 1 {
		pad = 1
	}
	line := label + strings.Repeat(" ", pad) + countStr

	rowStyle := lipgloss.NewStyle().Background(colBgCard)
	if selected {
		rowStyle = lipgloss.NewStyle().Background(colBgCardHi).Bold(true)
	}
	return rowStyle.Width(inner + 2).Render(" " + line)
}

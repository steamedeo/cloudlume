package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "starting cloudlume...\n"
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	sidebar := m.renderSidebar(bodyHeight)
	mainWidth := m.width - sidebarWidth - 1
	if mainWidth < 20 {
		mainWidth = 20
	}
	main := m.renderMain(mainWidth, bodyHeight)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", main)

	return pageStyle.Width(m.width).Render(strings.Join([]string{header, body, footer}, "\n"))
}

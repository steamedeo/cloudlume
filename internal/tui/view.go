package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "starting cloudlume...\n"
	}

	if m.isLoading() {
		return m.renderLoading()
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	// Sidebar/main are cardStyle boxes: Width()/Height() set the content
	// box, and the border adds cardBorderSize on top of that in each
	// dimension. Both cards render at the same requested height, so the
	// on-screen row they occupy is (contentHeight + cardBorderSize) tall —
	// subtract that here so the two cards plus header/footer fit exactly
	// inside m.height, or the outer page wraps and corrupts the layout.
	availableHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	contentHeight := availableHeight - cardBorderSize
	if contentHeight < 5 {
		contentHeight = 5
	}

	sidebar := m.renderSidebar(contentHeight)

	// Likewise for width: the sidebar renders at sidebarWidth+cardBorderSize
	// on screen, then a 1-column gap, then the main card at
	// mainWidth+cardBorderSize — all three must sum to m.width.
	mainWidth := m.width - (sidebarWidth + cardBorderSize) - 1 - cardBorderSize
	if mainWidth < 20 {
		mainWidth = 20
	}
	main := m.renderMain(mainWidth, contentHeight)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", main)

	return pageStyle.Width(m.width).Render(strings.Join([]string{header, body, footer}, "\n"))
}

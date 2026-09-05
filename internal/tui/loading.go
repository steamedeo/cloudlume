package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// isLoading reports whether the startup fetch is still in flight — accounts
// haven't been discovered yet, or at least one hasn't returned its first
// resource snapshot. The dashboard stays hidden behind a progress screen
// until this clears, so the first frame isn't a half-empty table.
func (m Model) isLoading() bool {
	return !m.accountsDiscovered || !m.loadDone
}

func (m Model) renderLoading() string {
	wordmark := gradientText("cloudlume", hexSky, hexViolet, true)

	var status string
	var ratio float64
	if !m.accountsDiscovered {
		status = "discovering local AWS profiles..."
		ratio = 0
	} else {
		total := len(m.accounts)
		settled := m.settledCount()
		if total == 0 {
			status = "no local accounts found"
			ratio = 1
		} else {
			status = fmt.Sprintf("loading resources from %d of %d accounts...", settled, total)
			ratio = float64(settled) / float64(total)
		}
	}

	const barWidth = 40
	bar := progressBar(barWidth, ratio)

	var accountLines []string
	for _, name := range m.sortedAccountNames() {
		st := m.statuses[name]
		health := colDim
		if st != nil {
			health = healthColor(st.Health)
		}
		dot := lipgloss.NewStyle().Foreground(health).Render("●")
		accountLines = append(accountLines, dot+" "+dimStyle.Background(colBgPage).Render(name))
	}

	lines := []string{
		"",
		"",
		wordmark,
		dimStyle.Background(colBgPage).Render(status),
		"",
		bar,
	}
	if len(accountLines) > 0 {
		lines = append(lines, "", strings.Join(accountLines, "   "))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, lines...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content,
		lipgloss.WithWhitespaceBackground(colBgPage))
}

package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/steamedeo/cloudlume/internal/model"
)

func (m Model) renderTabs() string {
	showAccountCol := m.selectedAccountName() == ""
	_ = showAccountCol

	var parts []string
	for i, cat := range model.AllCategories {
		label := " " + string(cat) + " "
		if i == m.tabIndex {
			parts = append(parts, pill(strings.TrimSpace(label), colSky, true))
		} else {
			parts = append(parts, dimStyle.Render(label))
		}
	}
	return strings.Join(parts, " ")
}

func (m Model) renderMain(width, height int) string {
	resources := m.visibleResources()
	showAccountCol := m.selectedAccountName() == ""

	tableHeight := height - 8 // tabs + divider + header row + detail card
	if tableHeight < 3 {
		tableHeight = 3
	}

	table := m.renderTable(resources, width-4, tableHeight, showAccountCol)
	detail := m.renderDetail(resources, width-4)

	body := m.renderTabs() + "\n" +
		dimStyle.Render(strings.Repeat("─", width-2)) + "\n" +
		table + "\n" +
		dimStyle.Render(strings.Repeat("─", width-2)) + "\n" +
		detail

	return cardStyleFocused.Width(width).Height(height).Render(body)
}

type col struct {
	label string
	width int
}

func (m Model) renderTable(resources []model.Resource, width, height int, showAccount bool) string {
	cols := tableColumns(width, showAccount)

	var b strings.Builder
	b.WriteString(renderRow(cols, headerCells(cols, showAccount), boldStyle, false))
	b.WriteString("\n")

	if len(resources) == 0 {
		b.WriteString(dimStyle.Render("  no resources in this category yet"))
		return b.String()
	}

	maxRows := height - 1
	for i, r := range resources {
		if i >= maxRows {
			break
		}
		cells := rowCells(r, showAccount)
		selected := i == m.cursor
		rowStyle := textStyle
		if selected {
			rowStyle = lipgloss.NewStyle().Foreground(colText).Background(colBgCardHi).Bold(true)
		}
		b.WriteString(renderRow(cols, cells, rowStyle, selected))
		if i < len(resources)-1 && i < maxRows-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func tableColumns(width int, showAccount bool) []col {
	var cols []col
	if showAccount {
		cols = append(cols, col{"ACCOUNT", 14})
	}
	cols = append(cols,
		col{"REGION", 14},
		col{"NAME", 0}, // flexible, filled below
		col{"TYPE", 14},
		col{"STATUS", 12},
	)

	fixed := 0
	for _, c := range cols {
		fixed += c.width
	}
	flex := width - fixed - (len(cols) * 1)
	if flex < 8 {
		flex = 8
	}
	for i := range cols {
		if cols[i].label == "NAME" {
			cols[i].width = flex
		}
	}
	return cols
}

func headerCells(cols []col, showAccount bool) []string {
	cells := make([]string, len(cols))
	for i, c := range cols {
		cells[i] = c.label
	}
	return cells
}

func rowCells(r model.Resource, showAccount bool) []string {
	var cells []string
	if showAccount {
		cells = append(cells, r.Account)
	}
	cells = append(cells, r.Region, r.Name, r.Type, r.Status)
	return cells
}

func renderRow(cols []col, cells []string, style lipgloss.Style, selected bool) string {
	var parts []string
	for i, c := range cols {
		val := ""
		if i < len(cells) {
			val = cells[i]
		}
		parts = append(parts, style.Copy().Width(c.width).MaxWidth(c.width).Render(truncate(val, c.width)))
	}
	line := strings.Join(parts, " ")
	return line
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	r := []rune(s)
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}

func (m Model) renderDetail(resources []model.Resource, width int) string {
	if m.cursor < 0 || m.cursor >= len(resources) {
		return dimStyle.Render("  select a resource to see details")
	}
	r := resources[m.cursor]

	title := boldStyle.Render(r.Name) + "  " + healthPill(r.Status, r.Health)

	const labelWidth = 18
	var lines []string
	for _, d := range r.Details {
		if d.Value == "" {
			continue
		}
		label := dimStyle.Width(labelWidth).Render(d.Label)
		lines = append(lines, label+textStyle.Render(d.Value))
	}
	return title + "\n" + strings.Join(lines, "\n")
}

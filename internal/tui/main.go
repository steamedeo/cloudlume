package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/steamedeo/cloudlume/internal/model"
)

func (m Model) renderTabs() string {
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

// renderTabsRow adds a right-aligned indicator next to the category tabs:
// a filter tag once confirmed, or otherwise a plain "N/total" position —
// otherwise scrolling past the first page, or an active search, has no
// visible cue. While actively typing a filter, renderFilterBar takes over
// the whole row instead (see there for why).
func (m Model) renderTabsRow(resources []model.Resource, width int) string {
	if m.filtering {
		return m.renderFilterBar(width)
	}

	tabs := m.renderTabs()
	if m.detailOpen {
		return tabs
	}

	var right string
	switch {
	case m.filterQuery != "" && len(resources) == 0:
		right = dimStyle.Render(fmt.Sprintf("/%s  no matches", m.filterQuery))
	case m.filterQuery != "":
		right = dimStyle.Render(fmt.Sprintf("/%s  %d/%d", m.filterQuery, m.cursor+1, len(resources)))
	case len(resources) > 0:
		right = dimStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(resources)))
	default:
		return tabs
	}

	gap := width - lipgloss.Width(tabs) - lipgloss.Width(right)
	if gap < 1 {
		return tabs
	}
	return tabs + strings.Repeat(" ", gap) + right
}

// renderFilterBar renders a full-width, solid-color bar for the fuzzy
// filter input. A small corner label was too easy to miss while typing —
// this can't be — and it matches the pill convention of using the page
// background as foreground so it visually "punches through".
func (m Model) renderFilterBar(width int) string {
	text := truncate(fmt.Sprintf(" Filter: %s▏", m.filterQuery), width)
	return lipgloss.NewStyle().
		Background(colSky).
		Foreground(colBgPage).
		Bold(true).
		Width(width).
		Render(text)
}

func (m Model) renderMain(width, height int) string {
	resources := m.visibleResources()
	showAccountCol := m.selectedAccountName() == ""

	innerWidth := width - cardPaddingX // usable width inside the card's padding

	contentHeight := height - 3 // tabs + divider + one spare line
	if contentHeight < 3 {
		contentHeight = 3
	}

	var content string
	if m.detailOpen {
		content = m.renderDetailView(resources, innerWidth, contentHeight)
	} else {
		content = m.renderTable(resources, innerWidth, contentHeight, showAccountCol)
	}

	tabsRow := m.renderTabsRow(resources, innerWidth)

	body := tabsRow + "\n" +
		dimStyle.Render(strings.Repeat("─", innerWidth)) + "\n" +
		content

	// Never emit more lines than the card has room for — an overflowing
	// main panel would spill past its box and break the side-by-side join
	// with the sidebar.
	body = clampLines(body, height)

	style := cardStyle
	if m.focus == focusTable {
		style = cardStyleFocused
	}
	return style.Width(width).Height(height).Render(body)
}

func clampLines(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
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
		msg := "  no resources in this category yet"
		if m.filterQuery != "" {
			msg = fmt.Sprintf("  no resources match \"%s\"", m.filterQuery)
		}
		b.WriteString(dimStyle.Render(msg))
		return b.String()
	}

	maxRows := height - 1
	if maxRows < 1 {
		maxRows = 1
	}

	// Scroll the visible window so the cursor is always on screen — without
	// this, moving the cursor past the first page of rows leaves it
	// selected but never rendered, since the loop below only ever drew
	// resources[0:maxRows].
	offset := 0
	if len(resources) > maxRows {
		if m.cursor >= maxRows {
			offset = m.cursor - maxRows + 1
		}
		if offset > len(resources)-maxRows {
			offset = len(resources) - maxRows
		}
	}

	end := offset + maxRows
	if end > len(resources) {
		end = len(resources)
	}

	for i := offset; i < end; i++ {
		r := resources[i]
		cells := rowCells(r, showAccount)
		selected := i == m.cursor
		rowStyle := textStyle
		if selected {
			rowStyle = lipgloss.NewStyle().Foreground(colText).Background(colBgCardHi).Bold(true)
		}
		b.WriteString(renderRow(cols, cells, rowStyle, selected))
		if i < end-1 {
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
		parts = append(parts, style.Width(c.width).MaxWidth(c.width).Render(truncate(val, c.width)))
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

// renderDetailView renders the full-height detail card for the selected
// resource, shown in place of the table when detailOpen is toggled on —
// giving a long field list room to breathe instead of the few lines a
// permanently-visible strip could spare.
func (m Model) renderDetailView(resources []model.Resource, width, height int) string {
	if m.cursor < 0 || m.cursor >= len(resources) {
		return dimStyle.Render(truncate("  no resource selected", width))
	}
	r := resources[m.cursor]

	nameWidth := width - 14 // leave room for the status pill
	if nameWidth < 4 {
		nameWidth = 4
	}
	title := boldStyle.Render(truncate(r.Name, nameWidth)) + "  " + healthPill(r.Status, r.Health)

	const labelWidth = 18
	valueWidth := width - labelWidth
	if valueWidth < 4 {
		valueWidth = 4
	}

	lines := []string{title, ""}
	for _, d := range r.Details {
		if d.Value == "" {
			continue
		}
		label := dimStyle.Width(labelWidth).Render(truncate(d.Label, labelWidth))
		value := textStyle.Render(truncate(d.Value, valueWidth))
		lines = append(lines, label+value)
	}
	lines = append(lines, "", dimStyle.Render("esc/enter to go back"))

	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

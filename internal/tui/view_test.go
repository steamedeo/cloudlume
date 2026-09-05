package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/steamedeo/cloudlume/internal/model"
)

// TestViewFitsTerminalWidth guards against the sidebar+main layout math
// silently drifting out of sync with lipgloss's border/padding overhead
// again: every rendered line must be exactly m.width wide, or the outer
// page style soft-wraps it and the whole layout corrupts below that row.
func TestViewFitsTerminalWidth(t *testing.T) {
	m := New(nil, time.Minute)
	m.accounts = []model.Account{{Provider: "aws", Name: "default", Region: "us-east-1"}}
	m.statuses["default"] = &model.AccountStatus{
		Account: m.accounts[0], Health: model.HealthOK,
		LastRefresh: time.Now(), ResourceCount: 1,
	}
	m.byAccount["default"] = []model.Resource{{
		Provider: "aws", Account: "default", Region: "us-east-1",
		Category: model.CategoryEC2, Type: "ec2-instance",
		ID: "i-1", Name: "web-1", Status: "running", Health: model.HealthOK,
	}}
	m.accountsDiscovered = true
	m.loadDone = true

	for _, size := range []struct{ w, h int }{
		{100, 30}, {80, 24}, {160, 45}, {28, 10},
	} {
		res, _ := m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
		mm := res.(Model)

		out := mm.View()
		for i, line := range strings.Split(out, "\n") {
			if w := lipgloss.Width(line); w > size.w {
				t.Fatalf("size %dx%d: line %d is %d cols wide, wider than terminal width %d: %q",
					size.w, size.h, i, w, size.w, line)
			}
		}
	}
}

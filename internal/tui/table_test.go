package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/steamedeo/cloudlume/internal/model"
)

func TestRenderTableScrollsToCursor(t *testing.T) {
	var resources []model.Resource
	for i := 0; i < 50; i++ {
		resources = append(resources, model.Resource{
			Name: fmt.Sprintf("resource-%02d", i),
		})
	}

	m := Model{cursor: 40}
	out := m.renderTable(resources, 60, 10, false)

	if !strings.Contains(out, "resource-40") {
		t.Fatalf("expected the selected row (cursor=40) to be visible in the rendered table, got:\n%s", out)
	}
}

func TestRenderTableScrollDoesNotOverscroll(t *testing.T) {
	var resources []model.Resource
	for i := 0; i < 5; i++ {
		resources = append(resources, model.Resource{
			Name: fmt.Sprintf("resource-%02d", i),
		})
	}

	// Cursor near the end of a short list shouldn't push the window past
	// the last row.
	m := Model{cursor: 4}
	out := m.renderTable(resources, 60, 10, false)

	if !strings.Contains(out, "resource-00") || !strings.Contains(out, "resource-04") {
		t.Fatalf("expected all rows visible when list fits the viewport, got:\n%s", out)
	}
}

package tui

import (
	"testing"

	"github.com/steamedeo/cloudlume/internal/model"
)

func TestFilterResources(t *testing.T) {
	resources := []model.Resource{
		{Name: "web-frontend-prod"},
		{Name: "worker-batch-jobs"},
		{Name: "db-replica-1"},
	}

	got := filterResources(resources, "wfp")
	if len(got) != 1 || got[0].Name != "web-frontend-prod" {
		t.Fatalf("expected fuzzy match on subsequence \"wfp\" to find web-frontend-prod, got %+v", got)
	}

	got = filterResources(resources, "zzz")
	if len(got) != 0 {
		t.Fatalf("expected no matches for \"zzz\", got %+v", got)
	}

	got = filterResources(resources, "")
	if len(got) != 0 {
		t.Fatalf("expected empty query to produce no matches (visibleResources skips filtering entirely instead), got %+v", got)
	}
}

package config

import "testing"

func TestRegionsFor(t *testing.T) {
	cfg := Config{
		Regions: map[string][]string{
			"work-prod": {"us-east-1", "eu-west-1"},
			"*":         {"us-west-2"},
		},
	}

	t.Run("explicit profile override wins", func(t *testing.T) {
		regions, ok := cfg.RegionsFor("work-prod")
		if !ok || len(regions) != 2 || regions[0] != "us-east-1" {
			t.Fatalf("got %v, ok=%v", regions, ok)
		}
	})

	t.Run("falls back to wildcard", func(t *testing.T) {
		regions, ok := cfg.RegionsFor("some-other-profile")
		if !ok || len(regions) != 1 || regions[0] != "us-west-2" {
			t.Fatalf("got %v, ok=%v", regions, ok)
		}
	})

	t.Run("no override at all", func(t *testing.T) {
		empty := Config{}
		_, ok := empty.RegionsFor("anything")
		if ok {
			t.Fatalf("expected no override, got one")
		}
	})
}

func TestIsExcluded(t *testing.T) {
	cfg := Config{Exclude: []string{"sandbox", "old-account"}}

	if !cfg.IsExcluded("sandbox") {
		t.Fatal("expected sandbox to be excluded")
	}
	if cfg.IsExcluded("work-prod") {
		t.Fatal("did not expect work-prod to be excluded")
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.RefreshInterval <= 0 {
		t.Fatalf("expected a positive default refresh interval, got %v", cfg.RefreshInterval)
	}
}

package tui

import (
	"github.com/sahilm/fuzzy"

	"github.com/steamedeo/cloudlume/internal/model"
)

// filterResources fuzzy-matches query against each resource's name and
// returns the matches ordered by relevance (best match first) — this
// replaces the usual alphabetical sort while a search is active, since
// relevance is what you're scanning for at that point.
func filterResources(resources []model.Resource, query string) []model.Resource {
	names := make([]string, len(resources))
	for i, r := range resources {
		names[i] = r.Name
	}

	matches := fuzzy.Find(query, names)
	filtered := make([]model.Resource, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, resources[match.Index])
	}
	return filtered
}

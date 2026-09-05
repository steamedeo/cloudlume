package tui

import (
	"time"

	"github.com/steamedeo/cloudlume/internal/model"
)

// accountsMsg carries the full set of accounts discovered across every
// registered provider, delivered once at startup.
type accountsMsg struct {
	accounts []model.Account
}

// resourcesMsg carries one account's freshly-fetched resource snapshot,
// or an error if the fetch failed.
type resourcesMsg struct {
	account   model.Account
	resources []model.Resource
	err       error
	fetchedAt time.Time
}

// tickMsg drives the periodic refresh loop.
type tickMsg struct{}

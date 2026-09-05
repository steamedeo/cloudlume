// Package provider defines the pluggable interface cloud backends implement,
// and a registry so the TUI layer can iterate over "whatever providers are
// wired in" without a hardcoded switch on AWS/GCP/Azure.
package provider

import (
	"context"

	"github.com/steamedeo/cloudlume/internal/model"
)

// Provider is one cloud backend (AWS, GCP, Azure, ...). Implementations
// discover local credentials themselves (profiles, ADC, CLI config) rather
// than accepting explicit secrets, matching cloudlume's "local creds only"
// design.
type Provider interface {
	// Name is the short provider identifier, e.g. "aws".
	Name() string

	// Accounts returns every locally-configured identity this provider
	// found (e.g. every AWS profile in ~/.aws/config).
	Accounts(ctx context.Context) ([]model.Account, error)

	// FetchResources returns the current snapshot of resources for one
	// account, across every category the provider supports.
	FetchResources(ctx context.Context, account model.Account) ([]model.Resource, error)
}

var registry = map[string]Provider{}

// Register adds a provider implementation to the global registry. Providers
// call this from an init() in their own package so wiring a new cloud in
// means importing its package for side effects, nothing more.
func Register(p Provider) {
	registry[p.Name()] = p
}

// All returns every registered provider, in no particular order.
func All() []Provider {
	out := make([]Provider, 0, len(registry))
	for _, p := range registry {
		out = append(out, p)
	}
	return out
}

// Package aws implements the cloudlume provider.Provider interface for
// Amazon Web Services, using local credentials only (profiles in
// ~/.aws/config and ~/.aws/credentials, env vars, SSO) — cloudlume never
// accepts or stores AWS secrets itself.
package aws

import (
	"context"
	"fmt"
	"sync"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"

	"github.com/steamedeo/cloudlume/internal/config"
	"github.com/steamedeo/cloudlume/internal/model"
)

const defaultRegion = "us-east-1"

// Provider is the AWS backend. It is stateful only in that it remembers,
// per profile, which regions to poll — everything else is fetched fresh.
type Provider struct {
	cfg config.Config

	mu             sync.RWMutex
	regionsByAccount map[string][]string
}

// New constructs an AWS provider using the given cloudlume settings for
// region overrides and profile exclusions.
func New(cfg config.Config) *Provider {
	return &Provider{
		cfg:              cfg,
		regionsByAccount: map[string][]string{},
	}
}

func (p *Provider) Name() string { return "aws" }

// Accounts discovers every local AWS profile and resolves the region(s)
// each one should be polled in: an explicit override from cloudlume's own
// config, else the profile's own configured region, else a hardcoded
// fallback so the tool never queries zero regions.
func (p *Provider) Accounts(ctx context.Context) ([]model.Account, error) {
	names := discoverProfiles()
	if len(names) == 0 {
		names = []string{"default"}
	}

	accounts := make([]model.Account, 0, len(names))
	for _, name := range names {
		if p.cfg.IsExcluded(name) {
			continue
		}

		region := profileRegion(name)
		if region == "" {
			region = defaultRegion
		}

		regions := []string{region}
		if override, ok := p.cfg.RegionsFor(name); ok {
			regions = override
		}

		p.mu.Lock()
		p.regionsByAccount[name] = regions
		p.mu.Unlock()

		accounts = append(accounts, model.Account{
			Provider: p.Name(),
			Name:     name,
			Region:   regions[0],
		})
	}
	return accounts, nil
}

func (p *Provider) regionsFor(account model.Account) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if r, ok := p.regionsByAccount[account.Name]; ok && len(r) > 0 {
		return r
	}
	if account.Region != "" {
		return []string{account.Region}
	}
	return []string{defaultRegion}
}

// loadConfig builds an aws.Config for one profile+region, using the
// standard local credential chain (shared config/credentials, env vars,
// SSO, instance/container roles) — never explicit keys passed in by
// cloudlume.
func loadConfig(ctx context.Context, profile, region string) (awssdk.Config, error) {
	return awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithSharedConfigProfile(profile),
		awsconfig.WithRegion(region),
	)
}

// FetchResources queries every supported resource category, across every
// configured region, for one account, and merges the results. Per-region
// failures (e.g. a service not available, or a permissions gap) are
// attached to the account status rather than aborting the whole fetch.
func (p *Provider) FetchResources(ctx context.Context, account model.Account) ([]model.Resource, error) {
	regions := p.regionsFor(account)

	var (
		mu        sync.Mutex
		resources []model.Resource
		firstErr  error
	)

	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()

			cfg, err := loadConfig(ctx, account.Name, region)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s/%s: %w", account.Name, region, err)
				}
				mu.Unlock()
				return
			}

			var regionResources []model.Resource
			regionResources = append(regionResources, fetchCompute(ctx, cfg, account, region)...)
			regionResources = append(regionResources, fetchDatabases(ctx, cfg, account, region)...)

			mu.Lock()
			resources = append(resources, regionResources...)
			mu.Unlock()
		}(region)
	}

	// S3 is a global(-ish) listing API, not per-region — fetch it once
	// using the first region's credentials/config.
	wg.Add(1)
	go func() {
		defer wg.Done()
		cfg, err := loadConfig(ctx, account.Name, regions[0])
		if err != nil {
			return
		}
		mu.Lock()
		resources = append(resources, fetchStorage(ctx, cfg, account)...)
		mu.Unlock()
	}()

	wg.Wait()

	if len(resources) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return resources, nil
}

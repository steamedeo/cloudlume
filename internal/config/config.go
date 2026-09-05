// Package config loads cloudlume's own optional settings file, layered on
// top of whatever local cloud credentials/profiles are discovered — it
// never holds secrets, only display and query preferences.
package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is cloudlume's own settings, loaded from cloudlume.yaml in the
// current directory or ~/.cloudlume.yaml. Every field is optional; a
// missing file just means "use provider defaults everywhere".
type Config struct {
	// RefreshInterval controls how often resources are re-polled.
	RefreshInterval time.Duration `yaml:"refreshInterval"`

	// Regions maps an account/profile name to an explicit list of regions
	// to query, overriding that profile's own default region. A "*" key
	// sets the default override for any profile not otherwise listed.
	Regions map[string][]string `yaml:"regions"`

	// Exclude lists profile names to skip entirely during auto-discovery.
	Exclude []string `yaml:"exclude"`
}

const defaultRefresh = 30 * time.Second

// Default returns the config used when no file is found.
func Default() Config {
	return Config{RefreshInterval: defaultRefresh}
}

// Load reads cloudlume.yaml from the working directory, falling back to
// ~/.cloudlume.yaml, then applies defaults for any unset field.
func Load() (Config, error) {
	cfg := Default()

	path := findConfigFile()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = defaultRefresh
	}
	return cfg, nil
}

func findConfigFile() string {
	candidates := []string{"cloudlume.yaml", "cloudlume.yml"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		for _, c := range []string{".cloudlume.yaml", ".cloudlume.yml"} {
			p := filepath.Join(home, c)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

// RegionsFor returns the region override list for a profile, if any, and
// whether one was configured at all.
func (c Config) RegionsFor(profile string) ([]string, bool) {
	if r, ok := c.Regions[profile]; ok && len(r) > 0 {
		return r, true
	}
	if r, ok := c.Regions["*"]; ok && len(r) > 0 {
		return r, true
	}
	return nil, false
}

// IsExcluded reports whether a profile name should be skipped.
func (c Config) IsExcluded(profile string) bool {
	for _, e := range c.Exclude {
		if e == profile {
			return true
		}
	}
	return false
}

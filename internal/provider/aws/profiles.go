package aws

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/ini.v1"
)

// discoverProfiles reads ~/.aws/config and ~/.aws/credentials and returns
// every profile name found across both files, deduplicated. This is local
// credential discovery only — no network calls, no secrets are read out.
func discoverProfiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	seen := map[string]bool{}

	addFrom := func(path string, stripProfilePrefix bool) {
		f, err := ini.Load(path)
		if err != nil {
			return
		}
		for _, section := range f.Sections() {
			name := section.Name()
			if name == ini.DefaultSection {
				continue
			}
			if stripProfilePrefix {
				// ~/.aws/config names sections "profile foo" (except
				// "default", which has no prefix).
				const prefix = "profile "
				if len(name) > len(prefix) && name[:len(prefix)] == prefix {
					name = name[len(prefix):]
				}
			}
			seen[name] = true
		}
	}

	addFrom(filepath.Join(home, ".aws", "config"), true)
	addFrom(filepath.Join(home, ".aws", "credentials"), false)

	if envFile := os.Getenv("AWS_CONFIG_FILE"); envFile != "" {
		addFrom(envFile, true)
	}
	if envFile := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); envFile != "" {
		addFrom(envFile, false)
	}

	profiles := make([]string, 0, len(seen))
	for name := range seen {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	return profiles
}

// profileRegion reads the "region" key for a profile directly out of the
// shared config file, without establishing any AWS session. Returns "" if
// unset.
func profileRegion(profile string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".aws", "config")
	if envFile := os.Getenv("AWS_CONFIG_FILE"); envFile != "" {
		path = envFile
	}
	f, err := ini.Load(path)
	if err != nil {
		return ""
	}
	sectionName := "profile " + profile
	if profile == "default" {
		sectionName = "default"
	}
	section, err := f.GetSection(sectionName)
	if err != nil {
		return ""
	}
	return section.Key("region").String()
}

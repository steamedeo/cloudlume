package aws

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}

func TestDiscoverProfiles(t *testing.T) {
	dir := t.TempDir()

	configPath := writeTempFile(t, dir, "config", `
[default]
region = us-east-1

[profile work-prod]
region = eu-west-1

[profile work-staging]
region = eu-west-1
`)
	credsPath := writeTempFile(t, dir, "credentials", `
[default]
aws_access_key_id = x

[work-prod]
aws_access_key_id = x

[personal]
aws_access_key_id = x
`)

	t.Setenv("AWS_CONFIG_FILE", configPath)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credsPath)

	got := discoverProfiles()
	sort.Strings(got)

	want := []string{"default", "personal", "work-prod", "work-staging"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestProfileRegion(t *testing.T) {
	dir := t.TempDir()
	configPath := writeTempFile(t, dir, "config", `
[default]
region = us-east-1

[profile work-prod]
region = eu-west-1
`)
	t.Setenv("AWS_CONFIG_FILE", configPath)

	if r := profileRegion("work-prod"); r != "eu-west-1" {
		t.Fatalf("got region %q, want eu-west-1", r)
	}
	if r := profileRegion("default"); r != "us-east-1" {
		t.Fatalf("got region %q, want us-east-1", r)
	}
	if r := profileRegion("does-not-exist"); r != "" {
		t.Fatalf("got region %q, want empty", r)
	}
}

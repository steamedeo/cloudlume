# cloudlume

A k9s-style terminal dashboard for cloud resources. Same family as
[serverlume](https://github.com/steamedeo/serverlume), but instead of an SSH
fleet it maps the cloud resources reachable from your **local credentials**
into one live, organized view — no agents, no server-side component, no
credentials of its own.

## What it does

- Auto-discovers every local AWS profile (`~/.aws/config`,
  `~/.aws/credentials`, SSO, env vars) and polls each one concurrently.
- Aggregates resources from **all accounts together** in one view — there's
  no account switcher to fight with.
- Groups resources by category (Compute, Storage, Databases, Networking)
  with a live-refreshing table and a detail card per resource.
- Never stores or transmits credentials; it only ever uses the standard AWS
  SDK credential chain, read locally.

## Usage

```bash
go run ./cmd/cloudlume
```

or build a binary:

```bash
go build -o cloudlume ./cmd/cloudlume
./cloudlume
```

### Keybindings

| Key | Action |
| --- | --- |
| `↑`/`k`, `↓`/`j` | Move selection |
| `←`/`h`, `→`/`l` | Switch account (or "All Accounts") |
| `Tab`/`Shift+Tab`, `1`-`4` | Switch resource category |
| `p` | Pause/resume live refresh |
| `r` | Refresh now |
| `q`/`Ctrl+C` | Quit |

### Configuration

Optional — cloudlume works with zero config, using every profile it finds
and each profile's own default region. To customize refresh interval,
region overrides per profile, or exclude certain profiles, copy
[`cloudlume.example.yaml`](cloudlume.example.yaml) to `cloudlume.yaml` in
the working directory or `~/.cloudlume.yaml`.

## Architecture

cloudlume is built to grow beyond AWS and beyond three resource types
without touching the TUI:

- [`internal/model`](internal/model/resource.go) — provider-agnostic
  `Resource`/`Account` types every backend normalizes into.
- [`internal/provider`](internal/provider/provider.go) — the `Provider`
  interface (`Accounts`, `FetchResources`) plus a registry. Adding a new
  cloud means implementing this interface and calling `provider.Register`
  from an `init()`.
- [`internal/provider/aws`](internal/provider/aws/aws.go) — the AWS
  implementation. Each resource category (compute, storage, database) is
  its own file with a `fetchX(ctx, cfg, account, region) []model.Resource`
  function — adding a new AWS service is adding one more such function and
  calling it from `FetchResources`.
- [`internal/tui`](internal/tui/app.go) — the Bubble Tea dashboard. It only
  ever talks to `model.Resource`/`model.Account`, so it never needs to
  change when a new provider or resource type is added — new categories
  just need a `model.Category` constant and a color mapping.

## Styling

Follows the serverlume TUI style guide: Bubble Tea + Lip Gloss, rounded
card borders, a LUV-blended gradient wordmark and gauges, status pills that
punch through in the page background color, and a single always-visible
keybinding legend. cloudlume's accent pair is sky blue + violet, distinct
from serverlume's pink + lavender but built from the same recipe.

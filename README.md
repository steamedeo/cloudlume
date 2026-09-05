# cloudlume

A k9s-style terminal dashboard for cloud resources. Uses your local AWS
credentials to map resources across accounts into one live, organized view.
No agents, no server-side component, no credentials of its own.

## What it does

- Auto-discovers every local AWS profile (`~/.aws/config`,
  `~/.aws/credentials`, SSO, env vars) and polls each one concurrently.
- Aggregates resources from all accounts together in one view.
- Groups resources by category with a live-refreshing table and a detail
  view per resource.

## Resources covered

| Category | Resources |
| --- | --- |
| EC2 | Instances |
| Lambda | Functions |
| Databases | RDS instances, DynamoDB tables |
| Storage | S3 buckets |
| Networking | CloudFront distributions |

## Usage

```bash
go build -o cloudlume ./cmd/cloudlume
./cloudlume
```

### Keybindings

| Key | Action |
| --- | --- |
| `↑`/`k`, `↓`/`j` | Move selection in the focused panel |
| `←`/`h`, `→`/`l` | Switch focus between the accounts sidebar and the resource table |
| `Enter` | Open/close the detail view for the selected resource |
| `/` | Fuzzy-filter resources in the current tab by name |
| `Esc` | Close the detail view, or clear the active filter |
| `Tab`/`Shift+Tab`, `1`-`5` | Switch resource category |
| `p` | Pause/resume live refresh |
| `r` | Refresh now |
| `q`/`Ctrl+C` | Quit |

### Configuration

Optional. Copy [`cloudlume.example.yaml`](cloudlume.example.yaml) to
`cloudlume.yaml` (working directory) or `~/.cloudlume.yaml` to set a refresh
interval, override regions per profile, or exclude profiles.

## Architecture

- [`internal/model`](internal/model/resource.go) — provider-agnostic
  `Resource`/`Account` types every backend normalizes into.
- [`internal/provider`](internal/provider/provider.go) — the `Provider`
  interface (`Accounts`, `FetchResources`) plus a registry. A new cloud
  means implementing this interface and registering it from an `init()`.
- [`internal/provider/aws`](internal/provider/aws/aws.go) — the AWS
  implementation. Each service is its own file
  (`ec2.go`, `lambda.go`, `database.go`, `dynamodb.go`, `storage.go`,
  `networking.go`) with a
  `fetchX(ctx, cfg, account, region) []model.Resource` function.
- [`internal/tui`](internal/tui/app.go) — the Bubble Tea dashboard. It only
  talks to `model.Resource`/`model.Account`, so it never changes when a
  provider or resource type is added.

## Styling

Bubble Tea + Lip Gloss, rounded card borders, a LUV-blended gradient
wordmark, status pills, and a single always-visible keybinding legend.
Accent pair: sky blue + violet.

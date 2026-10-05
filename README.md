# chora-cli

`chora` — command-line tool for Chora platform management. Expert and admin
access to the Chora Bimodal Atomic Learning Ecosystem from the terminal:
LearningAtom CRUD, tenant inspection, Familiar status, feature-flag
administration, and gateway health checks.

The CLI is cloud-neutral: it is a plain HTTP client of the Chora gateway REST
API. It listens on no ports, needs no database, and stores credentials in a
user-scoped file (`~/.config/chora/credentials.json`, mode 0600). API keys are
never logged or displayed in full.

## Install

```sh
go install github.com/apollo-chora/chora-cli/cmd/chora@latest
```

Or run from a checkout:

```sh
go run ./cmd/chora
```

Or via Docker:

```sh
docker run --rm walfa/chora-cli --help
```

## Usage

```sh
chora --help
```

All API calls carry the stored key in the `X-API-Key` header and target the
gateway URL stored at login (default `http://localhost:8093`).

| Command | Description |
| --- | --- |
| `chora auth login --api-key <key> [--gateway <url>]` | Authenticate and store credentials |
| `chora auth logout` | Clear local credentials |
| `chora atoms list [--topic <name>] [--cursor <c>] [--limit <n>]` | List LearningAtoms (cursor-paginated) |
| `chora atoms get <id>` | Fetch a single LearningAtom by ID |
| `chora atoms create --file <yaml>` | Create a LearningAtom from a YAML/JSON definition |
| `chora tenants info` | Display current tenant details |
| `chora familiars status` | Display Familiar status and stats |
| `chora health` | Check gateway health |
| `chora events tail` | Tail the event bus (streaming not yet implemented) |
| `chora flags list` | List feature flag overrides |
| `chora flags set <code>` | Enable a feature flag override |

## Configuration

Credentials live in `~/.config/chora/credentials.json` (override the directory
by pointing `HOME` elsewhere). The gateway URL defaults to
`http://localhost:8093` and can be set per-login with `--gateway`.

## Development

```sh
go build ./...
go vet ./...
go test ./...
```

## Layout

```
cmd/chora/                 entrypoint — cobra command tree
cmd/chora/internal/client/ HTTP gateway client (X-API-Key header, 30s timeout)
cmd/chora/internal/config/ credential file store (0600, key masking)
```

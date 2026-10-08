# chora-cli

## About

`chora-cli` is a Go command-line client for managing and inspecting the Chora platform through its gateway REST API. It provides authentication, LearningAtom operations, tenant and Familiar status, gateway health checks, feature-flag administration, and an events command. The CLI is a client only: it opens no listening port and does not require a database.

## Quick start

Requires Go 1.26.6, as declared in `go.mod`.

Install the CLI from the Go module:

```sh
go install github.com/apollo-chora/chora-cli/cmd/chora@latest
```

Then authenticate against the gateway:

```sh
chora auth login --api-key sk_test_your_key
```

The default gateway URL is `http://localhost:8093`. Supply another gateway with `--gateway`:

```sh
chora auth login   --api-key sk_test_your_key   --gateway http://localhost:8093
```

Check the installed command:

```sh
chora --help
chora --version
```

You can also run the CLI directly from a checkout:

```sh
git clone https://github.com/apollo-chora/chora-cli.git
cd chora-cli
go run ./cmd/chora --help
```

Or run the published container:

```sh
docker run --rm walfa/chora-cli --help
```

## Usage

### Authentication

Credentials are stored in `~/.config/chora/credentials.json` with file mode `0600`. The config directory follows the current user's home directory, so setting `HOME` changes the effective location.

```sh
chora auth login --api-key <key> [--gateway <url>]
chora auth logout
```

The API key is sent to the gateway as the `X-API-Key` header. The HTTP client also sends a `User-Agent` of `chora-cli/0.1`. API keys are masked when represented by the credential type and are not intended to be logged or printed in full.

### Commands

| Command | Description |
| --- | --- |
| `chora auth login --api-key <key> [--gateway <url>]` | Store an API key and gateway URL |
| `chora auth logout` | Remove stored credentials |
| `chora atoms list` | List LearningAtoms for the current tenant |
| `chora atoms get <id>` | Fetch one LearningAtom |
| `chora atoms create --file <path>` | Create a LearningAtom from a YAML/JSON file |
| `chora tenants info` | Fetch the current tenant |
| `chora familiars status` | Fetch Familiar status and statistics |
| `chora health` | Fetch gateway health |
| `chora events tail` | Print the current event-tail status; streaming is not implemented |
| `chora flags list` | List feature-flag overrides |
| `chora flags set <code>` | Enable a feature-flag override |

Examples:

```sh
chora atoms list
chora atoms get <id>
chora atoms create --file atom.yaml

chora tenants info
chora familiars status
chora health

chora flags list
chora flags set my_feature_code
```

The CLI calls these gateway paths:

| Command | HTTP path |
| --- | --- |
| `atoms list` | `GET /api/v1/atoms` |
| `atoms get` | `GET /api/v1/atoms/{id}` |
| `atoms create` | `POST /api/atoms` |
| `tenants info` | `GET /api/tenants/me` |
| `familiars status` | `GET /api/v1/familiars/me/stats` |
| `health` | `GET /health` |
| `flags list` | `GET /api/feature-flags` |
| `flags set` | `PUT /api/v1/admin/feature-flags/{code}` |

Responses are printed as pretty-printed JSON when the gateway returns JSON. HTTP responses with status 400 or higher are prefixed with the returned status code and their response body is still printed.

The HTTP client uses a 30-second request timeout. The CLI does not expose a separate timeout flag.

## Development

Build the CLI:

```sh
go build ./...
```

Run the same static and test checks used by CI:

```sh
gofmt -l .
go mod tidy
go vet ./...
go test ./...
```

CI also verifies that `go mod tidy` leaves `go.mod` and `go.sum` unchanged.

The project is organized as a single Cobra command tree:

```text
cmd/chora/
  main.go                     Cobra commands and gateway API calls
  internal/client/
    client.go                 HTTP client and request headers
    client_test.go             HTTP client tests
  internal/config/
    config.go                 Credential file storage and key masking
    config_test.go             Credential-store tests
go.mod                         Go module and dependencies
go.sum                         Dependency checksums
Dockerfile                     Multi-stage CLI image build
.github/workflows/
  ci.yml                      Formatting, module, vet, and test checks
  docker-publish.yml          Multi-architecture Docker image publishing
```

The Dockerfile builds a static Linux `amd64` binary and packages it in a non-root distroless runtime image. The GitHub Actions Docker workflow publishes `linux/amd64` and `linux/arm64` images as `walfa/chora-cli`.

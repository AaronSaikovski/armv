<div align="center">

# ARMV — <u>A</u>zure <u>R</u>esource <u>M</u>oveability <u>V</u>alidator



[![Build Status](https://github.com/AaronSaikovski/armv/workflows/build/badge.svg)](https://github.com/AaronSaikovski/armv/actions)
[![Release](https://github.com/AaronSaikovski/armv/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/AaronSaikovski/armv/actions/workflows/goreleaser.yml)
[![License](https://img.shields.io/github/license/AaronSaikovski/armv)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/AaronSaikovski/armv)](go.mod)

A lightweight Go utility for validating Azure resource moveability — **read-only**, no state changes.

</div>

> **⚠️ ARMV IS STRICTLY READ-ONLY.** It reports whether resources in a source resource group *could* be moved to a target group. It never performs the move.

---

## Overview

ARMV wraps Azure's [Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources?view=rest-resources-2021-04-01) and produces a timestamped Markdown validation report. It's the Go successor to the deprecated [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) Python utility — a single self-contained binary with no runtime dependencies.

Single CLI mode:

- **CLI mode** (`armv …`) — interactive terminal use with a progress bar, coloured summary banner, and a timestamped Markdown output file.

### Features

- **Non-destructive** — pure validation; no resources are ever mutated
- **Flexible auth** — `az login`, service principal secret, or the full `DefaultAzureCredential` chain (env vars, managed identity, workload identity)
- **Cross-subscription** — source and target may live in different subscriptions (same tenant)
- **Bounded polling** — long-running operation polled with a 30-minute ceiling and respects `Ctrl-C`
- **Markdown reports** — success/failure pages with per-resource failure tables and full JSON for forensics
- **Progress bar** (CLI) — renders live status for long-running calls
- **Hardened file I/O** — output files created with `0640` / directories with `0750` permissions
- **Cross-platform builds** — signed, reproducible binaries for Linux, macOS, Windows (amd64/arm64/386/armv7)
- **CI-enforced quality** — `go vet`, `staticcheck`, `golangci-lint`, `govulncheck`, race-enabled tests on every push

### Flow

1. Validate source/target subscription IDs (UUID format)
2. Resolve a credential: `DefaultAzureCredential` (`az login` / env vars / managed identity) or a service principal (tenant/client/secret)
3. Confirm access to the source subscription
4. Verify both resource groups exist; enumerate source resources
5. Start the Azure validate-move long-running operation
6. Poll with a progress bar until the operation completes or the 30-minute ceiling is hit
7. Write a timestamped Markdown file `output-YYYY-MM-DD-HH-MM-SS.md` and print a coloured summary banner.

### Response codes

| HTTP | Meaning | Output |
|------|---------|--------|
| **204** | All resources are movable | Success banner |
| **409** | Conflicts detected | Pretty-printed JSON error body |

### Example error report

```json
{
  "error": {
    "code": "ResourceMoveValidationFailed",
    "message": "The resource batch move request has '1' validation errors. Diagnostic information: timestamp '20240520T034539Z', tracking Id '8f53448f-e108-4f51-85d4-259e2137761d', request correlation Id '0a88b427-06ea-4045-98f1-7d2c4aaf2867'.",
    "details": [
      {
        "code": "ResourceMoveNotSupported",
        "target": "/subscriptions/<subID>/resourceGroups/src-rsg/providers/Microsoft.ContainerInstance/containerGroups/aciresource",
        "message": "Resource move is not supported for resource types 'Microsoft.ContainerInstance/containerGroups'."
      }
    ]
  }
}
```

---

## Installation

### Download a prebuilt binary

Prebuilt archives are published on every `v*` tag:

📦 [GitHub Releases](https://github.com/AaronSaikovski/armv/releases)

| OS | Architectures |
|----|---------------|
| Linux | amd64, arm64, 386, armv7 |
| macOS | amd64, arm64 |
| Windows | amd64, 386 |

Each release includes a `sha256` checksum file and per-archive SBOMs.

### macOS Security

If you downloaded a pre-built binary from a GitHub Release and macOS blocks it with "App can't be opened because Apple cannot check it for malicious software", run:

```bash
xattr -d com.apple.quarantine ./gogoodwe
```

Alternatively, right-click the binary and select **Open** from the context menu, then confirm when prompted.


### Install via `go install`

```bash
go install github.com/AaronSaikovski/armv/cmd/armv@latest
```

### Build from source

Requires Go **1.26+** and [Task](https://taskfile.dev/) (optional but recommended):

```bash
git clone https://github.com/AaronSaikovski/armv.git
cd armv
task release     # builds bin/armv (stripped, trimpath, version-injected)
# or: go build -trimpath -ldflags="-s -w -X main.version=dev" -o bin/armv ./cmd/armv
```

---

## Authentication

The **CLI** uses Azure's `DefaultAzureCredential` chain, which resolves credentials in this order: environment variables → workload identity → managed identity → Azure CLI. The simplest path is `az login`:

```bash
az login
az account set --subscription "<your-subscription-id>"
```

Service principal credentials work transparently when the standard Azure SDK environment variables are present (`AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET` or `AZURE_CLIENT_CERTIFICATE_PATH`); `DefaultAzureCredential` picks them up automatically.

---

## Usage

### Flags

Flag names are **kebab-case**. Required flags are marked with ⬤.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--source-subscription-id` ⬤ | string | — | Source Azure subscription ID (UUID) |
| `--source-resource-group` ⬤ | string | — | Source resource group name |
| `--target-subscription-id` ⬤ | string | — | Target Azure subscription ID (UUID) |
| `--target-resource-group` ⬤ | string | — | Target resource group name |
| `--output-path` | string | `./output` | Directory to write the report file |
| `--debug` | bool | `false` | Print elapsed time on exit |
| `--version` | — | — | Print version, commit and build date |
| `--help` | — | — | Show help |

### Examples

Validate a cross-subscription move:

```bash
armv \
  --source-subscription-id 12345678-1234-1234-1234-123456789012 \
  --source-resource-group  rg-prod-east \
  --target-subscription-id 87654321-4321-4321-4321-210987654321 \
  --target-resource-group  rg-dev-west
```

Same-subscription move with a custom output path and timing information:

```bash
armv \
  --source-subscription-id 12345678-1234-1234-1234-123456789012 \
  --source-resource-group  source-rg \
  --target-subscription-id 12345678-1234-1234-1234-123456789012 \
  --target-resource-group  target-rg \
  --output-path /var/log/armv-reports \
  --debug
```

The CLI prints progress to stdout, a coloured summary banner (green on success, red on failure), then writes the full Markdown report to the output directory:

```
Logged into Subscription Id: 12345678-1234-1234-1234-123456789012
 100% |████████████████████████████████| [2m45s]

*****************************************************************
*** SUCCESS - No Azure Resource Validation issues found. ***
*** Response Status OK - 204 No Content ***
*****************************************************************

***  Output file written to: - ./output ***
```

### Output file

On completion ARMV writes a timestamped **Markdown** report:

```
./output/output-2026-04-20-10-45-12.md
```

**Success report (HTTP 204)**

```markdown
# Azure Resource Move Validation Report

- **Generated:** 2026-04-20 10:45:12 UTC
- **Status:** SUCCESS
- **Source:** `<sub-id>` / `source-rg`
- **Target:** `<sub-id>` / `target-rg`
- **Resources validated:** 12
- **HTTP status:** 204 No Content

No validation issues found. All resources are eligible to move.
```

**Failure report (HTTP 409)**

```markdown
# Azure Resource Move Validation Report

- **Generated:** 2026-04-20 10:45:12 UTC
- **Status:** FAILED (1 error)
- **Source:** `<sub-id>` / `source-rg`
- **Target:** `<sub-id>` / `target-rg`
- **Resources validated:** 12
- **HTTP status:** 409 Conflict
- **Top-level code:** `ResourceMoveValidationFailed`

> The resource batch move request has '1' validation errors...

## Summary

| # | Resource Type | Name | Code |
|---|---|---|---|
| 1 | Microsoft.ContainerInstance/containerGroups | aciresource | ResourceMoveNotSupported |

## Details

### 1. aciresource
- **Type:** `Microsoft.ContainerInstance/containerGroups`
- **Resource ID:** `/subscriptions/.../aciresource`
- **Code:** `ResourceMoveNotSupported`
- **Message:** Resource move is not supported for resource types 'Microsoft.ContainerInstance/containerGroups'.

## Raw Azure API Response

\`\`\`json
{ ...full pretty-printed Azure API response... }
\`\`\`
```

The report contains:
- **Header** — timestamp, source/target subscriptions and resource groups, resource count, HTTP status
- **Summary table** — every failing resource with type, name, and error code
- **Details** — per-resource full resource ID, code, and message
- **Raw Azure response** — pretty-printed JSON for forensics

<!-- MCP Server Mode section disabled
---

## MCP Server Mode

In addition to running as a CLI, ARMV can expose its validation engine as a [Model Context Protocol](https://modelcontextprotocol.io) server. This lets LLM-based agents (Claude Desktop, Claude Code, VS Code MCP extensions, custom MCP clients) invoke resource-move validation as a tool.

Built on the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk), the server speaks MCP over **stdio** (standard input/output). The client is responsible for launching the `armv` binary as a subprocess; communication happens via newline-delimited JSON-RPC on the child's stdin/stdout.

### Starting the Server

```bash
./armv mcp serve
```

The server runs in the foreground and blocks until the client disconnects or the process is cancelled. It emits **no output on stdout** other than MCP protocol messages — any logs, errors, or debug information go to stderr.

### Exposed Tools

| Tool | Description |
|------|-------------|
| `validate_move` | Validate whether all resources in a source resource group can be moved to a target resource group (optionally in a different subscription) without performing the move. |
| `list_subscriptions` | List every Azure subscription the supplied credential can see. Used as the first step in a discovery flow so the LLM can offer the user a picklist instead of asking them to recall UUIDs. |
| `list_resource_groups` | List every resource group in a given subscription. |
| `list_resources` | List every Azure resource in a given resource group (name, type, location, ARM ID). Useful for inspecting what's in an RG before validating, or for pinpointing a likely blocker. |

All four tools share the same credential model — `bearer_token` > SP triple > `DefaultAzureCredential`. See [Credential selection](#credential-selection-priority-order) below.

#### Typical Discovery Flow

```
User:     "I want to validate moving something from one of my subs."
LLM:      → list_subscriptions
          "You have 3: prod-east, dev-west, sandbox. Which one?"
User:     "dev-west"
LLM:      → list_resource_groups(subscription_id: dev-west)
          "7 RGs: rg-app, rg-data, rg-network… which?"
User:     "rg-app, move to prod-east."
LLM:      → list_resource_groups(subscription_id: prod-east)     (confirms target RG exists)
          → validate_move(source/target …)
          "24 of 27 resources can move; the Container Instance is blocking."
```

The LLM chains the calls itself based on the user's natural-language intent and the tool descriptions exposed via `tools/list`.

#### `validate_move` — Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source_subscription_id` | string (UUID) | yes | Source Azure subscription ID |
| `source_resource_group` | string | yes | Source resource group name |
| `target_subscription_id` | string (UUID) | yes | Target Azure subscription ID |
| `target_resource_group` | string | yes | Target resource group name |
| `tenant_id` | string (UUID) | no | Service principal tenant ID |
| `client_id` | string (UUID) | no | Service principal client (application) ID |
| `client_secret` | string | no | Service principal client secret |
| `bearer_token` | string | no | Pre-fetched Azure AD bearer token for `https://management.azure.com` |

#### Credential selection (priority order)

1. **`bearer_token`** — if supplied, the server uses it directly and stores no credentials. The client is responsible for fetching the token (e.g. `az account get-access-token --resource https://management.azure.com`) and refreshing it when it expires (~1 hour). Mixing `bearer_token` with SP fields is rejected.
2. **Service principal** — all three of `tenant_id` / `client_id` / `client_secret` present. Supplying only one or two is rejected.
3. **`DefaultAzureCredential`** — fallback when no auth fields are supplied. Walks the standard Azure credential chain: environment variables, workload identity, managed identity, `az login`.

For local desktop use, option 3 with `az login` is the most ergonomic — no secrets anywhere. For sensitive environments where no credentials should ever reach the server process, option 1 (bearer token) is recommended.

All four tools accept the same optional auth fields, so a single credential strategy works across the whole discovery flow.

#### Discovery Tool Schemas

**`list_subscriptions`** — input is just the four auth fields (no resource parameters). Output contains `subscriptions[].subscription_id`, `subscriptions[].display_name`, `subscriptions[].state`, `subscriptions[].id`, and `count`.

**`list_resource_groups`** — additional required input: `subscription_id`. Output contains `resource_groups[].name`, `resource_groups[].id`, `resource_groups[].location`, plus the echoed `subscription_id` and `count`.

**`list_resources`** — additional required inputs: `subscription_id`, `resource_group`. Output contains `resources[].name`, `resources[].type`, `resources[].id`, `resources[].location`, plus echoed `subscription_id`, `resource_group`, and `count`.

#### `validate_move` — Output Schema

| Field | Type | Description |
|-------|------|-------------|
| `success` | bool | `true` when the Azure API returned HTTP 204 |
| `resource_ids` | string[] | Every resource enumerated in the source resource group |
| `target_resource_group_id` | string | Fully qualified ID of the target resource group |
| `http_status_code` | int | HTTP status code of the validate-move response (204 = ok, 409 = conflict) |
| `http_status` | string | HTTP status string |
| `diagnostics` | string | Raw response body — typically the 409 error payload when validation fails |

### Connecting a Client

Most clients drive ARMV through a configuration file — below are the common ones.

#### Claude Desktop

Edit `claude_desktop_config.json`:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "armv": {
      "command": "/absolute/path/to/armv",
      "args": ["mcp", "serve"]
    }
  }
}
```

Restart Claude Desktop. The tools appear in the picker.

#### Claude Code (CLI)

```bash
claude mcp add armv --command /absolute/path/to/armv --args mcp --args serve
```

Or edit `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "armv": { "command": "/absolute/path/to/armv", "args": ["mcp", "serve"] }
  }
}
```

#### VS Code (with an MCP extension)

`.vscode/mcp.json`:

```json
{
  "servers": {
    "armv": { "type": "stdio", "command": "/absolute/path/to/armv", "args": ["mcp", "serve"] }
  }
}
```

#### MCP Inspector (debugging)

```bash
npx @modelcontextprotocol/inspector /absolute/path/to/armv mcp serve
```

The inspector UI lists each tool with its input and output JSON schemas and lets you invoke it interactively.

#### Passing Credentials via Environment

Configure service principal credentials once at the client level and omit them from tool calls — `DefaultAzureCredential` picks them up:

```json
{
  "mcpServers": {
    "armv": {
      "command": "/absolute/path/to/armv",
      "args": ["mcp", "serve"],
      "env": {
        "AZURE_TENANT_ID": "<tenant-uuid>",
        "AZURE_CLIENT_ID": "<client-uuid>",
        "AZURE_CLIENT_SECRET": "<secret>"
      }
    }
  }
}
```

Swap `AZURE_CLIENT_SECRET` for `AZURE_CLIENT_CERTIFICATE_PATH` to use a cert-based SP.

#### Client-Supplied Bearer Token

Fetch a token client-side and pass it per-call — no Azure credential material lives on the server:

```bash
az account get-access-token --resource https://management.azure.com --query accessToken -o tsv
```

Pass the resulting string as `bearer_token` in the tool arguments. If the token is expired, the Azure API returns 401; the client fetches a fresh one and retries.

### Example Invocation

```json
{
  "name": "validate_move",
  "arguments": {
    "source_subscription_id": "12345678-1234-1234-1234-123456789012",
    "source_resource_group": "rg-prod-east",
    "target_subscription_id": "87654321-4321-4321-4321-210987654321",
    "target_resource_group": "rg-dev-west"
  }
}
```

A successful response:

```json
{
  "success": true,
  "resource_ids": ["/subscriptions/.../rg-prod-east/providers/..."],
  "target_resource_group_id": "/subscriptions/.../rg-dev-west",
  "http_status_code": 204,
  "http_status": "204 No Content"
}
```

A failed response sets `success: false`, `http_status_code: 409`, and includes the full Azure error payload in `diagnostics`.

### Progress Notifications

Azure validate-move can take minutes. The server emits MCP `notifications/progress` at every phase transition and on each 2-second poll tick, so clients can render a live status line:

```
Verifying Azure credentials
Enumerating resource groups and resources
Starting Azure validate-move for 27 resource(s)
Polling Azure validate-move (elapsed 2s)
…
Validation complete (HTTP 204)
```

Clients opt in by including a `progressToken` in the tool call (Claude Desktop, Claude Code, and MCP Inspector all do this automatically). Without a token, the server skips notifications entirely.

### Timeouts and Cancellation

| Layer | Limit |
|-------|-------|
| Server polling ceiling | **30 minutes** (hard cap via `context.WithTimeout` in `PollAndCollect`) |
| Poll interval | **2 seconds** — one progress update per tick |
| MCP client per-call deadline | **Client-specific** (Claude Desktop typically ~60 seconds) |

`notifications/cancelled` from the client propagates into the Azure SDK's `ctx`; the in-flight call aborts at the next poll boundary (within ~2 seconds) and returns a cancellation error. Validate-move is read-only, so no cleanup is required.

### Recommended LLM

Small tool-capable models are plenty — only four tools and short UUID-shaped inputs. **Claude Haiku 4.5** is the default pick (fast, cheap, high tool-use accuracy). Step up to **Sonnet 4.6** when the LLM needs to reason about large 409 diagnostics, propose remediations, or plan multi-RG migrations. Open-weight models work too (Qwen 2.5 Instruct 14B+, Llama 3.3 70B Instruct, Hermes 3) — see the [official MCP docs](https://modelcontextprotocol.io) for client configuration details.

---
-->

## Architecture

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **CLI** | `cmd/armv/app/` | Cobra root + flag parsing, CLI workflow orchestration |
| **Validator core** | `internal/pkg/validator/` | Library-friendly end-to-end validation flow — presentation-free |
| **Authentication** | `internal/pkg/auth/` | `DefaultAzureCredential`, `ClientSecretCredential`, `StaticTokenCredential` (bearer token) |
| **Validation** | `internal/pkg/validation/` | `AzureResourceMoveInfo` state + `BeginValidateMoveResources` wrapper |
| **Resource management** | `internal/pkg/resourcegroups/`, `internal/pkg/resources/` | RG + resource enumeration |
| **Polling** | `cmd/armv/poller/` | Interactive (`PollApi`) for CLI |
| **Utilities** | `pkg/utils/` | UUID validation, file I/O with hardened permissions, JSON helpers, console output |

```
cmd/armv/                          # Binary entry point
├── main.go                        # version/commit/date ldflags vars; bootstraps cobra
├── app/                           # Orchestration layer
│   ├── command.go                 # cobra root + flag binding
│   ├── root.go                    # run() — end-to-end CLI workflow + Config
│   ├── login.go                   # CheckLogin wrapper
│   └── resourcegroup.go           # RG lookup + resource enumeration driver
└── poller/                        # Azure long-running-operation handling
    ├── pollapi.go                 # Generic PollApi[T] — CLI progress bar + ctx-aware timer
    ├── report.go                  # ValidationReport / RenderMarkdown / ParseResourceID
    ├── pollresponse.go            # writeOutput: build ValidationReport, render .md
    ├── pollerresponsedata.go      # Response DTO
    ├── progressbar.go             # schollz/progressbar wiring
    └── constants.go               # StatusMoveOK/StatusMoveFailure, timings

internal/pkg/                      # Internal (module-private) packages
├── auth/
│   ├── auth.go                    # DefaultAzureCredential, ClientSecretCredential, client factories, ListSubscriptions
│   └── bearer.go                  # StaticTokenCredential for client-supplied bearer tokens
├── validator/
│   └── validator.go               # library-friendly Validate()
├── validation/
│   ├── azureresourcemoveinfo.go   # Workflow state struct
│   └── validatemove.go            # BeginValidateMoveResources caller
├── resourcegroups/resourcegroups.go
└── resources/resources.go

pkg/utils/                         # Public helpers (imported by tests)
├── args.go                        # Args struct + FormatVersion
├── validateinput.go               # UUID regex
├── outputfile.go                  # Mkdir/WriteFile with hardened permissions
├── output.go                      # OutputSuccess + OutputFailSummary console banners
└── jsonutils.go                   # any-based (un)marshal + pretty-print

test/                              # Black-box tests (separate package)
├── args_test.go
├── azureresourcemoveinfo_test.go
├── command_test.go
├── jsonutils_test.go
├── outputfile_test.go
├── pollerresponsedata_test.go
└── validateinput_test.go

.github/workflows/
├── build.yml                      # vet + golangci-lint + staticcheck + govulncheck + race tests + multi-OS build
└── goreleaser.yml                 # tag-triggered cross-platform release with SBOMs

.goreleaser.yaml                   # goreleaser v2 config (trimpath, -s -w, SBOMs, checksums)
Taskfile.yml                       # Cross-platform task runner
```

Credentials flow as the `azcore.TokenCredential` interface end-to-end so the credential implementation stays decoupled from the domain model.

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/Azure/azure-sdk-for-go/sdk/azcore` | v1.21.1 | Azure SDK core |
| `github.com/Azure/azure-sdk-for-go/sdk/azidentity` | v1.13.1 | `DefaultAzureCredential` |
| `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources` | v1.2.0 | Resources API client |
| `github.com/spf13/cobra` | v1.10.2 | CLI framework |
| `github.com/schollz/progressbar/v3` | v3.19.0 | Progress bar |
| `github.com/logrusorgru/aurora` | v2.0.3 | ANSI colour output |

See [`go.mod`](./go.mod) for the complete set, including transitive pins.

---

## Development

### Tooling

- **Go 1.26+**
- **Task** — [taskfile.dev](https://taskfile.dev/)
- Optional: `staticcheck`, `golangci-lint`, `govulncheck`, `goreleaser`

### Tasks

```bash
task                 # list all tasks
task build           # debug build → bin/armv
task release         # stripped, trimpath, version-injected build (runs vet+lint+seccheck first)
task run             # go run ./cmd/armv
task debug           # run with params sourced from envs/dev.env
task test            # race-enabled unit tests with coverage profile
task test-verbose    # same with -v
task cover           # per-function coverage summary (depends on task test)
task vet             # go vet
task lint            # go fmt + go mod tidy
task staticcheck     # staticcheck ./...
task golangci        # golangci-lint run ./...
task seccheck        # govulncheck ./...
task ci              # vet + staticcheck + seccheck + test (the local CI combo)
task goreleaser      # local cross-platform snapshot via goreleaser
task goreleaser-check  # validate .goreleaser.yaml
task deps            # go mod tidy + download + verify
task deps-upgrade    # go get -u ./... + tidy
task clean           # clean caches + remove bin/ dist/ coverage.out
```

Linux/macOS-only tasks (require `bash`):

```bash
task deploy          # deploy test Azure resources via Bicep
task destroy         # tear them down
```

Windows developers should run `task deploy-win` / `task destroy-win` (Git Bash or WSL required).

### Environment file for `task debug`

Create `./envs/dev.env` from the sample:

```bash
cp envs/sample.env envs/dev.env
# edit with your values
```

### Running a single test

```bash
go test -v ./test/ -run TestCheckValidSubscriptionID
go test -v ./test/ -run TestArgsFieldAssignment
```

---

## CI / Release pipeline

### `.github/workflows/build.yml`

Four jobs run on every push and pull request to `main`:

1. **lint** — `gofmt` drift check, `go mod tidy` drift check, `go vet`, `golangci-lint`, `staticcheck`
2. **vulncheck** — `govulncheck ./...`
3. **test** — race-enabled unit tests with coverage on Ubuntu / Windows / macOS (matrix, fail-fast: false)
4. **build** — full release-flag build on all three OSes to verify release binaries link correctly

### `.github/workflows/goreleaser.yml`

Triggered on `v*` tags (and manual dispatch). Builds the release matrix, generates SBOMs with `syft`, publishes archives and a `sha256` checksum file to the GitHub release.

### Release flags

Release builds use:

```
go build -trimpath \
  -ldflags="-s -w \
    -X main.version=<tag> \
    -X main.commit=<short-sha> \
    -X main.date=<commit-iso-date>" \
  -o bin/armv ./cmd/armv
```

`CGO_ENABLED=0` produces statically linked binaries; `-trimpath` and `mod_timestamp` make builds reproducible.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `DefaultAzureCredential: failed to acquire a token` | No active Azure login | `az login` then `az account set --subscription <id>` |
| `invalid source subscription ID format` | Malformed UUID | Match `00000000-0000-0000-0000-000000000000` |
| `source resource group "<name>" does not exist` | Typo or wrong subscription | Confirm with `az group show --name <name>` |
| `no resources found in source resource group` | Empty RG | Nothing to validate; add resources or choose another RG |
| `polling timeout or cancelled: context deadline exceeded` | 30-minute ceiling hit | Azure-side operation stalled. Check [status.azure.com](https://status.azure.com/) and retry |

---

## Limitations

- Same Azure tenant only
- Single source resource group per invocation
- Authentication is limited to the `DefaultAzureCredential` chain (no service-principal flag flow)

---

## Contributing

Issues and PRs are welcome:

- 🐛 [Report a bug](https://github.com/AaronSaikovski/armv/issues)
- 💡 [Suggest a feature](https://github.com/AaronSaikovski/armv/issues)

Before opening a PR:

```bash
task ci   # runs vet + staticcheck + seccheck + race tests locally
```

Please include in any bug report:

- Output of `armv --version`
- Go version (`go version`)
- Reproduction steps
- The generated output file, if one was produced

---

## License

[MIT](LICENSE) © Aaron Saikovski.

## Related

- [Azure — Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources)
- [Azure SDK for Go](https://learn.microsoft.com/en-us/azure/developer/go/overview)
- [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) — deprecated Python predecessor

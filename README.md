<div align="center">

# ARMV — <u>A</u>zure <u>R</u>esource <u>M</u>oveability <u>V</u>alidator

**v1.3.0**

[![Build Status](https://github.com/AaronSaikovski/armv/workflows/build/badge.svg)](https://github.com/AaronSaikovski/armv/actions)
[![Release](https://github.com/AaronSaikovski/armv/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/AaronSaikovski/armv/actions/workflows/goreleaser.yml)
[![License](https://img.shields.io/github/license/AaronSaikovski/armv)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/AaronSaikovski/armv)](go.mod)

A lightweight CLI for validating Azure resource moveability — **read-only**, no state changes.

</div>

> **⚠️ ARMV IS STRICTLY READ-ONLY.** It reports whether resources in a source resource group *could* be moved to a target group. It never performs the move.

---

## Overview

ARMV wraps Azure's [Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources?view=rest-resources-2021-04-01) and produces a timestamped validation report. It's the Go successor to the deprecated [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) Python utility — a single self-contained binary with no runtime dependencies.

### Features

- **Non-destructive** — pure validation; no resources are ever mutated
- **Cross-subscription** — source and target may live in different subscriptions (same tenant)
- **Bounded polling** — long-running operation polled with a 30-minute ceiling and respects `Ctrl-C`
- **Detailed diagnostics** — failed validations are written as pretty-printed JSON with Azure tracking/correlation IDs
- **Progress bar** — visual feedback during polling
- **Hardened file I/O** — output files created with `0640` / directories with `0750` permissions
- **Cross-platform builds** — signed, reproducible binaries for Linux, macOS, Windows (amd64/arm64/386/armv7)
- **CI-enforced quality** — `go vet`, `staticcheck`, `golangci-lint`, `govulncheck`, race-enabled tests on every push

### Flow

1. Validate source/target subscription IDs (UUID format)
2. Acquire `DefaultAzureCredential` (from `az login` context)
3. Confirm access to the source subscription
4. Verify both resource groups exist; enumerate source resources
5. Start the Azure validate-move long-running operation
6. Poll with progress bar until the operation completes or the 30-minute ceiling is hit
7. Write a timestamped output file: `output-YYYY-MM-DD-HH-MM-SS.txt`

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

ARMV uses Azure's `DefaultAzureCredential` chain, which resolves credentials in this order: environment variables → managed identity → Azure CLI. The simplest path is `az login`:

```bash
az login
az account set --subscription "<your-subscription-id>"
```

> ARMV does **not** accept service-principal credentials via flags or environment variables. Use `az login` (interactive or device-code) or an Azure-hosted managed identity.

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

---

## Architecture

```
cmd/armv/                          # Binary entry point
├── main.go                        # Version-injection var; bootstraps cobra
├── app/                           # Orchestration layer
│   ├── command.go                 # cobra flag definitions
│   ├── root.go                    # run() — end-to-end workflow + Config
│   ├── login.go                   # CheckLogin wrapper
│   └── resourcegroup.go           # RG lookup + resource enumeration driver
└── poller/                        # Azure long-running-operation handling
    ├── pollapi.go                 # Generic PollApi[T] with ctx-aware timer
    ├── pollresponse.go            # Response formatting (204 / 409 / empty)
    ├── pollerresponsedata.go      # Response DTO
    ├── progressbar.go             # schollz/progressbar wiring
    └── constants.go               # StatusMoveOK/StatusMoveFailure, timings

internal/pkg/                      # Internal (module-private) packages
├── auth/auth.go                   # DefaultAzureCredential + client factories
├── validation/
│   ├── azureresourcemoveinfo.go   # Workflow state struct
│   └── validatemove.go            # BeginValidateMoveResources caller
├── resourcegroups/resourcegroups.go
└── resources/resources.go

pkg/utils/                         # Public helpers (imported by tests)
├── args.go                        # Args struct + FormatVersion
├── validateinput.go               # UUID regex
├── outputfile.go                  # Mkdir/WriteFile with hardened permissions
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

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
- **Cross-platform builds** — reproducible, checksummed binaries for Linux, macOS, Windows (amd64/arm64/386/armv7)
- **CI-enforced quality** — `go vet`, `gofmt` drift check, `staticcheck`, `govulncheck`, and tests on every push

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

Each release includes a `sha256` checksum file.

### macOS Security

If you downloaded a pre-built binary from a GitHub Release and macOS blocks it with "App can't be opened because Apple cannot check it for malicious software", run:

```bash
xattr -d com.apple.quarantine ./armv
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

## Architecture

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **CLI** | `cmd/armv/app/` | Cobra root + flag parsing, CLI workflow orchestration (`run()`) |
| **Authentication** | `internal/pkg/auth/` | `DefaultAzureCredential` + Azure client factories |
| **Validation** | `internal/pkg/validation/` | `AzureResourceMoveInfo` state + `BeginValidateMoveResources` wrapper |
| **Resource management** | `internal/pkg/resourcegroups/`, `internal/pkg/resources/` | RG existence checks + resource enumeration |
| **Polling** | `cmd/armv/poller/` | `PollApi` — drives the long-running operation, renders the report |
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
│   └── auth.go                    # DefaultAzureCredential, login check, client factories
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
├── report_test.go
└── validateinput_test.go

.github/workflows/
├── build.yml                      # vet + gofmt drift + test + build (push to main, PRs)
├── test.yml                       # test + staticcheck + govulncheck (all branches, PRs)
├── goreleaser.yml                 # tag-triggered cross-platform release
└── release.yml                    # tag-triggered release (test + goreleaser)

.goreleaser.yaml                   # goreleaser v2 config (trimpath, -s -w, checksums)
Taskfile.yml                       # Cross-platform task runner
```

Credentials flow as the `azcore.TokenCredential` interface end-to-end so the credential implementation stays decoupled from the domain model.

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/Azure/azure-sdk-for-go/sdk/azcore` | v1.22.0 | Azure SDK core |
| `github.com/Azure/azure-sdk-for-go/sdk/azidentity` | v1.14.0 | `DefaultAzureCredential` |
| `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources` | v1.2.0 | Resources API client |
| `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription` | v1.2.0 | Subscription access check |
| `github.com/spf13/cobra` | v1.10.2 | CLI framework |
| `github.com/schollz/progressbar/v3` | v3.19.1 | Progress bar |
| `github.com/logrusorgru/aurora` | v2.0.3 | ANSI colour output |

See [`go.mod`](./go.mod) for the complete set, including transitive pins.

---

## Development

### Tooling

- **Go 1.26+**
- **Task** — [taskfile.dev](https://taskfile.dev/)
- Optional: `staticcheck`, `govulncheck`, `goreleaser`

### Tasks

Task names come straight from [`Taskfile.yml`](./Taskfile.yml):

```bash
task build           # debug build → bin/armv
task release         # stripped, trimpath, version-injected build (runs lint first)
task run             # go run ./cmd/armv --help
task test            # go test -v ./...
task vet             # go vet ./...
task lint            # go fmt + go mod tidy + go fix
task staticcheck     # staticcheck ./...
task seccheck        # govulncheck ./...
task generate        # go generate ./cmd/armv
task deps            # go mod tidy + download + go get -u
task goreleaser      # local cross-platform snapshot via goreleaser
task clean           # clean caches + remove bin/ dist/
```

### Running a single test

```bash
go test -v ./test/ -run TestCheckValidSubscriptionID
go test -v ./cmd/armv/poller/ -run TestWriteOutputEndToEnd
```

---

## CI / Release pipeline

### `.github/workflows/build.yml`

Runs on every push to `main` and on pull requests: `go vet`, a `gofmt` drift check (`go fmt ./... && git diff --exit-code`), `go test`, and a `go build` to confirm the binary links.

### `.github/workflows/test.yml`

Runs on every push (all branches) and on pull requests: `go test`, `staticcheck`, and `govulncheck`.

### `.github/workflows/goreleaser.yml` and `release.yml`

Both trigger on `v*` tags and run `goreleaser release --clean`, which builds the cross-platform matrix and publishes archives plus a `sha256` checksum file to the GitHub release.

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

Before opening a PR, run the same checks CI does:

```bash
task vet && task staticcheck && task seccheck && task test
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

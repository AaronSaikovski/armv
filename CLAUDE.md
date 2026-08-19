# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

ARMV (Azure Resource Movability Validator) is a single-binary Go CLI that wraps Azure's
[Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources).
It is **strictly read-only**: it reports whether resources in a source resource group
*could* move to a target group, and never performs the move. Output is a timestamped
Markdown report plus a coloured terminal summary banner.

## Commands

The project uses [Task](https://taskfile.dev/) (`Taskfile.yml`). Raw `go` commands work too.
Requires Go 1.26+ (see `go.mod`).

- `task build` — debug build to `bin/armv`
- `task release` — runs `lint` then a stripped (`-s -w`) release build
- `task run` — `go run ./cmd/armv --help`
- `task test` — `go test -v ./...`
- `task lint` — `go fmt ./...`, `go mod tidy -v`, `go fix ./...`
- `task vet` / `task staticcheck` / `task seccheck` — `go vet`, staticcheck, govulncheck
- `task goreleaser` — local snapshot cross-platform build
- `task generate` — `go generate ./cmd/armv` (build-version generation)

Run a single test: `go test -v -run TestName ./test/` (or the specific package path).

**CI gates** (`.github/workflows/`): `go vet`, `go fmt` with `git diff --exit-code`
(formatting drift fails the build), `go test`, `staticcheck`, and `govulncheck`. Keep code
`gofmt`-clean before committing.

## Running it

Requires an Azure credential — simplest is `az login`. Auth uses `DefaultAzureCredential`
(env vars → workload identity → managed identity → az CLI), so service-principal env vars
(`AZURE_TENANT_ID`/`AZURE_CLIENT_ID`/`AZURE_CLIENT_SECRET`) are picked up automatically.

All four flags are required and kebab-case: `--source-subscription-id`, `--source-resource-group`,
`--target-subscription-id`, `--target-resource-group`. Optional: `--output-path` (default
`./output`), `--debug` (prints elapsed time).

## Architecture

The run is a linear pipeline orchestrated in `cmd/armv/app/root.go` (`run()`), which is the
best entry point for understanding the whole flow:

1. **Validate flags** — subscription IDs must be UUIDs (`utils.CheckValidSubscriptionID`).
2. **Resolve credential** — `auth.GetAzureDefaultCredential()`.
3. **State object** — `validation.NewAzureResourceMoveInfo(...)` carries subscription/RG/credential
   state through the rest of the pipeline.
4. **Check login** (`app/login.go`) — confirms source-subscription access.
5. **Resolve resource groups** (`app/resourcegroup.go`) — verifies both RGs exist, enumerates
   source resource IDs, resolves the target RG's ID. Errors if the source RG is empty.
6. **Start LRO** — `AzureResourceMoveInfo.ValidateMove()` begins the long-running validate-move
   operation and returns an Azure SDK `runtime.Poller`.
7. **Poll + report** — `poller.PollApi()` drives the poller to completion and writes the report.

### Package layout

- `cmd/armv/main.go` — entry point; injects `version`/`commit`/`date` via `-ldflags -X`.
- `cmd/armv/app/` — cobra command wiring (`command.go` builds the root command) and the
  workflow orchestration steps (`root.go`, `login.go`, `resourcegroup.go`).
- `cmd/armv/poller/` — polling loop and report generation. `PollApi` (`pollapi.go`) is bounded
  by a 30-minute timeout and honours context cancellation at every wait point. `report.go`
  turns the raw API response into a `ValidationReport` and renders Markdown. Status codes:
  **204** = movable (success), **409** = conflicts (JSON error body → failure report).
- `internal/pkg/auth/` — credential and client construction (subscriptions, resources clients).
- `internal/pkg/validation/` — the `AzureResourceMoveInfo` state struct and `ValidateMove`.
- `internal/pkg/resourcegroups/`, `internal/pkg/resources/` — RG existence checks and resource
  enumeration.
- `pkg/utils/` — arg struct, input validation, JSON pretty-printing, hardened file output
  (files `0640`, dirs `0750`).

### Tests

Tests live in **two** places: a top-level `test/` package (black-box tests importing the public
API) and `_test.go` files inside `cmd/armv/poller/` (`pollresponse_test.go`, `report_test.go` —
the only in-package tests). `go test ./...` covers both.

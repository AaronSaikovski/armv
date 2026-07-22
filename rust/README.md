# ARMV — Rust port

**Status: `0.0.1-alpha`.** A faithful Rust rewrite of the Go `armv` CLI
(Azure Resource Movability Validator), developed side-by-side under `rust/`.
The Go binary at the repo root stays canonical until this port reaches full
parity; behavior here is matched to it byte-for-byte.

ARMV is **strictly read-only**: it drives Azure's `validateMoveResources`
long-running operation to report whether every resource in a source resource
group *could* move to a target group, writes a timestamped Markdown report,
and prints a coloured summary banner. It never performs the move.

See [`CHANGELOG.md`](CHANGELOG.md) for release notes.

## Azure SDK usage

Auth and transport use the official new-generation Azure SDK crates
(`azure_identity` and `azure_core`, the `0.30` line). Every ARM request runs
through `azure_core`'s HTTP pipeline with a `BearerTokenAuthorizationPolicy`,
so token acquisition, caching, and refresh come from the SDK rather than
custom code.

The new-generation SDK ships no management-plane crate exposing
`validateMoveResources` and omits `DefaultAzureCredential`, so two small
pieces are hand-written on top of the official crates:

- **The 5 ARM endpoint definitions** (`src/azure/client.rs`) — subscription
  get, resource-group existence (HEAD) and get, resource listing (with
  `nextLink` pagination), and the `validateMoveResources` LRO. Each is a
  single request sent through the shared pipeline; all HTTP glue is isolated
  in one `send_raw` helper.
- **The credential chain** (`src/auth.rs`) — reproduces Go's
  `DefaultAzureCredential` ordering (environment → workload identity →
  managed identity → developer tools) by composing official `azure_identity`
  credentials and exposing them as one `azure_core::TokenCredential`.

### Why not a generated management crate?

Checked (Feb 2026): there is no usable new-generation resource-management
crate yet. The `azure_resourcemanager*` names on crates.io (including
`azure_resourcemanager` itself, plus 180+ per-provider crates) are all
published at `0.0.1` as **placeholders** — the crate contains only a
`main.rs` that prints "Coming soon", with no dependencies and no API. They
are reserved names, not shipped code.

The old-generation `azure_mgmt_resources` *does* provide generated
resource-group, resource-listing, and `validate_move_resources` clients, but
it is built on `azure_core` ~0.21 with an incompatible `TokenCredential`;
adopting it would mean reverting the entire stack (pipeline, credential
chain, transport) to the unmaintained old generation. Not worth it to delete
~150 lines.

**Migration path:** when the real `azure_resourcemanager` crate ships against
the `azure_core` 0.30+ line, replacing `src/azure/client.rs` and
`src/azure/models.rs` with its generated resources client should be a
self-contained change — the credential chain in `src/auth.rs` stays, and all
HTTP glue is already isolated in `send_raw` for exactly this swap.

## Parity with the Go binary

Error message strings, exit codes, the Markdown report format, output
filenames, file permissions, and the success/failure banner bytes all match
the Go binary — deliberately including its quirks:

- A 409 validation failure still exits **0** (only hard errors exit 1).
- `--target-subscription-id` is validated and shown in the report but is
  never used for any Azure call.
- The final console line prints the output **directory**, not the generated
  filename.
- The report's `**HTTP status:**` line duplicates the code
  (`409 409 Conflict`), because the status text is the full HTTP status line.

The port adds a few things the Go build lacks, for usability (see
deviations): the `--exclude-resource-types` filter, cyan per-step status
lines, `--`-prefixed missing-flag names, prompt Ctrl-C/SIGTERM cancellation
at every pipeline step, and a short managed-identity probe timeout so
off-Azure runs don't hang on IMDS.

## Usage

All four flags are required and kebab-case; auth resolves via the credential
chain above (simplest: `az login`).

```bash
armv \
  --source-subscription-id 00000000-0000-0000-0000-000000000000 \
  --source-resource-group  rg-source \
  --target-subscription-id 11111111-1111-1111-1111-111111111111 \
  --target-resource-group  rg-target \
  --output-path ./output \                              # optional, default ./output
  --exclude-resource-types Microsoft.Web/certificates \ # optional, repeatable
  --debug                                               # optional, elapsed time + verbose logging
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--source-subscription-id` | string | — | **Required.** Source subscription ID (bare UUID). |
| `--source-resource-group` | string | — | **Required.** Source resource group name. |
| `--target-subscription-id` | string | — | **Required.** Target subscription ID (bare UUID). |
| `--target-resource-group` | string | — | **Required.** Target resource group name. |
| `--output-path` | string | `./output` | Directory for the report (and, with `--debug`, the log file). |
| `--exclude-resource-types` | strings | *(none)* | Resource `provider/type`s to exclude before validation. Repeatable and comma-separated; matched case-insensitively. |
| `--debug` | bool | `false` | Print elapsed time and enable verbose logging (stderr + `armv-debug-*.log`). |
| `-v`, `--version` | flag | — | Print version and exit. |
| `-h`, `--help` | flag | — | Print help and exit. |

`--debug` prints the elapsed time (as in the Go build) and additionally
enables verbose `tracing` diagnostics — each HTTP request and response,
resource counts, exclusions, credential selection, and poll transitions.
The output is written to **both stderr and a `armv-debug-*.log` file** in the
output directory (alongside the report); the log path is printed at startup.
Set `RUST_LOG` (e.g. `RUST_LOG=armv=debug,azure_core=debug`) to widen the
filter; without `--debug` no logging subscriber is installed, so normal runs
stay quiet.

Subscription IDs must be bare UUIDs. On completion a
`output-YYYY-MM-DD-HH-MM-SS.md` report is written under `--output-path`
(files `0640`, directories `0750`), and a green (204) or red (409) banner is
printed. `--version` and `--help` behave as in the Go build.

`--exclude-resource-types` (Rust-only) drops resources of the given types
before validation — useful for types known not to be movable (e.g.
`Microsoft.Web/certificates`). It is repeatable and comma-separated
(`--exclude-resource-types A,B --exclude-resource-types C`), matched
case-insensitively against each resource's `provider/type`. If every
resource is excluded the run stops with a clear error. When any resources
are excluded, the report adds an `## Excluded Resources` table (and an
`Excluded (by type)` count in the header) listing them for reference.

## Build, test, lint

```bash
cd rust
cargo build                                # or: task rust:build  (from repo root)
cargo test                                 #     task rust:test
cargo clippy --all-targets -- -D warnings  #     task rust:lint
cargo build --release                      #     task rust:release
```

The integration tests (`tests/`) spin up a `wiremock` server and drive the
built binary end-to-end through the real `azure_core` pipeline via the
test-only `ARMV_ENDPOINT` override; the unit tests assert byte-for-byte
parity with the Go output.

Version metadata: `build.rs` embeds the version/commit/date shown by
`--version`. The version defaults to the Cargo package version
(`0.0.1-alpha` — idiomatic for Rust, the crate carries its own version), and
the commit/date come from `git` (`rev-parse --short HEAD` /
`show -s --format=%cI`). Each can be overridden by the
`ARMV_VERSION`/`ARMV_COMMIT`/`ARMV_DATE` env vars for release pipelines;
missing git values fall back to `none`/`unknown`.

## Module map

| File | Go counterpart | Responsibility |
|------|----------------|----------------|
| `src/cli.rs` | `cmd/armv/app/command.go` | flag parsing with cobra-parity error text |
| `src/app.rs` | `cmd/armv/app/{root,login,resourcegroup}.go` | the validation pipeline |
| `src/auth.rs` | `internal/pkg/auth` | `DefaultAzureCredential`-equivalent chain |
| `src/azure/` | `internal/pkg/{auth,resources,resourcegroups,validation}` | the 5 ARM calls + models |
| `src/lro.rs` | `cmd/armv/poller/{pollapi,pollresponse,constants}.go` | poll loop + report writing |
| `src/progress.rs` | `cmd/armv/poller/progressbar.go` | progress bar |
| `src/report.rs` | `cmd/armv/poller/report.go` | report build + Markdown render |
| `src/output.rs` | `pkg/utils/{outputfile,output}.go` | hardened file I/O + banners |
| `src/colors.rs` | `pkg/utils/output.go` (aurora) | ANSI escapes matching aurora |
| `src/jsonfmt.rs` | `pkg/utils/jsonutils.go` | `json.Indent` (4-space) port |
| `src/validate.rs` | `pkg/utils/validateinput.go` | subscription-ID UUID check |

## Accepted deviations from the Go binary

- `--help` layout (content matches; cobra's column formatting is not
  replicated exactly).
- The missing-required-flag error prefixes each flag name with `--`
  (e.g. `required flag(s) "--source-resource-group" … not set`) so the
  correct invocation is obvious; cobra omits the prefix.
- Cyan per-step status lines (`Authenticating to Azure…`, `Enumerating
  resources…`, etc.) are printed as the pipeline runs; the Go build is
  silent between the login line and the progress bar.
- `--debug` additionally installs a `tracing` subscriber that logs verbose
  diagnostics (HTTP requests/responses, counts, credential selection, poll
  transitions) to stderr and to a `armv-debug-*.log` file in the output
  directory; the Go build's `--debug` only prints elapsed time.
- Every pre-poll Azure call is raced against cancellation, so Ctrl-C /
  SIGTERM interrupts promptly even during credential acquisition (the Go
  build only unwinds cleanly once it reaches the poll loop).
- The managed-identity credential gets a 5-second probe timeout so the chain
  fails over to `az login` quickly off-Azure instead of blocking on the
  unreachable IMDS endpoint (other credentials get 30s).
- `--exclude-resource-types` drops resources of the named types before
  validation (repeatable, comma-separated, case-insensitive); the Go build
  has no such filter.
- The `azure_core` pipeline retry policy is disabled: the LRO poll loop is
  driven explicitly and every status is interpreted by the client, so SDK
  retries would only conflict (e.g. retrying a terminal 500 for 60s).
- Azure SDK error internals: transport/auth failures render with
  `azure_core`/`azure_identity` text, not the Go azcore `ResponseError`
  format. Our own wrap prefixes (`login error:`,
  `auth: subscription "x" get:`, …) are identical.
- The environment credential supports the client-secret service principal;
  client-certificate and username/password flows are not wired up.
- The credential-chain aggregation error text differs from Go's
  `DefaultAzureCredential` (only surfaces when no credential works).
- Go `%q` escaping of exotic control characters in resource-group names.
- Progress-bar visual bytes (indicatif vs schollz); cadence and wrap
  behaviour match.
- In-flight cancellation during an HTTP poll reports
  `polling timeout or cancelled: context canceled` instead of Go's
  transport-level `poll: … context canceled` wording.
- `ARMV_ENDPOINT` (test-only) points the ARM client at a mock server with a
  static token; it is unset in production use.

## Code review notes (`0.0.1-alpha`)

Reviewed as a senior-Rust pass. Summary of the current state:

**Applied**
- Credential cache uses `OnceCell` (no lock held across `.await`); the
  winning credential is remembered without a double token-fetch.
- The `--debug` file-log writer recovers from a poisoned mutex instead of
  panicking — a logging sink can never take down the process.
- `unsafe` is forbidden package-wide via `[lints] unsafe_code = "forbid"`
  in `Cargo.toml` (covers the lib, binary, tests, and build script); the
  codebase contains zero `unsafe`.
- `azure_core` retry policy disabled so the explicitly-driven poll loop
  isn't fought by SDK retries.
- Dead `Config.version` field removed.

**Performance pass** (I/O-bound app — allocation/clarity cleanups, not
hot-path work)
- `--exclude-resource-types` matching is allocation-free
  (`eq_ignore_ascii_case`) instead of lowercasing a copy per resource.
- The poll loop was split into `poll_api` / `poll_to_terminal`, removing a
  per-completion `ReportContext` clone and a spurious 2-second wait on an
  already-terminal initial response.
- Release builds set `codegen-units = 1` alongside `lto` and `strip`.

**Known limitations / deferred (acceptable for alpha)**
- `--exclude-resource-types` matches the top-level `provider/type` only
  (e.g. `Microsoft.Web/certificates`); a nested child type such as
  `Microsoft.Web/sites/slots` would not match.
- `--debug` creates the output directory (and an `armv-debug-*.log`) eagerly,
  so it appears even when a run fails before writing a report.
- `ARMV_ENDPOINT` is a deliberate test hook that lives in the production
  path (needed to drive the real binary from the wiremock e2e tests).
- The public API surface is intentionally broad so integration tests in
  `tests/` can reach internals; not all `pub` items are part of a stable
  contract.
- The tokio runtime is multi-threaded although the workload is largely
  sequential I/O.

**Coverage:** ~80 tests — pure-logic unit tests plus wiremock end-to-end
tests exercising the full pipeline through the real `azure_core` stack.

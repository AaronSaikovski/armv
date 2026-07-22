# Changelog

All notable changes to the ARMV **Rust port** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1-alpha] - 2026-07-22

Initial alpha of the Rust rewrite of the Go `armv` CLI, developed side-by-side
under `rust/`. Behaviour is matched byte-for-byte to the Go binary (error
strings, exit codes, Markdown report, output filenames, `0640`/`0750`
permissions, and the coloured banner) — including its quirks (a 409 validation
failure exits `0`; the final line prints the output directory).

### Added

- **Full pipeline port** — flag parsing, credential resolution, source/target
  resource-group checks, resource enumeration (with `nextLink` pagination),
  the `validateMoveResources` long-running operation, poll loop, timestamped
  Markdown report, and coloured summary banner.
- **Official Azure SDK integration** — auth and transport use `azure_identity`
  and `azure_core` (0.30). ARM requests run through `azure_core`'s HTTP
  pipeline with a `BearerTokenAuthorizationPolicy`; the five ARM endpoints are
  hand-written (no new-generation management crate exists yet — see the
  README) and isolated in one `send_raw` helper.
- **`DefaultAzureCredential`-equivalent chain** — environment → workload
  identity → managed identity → developer tools, composed from official
  `azure_identity` credentials (the new-gen SDK omits `DefaultAzureCredential`).
- **`--exclude-resource-types`** (new vs. Go) — drop resource types known not
  to be movable (e.g. `Microsoft.Web/certificates`) before validation.
  Repeatable and comma-separated, matched case-insensitively against each
  resource's `provider/type`. Excluded resources are listed in the report
  under an `## Excluded Resources` table.
- **`--debug` verbose logging** (new vs. Go) — in addition to the elapsed-time
  line, installs a `tracing` subscriber writing HTTP requests/responses,
  resource counts, exclusions, credential selection, and poll transitions to
  **both** stderr and a timestamped `armv-debug-*.log` file in the output
  directory. `RUST_LOG` overrides the filter.
- **Actionable not-logged-in error** — when the initial login check fails
  because no Azure credential is available, the error is a clean one-liner:
  `not logged into Azure subscription "<id>": please run \`az login\` and
  retry`. The SDK's noisy multi-line credential-chain cause is suppressed
  from the message and shown only under `--debug`. This surfaces the message
  Go defines in `checkLogin` but never actually reaches.
- **Cyan per-step status lines** printed while the pipeline runs.
- **Prompt cancellation** — every pre-poll Azure call is raced against
  Ctrl-C / SIGTERM, so the process interrupts even during credential
  acquisition.
- **`--`-prefixed missing-flag errors** for clearer usage hints.
- **Test suite** — 76 tests: 59 pure-logic unit tests plus 17 integration
  tests (`assert_cmd` CLI tests and `wiremock` end-to-end tests exercising the
  full pipeline through the real `azure_core` stack).
- **Tooling** — `rust` CI workflow (fmt/clippy/test/build), a cross-platform
  release workflow reproducing the goreleaser artifact naming, and
  `task rust:*` targets.

### Changed

- Argument parsing migrated to [`clap`](https://docs.rs/clap) (derive),
  replacing the earlier hand-rolled cobra-parity parser (now fully removed).
  `--help`, `--version`, and usage-error output are clap-native, and usage
  errors exit with clap's code `2` (pipeline errors still exit `1`). Short
  version flag is `-V`. The default output path (`./output`) is defined once
  as a shared constant used by both the flag default and its tests.
- Version metadata for `--version` is sourced from the Cargo package version
  (idiomatic for Rust) with git supplying commit/date; env vars override for
  release pipelines.

### Fixed

- **Off-Azure hang** — the managed-identity credential gets a 5-second probe
  timeout so the chain fails over to `az login` quickly instead of blocking on
  the unreachable IMDS endpoint.
- **Deterministic polling** — the `azure_core` retry policy is disabled, so a
  terminal `500` is reported immediately rather than retried for the SDK's
  default 60-second budget.

### Security

- `unsafe` is forbidden package-wide via `[lints] unsafe_code = "forbid"`
  (lib, binary, tests, and build script); the codebase contains zero `unsafe`.

### Internal / review

- Credential cache uses `OnceCell` (no lock held across `.await`, no double
  token-fetch).
- The `--debug` file-log writer recovers from a poisoned mutex instead of
  panicking.
- `--exclude-resource-types` matching is allocation-free (`eq_ignore_ascii_case`).
- The poll loop is split into `poll_api` / `poll_to_terminal`, removing a
  per-completion `ReportContext` clone and a spurious 2-second wait on an
  already-terminal response.
- Release builds set `codegen-units = 1` alongside `lto` and `strip`.

### Known limitations

- `--exclude-resource-types` matches the top-level `provider/type` only; a
  nested child type such as `Microsoft.Web/sites/slots` will not match.
- `--debug` creates the output directory (and log file) eagerly, so they
  appear even when a run fails before writing a report.
- The environment credential supports the client-secret service principal;
  client-certificate and username/password flows are not wired up.

[0.0.1-alpha]: https://github.com/AaronSaikovski/armv/tree/feature/rust-rewrite/rust

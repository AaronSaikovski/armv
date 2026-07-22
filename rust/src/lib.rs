// ARMV - Azure Resource Movability Validator (Rust port of the Go CLI).
//
// Faithful behavioral port: console output, error strings, exit codes,
// report format, and file permissions match the Go binary byte-for-byte.
// Module mapping (Go -> Rust):
//   cmd/armv/app/{command,root,login,resourcegroup}.go -> cli.rs + app.rs
//   cmd/armv/poller/{pollapi,progressbar}.go           -> lro.rs + progress.rs
//   cmd/armv/poller/{report,pollresponse}.go           -> report.rs + lro.rs
//   internal/pkg/auth                                  -> auth.rs
//   internal/pkg/{resources,resourcegroups,validation} -> azure/
//   pkg/utils/{output,outputfile}.go                   -> output.rs + colors.rs
//   pkg/utils/jsonutils.go (json.Indent)               -> jsonfmt.rs
//   pkg/utils/validateinput.go                         -> validate.rs
//
// `unsafe` is forbidden package-wide via `[lints]` in Cargo.toml.

pub mod app;
pub mod auth;
pub mod azure;
pub mod cli;
pub mod colors;
pub mod jsonfmt;
pub mod lro;
pub mod output;
pub mod progress;
pub mod report;
pub mod validate;

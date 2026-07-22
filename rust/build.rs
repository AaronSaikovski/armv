// Build metadata injected at compile time (parity with the Go binary's
// -ldflags -X). The version is the Cargo package version by default
// (idiomatic for Rust — the crate carries its own version, e.g.
// "0.0.1-alpha"); commit/date come from git. An explicit env var overrides
// each for release pipelines.
use std::process::Command;

fn git(args: &[&str]) -> Option<String> {
    let out = Command::new("git").args(args).output().ok()?;
    if !out.status.success() {
        return None;
    }
    let s = String::from_utf8(out.stdout).ok()?;
    let s = s.trim().to_string();
    if s.is_empty() {
        None
    } else {
        Some(s)
    }
}

fn resolve(env_key: &str, git_args: &[&str], fallback: &str) -> String {
    println!("cargo:rerun-if-env-changed={env_key}");
    std::env::var(env_key)
        .ok()
        .filter(|v| !v.is_empty())
        .or_else(|| git(git_args))
        .unwrap_or_else(|| fallback.to_string())
}

fn main() {
    // Re-run when HEAD moves so the embedded commit stays fresh.
    println!("cargo:rerun-if-changed=../.git/HEAD");

    // Version: env override -> Cargo package version (CARGO_PKG_VERSION is
    // set by cargo for build scripts) -> "dev".
    println!("cargo:rerun-if-env-changed=ARMV_VERSION");
    let version = std::env::var("ARMV_VERSION")
        .ok()
        .filter(|v| !v.is_empty())
        .or_else(|| std::env::var("CARGO_PKG_VERSION").ok())
        .unwrap_or_else(|| "dev".to_string());
    let commit = resolve("ARMV_COMMIT", &["rev-parse", "--short", "HEAD"], "none");
    let date = resolve(
        "ARMV_DATE",
        &["show", "-s", "--format=%cI", "HEAD"],
        "unknown",
    );

    println!("cargo:rustc-env=ARMV_VERSION={version}");
    println!("cargo:rustc-env=ARMV_COMMIT={commit}");
    println!("cargo:rustc-env=ARMV_DATE={date}");
}

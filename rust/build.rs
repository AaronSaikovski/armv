// Build metadata parity with the Go binary, which injects version/commit/date
// via -ldflags -X. Resolution order per value: explicit env var (release
// pipelines) -> git (matching the Taskfile's `git describe` / `rev-parse` /
// `show -s --format=%cI`) -> the Go zero-values dev/none/unknown.
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

    let version = resolve(
        "ARMV_VERSION",
        &["describe", "--tags", "--always", "--dirty"],
        "dev",
    );
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

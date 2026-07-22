// Argument-level integration tests via the real binary (no network).
// Flag parsing is clap: --help/--version and usage errors are clap-native,
// and usage errors exit with code 2. Subscription-ID validation happens in
// the pipeline (not clap), so those still exit 1 with the Go-parity message.

use assert_cmd::Command;
use predicates::str::contains;

fn armv() -> Command {
    Command::cargo_bin("armv").unwrap()
}

const VALID_UUID: &str = "12345678-1234-1234-1234-123456789012";

/// The four required flags with valid values, so the only error is whatever
/// the test adds.
fn required_args() -> [&'static str; 8] {
    [
        "--source-subscription-id",
        VALID_UUID,
        "--source-resource-group",
        "rg1",
        "--target-subscription-id",
        VALID_UUID,
        "--target-resource-group",
        "rg2",
    ]
}

#[test]
fn version_flag() {
    let out = armv().arg("--version").assert().success();
    let stdout = String::from_utf8(out.get_output().stdout.clone()).unwrap();
    // clap prints "armv <version>"; our version embeds commit/date.
    assert!(stdout.starts_with("armv "), "{stdout}");
    assert!(stdout.contains("(commit "), "{stdout}");
    assert!(stdout.contains(", built "), "{stdout}");
    assert!(stdout.ends_with(")\n"), "{stdout}");
}

#[test]
fn help_flag() {
    let out = armv().arg("--help").assert().success();
    let stdout = String::from_utf8(out.get_output().stdout.clone()).unwrap();
    assert!(stdout.contains("ARMV - Azure Resource Movability Validator"));
    assert!(stdout.contains("Usage:"));
    assert!(stdout.contains("--source-subscription-id"));
    assert!(stdout.contains("--exclude-resource-types"));
}

#[test]
fn missing_all_required_flags() {
    armv()
        .assert()
        .failure()
        .code(2)
        .stderr(contains("required arguments were not provided"))
        .stderr(contains("--source-subscription-id"));
}

#[test]
fn missing_one_required_flag() {
    armv()
        .args([
            "--source-subscription-id",
            VALID_UUID,
            "--source-resource-group",
            "rg1",
            "--target-subscription-id",
            VALID_UUID,
        ])
        .assert()
        .failure()
        .code(2)
        .stderr(contains("required arguments were not provided"))
        .stderr(contains("--target-resource-group"));
}

#[test]
fn invalid_source_subscription_id() {
    // Passes clap (valid string), rejected by the pipeline's UUID check.
    armv()
        .args([
            "--source-subscription-id",
            "not-a-uuid",
            "--source-resource-group",
            "rg1",
            "--target-subscription-id",
            VALID_UUID,
            "--target-resource-group",
            "rg2",
        ])
        .assert()
        .failure()
        .code(1)
        .stderr("Error: invalid source subscription ID format: expected '00000000-0000-0000-0000-000000000000'\n");
}

#[test]
fn invalid_target_subscription_id() {
    armv()
        .args([
            "--source-subscription-id",
            VALID_UUID,
            "--source-resource-group",
            "rg1",
            "--target-subscription-id",
            "{12345678-1234-1234-1234-123456789012}",
            "--target-resource-group",
            "rg2",
        ])
        .assert()
        .failure()
        .code(1)
        .stderr("Error: invalid target subscription ID format: expected '00000000-0000-0000-0000-000000000000'\n");
}

#[test]
fn unknown_flag() {
    armv()
        .args(required_args())
        .arg("--bogus")
        .assert()
        .failure()
        .code(2)
        .stderr(contains("--bogus"));
}

#[test]
fn unknown_shorthand_flag() {
    armv()
        .args(required_args())
        .arg("-x")
        .assert()
        .failure()
        .code(2)
        .stderr(contains("-x"));
}

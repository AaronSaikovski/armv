// Argument-level integration tests via the real binary (no network):
// exit codes and exact stderr/stdout strings for everything that fails
// before authentication.

use assert_cmd::Command;

fn armv() -> Command {
    Command::cargo_bin("armv").unwrap()
}

const VALID_UUID: &str = "12345678-1234-1234-1234-123456789012";

#[test]
fn version_flag() {
    let out = armv().arg("--version").assert().success();
    let stdout = String::from_utf8(out.get_output().stdout.clone()).unwrap();
    assert!(stdout.starts_with("armv version "), "{stdout}");
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
}

#[test]
fn missing_all_required_flags() {
    armv().assert().failure().code(1).stderr(
        "Error: required flag(s) \"--source-resource-group\", \"--source-subscription-id\", \"--target-resource-group\", \"--target-subscription-id\" not set\n",
    );
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
        .code(1)
        .stderr("Error: required flag(s) \"--target-resource-group\" not set\n");
}

#[test]
fn invalid_source_subscription_id() {
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
        .arg("--bogus")
        .assert()
        .failure()
        .code(1)
        .stderr("Error: unknown flag: --bogus\n");
}

#[test]
fn unknown_shorthand_flag() {
    armv()
        .arg("-x")
        .assert()
        .failure()
        .code(1)
        .stderr("Error: unknown shorthand flag: 'x' in -x\n");
}

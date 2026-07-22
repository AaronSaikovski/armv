// End-to-end pipeline tests: the real binary against a wiremock ARM server
// (via the ARMV_ENDPOINT test override). Each successful-poll run takes a
// few seconds of real time because the Go-parity poll loop sleeps 2s per
// tick before its first poll.

use assert_cmd::Command;
use wiremock::matchers::{method, path, query_param};
use wiremock::{Mock, MockServer, ResponseTemplate};

const SUB: &str = "12345678-1234-1234-1234-123456789012";
const TGT_SUB: &str = "87654321-4321-4321-4321-210987654321";

fn armv(server_uri: &str, output_dir: &str) -> Command {
    let mut cmd = Command::cargo_bin("armv").unwrap();
    cmd.env("ARMV_ENDPOINT", server_uri)
        .args([
            "--source-subscription-id",
            SUB,
            "--source-resource-group",
            "src-rg",
            "--target-subscription-id",
            TGT_SUB,
            "--target-resource-group",
            "tgt-rg",
            "--output-path",
            output_dir,
        ])
        .timeout(std::time::Duration::from_secs(60));
    cmd
}

/// Mounts the happy-path mocks up to (and including) starting the LRO,
/// with two pages of resources exercising nextLink pagination.
async fn mount_pipeline(server: &MockServer) {
    Mock::given(method("GET"))
        .and(path(format!("/subscriptions/{SUB}")))
        .respond_with(ResponseTemplate::new(200).set_body_raw("{}", "application/json"))
        .mount(server)
        .await;

    for rg in ["src-rg", "tgt-rg"] {
        Mock::given(method("HEAD"))
            .and(path(format!("/subscriptions/{SUB}/resourcegroups/{rg}")))
            .respond_with(ResponseTemplate::new(204))
            .mount(server)
            .await;
    }

    let page2_url = format!(
        "{}/subscriptions/{SUB}/resourceGroups/src-rg/resources",
        server.uri()
    );
    Mock::given(method("GET"))
        .and(path(format!(
            "/subscriptions/{SUB}/resourceGroups/src-rg/resources"
        )))
        .and(query_param("page", "2"))
        .respond_with(ResponseTemplate::new(200).set_body_raw(
            r#"{"value":[{"id":"/subscriptions/s/resourceGroups/src-rg/providers/Microsoft.Web/sites/app1"},{"noid":true}]}"#,
            "application/json",
        ))
        .mount(server)
        .await;
    Mock::given(method("GET"))
        .and(path(format!(
            "/subscriptions/{SUB}/resourceGroups/src-rg/resources"
        )))
        .respond_with(
            ResponseTemplate::new(200).set_body_raw(
                format!(
                    r#"{{"value":[{{"id":"/subscriptions/s/resourceGroups/src-rg/providers/Microsoft.Storage/storageAccounts/stg1"}}],"nextLink":"{page2_url}?page=2&api-version=2022-09-01"}}"#,
                ),
                "application/json",
            ),
        )
        .mount(server)
        .await;

    Mock::given(method("GET"))
        .and(path(format!("/subscriptions/{SUB}/resourcegroups/tgt-rg")))
        .respond_with(ResponseTemplate::new(200).set_body_raw(
            r#"{"id":"/subscriptions/s/resourceGroups/tgt-rg","name":"tgt-rg"}"#,
            "application/json",
        ))
        .mount(server)
        .await;

    Mock::given(method("POST"))
        .and(path(format!(
            "/subscriptions/{SUB}/resourceGroups/src-rg/validateMoveResources"
        )))
        .respond_with(
            ResponseTemplate::new(202)
                .insert_header("Location", format!("{}/lro/poll", server.uri()).as_str()),
        )
        .mount(server)
        .await;
}

fn report_files(dir: &std::path::Path) -> Vec<String> {
    std::fs::read_dir(dir)
        .map(|entries| {
            entries
                .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
                .collect()
        })
        .unwrap_or_default()
}

#[tokio::test]
async fn success_204_end_to_end() {
    let server = MockServer::start().await;
    mount_pipeline(&server).await;
    // First poll in-progress, then terminal 204.
    Mock::given(method("GET"))
        .and(path("/lro/poll"))
        .respond_with(ResponseTemplate::new(202))
        .up_to_n_times(1)
        .mount(&server)
        .await;
    Mock::given(method("GET"))
        .and(path("/lro/poll"))
        .respond_with(ResponseTemplate::new(204))
        .mount(&server)
        .await;

    let dir = tempfile::tempdir().unwrap();
    let out = dir.path().join("reports");
    let uri = server.uri();
    let out_str = out.to_str().unwrap().to_string();

    let assert = tokio::task::spawn_blocking(move || {
        armv(&uri, &out_str).assert().success()
    })
    .await
    .unwrap();

    let stdout = String::from_utf8(assert.get_output().stdout.clone()).unwrap();
    assert!(stdout.contains(&format!(
        "\x1b[33mLogged into Subscription Id: {SUB}\x1b[0m\n"
    )));
    assert!(stdout.contains(
        "\x1b[1;32m*** SUCCESS - No Azure Resource Validation issues found. ***\x1b[0m"
    ));
    assert!(stdout.contains(
        "\x1b[32m*** Response Status OK - \x1b[0;32m204 No Content\x1b[0;32m ***\x1b[0m"
    ));
    assert!(stdout.contains("***  Output file written to: - "));

    let files = report_files(&out);
    assert_eq!(files.len(), 1);
    let content = std::fs::read_to_string(out.join(&files[0])).unwrap();
    assert!(content.contains("- **Status:** SUCCESS"));
    // Both pages of resources were counted (nil-id entry skipped).
    assert!(content.contains("- **Resources validated:** 2"));
    assert!(content.contains(&format!("- **Source:** `{SUB}` / `src-rg`")));
    assert!(content.contains(&format!("- **Target:** `{TGT_SUB}` / `tgt-rg`")));
}

#[tokio::test]
async fn conflict_409_exits_zero_with_failure_banner() {
    let server = MockServer::start().await;
    mount_pipeline(&server).await;
    let error_body = r#"{"error":{"code":"ResourceMoveValidationFailed","message":"The resource batch move request has '1' validation errors.","details":[{"code":"ResourceMoveNotSupported","target":"/subscriptions/s/resourceGroups/src-rg/providers/Microsoft.ContainerInstance/containerGroups/aci1","message":"Resource move is not supported."}]}}"#;
    Mock::given(method("GET"))
        .and(path("/lro/poll"))
        .respond_with(ResponseTemplate::new(409).set_body_raw(error_body, "application/json"))
        .mount(&server)
        .await;

    let dir = tempfile::tempdir().unwrap();
    let out = dir.path().join("reports");
    let uri = server.uri();
    let out_str = out.to_str().unwrap().to_string();

    // Validation failure still exits 0 (Go quirk, preserved).
    let assert = tokio::task::spawn_blocking(move || {
        armv(&uri, &out_str).assert().success()
    })
    .await
    .unwrap();

    let stdout = String::from_utf8(assert.get_output().stdout.clone()).unwrap();
    assert!(stdout.contains("\x1b[1;31m*** Validation FAILED ***\x1b[0m"));
    assert!(stdout.contains("\x1b[31m*** 1 resource(s) reported errors ***\x1b[0m"));
    assert!(stdout.contains("\x1b[31m*** Top failures: aci1 ***\x1b[0m"));

    let files = report_files(&out);
    assert_eq!(files.len(), 1);
    let content = std::fs::read_to_string(out.join(&files[0])).unwrap();
    assert!(content.contains("- **Status:** FAILED (1 error)"));
    assert!(content.contains("- **HTTP status:** 409 409 Conflict"));
    assert!(content.contains("- **Top-level code:** `ResourceMoveValidationFailed`"));
    assert!(content.contains("| 1 | Microsoft.ContainerInstance/containerGroups | aci1 | ResourceMoveNotSupported |"));
    assert!(content.contains("## Raw Azure API Response"));
    assert!(content.contains("    \"error\": {"));
}

#[tokio::test]
async fn not_logged_in_shows_login_error() {
    let server = MockServer::start().await;
    // The very first call (the login check) is rejected as unauthorized,
    // which is what a caller who hasn't run `az login` effectively sees.
    Mock::given(method("GET"))
        .and(path(format!("/subscriptions/{SUB}")))
        .respond_with(ResponseTemplate::new(401).set_body_raw(
            r#"{"error":{"code":"AuthenticationFailed","message":"no credential"}}"#,
            "application/json",
        ))
        .mount(&server)
        .await;

    let dir = tempfile::tempdir().unwrap();
    let uri = server.uri();
    let out_str = dir.path().to_str().unwrap().to_string();
    tokio::task::spawn_blocking(move || {
        armv(&uri, &out_str)
            .assert()
            .failure()
            .code(1)
            // Exact match: the message is clean (no SDK credential-chain dump
            // trailing it); that noise is only emitted under --debug.
            .stderr(format!(
                "Error: not logged into Azure subscription \"{SUB}\": please run `az login` and retry\n"
            ))
    })
    .await
    .unwrap();
}

#[tokio::test]
async fn missing_source_resource_group() {
    let server = MockServer::start().await;
    Mock::given(method("GET"))
        .and(path(format!("/subscriptions/{SUB}")))
        .respond_with(ResponseTemplate::new(200).set_body_raw("{}", "application/json"))
        .mount(&server)
        .await;
    Mock::given(method("HEAD"))
        .and(path(format!("/subscriptions/{SUB}/resourcegroups/src-rg")))
        .respond_with(ResponseTemplate::new(404))
        .mount(&server)
        .await;

    let dir = tempfile::tempdir().unwrap();
    let uri = server.uri();
    let out_str = dir.path().to_str().unwrap().to_string();
    tokio::task::spawn_blocking(move || {
        armv(&uri, &out_str)
            .assert()
            .failure()
            .code(1)
            .stderr("Error: source resource group \"src-rg\" does not exist\n")
    })
    .await
    .unwrap();
}

#[tokio::test]
async fn empty_source_resource_group() {
    let server = MockServer::start().await;
    Mock::given(method("GET"))
        .and(path(format!("/subscriptions/{SUB}")))
        .respond_with(ResponseTemplate::new(200).set_body_raw("{}", "application/json"))
        .mount(&server)
        .await;
    for rg in ["src-rg", "tgt-rg"] {
        Mock::given(method("HEAD"))
            .and(path(format!("/subscriptions/{SUB}/resourcegroups/{rg}")))
            .respond_with(ResponseTemplate::new(204))
            .mount(&server)
            .await;
    }
    Mock::given(method("GET"))
        .and(path(format!(
            "/subscriptions/{SUB}/resourceGroups/src-rg/resources"
        )))
        .respond_with(
            ResponseTemplate::new(200).set_body_raw(r#"{"value":[]}"#, "application/json"),
        )
        .mount(&server)
        .await;

    let dir = tempfile::tempdir().unwrap();
    let uri = server.uri();
    let out_str = dir.path().to_str().unwrap().to_string();
    tokio::task::spawn_blocking(move || {
        armv(&uri, &out_str)
            .assert()
            .failure()
            .code(1)
            .stderr("Error: no resources found in source resource group \"src-rg\"\n")
    })
    .await
    .unwrap();
}

#[tokio::test]
async fn terminal_500_becomes_failed_report_exit_zero() {
    let server = MockServer::start().await;
    mount_pipeline(&server).await;
    Mock::given(method("GET"))
        .and(path("/lro/poll"))
        .respond_with(
            ResponseTemplate::new(500).set_body_raw("<html>oops</html>", "text/html"),
        )
        .mount(&server)
        .await;

    let dir = tempfile::tempdir().unwrap();
    let out = dir.path().join("reports");
    let uri = server.uri();
    let out_str = out.to_str().unwrap().to_string();
    tokio::task::spawn_blocking(move || {
        armv(&uri, &out_str).assert().success()
    })
    .await
    .unwrap();

    let files = report_files(&out);
    assert_eq!(files.len(), 1);
    let content = std::fs::read_to_string(out.join(&files[0])).unwrap();
    assert!(content.contains("- **Status:** FAILED (0 errors)"));
    assert!(content.contains("- **HTTP status:** 500 500 Internal Server Error"));
    assert!(content.contains("```json\n<html>oops</html>\n```"));
}

#[tokio::test]
async fn debug_flag_prints_elapsed_time() {
    let server = MockServer::start().await;
    mount_pipeline(&server).await;
    Mock::given(method("GET"))
        .and(path("/lro/poll"))
        .respond_with(ResponseTemplate::new(204))
        .mount(&server)
        .await;

    let dir = tempfile::tempdir().unwrap();
    let uri = server.uri();
    let out_str = dir.path().join("r").to_str().unwrap().to_string();
    let assert = tokio::task::spawn_blocking(move || {
        let mut cmd = armv(&uri, &out_str);
        cmd.arg("--debug");
        cmd.assert().success()
    })
    .await
    .unwrap();
    let stdout = String::from_utf8(assert.get_output().stdout.clone()).unwrap();
    let re_ok = stdout
        .lines()
        .any(|l| l.starts_with("Elapsed time: ") && l.ends_with(" seconds"));
    assert!(re_ok, "expected elapsed-time line in: {stdout}");

    // --debug also turns on verbose tracing to stderr.
    let stderr = String::from_utf8(assert.get_output().stderr.clone()).unwrap();
    assert!(
        stderr.contains("confirmed access to subscription"),
        "expected verbose debug logging in stderr: {stderr}"
    );

    // ...and a debug log file is written to the output directory.
    let log_dir = dir.path().join("r");
    let logs: Vec<_> = std::fs::read_dir(&log_dir)
        .unwrap()
        .filter_map(Result::ok)
        .filter(|e| e.file_name().to_string_lossy().starts_with("armv-debug-"))
        .collect();
    assert_eq!(logs.len(), 1, "expected one debug log file");
    let log_content = std::fs::read_to_string(logs[0].path()).unwrap();
    assert!(log_content.contains("confirmed access to subscription"));
}

#[tokio::test]
async fn excludes_resource_types_from_validation() {
    let server = MockServer::start().await;
    mount_pipeline(&server).await; // src-rg has 2 resources: a Storage account and a Web site
    Mock::given(method("GET"))
        .and(path("/lro/poll"))
        .respond_with(ResponseTemplate::new(204))
        .mount(&server)
        .await;

    let dir = tempfile::tempdir().unwrap();
    let out = dir.path().join("reports");
    let uri = server.uri();
    let out_str = out.to_str().unwrap().to_string();

    let assert = tokio::task::spawn_blocking(move || {
        let mut cmd = armv(&uri, &out_str);
        cmd.args(["--exclude-resource-types", "Microsoft.Storage/storageAccounts"]);
        cmd.assert().success()
    })
    .await
    .unwrap();

    let stdout = String::from_utf8(assert.get_output().stdout.clone()).unwrap();
    assert!(stdout.contains("Excluded 1 resource(s) matching --exclude-resource-types"));

    let files = report_files(&out);
    assert_eq!(files.len(), 1);
    let content = std::fs::read_to_string(out.join(&files[0])).unwrap();
    // One of the two resources was excluded, so only one is validated.
    assert!(content.contains("- **Resources validated:** 1"));
    // ...and the excluded resource is listed in the report for reference.
    assert!(content.contains("- **Excluded (by type):** 1"));
    assert!(content.contains("## Excluded Resources"));
    assert!(content.contains("| 1 | Microsoft.Storage/storageAccounts | stg1 |"));
}

#[tokio::test]
async fn all_resources_excluded_errors() {
    let server = MockServer::start().await;
    mount_pipeline(&server).await; // src-rg's only two types are Storage + Web/sites

    let dir = tempfile::tempdir().unwrap();
    let uri = server.uri();
    let out_str = dir.path().to_str().unwrap().to_string();

    // Excluding both types empties the list; the run stops before validating.
    tokio::task::spawn_blocking(move || {
        let mut cmd = armv(&uri, &out_str);
        cmd.args([
            "--exclude-resource-types",
            "Microsoft.Storage/storageAccounts,Microsoft.Web/sites",
        ]);
        cmd.assert().failure().code(1).stderr(
            "Error: all resources in source resource group \"src-rg\" were excluded by --exclude-resource-types\n",
        )
    })
    .await
    .unwrap();
}

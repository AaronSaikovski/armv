// Port of cmd/armv/poller/{pollapi.go, pollresponse.go, constants.go}.

use std::time::Duration;

use anyhow::Context as _;
use tokio_util::sync::CancellationToken;

use crate::azure::{ArmClient, BeginMove, PollStatus};
use crate::progress::Progress;
use crate::report::{
    build_validation_report, render_markdown, ReportContext, ValidationReport, STATUS_MOVE_OK,
};

pub const PROGRESS_BAR_MAX: u32 = 100;
/// Azure long-running operations typically take minutes; poll every 2s.
pub const SLEEP_DURATION: Duration = Duration::from_secs(2);
pub const POLLING_TIMEOUT: Duration = Duration::from_secs(30 * 60);

/// Drives the LRO to completion, writing the Markdown report to
/// output_path and returning the parsed ValidationReport (Go PollApi). An
/// already-terminal initial response is written straight out; otherwise the
/// poll cycle is bounded by POLLING_TIMEOUT and honours cancellation at
/// every wait point.
pub async fn poll_api(
    cancel: &CancellationToken,
    client: &ArmClient,
    begin: BeginMove,
    output_path: &str,
    report_ctx: ReportContext,
) -> anyhow::Result<ValidationReport> {
    let (status_code, status_text, body) = match begin {
        BeginMove::Immediate {
            status_code,
            status_text,
            body,
        } => (status_code, status_text, body),
        BeginMove::Poller(url) => poll_to_terminal(cancel, client, &url).await?,
    };
    write_output(&body, status_code, &status_text, output_path, report_ctx)
}

/// Polls `url` until the operation reaches a terminal status, returning
/// `(status_code, status_text, body)`. Bounded by POLLING_TIMEOUT and
/// honouring cancellation at every wait point; the cancellation/timeout
/// error embeds the literal Go ctx.Err() strings for message parity.
async fn poll_to_terminal(
    cancel: &CancellationToken,
    client: &ArmClient,
    url: &str,
) -> anyhow::Result<(u16, String, Vec<u8>)> {
    let deadline = tokio::time::Instant::now() + POLLING_TIMEOUT;
    let mut bar = Progress::new();

    loop {
        tokio::select! {
            biased;
            _ = cancel.cancelled() => {
                bar.finish();
                anyhow::bail!("polling timeout or cancelled: context canceled");
            }
            _ = tokio::time::sleep_until(deadline) => {
                bar.finish();
                anyhow::bail!("polling timeout or cancelled: context deadline exceeded");
            }
            _ = tokio::time::sleep(SLEEP_DURATION) => {}
        }

        bar.tick();

        let status = tokio::select! {
            biased;
            _ = cancel.cancelled() => {
                bar.finish();
                anyhow::bail!("polling timeout or cancelled: context canceled");
            }
            _ = tokio::time::sleep_until(deadline) => {
                bar.finish();
                anyhow::bail!("polling timeout or cancelled: context deadline exceeded");
            }
            result = client.poll_once(url) => result.context("poll")?
        };

        match status {
            PollStatus::InProgress => {
                tracing::debug!("poll: operation in progress");
                continue;
            }
            PollStatus::Terminal {
                status_code,
                status_text,
                body,
            } => {
                tracing::debug!("poll: terminal status {status_code}");
                bar.finish();
                return Ok((status_code, status_text, body));
            }
        }
    }
}

/// Builds the report, renders Markdown, and writes the timestamped output
/// file (Go writeOutput). The filename uses LOCAL time; the report's
/// Generated timestamp uses UTC - two different clocks, matching Go.
pub fn write_output(
    body: &[u8],
    status_code: u16,
    status_text: &str,
    output_path: &str,
    ctx: ReportContext,
) -> anyhow::Result<ValidationReport> {
    let file_name = chrono::Local::now()
        .format("output-%Y-%m-%d-%H-%M-%S.md")
        .to_string();

    // Pretty-print the raw Azure body (if any). Non-JSON bodies are kept
    // verbatim rather than failing the operation.
    let mut pretty_json = String::new();
    if status_code != STATUS_MOVE_OK && !body.is_empty() {
        let raw = String::from_utf8_lossy(body);
        pretty_json = match crate::jsonfmt::pretty_json_string(&raw) {
            Ok(pj) => pj,
            Err(_) => raw.into_owned(),
        };
    }

    let report = build_validation_report(status_code, status_text, body, &pretty_json, ctx);
    let markdown = render_markdown(&report);

    crate::output::write_output_file(output_path, &file_name, &markdown)
        .context("failed to write output file")?;
    Ok(report)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn ctx() -> ReportContext {
        ReportContext {
            source_subscription_id: "s".into(),
            source_resource_group: "src".into(),
            target_subscription_id: "t".into(),
            target_resource_group: "tgt".into(),
            resource_count: 1,
            excluded_resources: Vec::new(),
        }
    }

    fn md_files(dir: &std::path::Path) -> Vec<String> {
        let mut names: Vec<String> = std::fs::read_dir(dir)
            .unwrap()
            .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
            .collect();
        names.sort();
        names
    }

    fn assert_filename_shape(name: &str) {
        // output-YYYY-MM-DD-HH-MM-SS.md: stamp is 19 chars, digits and dashes.
        assert!(name.starts_with("output-") && name.ends_with(".md"), "{name}");
        let stamp = &name["output-".len()..name.len() - ".md".len()];
        assert_eq!(stamp.len(), 19, "{name}");
        assert!(stamp.bytes().all(|b| b.is_ascii_digit() || b == b'-'));
    }

    #[test]
    fn write_output_204_success() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().to_str().unwrap();
        let report = write_output(b"", 204, "204 No Content", out, ctx()).unwrap();
        assert!(report.success);
        let files = md_files(dir.path());
        assert_eq!(files.len(), 1, "exactly one file expected");
        assert_filename_shape(&files[0]);
        let content = std::fs::read_to_string(dir.path().join(&files[0])).unwrap();
        assert!(content.contains("- **Status:** SUCCESS"));
        assert!(content.contains("No validation issues found."));
    }

    #[test]
    fn write_output_409_parsed() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().to_str().unwrap();
        let body = br#"{"error":{"code":"C","message":"M","details":[{"code":"D","target":"/providers/A/B/n","message":"m"}]}}"#;
        let report = write_output(body, 409, "409 Conflict", out, ctx()).unwrap();
        assert!(!report.success);
        assert_eq!(report.errors.len(), 1);
        let files = md_files(dir.path());
        let content = std::fs::read_to_string(dir.path().join(&files[0])).unwrap();
        assert!(content.contains("- **Status:** FAILED (1 error)"));
        // Body was pretty-printed with 4-space indent into the raw block.
        assert!(content.contains("## Raw Azure API Response"));
        assert!(content.contains("    \"error\": {"));
    }

    #[test]
    fn write_output_409_empty_body() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().to_str().unwrap();
        let report = write_output(b"", 409, "409 Conflict", out, ctx()).unwrap();
        assert!(!report.success);
        assert!(report.errors.is_empty());
        let files = md_files(dir.path());
        let content = std::fs::read_to_string(dir.path().join(&files[0])).unwrap();
        assert!(content.contains("- **Status:** FAILED (0 errors)"));
        assert!(!content.contains("## Raw Azure API Response"));
    }

    #[test]
    fn write_output_409_non_json_body() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().to_str().unwrap();
        let body = b"<html>bad gateway</html>";
        let report = write_output(body, 502, "502 Bad Gateway", out, ctx()).unwrap();
        assert!(!report.success);
        assert!(report.errors.is_empty());
        let files = md_files(dir.path());
        let content = std::fs::read_to_string(dir.path().join(&files[0])).unwrap();
        // Raw body preserved verbatim in the fenced block.
        assert!(content.contains("```json\n<html>bad gateway</html>\n```"));
    }

    #[test]
    fn write_output_creates_nested_dir() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().join("nested/deeper");
        let report = write_output(b"", 204, "204 No Content", out.to_str().unwrap(), ctx()).unwrap();
        assert!(report.success);
        assert_eq!(md_files(&out).len(), 1);
    }

    #[tokio::test(start_paused = true)]
    async fn poll_api_immediate_terminal() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().to_str().unwrap();
        let cancel = CancellationToken::new();
        let client = ArmClient::with_endpoint(
            std::sync::Arc::new(crate::auth::StaticCredential("test-token".into())),
            "http://127.0.0.1:1", // never contacted on the Immediate path
        );
        let begin = BeginMove::Immediate {
            status_code: 204,
            status_text: "204 No Content".into(),
            body: Vec::new(),
        };
        let report = poll_api(&cancel, &client, begin, out, ctx()).await.unwrap();
        assert!(report.success);
        assert_eq!(md_files(dir.path()).len(), 1);
    }

    #[tokio::test(start_paused = true)]
    async fn poll_api_cancellation_message() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().to_str().unwrap();
        let cancel = CancellationToken::new();
        cancel.cancel();
        let client = ArmClient::with_endpoint(
            std::sync::Arc::new(crate::auth::StaticCredential("test-token".into())),
            "http://127.0.0.1:1",
        );
        let begin = BeginMove::Poller("http://127.0.0.1:1/poll".into());
        let err = poll_api(&cancel, &client, begin, out, ctx()).await.unwrap_err();
        assert_eq!(format!("{err:#}"), "polling timeout or cancelled: context canceled");
        assert!(md_files(dir.path()).is_empty());
    }
}

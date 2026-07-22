// Port of cmd/armv/poller/report.go. The rendered Markdown matches the Go
// output byte-for-byte, including the duplicated status code in the
// "- **HTTP status:** {code} {status_text}" line (status_text is the full
// Go http.Response.Status, e.g. "409 Conflict").

use std::fmt::Write as _;

use chrono::{DateTime, Utc};
use serde::Deserialize;

/// HTTP status codes returned by the validate-move API (poller/constants.go).
pub const STATUS_MOVE_OK: u16 = 204;
pub const STATUS_MOVE_FAILURE: u16 = 409;

/// Validation-run metadata used to populate the report header.
#[derive(Debug, Clone, Default)]
pub struct ReportContext {
    pub source_subscription_id: String,
    pub source_resource_group: String,
    pub target_subscription_id: String,
    pub target_resource_group: String,
    pub resource_count: usize,
}

/// One entry from the Azure error response's details array.
#[derive(Debug, Clone, Default, Deserialize)]
pub struct AzureErrorDetail {
    #[serde(default)]
    pub code: String,
    #[serde(default)]
    pub target: String,
    #[serde(default)]
    pub message: String,
}

/// Shape of the JSON returned by the API on a 409 response:
/// {"error": {"code", "message", "details": [{code, target, message}]}}
#[derive(Debug, Clone, Default, Deserialize)]
pub struct AzureErrorResponse {
    #[serde(default)]
    pub error: AzureErrorBody,
}

#[derive(Debug, Clone, Default, Deserialize)]
pub struct AzureErrorBody {
    #[serde(default)]
    pub code: String,
    #[serde(default)]
    pub message: String,
    #[serde(default)]
    pub details: Vec<AzureErrorDetail>,
}

/// One failing resource, flattened from AzureErrorDetail.
#[derive(Debug, Clone, Default)]
pub struct ValidationError {
    pub resource_id: String,
    pub resource_type: String,
    pub resource_name: String,
    pub code: String,
    pub message: String,
}

/// The parsed, rendered form of the API response.
#[derive(Debug, Clone)]
pub struct ValidationReport {
    pub success: bool,
    pub generated_at: DateTime<Utc>,
    pub context: ReportContext,
    pub status_code: u16,
    pub status_text: String,
    /// code+message summarising the failure (empty on success)
    pub top_level: AzureErrorDetail,
    pub errors: Vec<ValidationError>,
    pub raw_json: String,
}

/// Turns a raw API response into a ValidationReport (BuildValidationReport).
/// Success is exactly `status_code == STATUS_MOVE_OK`. On success or an
/// empty body no error parsing happens; a body that fails to parse leaves
/// top_level/errors empty with raw_json preserved for diagnosis.
pub fn build_validation_report(
    status_code: u16,
    status_text: &str,
    raw_body: &[u8],
    pretty_json: &str,
    ctx: ReportContext,
) -> ValidationReport {
    let mut report = ValidationReport {
        success: status_code == STATUS_MOVE_OK,
        generated_at: Utc::now(),
        context: ctx,
        status_code,
        status_text: status_text.to_string(),
        top_level: AzureErrorDetail::default(),
        errors: Vec::new(),
        raw_json: pretty_json.to_string(),
    };

    if report.success || raw_body.is_empty() {
        return report;
    }

    let parsed: AzureErrorResponse = match serde_json::from_slice(raw_body) {
        Ok(p) => p,
        // Parsing failed - keep the raw JSON so operators can still diagnose.
        Err(_) => return report,
    };

    report.top_level = AzureErrorDetail {
        code: parsed.error.code,
        message: parsed.error.message,
        target: String::new(),
    };
    report.errors = parsed
        .error
        .details
        .into_iter()
        .map(|d| {
            let (resource_type, resource_name) = parse_resource_id(&d.target);
            ValidationError {
                resource_id: d.target,
                resource_type,
                resource_name,
                code: d.code,
                message: d.message,
            }
        })
        .collect();
    report
}

/// Extracts (resource_type, resource_name) from an Azure resource ID like
/// /subscriptions/<sub>/resourceGroups/<rg>/providers/<ns>/<type>/<name>.
/// If the shape is not recognised, both values fall back to the original.
pub fn parse_resource_id(target: &str) -> (String, String) {
    if target.is_empty() {
        return (String::new(), String::new());
    }
    let Some(idx) = target.find("/providers/") else {
        return (target.to_string(), target.to_string());
    };
    let after = &target[idx + "/providers/".len()..];
    let parts: Vec<&str> = after.split('/').collect();
    // Expect at least: <namespace>/<type>/<name>
    if parts.len() < 3 {
        return (target.to_string(), target.to_string());
    }
    let resource_type = format!("{}/{}", parts[0], parts[1]);
    let resource_name = parts[parts.len() - 1].to_string();
    (resource_type, resource_name)
}

/// Produces the Markdown report body (RenderMarkdown).
pub fn render_markdown(r: &ValidationReport) -> String {
    let mut b = String::new();
    b.push_str("# Azure Resource Move Validation Report\n\n");

    let _ = writeln!(
        b,
        "- **Generated:** {}",
        r.generated_at.format("%Y-%m-%d %H:%M:%S UTC")
    );
    if r.success {
        b.push_str("- **Status:** SUCCESS\n");
    } else {
        let _ = writeln!(
            b,
            "- **Status:** FAILED ({} {})",
            r.errors.len(),
            pluralise("error", r.errors.len())
        );
    }
    let _ = writeln!(
        b,
        "- **Source:** `{}` / `{}`",
        r.context.source_subscription_id, r.context.source_resource_group
    );
    let _ = writeln!(
        b,
        "- **Target:** `{}` / `{}`",
        r.context.target_subscription_id, r.context.target_resource_group
    );
    let _ = writeln!(b, "- **Resources validated:** {}", r.context.resource_count);
    let _ = writeln!(b, "- **HTTP status:** {} {}", r.status_code, r.status_text);
    if !r.success && !r.top_level.code.is_empty() {
        let _ = writeln!(b, "- **Top-level code:** `{}`", r.top_level.code);
    }
    b.push('\n');

    if r.success {
        b.push_str("No validation issues found. All resources are eligible to move.\n");
        return b;
    }

    if !r.top_level.message.is_empty() {
        b.push_str("> ");
        b.push_str(&r.top_level.message);
        b.push_str("\n\n");
    }

    if !r.errors.is_empty() {
        b.push_str("## Summary\n\n");
        b.push_str("| # | Resource Type | Name | Code |\n");
        b.push_str("|---|---|---|---|\n");
        for (i, e) in r.errors.iter().enumerate() {
            let _ = writeln!(
                b,
                "| {} | {} | {} | {} |",
                i + 1,
                md_escape(&e.resource_type),
                md_escape(&e.resource_name),
                md_escape(&e.code)
            );
        }
        b.push_str("\n## Details\n\n");
        for (i, e) in r.errors.iter().enumerate() {
            let _ = writeln!(b, "### {}. {}", i + 1, e.resource_name);
            let _ = writeln!(b, "- **Type:** `{}`", e.resource_type);
            let _ = writeln!(b, "- **Resource ID:** `{}`", e.resource_id);
            let _ = writeln!(b, "- **Code:** `{}`", e.code);
            let _ = writeln!(b, "- **Message:** {}\n", e.message);
        }
    }

    if !r.raw_json.is_empty() {
        b.push_str("## Raw Azure API Response\n\n");
        b.push_str("```json\n");
        b.push_str(&r.raw_json);
        if !r.raw_json.ends_with('\n') {
            b.push('\n');
        }
        b.push_str("```\n");
    }

    b
}

/// Up to n resource names from the failed details for the console summary.
pub fn top_failure_names(r: &ValidationReport, n: usize) -> Vec<String> {
    if n == 0 || r.errors.is_empty() {
        return Vec::new();
    }
    r.errors
        .iter()
        .take(n)
        .map(|e| e.resource_name.clone())
        .collect()
}

/// Escapes the pipe character inside Markdown table cells.
pub fn md_escape(s: &str) -> String {
    s.replace('|', "\\|")
}

/// "error" for n == 1, otherwise word + "s".
pub fn pluralise(word: &str, n: usize) -> String {
    if n == 1 {
        word.to_string()
    } else {
        format!("{word}s")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn ctx() -> ReportContext {
        ReportContext {
            source_subscription_id: "src-sub".into(),
            source_resource_group: "src-rg".into(),
            target_subscription_id: "tgt-sub".into(),
            target_resource_group: "tgt-rg".into(),
            resource_count: 12,
        }
    }

    const FULL_ID: &str = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ContainerInstance/containerGroups/aciresource";

    #[test]
    fn parse_resource_id_table() {
        assert_eq!(
            parse_resource_id(FULL_ID),
            (
                "Microsoft.ContainerInstance/containerGroups".to_string(),
                "aciresource".to_string()
            )
        );
        // Child/nested resource: type from the first two segments, name from the LAST.
        assert_eq!(
            parse_resource_id("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1/extensions/ext1"),
            ("Microsoft.Compute/virtualMachines".to_string(), "ext1".to_string())
        );
        // No /providers/ -> both fall back to the raw target.
        assert_eq!(
            parse_resource_id("/subscriptions/s/resourceGroups/rg"),
            (
                "/subscriptions/s/resourceGroups/rg".to_string(),
                "/subscriptions/s/resourceGroups/rg".to_string()
            )
        );
        // Empty -> ("", "").
        assert_eq!(parse_resource_id(""), (String::new(), String::new()));
        // Fewer than 3 segments after /providers/ -> fallback.
        assert_eq!(
            parse_resource_id("/providers/Microsoft.Compute"),
            (
                "/providers/Microsoft.Compute".to_string(),
                "/providers/Microsoft.Compute".to_string()
            )
        );
        // Exactly 3 segments parses.
        assert_eq!(
            parse_resource_id("/providers/Microsoft.Compute/virtualMachines/vm1"),
            ("Microsoft.Compute/virtualMachines".to_string(), "vm1".to_string())
        );
    }

    #[test]
    fn build_204_success() {
        let r = build_validation_report(204, "204 No Content", b"", "", ctx());
        assert!(r.success);
        assert!(r.errors.is_empty());
        assert!(r.top_level.code.is_empty());
        assert!(r.raw_json.is_empty());
    }

    #[test]
    fn build_409_empty_body() {
        let r = build_validation_report(409, "409 Conflict", b"", "", ctx());
        assert!(!r.success);
        assert!(r.errors.is_empty());
        assert!(r.top_level.code.is_empty());
    }

    #[test]
    fn build_409_malformed_json_keeps_raw() {
        let r = build_validation_report(409, "409 Conflict", b"<html>bad</html>", "<html>bad</html>", ctx());
        assert!(!r.success);
        assert!(r.errors.is_empty());
        assert!(r.top_level.code.is_empty());
        assert_eq!(r.raw_json, "<html>bad</html>");
    }

    #[test]
    fn build_409_empty_details() {
        let body = br#"{"error":{"code":"C","message":"M","details":[]}}"#;
        let r = build_validation_report(409, "409 Conflict", body, "{}", ctx());
        assert!(!r.success);
        assert_eq!(r.top_level.code, "C");
        assert_eq!(r.top_level.message, "M");
        assert!(r.errors.is_empty());
    }

    #[test]
    fn build_409_missing_code_field() {
        let body = br#"{"error":{"message":"only message"}}"#;
        let r = build_validation_report(409, "409 Conflict", body, "{}", ctx());
        assert_eq!(r.top_level.code, "");
        assert_eq!(r.top_level.message, "only message");
    }

    fn failure_report(n: usize) -> ValidationReport {
        let details: Vec<String> = (0..n)
            .map(|i| {
                format!(
                    r#"{{"code":"ResourceMoveNotSupported","target":"/subscriptions/s/resourceGroups/rg/providers/Microsoft.ContainerInstance/containerGroups/res{i}","message":"msg {i}"}}"#
                )
            })
            .collect();
        let body = format!(
            r#"{{"error":{{"code":"ResourceMoveValidationFailed","message":"top message","details":[{}]}}}}"#,
            details.join(",")
        );
        build_validation_report(409, "409 Conflict", body.as_bytes(), "{\n    \"x\": 1\n}", ctx())
    }

    #[test]
    fn render_success_markdown() {
        let r = build_validation_report(204, "204 No Content", b"", "", ctx());
        let md = render_markdown(&r);
        assert!(md.starts_with("# Azure Resource Move Validation Report\n\n"));
        assert!(md.contains("- **Status:** SUCCESS\n"));
        assert!(md.contains("- **Source:** `src-sub` / `src-rg`\n"));
        assert!(md.contains("- **Target:** `tgt-sub` / `tgt-rg`\n"));
        assert!(md.contains("- **Resources validated:** 12\n"));
        assert!(md.contains("- **HTTP status:** 204 204 No Content\n"));
        assert!(md.ends_with("No validation issues found. All resources are eligible to move.\n"));
        assert!(!md.contains("## Summary"));
        assert!(!md.contains("## Details"));
        assert!(!md.contains("Top-level code"));
    }

    #[test]
    fn render_failure_markdown_single_error() {
        let md = render_markdown(&failure_report(1));
        assert!(md.contains("- **Status:** FAILED (1 error)\n"));
        assert!(md.contains("- **Top-level code:** `ResourceMoveValidationFailed`\n"));
        assert!(md.contains("> top message\n\n"));
        assert!(md.contains("## Summary\n\n| # | Resource Type | Name | Code |\n|---|---|---|---|\n"));
        assert!(md.contains("| 1 | Microsoft.ContainerInstance/containerGroups | res0 | ResourceMoveNotSupported |\n"));
        assert!(md.contains("## Details\n\n### 1. res0\n"));
        assert!(md.contains("- **Message:** msg 0\n\n"));
        assert!(md.contains("## Raw Azure API Response\n\n```json\n{\n    \"x\": 1\n}\n```\n"));
    }

    #[test]
    fn render_failure_markdown_plural() {
        let md = render_markdown(&failure_report(3));
        assert!(md.contains("- **Status:** FAILED (3 errors)\n"));
        assert!(md.contains("### 1. res0\n"));
        assert!(md.contains("### 2. res1\n"));
        assert!(md.contains("### 3. res2\n"));
    }

    #[test]
    fn render_escapes_pipes_in_table() {
        let body = br#"{"error":{"code":"c|d","message":"m","details":[{"code":"c|d","target":"/providers/A/B/n|ame","message":"x"}]}}"#;
        let r = build_validation_report(409, "409 Conflict", body, "{}", ctx());
        let md = render_markdown(&r);
        assert!(md.contains("| 1 | A/B | n\\|ame | c\\|d |\n"));
        // Details section is NOT escaped (matches Go).
        assert!(md.contains("### 1. n|ame\n"));
    }

    #[test]
    fn render_raw_json_trailing_newline() {
        let mut r = build_validation_report(409, "409 Conflict", b"", "", ctx());
        r.raw_json = "{\"a\":1}\n".into();
        assert!(render_markdown(&r).contains("```json\n{\"a\":1}\n```\n"));
        r.raw_json = "{\"a\":1}".into();
        assert!(render_markdown(&r).contains("```json\n{\"a\":1}\n```\n"));
    }

    #[test]
    fn top_failure_names_caps() {
        let r = failure_report(5);
        assert_eq!(top_failure_names(&r, 3), vec!["res0", "res1", "res2"]);
        assert_eq!(top_failure_names(&r, 10).len(), 5);
        assert_eq!(top_failure_names(&r, 1), vec!["res0"]);
        assert!(top_failure_names(&r, 0).is_empty());
        let empty = build_validation_report(204, "204 No Content", b"", "", ctx());
        assert!(top_failure_names(&empty, 3).is_empty());
    }

    #[test]
    fn helpers() {
        assert_eq!(md_escape("a|b|c"), "a\\|b\\|c");
        assert_eq!(md_escape("abc"), "abc");
        assert_eq!(pluralise("error", 1), "error");
        assert_eq!(pluralise("error", 0), "errors");
        assert_eq!(pluralise("error", 3), "errors");
    }

    #[test]
    fn generated_at_format() {
        let r = build_validation_report(204, "204 No Content", b"", "", ctx());
        let line = format!("{}", r.generated_at.format("%Y-%m-%d %H:%M:%S UTC"));
        // e.g. "2026-07-22 01:02:03 UTC"
        assert_eq!(line.len(), 23);
        assert!(line.ends_with(" UTC"));
    }
}

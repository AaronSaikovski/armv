// Port of pkg/utils/outputfile.go (hardened file output) and
// pkg/utils/output.go (console banners).

use std::path::Path;

use anyhow::Context as _;

use crate::colors;

/// Owner rwx, group r-x on created directories (Go DirPermission 0750).
pub const DIR_PERMISSION: u32 = 0o750;
/// Owner rw-, group r-- on created files (Go FilePermission 0640).
pub const FILE_PERMISSION: u32 = 0o640;

const BORDER: &str = "*****************************************************************";

/// Reports whether a file or directory exists at `path`. Parity with Go's
/// CheckExists: any stat error OTHER than not-found counts as "exists".
pub fn check_exists(path: &str) -> bool {
    match std::fs::metadata(path) {
        Ok(_) => true,
        Err(e) => e.kind() != std::io::ErrorKind::NotFound,
    }
}

/// Creates `folder_name` (and parents) with DIR_PERMISSION, idempotently.
pub fn make_folder(folder_name: &str) -> anyhow::Result<()> {
    let mut builder = std::fs::DirBuilder::new();
    builder.recursive(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::DirBuilderExt;
        builder.mode(DIR_PERMISSION);
    }
    builder
        .create(folder_name)
        .with_context(|| format!("failed to create directory {folder_name}"))?;
    Ok(())
}

/// Writes `output` to output_path/filename, creating the directory if
/// necessary. Files are created with FILE_PERMISSION; an existing file is
/// truncated and keeps its existing mode (Go os.WriteFile semantics).
pub fn write_output_file(output_path: &str, filename: &str, output: &str) -> anyhow::Result<()> {
    make_folder(output_path).context("output directory creation failed")?;

    let full_path = Path::new(output_path).join(filename);
    let mut options = std::fs::OpenOptions::new();
    options.write(true).create(true).truncate(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(FILE_PERMISSION);
    }
    let write = options
        .open(&full_path)
        .and_then(|mut f| std::io::Write::write_all(&mut f, output.as_bytes()));
    write.with_context(|| format!("failed to write file {}", full_path.display()))?;

    Ok(())
}

/// The four lines of the green success banner, with exact aurora ANSI bytes.
pub fn success_banner_lines(resp_status: &str) -> [String; 4] {
    [
        colors::bold_green(&format!("\n{BORDER}")),
        colors::bold_green("*** SUCCESS - No Azure Resource Validation issues found. ***"),
        colors::green_status_line(resp_status),
        colors::bold_green(BORDER),
    ]
}

/// Prints the green success banner (OutputSuccess).
pub fn output_success(resp_status: &str) {
    for line in success_banner_lines(resp_status) {
        println!("{line}");
    }
}

/// The lines of the red failure banner; the "Top failures" line is present
/// only when top_failures is non-empty.
pub fn fail_banner_lines(error_count: usize, top_failures: &[String]) -> Vec<String> {
    let mut lines = vec![
        colors::bold_red(&format!("\n{BORDER}")),
        colors::bold_red("*** Validation FAILED ***"),
        colors::red(&format!("*** {error_count} resource(s) reported errors ***")),
    ];
    if !top_failures.is_empty() {
        lines.push(colors::red(&format!(
            "*** Top failures: {} ***",
            top_failures.join(", ")
        )));
    }
    lines.push(colors::bold_red(BORDER));
    lines
}

/// Prints the concise red banner when validation fails (OutputFailSummary).
pub fn output_fail_summary(error_count: usize, top_failures: &[String]) {
    for line in fail_banner_lines(error_count, top_failures) {
        println!("{line}");
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn check_exists_cases() {
        let dir = tempfile::tempdir().unwrap();
        let file = dir.path().join("f.txt");
        std::fs::write(&file, "x").unwrap();
        assert!(check_exists(dir.path().to_str().unwrap()));
        assert!(check_exists(file.to_str().unwrap()));
        assert!(!check_exists(dir.path().join("missing").to_str().unwrap()));
    }

    #[test]
    fn make_folder_new_existing_idempotent_nested() {
        let dir = tempfile::tempdir().unwrap();
        let nested = dir.path().join("a/b/c");
        let nested_str = nested.to_str().unwrap();
        for _ in 0..3 {
            make_folder(nested_str).unwrap();
        }
        assert!(nested.is_dir());
    }

    #[cfg(unix)]
    #[test]
    fn created_dir_and_file_permissions() {
        use std::os::unix::fs::MetadataExt;
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().join("outdir");
        let out_str = out.to_str().unwrap();
        write_output_file(out_str, "r.md", "hello").unwrap();
        assert_eq!(std::fs::metadata(&out).unwrap().mode() & 0o777, DIR_PERMISSION);
        assert_eq!(
            std::fs::metadata(out.join("r.md")).unwrap().mode() & 0o777,
            FILE_PERMISSION
        );
    }

    #[test]
    fn write_output_file_writes_and_overwrites() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().to_str().unwrap().to_string();
        write_output_file(&out, "r.md", "first").unwrap();
        write_output_file(&out, "r.md", "second").unwrap();
        let content = std::fs::read_to_string(dir.path().join("r.md")).unwrap();
        assert_eq!(content, "second");
    }

    #[test]
    fn write_output_file_creates_nested_dir() {
        let dir = tempfile::tempdir().unwrap();
        let out = dir.path().join("x/y").to_str().unwrap().to_string();
        write_output_file(&out, "r.md", "data").unwrap();
        assert!(Path::new(&out).join("r.md").is_file());
    }

    #[test]
    fn success_banner_exact_bytes() {
        let lines = success_banner_lines("204 No Content");
        assert_eq!(lines[0], format!("\x1b[1;32m\n{BORDER}\x1b[0m"));
        assert_eq!(
            lines[1],
            "\x1b[1;32m*** SUCCESS - No Azure Resource Validation issues found. ***\x1b[0m"
        );
        assert_eq!(
            lines[2],
            "\x1b[32m*** Response Status OK - \x1b[0;32m204 No Content\x1b[0;32m ***\x1b[0m"
        );
        assert_eq!(lines[3], format!("\x1b[1;32m{BORDER}\x1b[0m"));
    }

    #[test]
    fn fail_banner_exact_bytes() {
        let lines = fail_banner_lines(2, &["a".into(), "b".into()]);
        assert_eq!(lines.len(), 5);
        assert_eq!(lines[0], format!("\x1b[1;31m\n{BORDER}\x1b[0m"));
        assert_eq!(lines[1], "\x1b[1;31m*** Validation FAILED ***\x1b[0m");
        assert_eq!(lines[2], "\x1b[31m*** 2 resource(s) reported errors ***\x1b[0m");
        assert_eq!(lines[3], "\x1b[31m*** Top failures: a, b ***\x1b[0m");
        assert_eq!(lines[4], format!("\x1b[1;31m{BORDER}\x1b[0m"));
    }

    #[test]
    fn fail_banner_no_top_failures_line_when_empty() {
        let lines = fail_banner_lines(0, &[]);
        assert_eq!(lines.len(), 4);
        assert!(!lines.iter().any(|l| l.contains("Top failures")));
    }

    #[test]
    fn border_is_65_stars() {
        assert_eq!(BORDER.len(), 65);
        assert!(BORDER.bytes().all(|b| b == b'*'));
    }
}

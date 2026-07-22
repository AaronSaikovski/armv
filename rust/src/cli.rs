// CLI definition using clap (derive). This replaces the earlier hand-rolled
// cobra-parity parser: `--help`/`--version`/usage-error text is now
// clap-native, and usage errors exit with clap's code 2 — an accepted
// deviation from the Go binary's cobra output.

use clap::Parser;

/// Default directory for output files.
pub const DEFAULT_OUTPUT_PATH: &str = "./output";

/// Application description shown in `--help`.
pub const APP_DESCRIPTION: &str = "ARMV - Azure Resource Movability Validator\n\n\
Performs a Read-Only check whether resources in a source resource group\n\
can be moved to a target resource group. The source and target may live in\n\
different subscriptions within the same tenant.";

/// Full version string shown by `--version` (build metadata injected by
/// build.rs: env override -> Cargo package version + git commit/date).
const FULL_VERSION: &str = concat!(
    env!("ARMV_VERSION"),
    " (commit ",
    env!("ARMV_COMMIT"),
    ", built ",
    env!("ARMV_DATE"),
    ")"
);

/// Command-line arguments (port of pkg/utils/args.go Args).
#[derive(Parser, Debug, Clone)]
#[command(name = "armv", about = APP_DESCRIPTION, version = FULL_VERSION)]
pub struct Args {
    /// Source Azure subscription ID (UUID)
    #[arg(long)]
    pub source_subscription_id: String,

    /// Source resource group name
    #[arg(long)]
    pub source_resource_group: String,

    /// Target Azure subscription ID (UUID)
    #[arg(long)]
    pub target_subscription_id: String,

    /// Target resource group name
    #[arg(long)]
    pub target_resource_group: String,

    /// Directory to write the report (and, with --debug, the log file)
    #[arg(long, default_value = DEFAULT_OUTPUT_PATH)]
    pub output_path: String,

    /// Resource types to exclude before validation (repeatable, comma-separated)
    #[arg(long, value_delimiter = ',')]
    pub exclude_resource_types: Vec<String>,

    /// Print elapsed time and enable verbose logging
    #[arg(long)]
    pub debug: bool,
}

/// Parses the process arguments. clap prints and exits on its own for
/// `--help`, `--version`, and usage errors (exit code 2).
pub fn parse() -> Args {
    Args::parse()
}

#[cfg(test)]
mod tests {
    use super::*;
    use clap::error::ErrorKind;

    fn base() -> Vec<&'static str> {
        vec![
            "armv",
            "--source-subscription-id",
            "sub-a",
            "--source-resource-group",
            "rg-a",
            "--target-subscription-id",
            "sub-b",
            "--target-resource-group",
            "rg-b",
        ]
    }

    #[test]
    fn all_required_parse_with_defaults() {
        let args = Args::try_parse_from(base()).unwrap();
        assert_eq!(args.source_subscription_id, "sub-a");
        assert_eq!(args.source_resource_group, "rg-a");
        assert_eq!(args.target_subscription_id, "sub-b");
        assert_eq!(args.target_resource_group, "rg-b");
        assert_eq!(args.output_path, DEFAULT_OUTPUT_PATH);
        assert!(!args.debug);
        assert!(args.exclude_resource_types.is_empty());
    }

    #[test]
    fn debug_output_path_and_exclude() {
        let mut a = base();
        a.extend([
            "--debug",
            "--output-path",
            "/tmp/x",
            "--exclude-resource-types",
            "Microsoft.Web/certificates,Microsoft.Foo/bar",
            "--exclude-resource-types",
            "Microsoft.Baz/qux",
        ]);
        let args = Args::try_parse_from(a).unwrap();
        assert!(args.debug);
        assert_eq!(args.output_path, "/tmp/x");
        assert_eq!(
            args.exclude_resource_types,
            vec![
                "Microsoft.Web/certificates".to_string(),
                "Microsoft.Foo/bar".to_string(),
                "Microsoft.Baz/qux".to_string(),
            ]
        );
    }

    #[test]
    fn missing_required_errors() {
        let err = Args::try_parse_from(["armv", "--source-subscription-id", "s"]).unwrap_err();
        assert_eq!(err.kind(), ErrorKind::MissingRequiredArgument);
    }

    #[test]
    fn unknown_flag_errors() {
        let mut a = base();
        a.push("--bogus");
        let err = Args::try_parse_from(a).unwrap_err();
        assert_eq!(err.kind(), ErrorKind::UnknownArgument);
    }

    #[test]
    fn help_and_version_are_display_errors() {
        assert_eq!(
            Args::try_parse_from(["armv", "--help"]).unwrap_err().kind(),
            ErrorKind::DisplayHelp
        );
        assert_eq!(
            Args::try_parse_from(["armv", "--version"])
                .unwrap_err()
                .kind(),
            ErrorKind::DisplayVersion
        );
    }

    #[test]
    fn app_description_guards() {
        assert!(APP_DESCRIPTION.contains("Azure Resource Movability Validator"));
        assert!(APP_DESCRIPTION.contains("Read-Only"));
        // Cross-subscription moves are supported; the description must not
        // claim otherwise.
        assert!(!APP_DESCRIPTION.contains("same subscription"));
    }
}

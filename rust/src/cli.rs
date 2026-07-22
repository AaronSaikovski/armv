// Port of cmd/armv/app/command.go (cobra/pflag). Hand-rolled parser so the
// error text matches cobra byte-for-byte: clap's messages differ and cannot
// be reshaped reliably. Behaviors replicated from cobra/pflag:
//   - "--flag=value" and "--flag value" forms; bools do not consume a
//     separate value argument
//   - "--" terminates flag parsing; positionals are accepted and ignored
//     (root command with no subcommands uses ArbitraryArgs)
//   - help beats version beats required-flag validation
//   - a flag counts as "set" if it appeared, even with an empty value
//   - missing required flags are reported in lexical order

/// Command-line arguments (port of pkg/utils/args.go Args).
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct Args {
    pub source_subscription_id: String,
    pub source_resource_group: String,
    pub target_subscription_id: String,
    pub target_resource_group: String,
    pub debug: bool,
    pub output_path: String,
    /// Resource types to drop from validation (e.g. types known not to be
    /// movable, such as "Microsoft.Web/certificates"). Repeatable and/or
    /// comma-separated; matched case-insensitively.
    pub exclude_resource_types: Vec<String>,
}

/// Default directory for output files (app/root.go DefaultOutputPath).
pub const DEFAULT_OUTPUT_PATH: &str = "./output";

/// Application description shown in help (pkg/utils/args.go AppDescription).
pub const APP_DESCRIPTION: &str = "ARMV - Azure Resource Movability Validator

Performs a Read-Only check whether resources in a source resource group
can be moved to a target resource group. The source and target may live in
different subscriptions within the same tenant.";

const REQUIRED_FLAGS: [&str; 4] = [
    "source-resource-group",
    "source-subscription-id",
    "target-resource-group",
    "target-subscription-id",
];

/// Parses argv (without the program name). Returns Ok(Some(args)) to run,
/// Ok(None) when help/version was printed (exit 0), or Err with a
/// cobra-compatible message (caller prints "Error: <msg>" and exits 1).
pub fn parse(argv: &[String], version: &str) -> anyhow::Result<Option<Args>> {
    let mut args = Args {
        output_path: DEFAULT_OUTPUT_PATH.to_string(),
        ..Args::default()
    };
    let mut help = false;
    let mut show_version = false;
    let mut set: Vec<&'static str> = Vec::new();

    let mut i = 0;
    while i < argv.len() {
        let arg = &argv[i];
        if arg == "--" {
            break;
        }
        if let Some(rest) = arg.strip_prefix("--") {
            let (name, inline_value) = match rest.split_once('=') {
                Some((n, v)) => (n, Some(v.to_string())),
                None => (rest, None),
            };
            match name {
                "help" => set_bool(&mut help, name, inline_value)?,
                "version" => set_bool(&mut show_version, name, inline_value)?,
                "debug" => set_bool(&mut args.debug, name, inline_value)?,
                "source-subscription-id"
                | "source-resource-group"
                | "target-subscription-id"
                | "target-resource-group"
                | "output-path"
                | "exclude-resource-types" => {
                    let value = match inline_value {
                        Some(v) => v,
                        None => {
                            i += 1;
                            match argv.get(i) {
                                Some(v) => v.clone(),
                                None => {
                                    anyhow::bail!("flag needs an argument: --{name}")
                                }
                            }
                        }
                    };
                    match name {
                        "source-subscription-id" => {
                            args.source_subscription_id = value;
                            set.push("source-subscription-id");
                        }
                        "source-resource-group" => {
                            args.source_resource_group = value;
                            set.push("source-resource-group");
                        }
                        "target-subscription-id" => {
                            args.target_subscription_id = value;
                            set.push("target-subscription-id");
                        }
                        "target-resource-group" => {
                            args.target_resource_group = value;
                            set.push("target-resource-group");
                        }
                        "output-path" => args.output_path = value,
                        // Repeatable and comma-separated; accumulate.
                        "exclude-resource-types" => {
                            for t in value.split(',') {
                                let t = t.trim();
                                if !t.is_empty() {
                                    args.exclude_resource_types.push(t.to_string());
                                }
                            }
                        }
                        _ => unreachable!(),
                    }
                }
                _ => anyhow::bail!("unknown flag: --{name}"),
            }
        } else if let Some(cluster) = arg.strip_prefix('-') {
            if cluster.is_empty() {
                // Bare "-" is a positional argument; ignored.
                i += 1;
                continue;
            }
            for c in cluster.chars() {
                match c {
                    'h' => help = true,
                    'v' => show_version = true,
                    _ => anyhow::bail!("unknown shorthand flag: '{c}' in -{cluster}"),
                }
            }
        }
        // Positional arguments are accepted and ignored.
        i += 1;
    }

    if help {
        print!("{}", help_text());
        return Ok(None);
    }
    if show_version {
        println!("armv version {version}");
        return Ok(None);
    }

    // Names are printed with the "--" prefix so the message makes the
    // correct invocation obvious. This intentionally diverges from cobra,
    // which omits the prefix.
    let missing: Vec<String> = REQUIRED_FLAGS
        .iter()
        .filter(|f| !set.contains(*f))
        .map(|f| format!("--{f}"))
        .collect();
    if !missing.is_empty() {
        anyhow::bail!("required flag(s) \"{}\" not set", missing.join("\", \""));
    }

    Ok(Some(args))
}

/// Sets a bool flag, replicating pflag's strconv.ParseBool error text for
/// the "--flag=value" form.
fn set_bool(target: &mut bool, name: &str, inline_value: Option<String>) -> anyhow::Result<()> {
    match inline_value {
        None => {
            *target = true;
            Ok(())
        }
        Some(v) => match v.as_str() {
            "1" | "t" | "T" | "true" | "TRUE" | "True" => {
                *target = true;
                Ok(())
            }
            "0" | "f" | "F" | "false" | "FALSE" | "False" => {
                *target = false;
                Ok(())
            }
            _ => anyhow::bail!(
                "invalid argument \"{v}\" for \"--{name}\" flag: strconv.ParseBool: parsing \"{v}\": invalid syntax"
            ),
        },
    }
}

/// Cobra-style help output. Layout parity with cobra is an accepted
/// deviation; content (description, usage, flags, defaults) matches.
fn help_text() -> String {
    format!(
        "{APP_DESCRIPTION}\n\
\n\
Usage:\n\
  armv [flags]\n\
\n\
Flags:\n\
      --debug                             Enable debug mode with timing information\n\
      --exclude-resource-types strings    Resource types to exclude from validation, e.g.\n\
                                          Microsoft.Web/certificates (repeatable/comma-separated)\n\
  -h, --help                              help for armv\n\
      --output-path string                Output path to write results (default \"./output\")\n\
      --source-resource-group string      Source Resource Group (required)\n\
      --source-subscription-id string     Source Subscription Id (required)\n\
      --target-resource-group string      Target Resource Group (required)\n\
      --target-subscription-id string     Target Subscription Id (required)\n\
  -v, --version                           version for armv\n"
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    fn v(args: &[&str]) -> Vec<String> {
        args.iter().map(|s| s.to_string()).collect()
    }

    const ALL: [&str; 8] = [
        "--source-subscription-id",
        "sub-a",
        "--source-resource-group",
        "rg-a",
        "--target-subscription-id",
        "sub-b",
        "--target-resource-group",
        "rg-b",
    ];

    #[test]
    fn all_required_present() {
        let args = parse(&v(&ALL), "x").unwrap().unwrap();
        assert_eq!(args.source_subscription_id, "sub-a");
        assert_eq!(args.source_resource_group, "rg-a");
        assert_eq!(args.target_subscription_id, "sub-b");
        assert_eq!(args.target_resource_group, "rg-b");
        assert_eq!(args.output_path, "./output");
        assert!(!args.debug);
        assert!(args.exclude_resource_types.is_empty());
    }

    #[test]
    fn exclude_resource_types_repeatable_and_comma_separated() {
        let mut a = v(&ALL);
        a.extend(v(&[
            "--exclude-resource-types",
            "Microsoft.Web/certificates, Microsoft.Foo/bar",
            "--exclude-resource-types",
            "Microsoft.Baz/qux",
        ]));
        let args = parse(&a, "x").unwrap().unwrap();
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
    fn equals_form_and_debug_and_output_path() {
        let mut a = vec![
            "--source-subscription-id=s".to_string(),
            "--source-resource-group=r1".to_string(),
            "--target-subscription-id=t".to_string(),
            "--target-resource-group=r2".to_string(),
            "--debug".to_string(),
            "--output-path=/tmp/x".to_string(),
        ];
        let args = parse(&a, "x").unwrap().unwrap();
        assert!(args.debug);
        assert_eq!(args.output_path, "/tmp/x");
        a.pop();
        a.push("--debug=false".to_string());
        let args = parse(&a, "x").unwrap().unwrap();
        assert!(!args.debug);
    }

    #[test]
    fn empty_value_counts_as_set() {
        let args = parse(
            &v(&[
                "--source-subscription-id=",
                "--source-resource-group=",
                "--target-subscription-id=",
                "--target-resource-group=",
            ]),
            "x",
        )
        .unwrap()
        .unwrap();
        assert_eq!(args.source_subscription_id, "");
    }

    #[test]
    fn missing_all_required_exact_message() {
        let err = parse(&v(&[]), "x").unwrap_err();
        assert_eq!(
            format!("{err:#}"),
            "required flag(s) \"--source-resource-group\", \"--source-subscription-id\", \"--target-resource-group\", \"--target-subscription-id\" not set"
        );
    }

    #[test]
    fn missing_one_required() {
        let err = parse(&v(&ALL[..6]), "x").unwrap_err();
        assert_eq!(
            format!("{err:#}"),
            "required flag(s) \"--target-resource-group\" not set"
        );
    }

    #[test]
    fn unknown_flag_and_shorthand() {
        let err = parse(&v(&["--bogus"]), "x").unwrap_err();
        assert_eq!(format!("{err:#}"), "unknown flag: --bogus");
        let err = parse(&v(&["-x"]), "x").unwrap_err();
        assert_eq!(format!("{err:#}"), "unknown shorthand flag: 'x' in -x");
    }

    #[test]
    fn flag_needs_argument() {
        let err = parse(&v(&["--output-path"]), "x").unwrap_err();
        assert_eq!(format!("{err:#}"), "flag needs an argument: --output-path");
    }

    #[test]
    fn invalid_bool_value() {
        let err = parse(&v(&["--debug=notabool"]), "x").unwrap_err();
        assert_eq!(
            format!("{err:#}"),
            "invalid argument \"notabool\" for \"--debug\" flag: strconv.ParseBool: parsing \"notabool\": invalid syntax"
        );
    }

    #[test]
    fn help_and_version_return_none() {
        assert!(parse(&v(&["--help"]), "x").unwrap().is_none());
        assert!(parse(&v(&["-h"]), "x").unwrap().is_none());
        assert!(parse(&v(&["--version"]), "x").unwrap().is_none());
        assert!(parse(&v(&["-v"]), "x").unwrap().is_none());
        // help beats version beats required-flag validation
        assert!(parse(&v(&["--version", "--help"]), "x").unwrap().is_none());
    }

    #[test]
    fn positionals_ignored_and_double_dash_terminates() {
        let mut a = v(&ALL);
        a.push("positional".to_string());
        assert!(parse(&a, "x").unwrap().is_some());

        // Flags after "--" are ignored, so required flags are missing.
        let a = v(&["--", "--source-subscription-id", "s"]);
        let err = parse(&a, "x").unwrap_err();
        assert!(format!("{err:#}").starts_with("required flag(s)"));
    }

    #[test]
    fn app_description_guards() {
        assert!(APP_DESCRIPTION.contains("Azure Resource Movability Validator"));
        assert!(APP_DESCRIPTION.contains("Read-Only"));
        // Cross-subscription moves are supported; the description must not
        // claim otherwise (Go args_test.go doc-bug guard).
        assert!(!APP_DESCRIPTION.contains("same subscription"));
    }
}

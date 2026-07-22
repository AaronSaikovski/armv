// Port of cmd/armv/app/{root.go, login.go, resourcegroup.go}: the linear
// validation pipeline. Every user-visible string is exact Go parity; the
// quoted names replicate Go %q double-quoting. Preserved Go quirks:
// --target-subscription-id is never used for any Azure call, and the final
// console line prints the output DIRECTORY, not the generated filename.

use anyhow::Context as _;
use tokio_util::sync::CancellationToken;

use crate::azure::ArmClient;
use crate::cli::Args;
use crate::colors;
use crate::report::{top_failure_names, ReportContext};

/// Caps how many failing resource names appear in the terminal banner;
/// the full list is always in the Markdown report.
pub const CONSOLE_TOP_FAILURES: usize = 3;

/// Application configuration resolved from CLI flags (Go app.Config).
pub struct Config {
    pub version: String,
    pub args: Args,
}

/// Executes the validation workflow end-to-end (Go run()).
pub async fn run(cancel: CancellationToken, cfg: Config) -> anyhow::Result<()> {
    if !crate::validate::check_valid_subscription_id(&cfg.args.source_subscription_id) {
        anyhow::bail!(
            "invalid source subscription ID format: expected '00000000-0000-0000-0000-000000000000'"
        );
    }
    if !crate::validate::check_valid_subscription_id(&cfg.args.target_subscription_id) {
        anyhow::bail!(
            "invalid target subscription ID format: expected '00000000-0000-0000-0000-000000000000'"
        );
    }

    // The Go defer registers after subscription-id validation, so bad UUIDs
    // never print an elapsed time - and everything after this point does,
    // success or error, before main prints "Error:".
    let debug = cfg.args.debug;
    let start = std::time::Instant::now();
    let result = run_inner(cancel, &cfg).await;
    if debug {
        println!("Elapsed time: {:.2} seconds", start.elapsed().as_secs_f64());
    }
    result
}

async fn run_inner(cancel: CancellationToken, cfg: &Config) -> anyhow::Result<()> {
    let sub = &cfg.args.source_subscription_id;
    let src_rg = &cfg.args.source_resource_group;
    let tgt_rg = &cfg.args.target_resource_group;

    // ARMV_ENDPOINT is a test-only override so integration tests can point
    // the client at a mock server (with a static token - mock servers do
    // not validate credentials). Production always takes the else branch.
    let client = match std::env::var("ARMV_ENDPOINT") {
        Ok(endpoint) if !endpoint.is_empty() => ArmClient::with_endpoint(
            std::sync::Arc::new(crate::auth::StaticCredential("test-token".into())),
            &endpoint,
        ),
        _ => {
            let cred = crate::auth::default_azure_credential()
                .context("failed to get Azure default credential")?;
            ArmClient::new(cred)
        }
    };

    // Login check against the source subscription (login.go checkLogin).
    // Every pre-poll Azure call is raced against cancellation so Ctrl-C
    // interrupts even while a credential probe or HTTP request is in flight
    // (the poll loop honours cancellation on its own).
    status("Authenticating to Azure...");
    cancellable(&cancel, client.get_subscription(sub))
        .await
        .context("login error")?;
    println!(
        "{}",
        colors::yellow(&format!("Logged into Subscription Id: {sub}"))
    );

    // Resource-group resolution (resourcegroup.go getResourceGroupInfo).
    status("Verifying source and target resource groups...");
    let src_exists = cancellable(&cancel, client.resource_group_exists(sub, src_rg))
        .await
        .with_context(|| format!("checking source resource group \"{src_rg}\""))?;
    if !src_exists {
        anyhow::bail!("source resource group \"{src_rg}\" does not exist");
    }

    let dst_exists = cancellable(&cancel, client.resource_group_exists(sub, tgt_rg))
        .await
        .with_context(|| format!("checking target resource group \"{tgt_rg}\""))?;
    if !dst_exists {
        anyhow::bail!("destination resource group \"{tgt_rg}\" does not exist");
    }

    status(&format!(
        "Enumerating resources in source resource group \"{src_rg}\"..."
    ));
    let resource_ids = cancellable(&cancel, client.list_resource_ids(sub, src_rg))
        .await
        .context("failed to get resource IDs")?;
    if resource_ids.is_empty() {
        anyhow::bail!("no resources found in source resource group \"{src_rg}\"");
    }
    status(&format!("Found {} resource(s) to validate.", resource_ids.len()));

    let target_rg_id = cancellable(&cancel, client.get_resource_group_id(sub, tgt_rg))
        .await
        .context("failed to get target resource group ID")?;

    // Start the validate-move LRO (validatemove.go ValidateMove).
    status("Validating resource move (this can take a few minutes)...");
    let begin = cancellable(
        &cancel,
        client.begin_validate_move(sub, src_rg, &resource_ids, &target_rg_id),
    )
    .await
    .context("failed to validate resource move")?;

    let report_ctx = ReportContext {
        source_subscription_id: cfg.args.source_subscription_id.clone(),
        source_resource_group: cfg.args.source_resource_group.clone(),
        target_subscription_id: cfg.args.target_subscription_id.clone(),
        target_resource_group: cfg.args.target_resource_group.clone(),
        resource_count: resource_ids.len(),
    };

    let report = crate::lro::poll_api(&cancel, &client, begin, &cfg.args.output_path, report_ctx)
        .await
        .context("failed to poll API")?;

    if report.success {
        crate::output::output_success(&report.status_text);
    } else {
        crate::output::output_fail_summary(
            report.errors.len(),
            &top_failure_names(&report, CONSOLE_TOP_FAILURES),
        );
    }

    println!(
        "{}",
        colors::yellow(&format!(
            "\n***  Output file written to: - {} ***",
            cfg.args.output_path
        ))
    );
    Ok(())
}

/// Prints a cyan progress/status line (not present in the Go binary).
fn status(msg: &str) {
    println!("{}", colors::cyan(msg));
}

/// Races an async pipeline step against cancellation so Ctrl-C / SIGTERM
/// interrupt promptly even while an Azure call is in flight. On cancellation
/// the in-flight future is dropped and a "context canceled" error is
/// returned, which the caller's `.context(...)` wraps for a Go-like message.
async fn cancellable<F, T>(cancel: &CancellationToken, fut: F) -> anyhow::Result<T>
where
    F: std::future::Future<Output = anyhow::Result<T>>,
{
    tokio::select! {
        biased;
        () = cancel.cancelled() => anyhow::bail!("context canceled"),
        result = fut => result,
    }
}

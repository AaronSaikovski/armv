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
    tracing::debug!("pipeline start: source_sub={sub} source_rg={src_rg} target_rg={tgt_rg} exclude_types={:?}", cfg.args.exclude_resource_types);

    // Every pre-poll Azure call is raced against cancellation so Ctrl-C
    // interrupts even while a credential probe or HTTP request is in flight
    // (the poll loop honours cancellation on its own).
    // If the caller isn't logged into Azure, this first call fails (no usable
    // credential / 401). Show a clean, actionable message rather than the
    // SDK's multi-line credential-chain dump; the noisy cause is preserved for
    // `--debug` via tracing. Go's checkLogin defines this exact string but
    // never reaches it. A genuine cancellation (Ctrl-C) is passed through
    // unchanged, not mislabelled as a login problem.
    status("Authenticating to Azure...");
    if let Err(err) = cancellable(&cancel, client.get_subscription(sub)).await {
        if cancel.is_cancelled() {
            return Err(err.context("login error"));
        }
        tracing::debug!("login check failed: {err:#}");
        anyhow::bail!(
            "not logged into Azure subscription \"{sub}\": please run `az login` and retry"
        );
    }
    tracing::debug!("confirmed access to subscription {sub}");
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
    tracing::debug!("listed {} resource(s) in \"{src_rg}\"", resource_ids.len());

    // Drop any resource types the caller asked to exclude (e.g. types known
    // not to be movable); the excluded IDs are recorded in the report.
    let (resource_ids, excluded_resources) =
        exclude_by_type(resource_ids, &cfg.args.exclude_resource_types);
    if !excluded_resources.is_empty() {
        tracing::debug!(
            "excluded {} resource(s) by type; {} remaining",
            excluded_resources.len(),
            resource_ids.len()
        );
        status(&format!(
            "Excluded {} resource(s) matching --exclude-resource-types.",
            excluded_resources.len()
        ));
    }
    if resource_ids.is_empty() {
        anyhow::bail!(
            "all resources in source resource group \"{src_rg}\" were excluded by --exclude-resource-types"
        );
    }
    status(&format!("Found {} resource(s) to validate.", resource_ids.len()));

    let target_rg_id = cancellable(&cancel, client.get_resource_group_id(sub, tgt_rg))
        .await
        .context("failed to get target resource group ID")?;
    tracing::debug!("resolved target resource group id {target_rg_id}");

    // Start the validate-move LRO (validatemove.go ValidateMove).
    status("Validating resource move (this can take a few minutes)...");
    let begin = cancellable(
        &cancel,
        client.begin_validate_move(sub, src_rg, &resource_ids, &target_rg_id),
    )
    .await
    .context("failed to validate resource move")?;
    tracing::debug!("validate-move long-running operation started");

    let report_ctx = ReportContext {
        source_subscription_id: cfg.args.source_subscription_id.clone(),
        source_resource_group: cfg.args.source_resource_group.clone(),
        target_subscription_id: cfg.args.target_subscription_id.clone(),
        target_resource_group: cfg.args.target_resource_group.clone(),
        resource_count: resource_ids.len(),
        excluded_resources,
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

/// Partitions resource IDs by whether their type is in `excluded`
/// (case-insensitive). The type is the `provider/type` pair parsed from each
/// ID (e.g. `Microsoft.Web/certificates`). Returns `(kept, removed)`.
fn exclude_by_type(ids: Vec<String>, excluded: &[String]) -> (Vec<String>, Vec<String>) {
    if excluded.is_empty() {
        return (ids, Vec::new());
    }
    let mut kept = Vec::new();
    let mut removed = Vec::new();
    for id in ids {
        let (resource_type, _) = crate::report::parse_resource_id(&id);
        // Azure resource types are ASCII, so compare case-insensitively
        // without allocating a lowercased copy per resource.
        if excluded
            .iter()
            .any(|ex| ex.eq_ignore_ascii_case(&resource_type))
        {
            removed.push(id);
        } else {
            kept.push(id);
        }
    }
    (kept, removed)
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

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn cancellable_passes_result_through_when_not_cancelled() {
        let cancel = CancellationToken::new();
        let value = cancellable(&cancel, async { anyhow::Ok(42) }).await.unwrap();
        assert_eq!(value, 42);
    }

    #[tokio::test]
    async fn cancellable_interrupts_a_pending_future() {
        let cancel = CancellationToken::new();
        cancel.cancel();
        // A future that would never complete on its own; cancellation must win.
        let never = async {
            std::future::pending::<()>().await;
            anyhow::Ok(())
        };
        let err = cancellable(&cancel, never).await.unwrap_err();
        assert_eq!(format!("{err:#}"), "context canceled");
    }

    fn ids() -> Vec<String> {
        vec![
            "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Web/certificates/cert1".into(),
            "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/stg1"
                .into(),
            "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Web/sites/app1".into(),
        ]
    }

    #[test]
    fn exclude_by_type_empty_list_keeps_all() {
        let (kept, removed) = exclude_by_type(ids(), &[]);
        assert_eq!(kept.len(), 3);
        assert!(removed.is_empty());
    }

    #[test]
    fn exclude_by_type_removes_matching_types() {
        let (kept, removed) = exclude_by_type(ids(), &["Microsoft.Web/certificates".to_string()]);
        assert_eq!(removed.len(), 1);
        assert!(removed[0].contains("/certificates/"));
        assert_eq!(kept.len(), 2);
        assert!(kept.iter().all(|id| !id.contains("/certificates/")));
    }

    #[test]
    fn exclude_by_type_is_case_insensitive() {
        let (kept, removed) = exclude_by_type(ids(), &["microsoft.web/CERTIFICATES".to_string()]);
        assert_eq!(removed.len(), 1);
        assert_eq!(kept.len(), 2);
    }

    #[test]
    fn exclude_by_type_multiple_types() {
        let excluded = vec![
            "Microsoft.Web/certificates".to_string(),
            "Microsoft.Storage/storageAccounts".to_string(),
        ];
        let (kept, removed) = exclude_by_type(ids(), &excluded);
        assert_eq!(removed.len(), 2);
        assert_eq!(kept.len(), 1);
        assert!(kept[0].contains("/sites/"));
    }
}

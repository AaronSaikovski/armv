// Credential half of internal/pkg/auth/auth.go. Go uses
// azidentity.NewDefaultAzureCredential; the new-generation Rust
// azure_identity does not ship DefaultAzureCredential, so we hand-assemble
// the same chain from the official credentials and expose it as an
// azure_core `TokenCredential` so it plugs directly into the HTTP
// pipeline's BearerTokenAuthorizationPolicy (see azure/client.rs). Each
// credential attempt is bounded by a timeout so an unreachable probe
// (e.g. managed-identity IMDS off-Azure) fails over instead of hanging.

use std::sync::Arc;
use std::time::Duration;

use anyhow::Context as _;
use azure_core::credentials::{AccessToken, Secret, TokenCredential, TokenRequestOptions};
use tokio::sync::Mutex;

/// ARM scope used for every management-plane call.
pub const ARM_SCOPE: &str = "https://management.azure.com/.default";

/// Managed identity is probed against the IMDS endpoint, which is
/// unreachable off-Azure. A short timeout lets the chain fail over to the
/// developer-tools credential quickly instead of hanging (this mirrors the
/// short IMDS probe timeout in Go's DefaultAzureCredential).
const MANAGED_IDENTITY_TIMEOUT: Duration = Duration::from_secs(5);
/// Upper bound for the other credentials (AAD round-trip / az CLI startup).
const CREDENTIAL_TIMEOUT: Duration = Duration::from_secs(30);

type Candidate = (&'static str, Duration, Arc<dyn TokenCredential>);

/// DefaultAzureCredential-equivalent: tries each candidate on first
/// get_token and caches the first that works (Go caches the winner too).
/// Each attempt is bounded by a timeout so no single credential can hang
/// the whole chain.
struct ChainedTokenCredential {
    candidates: Vec<Candidate>,
    cached: Mutex<Option<Arc<dyn TokenCredential>>>,
}

impl std::fmt::Debug for ChainedTokenCredential {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("ChainedTokenCredential")
            .field(
                "candidates",
                &self
                    .candidates
                    .iter()
                    .map(|(n, _, _)| *n)
                    .collect::<Vec<_>>(),
            )
            .finish()
    }
}

#[async_trait::async_trait]
impl TokenCredential for ChainedTokenCredential {
    async fn get_token(
        &self,
        scopes: &[&str],
        options: Option<TokenRequestOptions<'_>>,
    ) -> azure_core::Result<AccessToken> {
        let mut cached = self.cached.lock().await;
        if let Some(cred) = cached.as_ref() {
            return cred.get_token(scopes, options).await;
        }

        let mut last_err: Option<azure_core::Error> = None;
        for (name, timeout, cred) in &self.candidates {
            match tokio::time::timeout(*timeout, cred.get_token(scopes, options.clone())).await {
                Ok(Ok(token)) => {
                    *cached = Some(cred.clone());
                    return Ok(token);
                }
                Ok(Err(e)) => last_err = Some(e),
                Err(_) => {
                    last_err = Some(azure_core::Error::with_message(
                        azure_core::error::ErrorKind::Credential,
                        format!("{name} timed out"),
                    ));
                }
            }
        }
        Err(last_err.unwrap_or_else(|| {
            azure_core::Error::with_message(
                azure_core::error::ErrorKind::Credential,
                "no credential in the chain produced a token",
            )
        }))
    }
}

fn env_nonempty(key: &str) -> Option<String> {
    std::env::var(key).ok().filter(|v| !v.is_empty())
}

/// Builds the DefaultAzureCredential-equivalent chain. Construction is
/// cheap and does not touch the network (Go defers acquisition to first
/// use). Candidates whose constructors fail (e.g. env vars absent) are
/// skipped; an empty chain is an error with context "auth: default
/// credential". Order mirrors Go: environment -> workload identity ->
/// managed identity -> developer tools (az CLI / azd).
pub fn default_azure_credential() -> anyhow::Result<Arc<dyn TokenCredential>> {
    let mut candidates: Vec<Candidate> = Vec::new();

    // 1. Environment: service principal with client secret.
    if let (Some(tenant), Some(client), Some(secret)) = (
        env_nonempty("AZURE_TENANT_ID"),
        env_nonempty("AZURE_CLIENT_ID"),
        env_nonempty("AZURE_CLIENT_SECRET"),
    ) && let Ok(cred) =
        azure_identity::ClientSecretCredential::new(&tenant, client, Secret::new(secret), None)
    {
        candidates.push(("EnvironmentCredential", CREDENTIAL_TIMEOUT, cred));
    }

    // 2. Workload identity (federated token file, e.g. AKS).
    if env_nonempty("AZURE_FEDERATED_TOKEN_FILE").is_some()
        && let Ok(cred) = azure_identity::WorkloadIdentityCredential::new(None)
    {
        candidates.push(("WorkloadIdentityCredential", CREDENTIAL_TIMEOUT, cred));
    }

    // 3. Managed identity (IMDS) - short timeout, unreachable off-Azure.
    if let Ok(cred) = azure_identity::ManagedIdentityCredential::new(None) {
        candidates.push(("ManagedIdentityCredential", MANAGED_IDENTITY_TIMEOUT, cred));
    }

    // 4. Developer tools: az CLI, azd.
    if let Ok(cred) = azure_identity::DeveloperToolsCredential::new(None) {
        candidates.push(("DeveloperToolsCredential", CREDENTIAL_TIMEOUT, cred));
    }

    if candidates.is_empty() {
        return Err(anyhow::anyhow!("no credential could be constructed"))
            .context("auth: default credential");
    }

    Ok(Arc::new(ChainedTokenCredential {
        candidates,
        cached: Mutex::new(None),
    }))
}

/// Fixed-token credential for tests (used with the ARMV_ENDPOINT override;
/// mock servers do not validate tokens).
#[derive(Debug)]
pub struct StaticCredential(pub String);

#[async_trait::async_trait]
impl TokenCredential for StaticCredential {
    async fn get_token(
        &self,
        _scopes: &[&str],
        _options: Option<TokenRequestOptions<'_>>,
    ) -> azure_core::Result<AccessToken> {
        // Expiry far in the future so the pipeline never tries to refresh.
        let expires_on =
            azure_core::time::OffsetDateTime::now_utc() + azure_core::time::Duration::hours(1);
        Ok(AccessToken::new(Secret::new(self.0.clone()), expires_on))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn chain_constructs_without_env() {
        // Developer-tools / managed-identity candidates need no env vars,
        // so the chain must construct on a bare machine.
        assert!(default_azure_credential().is_ok());
    }
}

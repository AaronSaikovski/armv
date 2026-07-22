// ARM client for the 5 operations the Go SDK clients cover
// (internal/pkg/{auth,resources,resourcegroups,validation}). No new-gen
// management crate exposes validateMoveResources, so the endpoints are
// hand-defined - but every call goes through azure_core's HTTP pipeline
// with a BearerTokenAuthorizationPolicy, so token acquisition/caching/refresh
// come from the official crate rather than custom code. All the azure_core
// HTTP glue is confined to the `send_raw` helper.
//
//   GET  /subscriptions/{sub}?api-version=2016-06-01                     (login check)
//   HEAD /subscriptions/{sub}/resourcegroups/{rg}?api-version=2022-09-01 (existence)
//   GET  /subscriptions/{sub}/resourcegroups/{rg}?api-version=2022-09-01 (rg id)
//   GET  /subscriptions/{sub}/resourceGroups/{rg}/resources?...         (+ nextLink)
//   POST /subscriptions/{sub}/resourceGroups/{rg}/validateMoveResources (LRO)

use std::sync::Arc;

use anyhow::Context as _;
use azure_core::credentials::TokenCredential;
use azure_core::http::headers::{HeaderName, CONTENT_TYPE};
use azure_core::http::policies::{BearerTokenAuthorizationPolicy, Policy};
use azure_core::http::{
    ClientOptions, Context as HttpContext, Method, Pipeline, PipelineSendOptions, RetryOptions,
    Request, Url,
};

use crate::auth::ARM_SCOPE;
use crate::azure::models::{MoveInfo, ResourceGroup, ResourceListPage};

pub const DEFAULT_ENDPOINT: &str = "https://management.azure.com";
pub const SUBSCRIPTION_API_VERSION: &str = "2016-06-01";
pub const RESOURCES_API_VERSION: &str = "2022-09-01";

/// Outcome of one poll of the LRO URL: 202 keeps polling; ANY other status
/// is terminal (including 4xx/5xx - the report is built from it, never an
/// error; this preserves the Go 409 -> report -> exit 0 behavior).
#[derive(Debug, Clone)]
pub enum PollStatus {
    InProgress,
    Terminal {
        status_code: u16,
        /// Go http.Response.Status shape: "{code} {reason}".
        status_text: String,
        body: Vec<u8>,
    },
}

/// Result of starting the validate-move LRO.
#[derive(Debug, Clone)]
pub enum BeginMove {
    /// 202 Accepted with a poll URL (Azure-AsyncOperation preferred, else
    /// Location - matching azcore's poller selection order).
    Poller(String),
    /// The initial response was already terminal (2xx other than 202).
    Immediate {
        status_code: u16,
        status_text: String,
        body: Vec<u8>,
    },
}

/// A raw HTTP outcome: status code, header lookups, and body bytes.
struct RawOutcome {
    status: u16,
    location: Option<String>,
    async_operation: Option<String>,
    body: Vec<u8>,
}

/// Go http.StatusText mapping, for byte-parity in the report's status line
/// ("204 No Content", "409 Conflict"). Unknown codes fall back to Go's
/// http.Response.Status shape for an unmapped code.
pub fn status_text(code: u16) -> String {
    let reason = match code {
        200 => "OK",
        202 => "Accepted",
        204 => "No Content",
        400 => "Bad Request",
        401 => "Unauthorized",
        403 => "Forbidden",
        404 => "Not Found",
        409 => "Conflict",
        429 => "Too Many Requests",
        500 => "Internal Server Error",
        502 => "Bad Gateway",
        503 => "Service Unavailable",
        504 => "Gateway Timeout",
        _ => return format!("{code} status code {code}"),
    };
    format!("{code} {reason}")
}

/// Thin ARM client over the azure_core pipeline; endpoint is overridable so
/// integration tests can point at a mock server.
pub struct ArmClient {
    pipeline: Pipeline,
    endpoint: String,
}

impl ArmClient {
    pub fn new(credential: Arc<dyn TokenCredential>) -> Self {
        Self::with_endpoint(credential, DEFAULT_ENDPOINT)
    }

    pub fn with_endpoint(credential: Arc<dyn TokenCredential>, endpoint: &str) -> Self {
        // BearerTokenCredentialPolicy adds "Authorization: Bearer <token>"
        // to every request and caches the token across calls. It only
        // attaches tokens over TLS, so for a plain-http endpoint (only ever
        // a test mock, which ignores auth) we omit it.
        let mut per_try: Vec<Arc<dyn Policy>> = Vec::new();
        if !endpoint.starts_with("http://") {
            per_try.push(Arc::new(BearerTokenAuthorizationPolicy::new(
                credential,
                [ARM_SCOPE],
            )));
        }
        // Disable the pipeline's retry policy: we drive the LRO poll loop
        // ourselves and interpret every status explicitly, so SDK retries
        // conflict with that (e.g. a terminal 500 would otherwise be retried
        // for the default 60s budget instead of being reported).
        let options = ClientOptions {
            retry: RetryOptions::none(),
            ..Default::default()
        };
        let pipeline = Pipeline::new(
            option_env!("CARGO_PKG_NAME"),
            option_env!("CARGO_PKG_VERSION"),
            options,
            Vec::new(),
            per_try,
            None,
        );
        Self {
            pipeline,
            endpoint: endpoint.trim_end_matches('/').to_string(),
        }
    }

    /// The single point of contact with the azure_core HTTP surface. Sends
    /// one request and returns the status, the two LRO headers, and the
    /// body bytes. The pipeline does NOT error on non-2xx status (status
    /// errors are a higher-level concern), which is exactly what the LRO
    /// terminal-status handling needs.
    async fn send_raw(
        &self,
        method: Method,
        url: &str,
        json_body: Option<Vec<u8>>,
    ) -> anyhow::Result<RawOutcome> {
        let parsed: Url = url
            .parse()
            .with_context(|| format!("invalid request URL {url}"))?;
        tracing::debug!("request {method:?} {url}");
        let mut request = Request::new(parsed, method);
        if let Some(bytes) = json_body {
            request.insert_header(CONTENT_TYPE, "application/json");
            request.set_body(bytes);
        }

        // skip_checks: true returns the RawResponse for ANY status (the
        // pipeline would otherwise error on non-2xx). We interpret status
        // ourselves so 409/500 become reports, not errors (Go parity).
        let response = self
            .pipeline
            .send(
                &HttpContext::default(),
                &mut request,
                Some(PipelineSendOptions {
                    skip_checks: true,
                    ..Default::default()
                }),
            )
            .await
            .context("http request failed")?;

        let (status, headers, body) = response.deconstruct();
        let status = u16::from(status);
        tracing::debug!("response {status} ({} bytes)", body.len());
        let location = headers
            .get_optional_str(&HeaderName::from_static("location"))
            .map(str::to_string);
        let async_operation = headers
            .get_optional_str(&HeaderName::from_static("azure-asyncoperation"))
            .map(str::to_string);

        Ok(RawOutcome {
            status,
            location,
            async_operation,
            body: body.to_vec(),
        })
    }

    /// Login check: GET the subscription; any non-2xx is an error.
    pub async fn get_subscription(&self, subscription_id: &str) -> anyhow::Result<()> {
        let url = format!(
            "{}/subscriptions/{}?api-version={}",
            self.endpoint, subscription_id, SUBSCRIPTION_API_VERSION
        );
        let out = self
            .send_raw(Method::Get, &url, None)
            .await
            .with_context(|| format!("auth: subscription \"{subscription_id}\" get"))?;
        if is_success(out.status) {
            Ok(())
        } else {
            Err(http_status_error(&out))
                .with_context(|| format!("auth: subscription \"{subscription_id}\" get"))
        }
    }

    /// HEAD existence check: 2xx -> true, 404 -> false, other -> error.
    pub async fn resource_group_exists(
        &self,
        subscription_id: &str,
        resource_group: &str,
    ) -> anyhow::Result<bool> {
        let url = format!(
            "{}/subscriptions/{}/resourcegroups/{}?api-version={}",
            self.endpoint, subscription_id, resource_group, RESOURCES_API_VERSION
        );
        let out = self
            .send_raw(Method::Head, &url, None)
            .await
            .with_context(|| format!("resourcegroups: check existence of \"{resource_group}\""))?;
        if is_success(out.status) {
            Ok(true)
        } else if out.status == 404 {
            Ok(false)
        } else {
            Err(http_status_error(&out)).with_context(|| {
                format!("resourcegroups: check existence of \"{resource_group}\"")
            })
        }
    }

    /// GET the resource group and return its full resource id.
    pub async fn get_resource_group_id(
        &self,
        subscription_id: &str,
        resource_group: &str,
    ) -> anyhow::Result<String> {
        let url = format!(
            "{}/subscriptions/{}/resourcegroups/{}?api-version={}",
            self.endpoint, subscription_id, resource_group, RESOURCES_API_VERSION
        );
        let ctx = || format!("resourcegroups: get \"{resource_group}\"");
        let out = self.send_raw(Method::Get, &url, None).await.with_context(ctx)?;
        if !is_success(out.status) {
            return Err(http_status_error(&out)).with_context(ctx);
        }
        let rg: ResourceGroup = serde_json::from_slice(&out.body).with_context(ctx)?;
        rg.id
            .ok_or_else(|| anyhow::anyhow!("response missing resource group id"))
            .with_context(ctx)
    }

    /// Lists every resource id in the group, following nextLink pages and
    /// skipping entries without an id.
    pub async fn list_resource_ids(
        &self,
        subscription_id: &str,
        resource_group: &str,
    ) -> anyhow::Result<Vec<String>> {
        let ctx = || format!("resources: list page for \"{resource_group}\"");
        let mut ids = Vec::new();
        let mut url = format!(
            "{}/subscriptions/{}/resourceGroups/{}/resources?api-version={}",
            self.endpoint, subscription_id, resource_group, RESOURCES_API_VERSION
        );
        loop {
            let out = self.send_raw(Method::Get, &url, None).await.with_context(ctx)?;
            if !is_success(out.status) {
                return Err(http_status_error(&out)).with_context(ctx);
            }
            let page: ResourceListPage = serde_json::from_slice(&out.body).with_context(ctx)?;
            ids.extend(page.value.into_iter().filter_map(|r| r.id));
            match page.next_link {
                Some(next) if !next.is_empty() => url = next,
                _ => break,
            }
        }
        Ok(ids)
    }

    /// Starts the validate-move LRO.
    pub async fn begin_validate_move(
        &self,
        subscription_id: &str,
        source_resource_group: &str,
        resource_ids: &[String],
        target_resource_group_id: &str,
    ) -> anyhow::Result<BeginMove> {
        let url = format!(
            "{}/subscriptions/{}/resourceGroups/{}/validateMoveResources?api-version={}",
            self.endpoint, subscription_id, source_resource_group, RESOURCES_API_VERSION
        );
        let body = serde_json::to_vec(&MoveInfo {
            resources: resource_ids,
            target_resource_group: target_resource_group_id,
        })
        .context("validation: begin validate move")?;

        let out = self
            .send_raw(Method::Post, &url, Some(body))
            .await
            .context("validation: begin validate move")?;

        if out.status == 202 {
            match out.async_operation.or(out.location) {
                Some(u) => Ok(BeginMove::Poller(u)),
                None => Err(anyhow::anyhow!("202 response without a polling URL"))
                    .context("validation: begin validate move"),
            }
        } else if is_success(out.status) {
            // Synchronously terminal (not seen in practice for this API).
            Ok(BeginMove::Immediate {
                status_code: out.status,
                status_text: status_text(out.status),
                body: out.body,
            })
        } else {
            Err(http_status_error(&out)).context("validation: begin validate move")
        }
    }

    /// One GET of the LRO poll URL.
    pub async fn poll_once(&self, poll_url: &str) -> anyhow::Result<PollStatus> {
        let out = self.send_raw(Method::Get, poll_url, None).await?;
        if out.status == 202 {
            return Ok(PollStatus::InProgress);
        }
        Ok(PollStatus::Terminal {
            status_code: out.status,
            status_text: status_text(out.status),
            body: out.body,
        })
    }
}

fn is_success(status: u16) -> bool {
    (200..300).contains(&status)
}

fn http_status_error(out: &RawOutcome) -> anyhow::Error {
    let body = String::from_utf8_lossy(&out.body);
    anyhow::anyhow!("unexpected status {}: {}", status_text(out.status), body.trim())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn status_text_shapes() {
        assert_eq!(status_text(204), "204 No Content");
        assert_eq!(status_text(409), "409 Conflict");
        assert_eq!(status_text(500), "500 Internal Server Error");
        assert_eq!(status_text(599), "599 status code 599");
    }
}

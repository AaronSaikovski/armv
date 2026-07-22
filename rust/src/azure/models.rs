// Serde models for the hand-written ARM REST calls (replacing the Go
// armresources/armsubscription SDK types). Unknown fields must be ignored.

use serde::{Deserialize, Serialize};

/// Request body for validateMoveResources (armresources.MoveInfo):
/// {"resources": ["<id>", ...], "targetResourceGroup": "<rg resource id>"}
#[derive(Debug, Serialize)]
pub struct MoveInfo<'a> {
    pub resources: &'a [String],
    #[serde(rename = "targetResourceGroup")]
    pub target_resource_group: &'a str,
}

/// One page of GET .../resourceGroups/{rg}/resources.
#[derive(Debug, Deserialize)]
pub struct ResourceListPage {
    #[serde(default)]
    pub value: Vec<ResourceEntry>,
    #[serde(rename = "nextLink")]
    pub next_link: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ResourceEntry {
    /// Entries without an id are skipped (parity with the Go nil-ID filter).
    pub id: Option<String>,
}

/// GET .../resourcegroups/{name} - only the id is used.
#[derive(Debug, Deserialize)]
pub struct ResourceGroup {
    pub id: Option<String>,
}

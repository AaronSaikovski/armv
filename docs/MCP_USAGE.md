# Using the ARMV MCP server

A practical guide to driving ARMV through an LLM — what to ask, how to phrase it, and how to get useful answers out of the other side.

**Prereqs:** the server is registered with your MCP client. If not, see [MCP_INSTALL.md](MCP_INSTALL.md).

---

## The four tools, in one line each

| Tool | What it does | When the LLM will pick it |
|---|---|---|
| `list_subscriptions` | Returns every subscription the credential can see | You mention "my subs" / "subscriptions I have access to" |
| `list_resource_groups` | Returns every RG in a given subscription | You pick a subscription and want to see what's in it |
| `list_resources` | Returns every resource in a given RG | You want to know what's *in* an RG before validating |
| `validate_move` | The real payload — asks Azure whether everything in `source-rg` could be moved to `target-rg` | You ask whether a move is possible |

The LLM chains these together based on your intent — you don't invoke them manually. Your job is to describe what you want in English; the client's consent prompt shows you exactly what will be sent before it runs.

---

## The usual workflow

You typically don't start with all four UUIDs in hand. The flow looks like this:

```
1. "What subs do I have?"               → list_subscriptions
2. "What's in the prod one?"            → list_resource_groups(prod)
3. "Show me rg-app's contents."         → list_resources(prod, rg-app)
4. "Can I move rg-app to dev?"          → validate_move(...)
5. "Why did it fail?"                   → LLM reasons over the returned diagnostics
```

Each step narrows the scope. The LLM holds the IDs from earlier steps in its conversation context, so you don't have to paste them again.

---

## Sample questions

Copy-paste ready. Written for how people actually talk to an LLM — not stiff, not over-specified.

### Discovery

- *"What Azure subscriptions do I have access to through ARMV?"*
- *"List the subscriptions you can see, grouped by whether they're Enabled or Disabled."*
- *"Show me the resource groups in subscription `prod-east`."*
- *"Which resource groups in my sandbox sub contain storage accounts?"* — LLM will call `list_resource_groups` then `list_resources` on each, then filter.
- *"What's in `rg-app` in the prod subscription? I want types and counts, not a full dump."*
- *"Compare `rg-prod-east` and `rg-prod-west` — what's in each?"*

### Happy-path validation

- *"Can I move everything in `rg-app` (sub `12345678-1234-1234-1234-123456789012`) to `rg-app-dev` in sub `87654321-4321-4321-4321-210987654321`? Just tell me yes or no and the count."*
- *"Validate whether `rg-legacy` can move to `rg-archive` in the same subscription."*
- *"Run the move-validation for `rg-app` into `rg-app-v2`, and if it passes, summarise what would be moved."*

### Post-failure analysis

After a `validate_move` returns `success: false` with a 409 payload:

- *"Which resources are blocking the move? Group them by resource type."*
- *"What's the most common error code in that response?"*
- *"For each failing resource, give me a one-line recommendation on how to unblock it."*
- *"Of the 24 resources that can move, which one is highest-risk? Highlight anything stateful."*
- *"Write me a migration runbook based on that failure — which resources must be replaced vs. can be moved as-is."*

### Planning / scoping

- *"I need to consolidate three RGs (`rg-app1`, `rg-app2`, `rg-app3`) into `rg-app-merged`. Run validate_move on each and tell me which would need rework first."* — LLM will make three sequential calls.
- *"Check each RG in subscription `dev` and report which ones are fully movable to sub `prod`."* — this is expensive (one call per RG). The LLM should ask for confirmation before starting.
- *"Before I run validation, just list the blockers: any resources in `rg-app` whose type is known to reject moves?"* — LLM answers from its own Azure knowledge without making an MCP call.

### Using specific credentials

ARMV picks up `az login` / env var credentials automatically. To be explicit:

- *"Using the service principal with tenant `<tenant-uuid>`, client `<client-uuid>`, secret `<secret>`, validate moving `rg-a` to `rg-b`."* — don't do this in a persisted chat; secrets live in logs.
- *"I've set `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET` in the client config. Validate the move from `rg-a` to `rg-b`."* — LLM omits the auth fields; `DefaultAzureCredential` picks them up from the env.
- *"I'll paste a bearer token each call. First one: `eyJ…`. Validate `rg-a` → `rg-b`."* — the server uses it and stores nothing. When it expires (~1 hour), you'll get a 401 and the LLM will ask you for a fresh one.

---

## Tips for better results

### Be concrete about which subscription

If you have multiple subs, the LLM might guess or ask. Spell it out:

> ❌ *"Check if my app can move to the dev environment."*  — LLM has to ask follow-ups.
>
> ✅ *"Check if everything in `rg-app` in sub `dev-east` (`9876…`) can move to `rg-app` in sub `dev-west` (`5432…`)."*

### Ask the LLM to summarise, not paginate

The raw `list_resources` output on a large RG is noisy. Ask for a digest:

> *"List resources in `rg-data` grouped by type, with counts per type."*

### Use the consent prompt as a dry run

When the MCP client shows the consent prompt with the arguments about to be sent, *read it*. This is your chance to catch a subscription-ID typo before it hits Azure. If something looks off, deny, correct the chat message, and re-ask.

### Chain validations for bulk checks

Instead of running five separate chat sessions:

> *"For RGs `rg-a`, `rg-b`, `rg-c` in sub `prod`, validate moving each to their equivalent in sub `dev` (`rg-a-dev`, etc.). Give me a single table of results."*

The LLM will run three sequential `validate_move` calls and aggregate.

### Know when to step up the model

The default (Claude Haiku 4.5 / GPT-4o-mini) is fine for tool selection and extraction. Step up to Claude Sonnet 4.6 when:

- The 409 payload has **many** failing resources and you want them grouped intelligently
- You want remediation suggestions ("replace this Container Instance with…")
- You're planning a **multi-RG** migration and need the model to reason about ordering

---

## What the LLM *can't* do

A few things that look like they should work but don't:

- **Mid-call status questions.** If `validate_move` is running, the client shows progress notifications (`Polling Azure validate-move (elapsed 8s)`), but the LLM can't answer chat messages until the call returns. Your "how's it going?" is queued.
- **Retrieve previous run results.** Each call is stateless server-side; there's no `get_validation_result(job_id)` endpoint. The LLM remembers the last result via its conversation context — a new chat starts fresh.
- **Perform the actual move.** ARMV is read-only by design. Use `az resource move` or ARM deployment for the actual operation, armed with the validation result.
- **Validate across tenants.** The API is limited to same-tenant moves. The server surfaces this as a hard error.
- **Validate partial RG contents.** It's all-or-nothing per source RG. If you need to move a subset, move them into a staging RG first.

---

## Gotchas

| Symptom | Likely cause | Fix |
|---|---|---|
| `invalid source subscription ID format` | Typo in a UUID | Re-check; paste from `list_subscriptions` output |
| `no resources found in source resource group` | RG is empty or wrong subscription | Run `list_resources` on it first |
| Silent failure, no result | MCP client cancelled mid-call (often the 60s client timeout) | Ask again; large RGs may take 2–5 min. Check the client's per-call timeout. |
| `401 unauthorized` with `bearer_token` | Token expired (~1h TTL) | Fetch a fresh one: `az account get-access-token --resource https://management.azure.com` |
| LLM invents a subscription ID | Model guessed instead of asking | Correct it in chat, re-send; consider stepping up to Sonnet if it's frequent |
| Same resource appears as blocker every time | That resource type literally can't move (e.g. Container Instances, Log Analytics workspaces, some managed services) | Plan to recreate it in the target; the rest of the RG is likely fine |

---

## Example full session

A realistic transcript. The LLM is driving; the human narrates intent.

> **User:** *"I want to consolidate my app stack. What subscriptions do I have?"*
>
> **Claude:** *calls `list_subscriptions`* → *"You have three Enabled subscriptions: **prod-east** (`1111…`), **prod-west** (`2222…`), **sandbox** (`3333…`)."*
>
> **User:** *"Show me what's in `rg-app` in prod-east."*
>
> **Claude:** *calls `list_resources(prod-east, rg-app)`* → *"27 resources: 1 App Service Plan, 3 Web Apps, 1 Storage Account, 2 Key Vaults, 1 Container Instance, 19 Application Insights components. Anything specific you want to drill into?"*
>
> **User:** *"Can that whole thing move to `rg-app-consolidated` in prod-west?"*
>
> **Claude:** *consent prompt; user approves; calls `validate_move(...)`; shows progress notifications; returns result*
>
> → *"**Failed** (409 Conflict). 26 of 27 resources can move; the blocker is the Container Instance `aciresource` — `Microsoft.ContainerInstance/containerGroups` doesn't support resource moves at all. **Recommendation:** delete it from the source, deploy a new one in the target RG, then move the remaining 26. The storage account and key vaults can move without rebuild."*
>
> **User:** *"Great. Pretend I've done that. Re-validate without the Container Instance."*
>
> **Claude:** *"ARMV validates an existing RG — it can't skip resources. Once you delete `aciresource` from `rg-app`, re-run the same validation and it should pass."*
>
> **User:** *"OK, assume that's done. Now do the same validation for `rg-data` → `rg-data-consolidated` in the same sub pair."*
>
> **Claude:** *calls `validate_move` with the new RGs* → *"**Success** (204 No Content). All 14 resources in `rg-data` are eligible for move to `rg-data-consolidated`."*

---

## Related

- [MCP_INSTALL.md](MCP_INSTALL.md) — getting the server connected to Claude Code
- [README → MCP Server Mode](../README.md#mcp-server-mode) — protocol details, tool schemas, credential chain
- [Model Context Protocol](https://modelcontextprotocol.io/) — spec and client catalog

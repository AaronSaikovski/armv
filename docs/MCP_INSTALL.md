# Installing the ARMV MCP server in Claude Code

Step-by-step guide for wiring `armv mcp serve` into **Claude Code** on **Windows** and **macOS**.

This document assumes you already have a working `armv` binary. If you don't, see the [main README → Installation](../README.md#installation) — grab a prebuilt binary from [Releases](https://github.com/AaronSaikovski/armv/releases) or build from source with `task release`.

---

## TL;DR

```bash
# 1. verify the binary works
armv --version
armv mcp serve   # (Ctrl-C to exit — nothing will happen without an MCP client attached)

# 2. register it with Claude Code using the absolute path
claude mcp add armv -- /absolute/path/to/armv mcp serve

# 3. authenticate against Azure once
az login

# 4. restart Claude Code, verify it's loaded
claude mcp list
```

That's it. Skip to [Verifying the install](#verifying-the-install) if `claude mcp list` already shows `armv` as connected.

---

## Prerequisites

| Tool | Purpose | macOS | Windows |
|---|---|---|---|
| `armv` binary | The MCP server itself | [Releases](https://github.com/AaronSaikovski/armv/releases) (`*_Darwin_*.tar.gz`) or `task release` | [Releases](https://github.com/AaronSaikovski/armv/releases) (`*_Windows_*.zip`) or `task release` |
| Claude Code CLI | MCP host | `brew install anthropic-ai/tap/claude` | Download from [claude.com/download](https://claude.com/download) |
| Azure CLI (`az`) | Default credential source | `brew install azure-cli` | `winget install Microsoft.AzureCLI` |

Verify each is on your PATH:

**macOS / Linux:**
```bash
armv --version
claude --version
az --version
```

**Windows (PowerShell):**
```powershell
armv --version
claude --version
az --version
```

If `armv` isn't resolved, the registration step below works with an absolute path — just substitute the full location instead of relying on PATH.

---

## Installing the binary

### macOS

**Option A — prebuilt release (fastest):**

```bash
# Pick the archive matching your arch:
#   armv_<VERSION>_Darwin_x86_64.tar.gz   (Intel)
#   armv_<VERSION>_Darwin_arm64.tar.gz    (Apple Silicon)

curl -L -o armv.tar.gz \
  https://github.com/AaronSaikovski/armv/releases/latest/download/armv_<VERSION>_Darwin_arm64.tar.gz

tar -xzf armv.tar.gz
sudo mv armv /usr/local/bin/armv
chmod +x /usr/local/bin/armv
armv --version
```

**Option B — build from source:**

```bash
git clone https://github.com/AaronSaikovski/armv.git
cd armv
task release                       # produces bin/armv
sudo cp bin/armv /usr/local/bin/armv
```

### Windows

**Option A — prebuilt release (fastest):**

1. Download `armv_<VERSION>_Windows_x86_64.zip` from [Releases](https://github.com/AaronSaikovski/armv/releases).
2. Extract `armv.exe` somewhere permanent — recommended: `C:\Tools\armv\armv.exe`.
3. (Optional) Add `C:\Tools\armv` to PATH:
   - Start → "Edit the system environment variables" → **Environment Variables…** → under **User variables**, edit `Path`, add `C:\Tools\armv`.
4. Open a new PowerShell window and run `armv --version` to confirm.

**Option B — build from source:**

```powershell
git clone https://github.com/AaronSaikovski/armv.git
cd armv
task release                       # produces bin\armv.exe
Copy-Item bin\armv.exe C:\Tools\armv\armv.exe
```

---

## Registering with Claude Code

Claude Code talks to MCP servers defined in your user-scoped MCP config. The easiest way to add one is the built-in `claude mcp add` command — it edits the right file for you and handles escaping.

### macOS / Linux

```bash
claude mcp add armv -- /usr/local/bin/armv mcp serve
```

The `--` separates Claude Code's own flags from the command it should launch. Everything after `--` is the subprocess command-line.

Confirm it registered:

```bash
claude mcp list
```

Expected output (approximately):

```
NAME   COMMAND                       STATUS
armv   /usr/local/bin/armv mcp …     ✓ connected
```

### Windows (PowerShell)

```powershell
claude mcp add armv -- "C:\Tools\armv\armv.exe" mcp serve
```

The quotes around the path are required if it contains spaces (e.g. `"C:\Program Files\…"`).

Confirm:

```powershell
claude mcp list
```

### Manual config (either OS)

If you prefer editing JSON directly, the config file is:

- **macOS**: `~/.claude.json`
- **Windows**: `%USERPROFILE%\.claude.json` (typically `C:\Users\<you>\.claude.json`)

Add or edit the `mcpServers` key at the top level:

**macOS:**
```json
{
  "mcpServers": {
    "armv": {
      "command": "/usr/local/bin/armv",
      "args": ["mcp", "serve"]
    }
  }
}
```

**Windows** (note the escaped backslashes):
```json
{
  "mcpServers": {
    "armv": {
      "command": "C:\\Tools\\armv\\armv.exe",
      "args": ["mcp", "serve"]
    }
  }
}
```

Restart Claude Code (exit and relaunch) to pick up the change.

---

## Authenticating against Azure

The MCP server uses Azure's `DefaultAzureCredential` chain by default, so whatever credential Azure CLI is using is what the MCP server sees. Three sensible setups:

### 1. `az login` — recommended for personal machines

```bash
az login
az account set --subscription "<your-primary-sub>"
```

No config changes. The server picks up your cached token automatically; no secrets live in any config file.

### 2. Service principal via environment variables

Extend the MCP registration so the server process inherits the SP env vars. On either OS you want the `env` block in your `~/.claude.json`:

```json
{
  "mcpServers": {
    "armv": {
      "command": "/usr/local/bin/armv",
      "args": ["mcp", "serve"],
      "env": {
        "AZURE_TENANT_ID": "00000000-0000-0000-0000-000000000000",
        "AZURE_CLIENT_ID": "00000000-0000-0000-0000-000000000000",
        "AZURE_CLIENT_SECRET": "<secret>"
      }
    }
  }
}
```

> Any values here end up in plain text in your home directory. Use `az login` instead on personal machines; reserve SP env vars for locked-down workstations or CI-like scenarios.

For a cert-based SP, swap `AZURE_CLIENT_SECRET` for `AZURE_CLIENT_CERTIFICATE_PATH` pointing at a PEM file.

### 3. Per-call bearer token — zero credentials on disk

Fetch a short-lived token at invocation time (see the main README → [MCP Server Mode → Client-Supplied Bearer Token](../README.md#client-supplied-bearer-token)). Claude Code doesn't have a built-in token refresher; this flow is most useful for scripted callers or agents that already hold a token.

---

## Verifying the install

1. `claude mcp list` — `armv` should show `✓ connected`. If `✗`, jump to [Troubleshooting](#troubleshooting).
2. In Claude Code, ask:

   > *"What tools does the armv MCP server expose?"*

   Claude should enumerate `validate_move`, `list_subscriptions`, `list_resource_groups`, `list_resources`.

3. Try a discovery call:

   > *"Use armv to list my Azure subscriptions."*

   Claude Code will show a consent prompt with the arguments; approve it. The result should be a list of subscriptions your `az` credential can see.

4. Try a validation (pick a safe throwaway RG):

   > *"Can I move everything from resource group `rg-dev-test` in subscription `<your-sub>` to `rg-dev-staging`?"*

   Claude Code fills in the four required arguments from your message and asks for consent. After approval, you'll see progress notifications while ARMV polls the Azure API, then a final structured answer.

---

## Troubleshooting

### `armv` shows as disconnected in `claude mcp list`

Most common causes:

| Symptom | Check |
|---|---|
| `command not found` in the MCP log | Use an **absolute path** in the registration, not just `armv`. On Windows include the `.exe`. |
| Server exits immediately | Run `armv mcp serve` manually in a terminal. If it crashes at startup, it'll print the error to stderr. |
| `DefaultAzureCredential: no credential found` | Run `az login`; verify `az account show` succeeds. |
| Binary built for wrong arch | On Apple Silicon make sure you downloaded the `arm64` archive, not `x86_64`. |

View Claude Code's MCP log:

**macOS:**
```bash
tail -f ~/Library/Logs/Claude/mcp-server-armv.log
```

**Windows (PowerShell):**
```powershell
Get-Content -Wait "$env:APPDATA\Claude\logs\mcp-server-armv.log"
```

### `AADSTS70011` / `invalid scope` errors

You're hitting Azure AD token-scope validation. Easiest fix: let `DefaultAzureCredential` handle it — remove any `bearer_token` field from the tool call and rely on `az login` or SP env vars.

### "Permission denied" opening `/usr/local/bin/armv` on macOS

First run of an unsigned download. Either:

- `xattr -d com.apple.quarantine /usr/local/bin/armv` to drop Gatekeeper's quarantine flag, or
- build from source — `task release` products are never quarantined.

### `armv mcp serve` blocks indefinitely with no output

That's correct behaviour. MCP stdio servers are silent until a client connects and sends JSON-RPC. Stop it with `Ctrl-C`.

### Windows-specific: PowerShell escaping in `claude mcp add`

If a path contains spaces, double-quote it. Both of these are equivalent:

```powershell
claude mcp add armv -- "C:\Program Files\Armv\armv.exe" mcp serve

claude mcp add armv -- 'C:\Program Files\Armv\armv.exe' mcp serve
```

Avoid backslash-escaped quotes — PowerShell doesn't parse them the way `bash` does.

### `claude` isn't on PATH on Windows

The installer places it in `%LOCALAPPDATA%\AnthropicClaude\app-x.y.z\`. Either:

- Use the Start-menu shortcut to launch Claude Code, and run `mcp` commands from inside it via `/mcp` slash-commands, or
- Add the install directory to `Path` so `claude mcp add` works from any terminal.

---

## Uninstalling

**Remove the registration:**

```bash
claude mcp remove armv
```

**macOS — delete the binary:**
```bash
sudo rm /usr/local/bin/armv
```

**Windows — delete the binary:**
```powershell
Remove-Item C:\Tools\armv\armv.exe
# and remove C:\Tools\armv from PATH if you added it
```

---

## Related

- [Main README](../README.md) — full CLI and MCP reference
- [README → MCP Server Mode](../README.md#mcp-server-mode) — exposed tools, credential selection, progress notifications
- [Model Context Protocol](https://modelcontextprotocol.io/) — protocol spec and client catalog
- [Claude Code docs](https://docs.claude.com/en/docs/claude-code) — full MCP configuration reference

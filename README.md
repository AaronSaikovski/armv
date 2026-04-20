<div align="center">

# ARMV - <u>A</u>zure <u>R</u>esource <u>M</u>oveability <u>V</u>alidator

**v1.2.2**

[![Build Status](https://github.com/AaronSaikovski/armv/workflows/build/badge.svg)](https://github.com/AaronSaikovski/armv/actions)
[![Licence](https://img.shields.io/github/license/AaronSaikovski/armv)](LICENSE)

A lightweight, high-performance Go utility for validating Azure resource moveability without performing the actual move operation. Ships as both a **CLI** and a **Model Context Protocol (MCP) server** so it can be driven by a human at the terminal or by LLM agents like Claude Desktop and Claude Code.

</div>

> **⚠️ NOTE: THIS TOOL PERFORMS READ-ONLY VALIDATION AND WILL NOT PERFORM ANY ACTUAL MOVE OPERATIONS. IT GENERATES A DETAILED VALIDATION REPORT FOR ANALYSIS.**

## Overview

ARMV is a production-ready Go utility that validates whether Azure resources can be moved between resource groups and subscriptions. It acts as a comprehensive wrapper around Azure's [Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources?view=rest-resources-2021-04-01), providing intelligent polling, detailed error reporting, and secure file operations.

Two modes share the same core engine:

- **CLI mode** (`armv …`) — interactive terminal use with a progress bar and timestamped output files.
- **MCP server mode** (`armv mcp serve`) — exposes a `validate_move` tool over stdio so MCP clients (Claude Desktop, Claude Code, VS Code MCP extensions, MCP Inspector) can invoke it. Supports per-call service principal credentials, pre-fetched bearer tokens, or ambient `az login`. See [MCP Server Mode](#mcp-server-mode) below.

This tool is a complete rewrite of the deprecated [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) Python utility, now delivered as a fully standalone, self-contained Go binary with zero external dependencies.

### Key Features

- **Non-Destructive Validation**: Tests moveability without making any changes
- **Cross-Subscription Support**: Validate moves across different subscriptions (same tenant only)
- **Dual Mode**: CLI for terminal use + MCP server (`armv mcp serve`) for LLM-driven automation via Claude Desktop, Claude Code, and other MCP clients
- **Flexible Authentication**: `az login`, service principal (secret or certificate), client-supplied bearer tokens, and `DefaultAzureCredential` chain (managed identity, workload identity, env vars)
- **Intelligent Polling**: Long-running operation polling with timeout protection (max 30 minutes)
- **Comprehensive Error Reporting**: Detailed JSON diagnostics including tracking IDs and timestamps
- **Progress Visualization**: Real-time progress bar during validation operations (CLI mode)
- **Secure File Operations**: Timestamped output files with secure permissions (CLI mode)
- **Debug Mode**: Optional verbose logging and timing information
- **Production-Ready**: Fully tested, optimized memory usage, and comprehensive error handling
- **Multi-Platform Support**: Cross-platform binaries for Linux, Windows, and macOS
- **Security Focused**: Vulnerability scanning (govulncheck), static analysis (staticcheck), and comprehensive testing
- **Continuous Integration**: Automated builds, tests, and security checks on every commit

### How It Works

1. **Input Validation**: Validates subscription IDs and resource group references
2. **Authentication**: Resolves a credential via `az login`, service principal (secret/cert), a client-supplied bearer token, or the `DefaultAzureCredential` chain
3. **Resource Discovery**: Enumerates all resources in the source resource group
4. **Validation**: Calls Azure's Validate Move Resources API
5. **Polling**: Polls the long-running operation (with a progress bar in CLI mode; silently in MCP mode)
6. **Output**: CLI mode writes timestamped files; MCP mode returns a structured JSON result to the client

### Response Codes

- **HTTP 204 (Success)**: All resources are eligible for move
- **HTTP 409 (Conflict)**: One or more resources cannot be moved - detailed error report is generated

### Example Error Report

When validation fails, a detailed JSON report is generated:

```json
{
  "error": {
    "code": "ResourceMoveValidationFailed",
    "message": "The resource batch move request has '1' validation errors. Diagnostic information: timestamp '20240520T034539Z', tracking Id '8f53448f-e108-4f51-85d4-259e2137761d', request correlation Id '0a88b427-06ea-4045-98f1-7d2c4aaf2867'.",
    "details": [
      {
        "code": "ResourceMoveNotSupported",
        "target": "/subscriptions/<subID>/resourceGroups/src-rsg/providers/Microsoft.ContainerInstance/containerGroups/aciresource",
        "message": "Resource move is not supported for resource types 'Microsoft.ContainerInstance/containerGroups'."
      }
    ]
  }
}
```

## Architecture

ARMV follows a clean, layered architecture for maintainability and extensibility:

### Core Layers

| Layer | Location | Responsibility |
|-------|----------|-----------------|
| **CLI Layer** | `cmd/armv/app/` | Command-line interface, flag parsing, orchestration, `mcp serve` subcommand |
| **MCP Server** | `internal/pkg/mcpserver/` | Model Context Protocol server over stdio, tool registration, credential selection |
| **Validator Core** | `internal/pkg/validator/` | Library-friendly end-to-end validation flow (no stdout/file/progress-bar concerns) — shared by CLI and MCP server |
| **Authentication** | `internal/pkg/auth/` | Azure credential management: `DefaultAzureCredential`, `ClientSecretCredential`, `StaticTokenCredential` (bearer token) |
| **Validation** | `internal/pkg/validation/` | Resource move validation API wrapper |
| **Resource Management** | `internal/pkg/resourcegroups/`, `internal/pkg/resources/` | Azure resource group and resource operations |
| **Polling** | `cmd/armv/poller/` | Long-running operation polling — interactive (`PollApi`) for CLI and quiet (`PollAndCollect`) for MCP |
| **Utilities** | `pkg/utils/` | Input validation, error handling, file I/O, JSON processing |

### Project Structure

```
armv/
├── cmd/armv/                          # Application entry points
│   ├── main.go                        # Entry point with version embedding
│   ├── app/                           # Application logic
│   │   ├── root.go                   # CLI validation workflow (run)
│   │   ├── command.go                # Cobra root command + flag binding
│   │   ├── mcp.go                    # `armv mcp serve` subcommand
│   │   ├── login.go                  # CLI auth verification (with user output)
│   │   └── resourcegroup.go          # CLI resource group helpers
│   └── poller/                        # Long-running operation handling
│       ├── pollapi.go                # CLI polling with progress bar + file write
│       ├── pollquiet.go              # MCP polling (no stdout, returns data)
│       ├── pollresponse.go           # Response handling and output
│       ├── pollerresponsedata.go     # Response data structures
│       ├── progressbar.go            # Progress bar visualization
│       └── constants.go              # Configuration constants
│
├── internal/pkg/                      # Private internal packages
│   ├── auth/                          # Azure authentication
│   │   ├── auth.go                   # DefaultAzureCredential, ClientSecretCredential, SDK clients
│   │   └── bearer.go                 # StaticTokenCredential for pre-fetched bearer tokens
│   ├── mcpserver/                     # Model Context Protocol server
│   │   ├── server.go                 # validate_move tool, credential selection, stdio transport
│   │   └── server_test.go            # Credential-selection and input-validation tests
│   ├── validator/                     # Shared validation core (CLI + MCP)
│   │   └── validator.go              # Input, Result, Validate() — pure library function
│   ├── validation/                    # Resource move validation
│   │   ├── azureresourcemoveinfo.go  # Validation parameters struct
│   │   └── validatemove.go           # Validation API wrapper
│   ├── resourcegroups/                # Resource group operations
│   │   └── resourcegroups.go         # RG client and operations
│   └── resources/                     # Resource operations
│       └── resources.go              # Resource enumeration and filtering
│
├── pkg/utils/                         # Public utility functions
│   ├── args.go                       # CLI argument structures
│   ├── validateinput.go              # UUID/subscription ID validation
│   ├── errorhandler.go               # Error handling utilities
│   ├── output.go                     # Console output formatting
│   ├── outputfile.go                 # File I/O with secure permissions
│   └── jsonutils.go                  # JSON marshaling/unmarshaling
│
├── test/                              # Comprehensive test suite
│   ├── args_test.go
│   ├── validateinput_test.go
│   ├── azureresourcemoveinfo_test.go
│   ├── command_test.go
│   ├── pollerresponsedata_test.go
│   ├── jsonutils_test.go
│   └── outputfile_test.go
│
├── go.mod                             # Go module definition
├── go.sum                             # Dependency checksums
├── Taskfile.yml                       # Build task automation
├── README.md                          # This file
├── LICENSE                            # MIT License
├── CHANGELOG.md                       # Version history
└── TODO.md                            # Future enhancements
```

## Dependencies

The project uses the following key Go packages:

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/Azure/azure-sdk-for-go/sdk/azcore` | v1.20.0 | Azure SDK core functionality |
| `github.com/Azure/azure-sdk-for-go/sdk/azidentity` | v1.13.1 | Azure credential authentication |
| `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources` | v1.2.0 | Azure resources API client |
| `github.com/modelcontextprotocol/go-sdk` | v1.5.0 | Official Go SDK for Model Context Protocol server |
| `github.com/spf13/cobra` | v1.10.2 | CLI command framework |
| `github.com/schollz/progressbar/v3` | v3.18.0 | Progress bar visualization |

## Requirements

- **Go**: [v1.25.3](https://www.go.dev/dl/) or later for building from source
- **Azure CLI**: [v2.50](https://learn.microsoft.com/en-us/cli/azure/) or later
- **Taskfile**: [v3.0+](https://taskfile.dev/) for build automation (optional but recommended)
- **Azure Credentials**: Valid Azure CLI login context (`az login`)

## CI/CD & Quality Assurance

The project uses comprehensive GitHub Actions workflows for:

- **Multi-Platform Testing**: Automated builds and tests on Ubuntu, Windows, and macOS
- **Multi-Architecture Support**: Cross-platform binary releases for amd64, arm64, armv7, and 386
- **Code Quality**: Automated vulnerability scanning, static analysis, and linting
- **Release Automation**: GoReleaser for building and publishing multi-platform binaries

### Quality Checks

All commits run through:
- ✅ **govulncheck**: Vulnerability scanning
- ✅ **staticcheck**: Advanced static code analysis
- ✅ **go vet**: Go's built-in static analysis
- ✅ **go test**: Comprehensive unit tests (38+ test cases, 100% pass rate)

## Installation & Setup

### Prerequisites: Azure Authentication

Before using ARMV, you must authenticate with Azure:

```bash
# Log in to Azure
az login

# Set the active subscription (replace XXXX-XXXX-XXXX-XXXX with your subscription ID)
az account set --subscription "XXXX-XXXX-XXXX-XXXX"
```

ARMV uses the DefaultAzureCredential chain from the Azure SDK, which will use your authenticated Azure CLI context.

### Building from Source

This project uses **Taskfile** for build automation. All build commands are defined in `Taskfile.yml`.

#### Available Tasks

```bash
build            # Build debug version for your platform
clean            # Remove builds and debug artifacts
debug            # Run debug version with environment variables
deploy           # Deploy test Azure resources using Bicep
destroy          # Destroy test Azure resources
deps             # Fetch and update dependencies
generate         # Update binary version using go:generate
goreleaser       # Build cross-platform release binaries
lint             # Format, lint, and tidy code
release          # Build optimized release binary (outputs to /bin)
run              # Build and run the application
seccheck         # Security vulnerability scanner
staticcheck      # Static code analysis
test             # Run unit tests
version          # Display Go version
vet              # Examine code for suspicious constructs
watch            # Enable hot reload with air
```

#### Quick Start

```bash
# 1. Fetch dependencies
task deps

# 2. Build debug version
task build

# 3. Build release version (outputs to ./bin)
task release
```

#### Testing

```bash
# Run all unit tests
task test

# Run tests with verbose output
go test ./... -v

# Run tests with coverage report
go test ./... -cover
```

## Usage

### Basic Command

```bash
./armv --SourceSubscriptionId <SOURCE_SUB_ID> \
       --SourceResourceGroup <SOURCE_RG> \
       --TargetSubscriptionId <TARGET_SUB_ID> \
       --TargetResourceGroup <TARGET_RG>
```

### Command-Line Flags

#### Required Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--SourceSubscriptionId` | Source Azure subscription ID (UUID format) | `00000000-0000-0000-0000-000000000000` |
| `--SourceResourceGroup` | Source resource group name | `my-source-rg` |
| `--TargetSubscriptionId` | Target Azure subscription ID (UUID format) | `00000000-0000-0000-0000-000000000001` |
| `--TargetResourceGroup` | Target resource group name | `my-target-rg` |

#### Optional Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--outpath` | Directory path for output files | `./output` |
| `--debug` | Enable debug mode with timing information | `false` |
| `--help` | Display help information | - |
| `--version` | Display version information | - |

### Examples

#### Example 1: Validate Move (Same Tenant, Different Subscription)

```bash
./armv \
  --SourceSubscriptionId "12345678-1234-1234-1234-123456789012" \
  --SourceResourceGroup "rg-prod-east" \
  --TargetSubscriptionId "87654321-4321-4321-4321-210987654321" \
  --TargetResourceGroup "rg-dev-west"
```

#### Example 2: Validate Move with Custom Output Directory

```bash
./armv \
  --SourceSubscriptionId "12345678-1234-1234-1234-123456789012" \
  --SourceResourceGroup "source-rg" \
  --TargetSubscriptionId "12345678-1234-1234-1234-123456789012" \
  --TargetResourceGroup "target-rg" \
  --outpath "/var/log/armv-reports"
```

#### Example 3: Validate Move with Debug Output

```bash
./armv \
  --SourceSubscriptionId "12345678-1234-1234-1234-123456789012" \
  --SourceResourceGroup "source-rg" \
  --TargetSubscriptionId "12345678-1234-1234-1234-123456789012" \
  --TargetResourceGroup "target-rg" \
  --debug
```

### Output

When validation completes, a timestamped report file is created in the output directory:

```
./output/output-2024-10-30-14-30-45.txt
```

**Success Output (HTTP 204):**
```
Validation succeeded with HTTP code: 204
All resources are eligible for move
```

**Failure Output (HTTP 409):**
The output file contains a detailed JSON error report with:
- Error code and message
- Diagnostic information (timestamp, tracking ID, correlation ID)
- Per-resource validation failures with specific reasons

## MCP Server Mode

In addition to running as a CLI, ARMV can expose its validation engine as a [Model Context Protocol](https://modelcontextprotocol.io) server. This lets LLM-based agents (Claude Desktop, Claude Code, VS Code MCP extensions, custom MCP clients) invoke resource-move validation as a tool.

Built on the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk), the server speaks MCP over **stdio** (standard input/output). The client is responsible for launching the `armv` binary as a subprocess; communication happens via newline-delimited JSON-RPC on the child's stdin/stdout.

### Starting the Server

```bash
./armv mcp serve
```

The server runs in the foreground and blocks until the client disconnects or the process is cancelled. It emits **no output on stdout** other than MCP protocol messages — any logs, errors, or debug information go to stderr.

### Exposed Tools

| Tool | Description |
|------|-------------|
| `validate_move` | Validate whether all resources in a source resource group can be moved to a target resource group (optionally in a different subscription) without performing the move. |
| `list_subscriptions` | List every Azure subscription the supplied credential can see. Used as the first step in a discovery flow so the LLM can offer the user a picklist instead of asking them to recall UUIDs. |
| `list_resource_groups` | List every resource group in a given subscription. |
| `list_resources` | List every Azure resource in a given resource group (name, type, location, ARM ID). Useful for inspecting what's in an RG before validating, or for pinpointing a likely blocker. |

All four tools share the same credential model — `bearer_token` > SP triple > `DefaultAzureCredential`. See [Credential selection](#credential-selection-priority-order) below.

#### Typical Discovery Flow

```
User:     "I want to validate moving something from one of my subs."
LLM:      → list_subscriptions
          "You have 3: prod-east, dev-west, sandbox. Which one?"
User:     "dev-west"
LLM:      → list_resource_groups(subscription_id: dev-west)
          "7 RGs: rg-app, rg-data, rg-network… which?"
User:     "rg-app, move to prod-east."
LLM:      → list_resource_groups(subscription_id: prod-east)     (confirms target RG exists)
          → validate_move(source/target …)
          "24 of 27 resources can move; the Container Instance is blocking."
```

The LLM chains the calls itself based on the user's natural-language intent and the tool descriptions exposed via `tools/list`.

#### `validate_move` — Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source_subscription_id` | string (UUID) | yes | Source Azure subscription ID |
| `source_resource_group` | string | yes | Source resource group name |
| `target_subscription_id` | string (UUID) | yes | Target Azure subscription ID |
| `target_resource_group` | string | yes | Target resource group name |
| `tenant_id` | string (UUID) | no | Service principal tenant ID |
| `client_id` | string (UUID) | no | Service principal client (application) ID |
| `client_secret` | string | no | Service principal client secret |
| `bearer_token` | string | no | Pre-fetched Azure AD bearer token for `https://management.azure.com` |

#### Credential selection (priority order)

1. **`bearer_token`** — if supplied, the server uses it directly and stores no credentials. The client is responsible for fetching the token (e.g. `az account get-access-token --resource https://management.azure.com`) and refreshing it when it expires (~1 hour). Mixing `bearer_token` with SP fields is rejected.
2. **Service principal** — all three of `tenant_id` / `client_id` / `client_secret` present. Supplying only one or two is rejected.
3. **`DefaultAzureCredential`** — fallback when no auth fields are supplied. Walks the standard Azure credential chain: environment variables, workload identity, managed identity, `az login`.

For local desktop use, option 3 with `az login` is the most ergonomic — no secrets anywhere. For sensitive environments where no credentials should ever reach the server process, option 1 (bearer token) is recommended.

All four tools (`validate_move`, `list_subscriptions`, `list_resource_groups`, `list_resources`) accept these same four optional auth fields, so a single credential strategy works across the whole discovery flow.

#### Discovery Tool Schemas

**`list_subscriptions`** — input is just the four auth fields (no resource parameters). Output:

| Field | Type | Description |
|-------|------|-------------|
| `subscriptions[].subscription_id` | string (UUID) | Subscription UUID — use as `source_subscription_id` / `target_subscription_id` in `validate_move` |
| `subscriptions[].display_name` | string | Human-readable name |
| `subscriptions[].state` | string | Enabled / Disabled / Warned |
| `subscriptions[].id` | string | Fully qualified ARM ID |
| `count` | int | Number of subscriptions returned |

**`list_resource_groups`** — additional required input:

| Field | Type | Description |
|-------|------|-------------|
| `subscription_id` | string (UUID) | Subscription to enumerate |

Output contains `resource_groups[].name`, `resource_groups[].id`, `resource_groups[].location`, plus the echoed `subscription_id` and `count`.

**`list_resources`** — additional required inputs:

| Field | Type | Description |
|-------|------|-------------|
| `subscription_id` | string (UUID) | Subscription containing the RG |
| `resource_group` | string | Resource group name |

Output contains `resources[].name`, `resources[].type`, `resources[].id`, `resources[].location`, plus echoed `subscription_id`, `resource_group`, and `count`.

#### Output Schema

| Field | Type | Description |
|-------|------|-------------|
| `success` | bool | `true` when the Azure API returned HTTP 204 |
| `resource_ids` | string[] | Every resource enumerated in the source resource group |
| `target_resource_group_id` | string | Fully qualified ID of the target resource group |
| `http_status_code` | int | HTTP status code of the validate-move response (204 = ok, 409 = conflict) |
| `http_status` | string | HTTP status string |
| `diagnostics` | string | Raw response body — typically the 409 error payload when validation fails |

### Connecting a Client

An MCP client connects by spawning `armv mcp serve` as a child process and exchanging JSON-RPC messages over the child's stdio streams. Most clients drive this through a configuration file — below are the common ones.

#### Claude Desktop

Edit `claude_desktop_config.json`:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "armv": {
      "command": "/absolute/path/to/armv",
      "args": ["mcp", "serve"]
    }
  }
}
```

Restart Claude Desktop. The `validate_move` tool appears in the tool picker, and Claude can invoke it in conversation.

#### Claude Code (CLI)

```bash
claude mcp add armv --command /absolute/path/to/armv --args mcp --args serve
```

Or edit `~/.claude/mcp.json` directly:

```json
{
  "mcpServers": {
    "armv": {
      "command": "/absolute/path/to/armv",
      "args": ["mcp", "serve"]
    }
  }
}
```

#### VS Code (with an MCP extension)

In `.vscode/mcp.json` or the extension's settings:

```json
{
  "servers": {
    "armv": {
      "type": "stdio",
      "command": "/absolute/path/to/armv",
      "args": ["mcp", "serve"]
    }
  }
}
```

#### MCP Inspector (debugging)

The official [MCP Inspector](https://github.com/modelcontextprotocol/inspector) is the fastest way to verify the server and explore the tool schema:

```bash
npx @modelcontextprotocol/inspector /absolute/path/to/armv mcp serve
```

The inspector UI lists the `validate_move` tool with its input and output JSON schemas, and lets you invoke it interactively with arbitrary arguments.

#### Passing Credentials via Environment

If you prefer to configure service principal credentials once at the client level rather than passing them on every tool call, set the standard Azure SDK environment variables on the spawned process and omit `tenant_id` / `client_id` / `client_secret` from the tool call — `DefaultAzureCredential` will pick them up:

```json
{
  "mcpServers": {
    "armv": {
      "command": "/absolute/path/to/armv",
      "args": ["mcp", "serve"],
      "env": {
        "AZURE_TENANT_ID": "<tenant-uuid>",
        "AZURE_CLIENT_ID": "<client-uuid>",
        "AZURE_CLIENT_SECRET": "<secret>"
      }
    }
  }
}
```

For a cert-based SP (no client secret on disk), swap `AZURE_CLIENT_SECRET` for `AZURE_CLIENT_CERTIFICATE_PATH` pointing at a PEM file. The Azure SDK picks up the cert automatically.

#### Using `az login` (No Secrets Anywhere)

For local desktop use, run `az login` once in a terminal. Tokens are cached under `~/.azure/` and refresh themselves. Configure the MCP client with no auth fields at all:

```json
{ "mcpServers": { "armv": { "command": "/absolute/path/to/armv", "args": ["mcp", "serve"] } } }
```

`DefaultAzureCredential` finds the `az` session and uses it. No secrets in config files, no env vars, no per-call parameters.

#### Client-Supplied Bearer Token (Zero Credentials on the Server)

If you don't want any Azure credential material — secret, cert, or cached session — to reach the server process, the client fetches an access token itself and passes it per-call. The server never stores anything.

Client-side, before each call (or once per hour):

```bash
az account get-access-token --resource https://management.azure.com --query accessToken -o tsv
```

Pass the resulting string as `bearer_token` in the `validate_move` arguments. The server wraps it in a static credential, makes the Azure call, and discards it. If the token is expired, the Azure API returns 401 and the error surfaces to the client — fetch a fresh token and retry. The client carries the full responsibility for token acquisition and refresh.

This path is useful when:
- The server runs in a shared or less-trusted context and you don't want credentials living there.
- The caller already has a token from an upstream OAuth exchange (e.g. an agent that obtained one through on-behalf-of flow).
- Different tool calls need to run as different principals.

### How Users Invoke the Tool

Users don't fill in `source_resource_group`, `source_subscription_id`, etc. in a form. They describe the intent to the LLM in natural language; the LLM reads the tool's JSON Schema (with the field descriptions declared via `jsonschema:"..."` tags in the input struct) and extracts the arguments from the conversation.

#### From an LLM Chat Client (Claude Desktop, Claude Code, etc.)

Typical interaction:

> **User:** "Can you check if I can move everything in `rg-prod-east` (sub `12345678-1234-1234-1234-123456789012`) to `rg-dev-west` in subscription `87654321-4321-4321-4321-210987654321`?"

What happens:

1. The LLM has already fetched the tool list (including `validate_move`'s input schema) when the MCP client connected.
2. It extracts each required field from the message:
   - `source_resource_group` ← `"rg-prod-east"`
   - `source_subscription_id` ← `"12345678-1234-1234-1234-123456789012"`
   - `target_resource_group` ← `"rg-dev-west"`
   - `target_subscription_id` ← `"87654321-4321-4321-4321-210987654321"`
3. The MCP client shows a **consent prompt** listing the tool name and exact arguments about to be sent. The user approves or denies per call.
4. On approval, the server runs the validation and returns the structured result.
5. The LLM reads both the structured JSON and its text form, then replies in natural language (e.g. *"24 of 27 resources are movable; the Container Instance `.../aciresource` is blocking the move because `Microsoft.ContainerInstance/containerGroups` doesn't support resource moves."*).

If the message is ambiguous — *"validate the prod-to-dev move"* — the LLM will ask follow-up questions, because the schema marks those four fields as required.

#### Where Each Field Comes From in Practice

| Field | Typical source |
|-------|----------------|
| `source_subscription_id`, `source_resource_group`, `target_subscription_id`, `target_resource_group` | Extracted by the LLM from the user's chat message. |
| `tenant_id` / `client_id` / `client_secret` | **Almost never typed into chat** (leaks into conversation logs). Instead: set as env vars on the MCP server process (`AZURE_*` in the client config), and the LLM omits these fields — `DefaultAzureCredential` picks them up automatically. |
| `bearer_token` | An upstream agent or client-side helper shells out to `az account get-access-token --resource https://management.azure.com` and injects the token per call. Users don't paste tokens into chat. |
| No auth fields at all | User ran `az login` once. `DefaultAzureCredential` uses the cached session. Simplest and recommended for local desktop use. |

#### Drafting Effective `jsonschema` Descriptions

The `jsonschema:"..."` tag on each input field is the instruction the LLM reads when deciding how to call the tool. If you see misextracted arguments, tune the description to be more specific:

```go
SourceSubscriptionID string `json:"source_subscription_id" jsonschema:"source Azure subscription UUID (required)"`
```

Mentioning the format (UUID, ARM path, resource group name) and whether it's required helps the model extract correctly from less-structured prompts.

#### Direct Form Entry (No LLM)

For debugging or scripted callers that don't want to involve an LLM at all, use the **MCP Inspector**:

```bash
npx @modelcontextprotocol/inspector /absolute/path/to/armv mcp serve
```

The inspector renders a web form with one labelled input per schema field (`source_resource_group`, `source_subscription_id`, etc.), so you type arguments directly. It's the fastest way to verify the tool schema, exercise edge cases, and reproduce bugs without round-tripping through a model.

Custom MCP clients (VS Code extensions, first-party agent UIs, scripted clients using the MCP SDKs) can render any UI they want — the server exposes the same JSON Schema, and clients decide how to collect inputs.

### Example Tool Invocation

From an MCP client, a call looks like this (the client marshals it into JSON-RPC; you don't write this by hand):

```json
{
  "name": "validate_move",
  "arguments": {
    "source_subscription_id": "12345678-1234-1234-1234-123456789012",
    "source_resource_group": "rg-prod-east",
    "target_subscription_id": "87654321-4321-4321-4321-210987654321",
    "target_resource_group": "rg-dev-west"
  }
}
```

A successful response (resources are eligible for move):

```json
{
  "success": true,
  "resource_ids": ["/subscriptions/.../resourceGroups/rg-prod-east/providers/..."],
  "target_resource_group_id": "/subscriptions/.../resourceGroups/rg-dev-west",
  "http_status_code": 204,
  "http_status": "204 No Content"
}
```

A failed response (at least one resource cannot be moved) sets `success: false`, populates `http_status_code: 409`, and includes the full Azure error payload in `diagnostics`.

### Progress Updates (Long-Running Operations)

Azure validate-move can take anywhere from a few seconds to several minutes depending on the number and type of resources in the source group. Rather than leaving the user staring at a silent spinner, the server emits **MCP progress notifications** during the call so the client can display live status in the UI.

#### What the user sees

If the MCP client opts in, it renders updates inline while the tool is running:

```
Verifying Azure credentials
Enumerating resource groups and resources
Starting Azure validate-move for 27 resource(s)
Polling Azure validate-move (elapsed 2s)
Polling Azure validate-move (elapsed 4s)
Polling Azure validate-move (elapsed 6s)
…
Validation complete (HTTP 204)
```

The LLM sees the final structured result as before; the progress stream is a client-UI concern and doesn't enter the model's context.

#### How it works

1. **Client opt-in** — the MCP client includes a `progressToken` in the `_meta` block of the initial tool call. Claude Desktop, Claude Code, and MCP Inspector all do this automatically for any tool invocation.
2. **Server emits `notifications/progress`** — per MCP spec, each notification carries the original `progressToken`, a monotonically increasing `progress` value, and a human-readable `message`. Emitted at every phase transition (credentials, RG enumeration, Azure call start) and once per 2-second poll tick.
3. **Client matches token → UI update** — the client correlates the token with the in-flight tool call and renders each update in its status area.

If the client did **not** send a `progressToken`, the server skips notifications entirely (zero overhead) and the tool call still returns the final result.

#### Timeouts to be aware of

| Layer | Limit |
|-------|-------|
| Server-side polling ceiling | **30 minutes** (hard cap; applied via `context.WithTimeout` in `PollAndCollect`) |
| Poll interval | **2 seconds** — one progress update per tick |
| Azure validate-move API | seconds to a few minutes for typical RGs |
| MCP client per-call deadline | **Client-specific** (Claude Desktop typically ~60 seconds). If the client cancels, `ctx.Done()` propagates through to the poller; the call returns a cancellation error and the LLM sees it as a tool failure. |

The weakest ceiling wins. For large RGs or slow resource types, tune the client's tool-call timeout upward — progress notifications don't extend the client's deadline, they just keep the user informed.

#### Cancelling in-flight calls

MCP supports `notifications/cancelled` from the client. When received, the SDK cancels the handler's `ctx`, which propagates through `auth.CheckLogin`, the Azure SDK's HTTP client, and `PollAndCollect`. The in-flight Azure call will abort at the next poll boundary (within ~2 seconds) and return a context-cancellation error to the client. No cleanup on the Azure side is needed — validate-move is read-only.

#### What Users Can and Can't Ask About Status

Progress notifications go to the **MCP client UI**, not to the LLM's context. That distinction matters for what the user can actually do with them.

**✅ During an in-flight call:**

- **Watch the UI** — clients render each notification inline automatically (`Polling Azure validate-move (elapsed 8s)`, etc.). No interaction with the model needed.
- **Cancel** — the client's stop/cancel control sends `notifications/cancelled`; the handler's `ctx` is cancelled, the Azure call aborts at the next poll boundary, and the LLM receives a cancellation error as the tool result.

**✅ After the call returns**, the LLM can answer any question grounded in the structured output (`success`, `resource_ids`, `http_status_code`, `diagnostics`):

- *"Did the validation succeed?"*
- *"Which resources are blocking the move and why?"*
- *"How many resources were checked?"*
- *"What's the Azure tracking ID for follow-up?"*

**❌ What doesn't work:**

- *"How's it going?"* asked mid-call. From the model's perspective, `validate_move` is a single blocking function call — it can't reply to chat messages until that call returns. Anything the user types during the call is queued and handled afterwards.
- Issuing a parallel tool call to query status. Each tool call is stateless and independent; there is no `get_progress(call_id)` endpoint because no server-side state ties calls together.
- Progress updates entering the LLM's reasoning context. By design — a 2-second-interval poll stream would bloat the context for no gain.

**When you'd want a real "status query" tool:** only if the LLM needs to *reason* about intermediate state (e.g. "if it's still running after 5 minutes, summarise the resources enumerated so far"). That requires an async job pattern — `start_validate_move` returns a `job_id`, plus `get_validation_status(job_id)` and `get_validation_result(job_id)` tools, plus a server-side job registry. Not currently implemented; open an issue if you have a concrete use case.

### Choosing an LLM / Model

**Short version:** a small, fast, tool-capable model is plenty. This server's workload is tool selection and field extraction from chat — not heavy reasoning — so you don't need a frontier model.

#### Recommended (works well, fast and cheap)

| Model | Notes |
|-------|-------|
| **Claude Haiku 4.5** | Ideal default. Very good at MCP tool selection, minimal latency, lowest cost. What I'd wire into Claude Desktop/Code for day-to-day use. |
| **Claude Sonnet 4.6** | Step up when you want the model to *analyse* 409 diagnostics or propose remediations, not just surface them. |
| **GPT-4o-mini / GPT-4.1-mini** | Comparable to Haiku for tool-use. Good option in OpenAI-backed MCP clients. |
| **Self-hosted Llama 3.1 70B (or larger)** | Works if the client has reliable tool-use support for it. Smaller Llamas tend to struggle with correct JSON argument formatting. |

#### Why a small model is enough

- **Only 4 tools** exposed. Tool-selection ambiguity is low — any model trained on MCP/function-calling should pick correctly from the descriptions.
- **Inputs are short strings** (UUIDs, RG names). Extraction is pattern-matching, not reasoning.
- **Outputs are structured JSON** with labelled fields — the model mostly summarises, it doesn't compute.
- **Discovery flows are 3–4 sequential calls** — well within any modern agent's orchestration ability.

#### When to step up to a larger model

Reach for **Sonnet 4.6** (or equivalent) if you regularly hit:

- **Large 409 diagnostics** — validation failures with dozens or hundreds of blocked resources. Small models miss patterns when grouping failures; larger ones cluster by resource type and produce more useful summaries.
- **Remediation suggestions** — *"this Container Instance can't move; delete and recreate in the target RG"* needs broader Azure knowledge than raw diagnostics provide. Small models will parrot the error message; larger ones explain the fix.
- **Multi-RG migration planning** — reasoning across many RGs, resource dependencies, or move-order sequencing. Context and reasoning load both grow.

#### Rule of thumb

- **Interactive validation + discovery loop** → Haiku 4.5 (or peers). Fast, cheap, sufficient.
- **Analysis + remediation + planning on top of raw validation** → Sonnet 4.6 (or peers).
- **Anything below this tier** (Claude Instant, older Llama, GPT-3.5) — workable for `validate_move` alone with explicit prompts, but tool-use reliability degrades noticeably on the discovery chain.

The `jsonschema:"…"` descriptions on every input field are written to be terse and unambiguous — deliberate, to keep small models on the happy path. If you see a model consistently misinterpreting a field, tune its description first before reaching for a bigger model.

#### Self-Hosted Models

The server is transport- and model-agnostic — nothing forces you to use a hosted frontier model. But tool-use reliability varies more across open models than hosted ones, so the model you pick matters more.

**What you need on the model side:**

1. **Instruction-tuned** (not base/completion).
2. **Function-calling / tool-use fine-tune.** Many open models had this added via JSON-mode training or ToolLlama-style fine-tunes. Base chat models without tool-use training often produce malformed JSON arguments.
3. **Enough capability** for structured JSON output and chaining 3–4 tool calls in order. In practice that's ≥ ~8–14B parameters for the discovery flow to work reliably.

**Models that work well (ranked loosely by reliability on MCP workloads):**

| Model | Notes |
|-------|-------|
| **Qwen 2.5 Instruct** (14B / 32B / 72B) | Explicitly trained for function calling. 14B is the smallest I'd trust for the full discovery chain; 32B is a sweet spot on a single 24 GB GPU at 4-bit quant. |
| **Llama 3.3 70B Instruct** | Current Meta flagship for open tool-use. Needs multi-GPU or H100 at full precision, fits on 48 GB at 4-bit. |
| **Hermes 3** (Llama 3.1 8B / 70B) | Nous's function-calling fine-tune. The 8B variant is surprisingly reliable for single-tool calls; 70B handles the full chain. |
| **Mistral / Mixtral 8x22B Instruct** | Good tool use via fine-tune. Memory-heavy — plan for ≥80 GB VRAM or a beefy 4-bit quant. |
| **Command R+** (Cohere, open weights) | Designed specifically for RAG and tool use. Strong function calling out of the box. |
| **Llama 3.1 8B Instruct** | Fine for `validate_move` alone with direct prompts; chain-of-tools (discovery flow) is flakier — it'll occasionally forget `subscription_id` between calls. |

**Models to avoid for this workload:**

- Base (non-instruction) models.
- Models < 7 B — JSON argument formatting becomes unreliable.
- Code-only models (DeepSeek-Coder-base, CodeLlama-base).
- Older Llama/Mistral checkpoints without a tool-use fine-tune — they often invent fields or produce free text when strict JSON is required.

**Inference stacks that speak MCP cleanly:**

| Stack | Why it fits |
|-------|-------------|
| **vLLM** (OpenAI-compatible API + native tool calling) | Production-grade; most MCP-capable clients connect via its OpenAI-compatible endpoint. Best throughput. |
| **llama.cpp server** | Lightweight, OpenAI-compatible, runs on CPU or small GPUs. Good for a single-user local dev box. |
| **LM Studio** | GUI-based, exposes an OpenAI-compatible server. Easiest way to try a model before committing. |
| **Ollama** | Easy to run; tool-use support varies by model and by client. Works with Cline, Continue, Open WebUI, LibreChat. |
| **Text Generation Inference (HuggingFace TGI)** | OpenAI-compatible, strong multi-GPU scaling. |

**MCP-capable clients that support local LLMs:**

- **Continue** (VS Code / JetBrains extension) — Ollama, vLLM, OpenAI-compatible endpoints.
- **Cline** (VS Code extension, formerly Claude Dev) — Ollama-first.
- **LibreChat** — self-hostable ChatGPT clone with MCP support and provider plugins.
- **Open WebUI** — has MCP tool integration.
- **Custom agents** — you can drive the server directly from any Python/TypeScript MCP SDK plus your LLM of choice.

**Practical starting points:**

| Setup | Recommended model | Why |
|-------|-------------------|-----|
| Single-user local dev (24 GB GPU) | Qwen 2.5 14B Instruct (4-bit via vLLM or llama.cpp) | Fits comfortably, reliable tool use, fast enough interactively. |
| Single-user local dev (48 GB GPU) | Qwen 2.5 32B Instruct or Hermes 3 Llama 3.1 70B (4-bit) | Noticeable step up on chained calls and diagnostic summarisation. |
| Multi-user / team (server GPU) | Llama 3.3 70B Instruct or Qwen 2.5 72B via vLLM | Throughput and reliability for a small team. |
| CPU-only / tiny | Qwen 2.5 7B Instruct via llama.cpp (Q4_K_M) | Workable for `validate_move` with direct prompts; expect the occasional retry on discovery chains. |

**Safety net**: The server's server-side input validators (`validateInputs`, `selectCredential`, `validateListResourcesInput`) reject malformed arguments with clear error text. When a flaky local model produces bad JSON or forgets a field, the tool returns an error the model can see and retry, rather than silently doing the wrong thing. This makes the whole setup notably more forgiving of lower-quality models than it would be otherwise.

## Testing

The project includes comprehensive unit tests with 38+ test cases covering:

### Test Modules

| Test File | Focus | Test Cases |
|-----------|-------|-----------|
| `validateinput_test.go` | UUID/subscription ID validation | 8 tests |
| `args_test.go` | Command-line argument parsing | 3 tests |
| `azureresourcemoveinfo_test.go` | Validation parameter initialization | 5 tests |
| `command_test.go` | Cobra command configuration | 10 tests |
| `pollerresponsedata_test.go` | API response processing | 4 tests |
| `jsonutils_test.go` | JSON serialization/deserialization | 14 tests |
| `outputfile_test.go` | File I/O and permissions | 7 tests |

**Test Status**: ✅ 100% Pass Rate

Run tests with:

```bash
# Run all tests
task test

# Verbose output
go test ./... -v

# With coverage
go test ./... -cover

# Run security checks
task seccheck          # govulncheck
task staticcheck       # Static analysis
task vet               # go vet
```

## Limitations

- **Single Tenant**: Currently supports only subscriptions and resource groups within the same Azure tenant
- **Source Resource Group**: All resources to be moved must be in the same source resource group
- **Target Flexibility**: Target resource group can be in a different subscription (within the same tenant)

## Troubleshooting

### Azure Authentication Issues

**Error:** `No credentials found`

**Solution:** Ensure you're logged in to Azure CLI:
```bash
az login
az account show  # Verify active subscription
```

### Input Validation Errors

**Error:** `Invalid subscription ID format`

**Solution:** Ensure subscription IDs are valid UUIDs:
```
Format: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
Example: 12345678-1234-1234-1234-123456789012
```

### Resource Group Not Found

**Error:** `Resource group not found`

**Solution:**
1. Verify the resource group exists in the specified subscription
2. Ensure you have permissions to access the resource group
3. Check that you're using the correct subscription context

### API Timeout

**Error:** `Validation operation timed out after 30 minutes`

**Solution:**
- This indicates the Azure API is taking longer than expected
- Check Azure service health: https://status.azure.com/
- Retry the validation operation
- Consider validating smaller batches of resources

## Reporting Issues

Found a bug or have a feature request? Please report it on GitHub:

📋 **[Report Issues Here](https://github.com/AaronSaikovski/armv/issues)**

Include the following information:
- ARMV version (`./armv --version`)
- Go version (`go version`)
- Error message and output file
- Steps to reproduce

## Development

### Code Quality

The project maintains high code quality standards:

```bash
# Format and lint
task lint

# Security vulnerability scan
task seccheck

# Static analysis
task staticcheck

# Vet (examines code for suspicious constructs)
task vet
```

### Git Workflow

1. Create a feature branch: `git checkout -b feature/my-feature`
2. Make changes and commit: `git commit -am "Add new feature"`
3. Push to remote: `git push origin feature/my-feature`
4. Create a pull request

### GitHub Actions Workflows

The project includes automated CI/CD workflows:

#### Build Workflow (`.github/workflows/build.yml`)
Runs on every push and pull request to `main`:
- Multi-platform matrix testing (Ubuntu, Windows, macOS)
- Dependency verification
- Code quality checks (go vet, staticcheck, govulncheck)
- Comprehensive unit tests
- Binary build verification

#### GoReleaser Workflow (`.github/workflows/goreleaser.yml`)
Runs on release creation:
- Pre-release multi-platform testing
- Cross-platform binary compilation (amd64, arm64, armv7, 386)
- Automatic release notes generation
- Binary artifact publishing

## Releases & Binary Availability

Pre-compiled binaries are available for multiple platforms:

| OS | Architectures |
|----|---|
| Linux | amd64, arm64, armv7, 386 |
| Windows | amd64, 386 |
| macOS | amd64, arm64 |

Download latest binaries from [GitHub Releases](https://github.com/AaronSaikovski/armv/releases)

## License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

## Related Projects

- **Original Tool**: [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) (Python, deprecated)
- **Azure Documentation**: [Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources)
- **Azure Go SDK**: [Azure SDK for Go](https://learn.microsoft.com/en-us/azure/developer/go/overview)

---

**Questions or feedback?** Please open an issue or check the [project repository](https://github.com/AaronSaikovski/armv)

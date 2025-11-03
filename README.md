<div align="center">

# ARMV - <u>A</u>zure <u>R</u>esource <u>M</u>oveability <u>V</u>alidator

**v1.2.2**

[![Build Status](https://github.com/AaronSaikovski/armv/workflows/build/badge.svg)](https://github.com/AaronSaikovski/armv/actions)
[![Licence](https://img.shields.io/github/license/AaronSaikovski/armv)](LICENSE)

A lightweight, high-performance CLI utility for validating Azure resource moveability without performing the actual move operation.

</div>

> **⚠️ NOTE: THIS TOOL PERFORMS READ-ONLY VALIDATION AND WILL NOT PERFORM ANY ACTUAL MOVE OPERATIONS. IT GENERATES A DETAILED VALIDATION REPORT FOR ANALYSIS.**

## Overview

ARMV is a production-ready Go CLI utility that validates whether Azure resources can be moved between resource groups and subscriptions. It acts as a comprehensive wrapper around Azure's [Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources?view=rest-resources-2021-04-01), providing intelligent polling, detailed error reporting, and secure file operations.

This tool is a complete rewrite of the deprecated [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) Python utility, now delivered as a fully standalone, self-contained Go binary with zero external dependencies.

### Key Features

- **Non-Destructive Validation**: Tests moveability without making any changes
- **Cross-Subscription Support**: Validate moves across different subscriptions (same tenant only)
- **Intelligent Polling**: Long-running operation polling with timeout protection (max 30 minutes)
- **Comprehensive Error Reporting**: Detailed JSON diagnostics including tracking IDs and timestamps
- **Progress Visualization**: Real-time progress bar during validation operations
- **Secure File Operations**: Timestamped output files with secure permissions
- **Debug Mode**: Optional verbose logging and timing information
- **Production-Ready**: Fully tested, optimized memory usage, and comprehensive error handling
- **Multi-Platform Support**: Cross-platform binaries for Linux, Windows, and macOS
- **Security Focused**: Vulnerability scanning (govulncheck), static analysis (staticcheck), and comprehensive testing
- **Continuous Integration**: Automated builds, tests, and security checks on every commit

### How It Works

1. **Input Validation**: Validates subscription IDs and resource group references
2. **Authentication**: Uses Azure CLI's stored credentials (`az login` context)
3. **Resource Discovery**: Enumerates all resources in the source resource group
4. **Validation**: Calls Azure's Validate Move Resources API
5. **Polling**: Polls the long-running operation with visual progress indication
6. **Output Generation**: Writes timestamped results to the output directory

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
| **CLI Layer** | `cmd/armv/app/` | Command-line interface, flag parsing, orchestration |
| **Authentication** | `internal/pkg/auth/` | Azure credential management and SDK clients |
| **Validation** | `internal/pkg/validation/` | Resource move validation API wrapper |
| **Resource Management** | `internal/pkg/resourcegroups/`, `internal/pkg/resources/` | Azure resource group and resource operations |
| **Polling** | `cmd/armv/poller/` | Long-running operation polling and result handling |
| **Utilities** | `pkg/utils/` | Input validation, error handling, file I/O, JSON processing |

### Project Structure

```
armv/
├── cmd/armv/                          # Application entry points
│   ├── main.go                        # CLI entry point with version embedding
│   ├── app/                           # Application logic
│   │   ├── root.go                   # Cobra CLI command setup
│   │   ├── command.go                # Main validation workflow orchestration
│   │   ├── login.go                  # Azure authentication verification
│   │   └── resourcegroup.go          # Resource group operations
│   └── poller/                        # Long-running operation handling
│       ├── pollapi.go                # API polling logic with timeout
│       ├── pollresponse.go           # Response handling and output
│       ├── pollerresponsedata.go     # Response data structures
│       ├── progressbar.go            # Progress bar visualization
│       └── constants.go              # Configuration constants
│
├── internal/pkg/                      # Private internal packages
│   ├── auth/                          # Azure authentication
│   │   └── auth.go                   # Credential and client creation
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
| `github.com/Azure/azure-sdk-for-go/sdk/azcore` | v1.19.1 | Azure SDK core functionality |
| `github.com/Azure/azure-sdk-for-go/sdk/azidentity` | v1.13.0 | Azure credential authentication |
| `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources` | v1.2.0 | Azure resources API client |
| `github.com/spf13/cobra` | v1.10.1 | CLI command framework |
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

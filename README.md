<div align="center">

# ARMV - Azure Resource Moveability Validator

[![Build Status](https://github.com/AaronSaikovski/armv/workflows/build/badge.svg)](https://github.com/AaronSaikovski/armv/actions)
[![Licence](https://img.shields.io/github/license/AaronSaikovski/armv)](LICENSE)

</div>

### Description

Provides a command line interface to validate moving of Azure resources from one subscription & resource group to another within a Single Azure tenant.

The expected inputs are a source Azure SubscriptionID and Resource Group, passes these to the [Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources?view=rest-resources-2021-04-01) and validates these source resources against the target SubscriptionID and ResourceGroup and reports accordingly.

**NOTE: This tool will not perform the move operation but will generate a report for further analysis.**

This tool is a rewrite of the [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) tool originally written in Python and now greatly enhanced in 100% Go. This tool utilisies the [Microsoft Go SDK](https://learn.microsoft.com/en-us/azure/developer/go/overview) and is cross platform and designed to be self-contained.

### Software Requirements

- [Go v1.22](https://www.go.dev/dl/) or later needs to be installed to build the code.
- [Azure CLI tools](https://learn.microsoft.com/en-us/cli/azure/) 2.50 or later

## Azure Setup

You must be logged into Azure from the CLI (az login) for this program to work. This program will use the CLIs current logged in identity context.  
Ensure you have run the following:

```bash
az login

# Where "XXXX-XXXX-XXXX-XXXX" is your subscriptionID
az account set --subscription "XXXX-XXXX-XXXX-XXXX"
```

## Installation

The toolchain is mainly driven by the Makefile.

```bash
help         - Display help about make targets for this Makefile
release      - Builds the project in preparation for (local)release
build        - Builds the project in preparation for debug
run          - builds and runs the program on the target platform
clean        - Remove the old builds and any debug information
test         - executes unit tests
deps         - fetches any external dependencies and updates
vet          - Vet examines Go source code and reports suspicious constructs
staticcheck  - Runs static code analyzer staticcheck - currently broken
seccheck     - Code vulnerability check
lint         - format code and tidy modules
```

To get started type,

- make dep - to fetch all dependencies
- make build - to build debug version for your target environment architecture
- make release - Builds a release version for your target environment architecture

## Usage

```bash
./armv --SourceSubId "XXXX-XXXX-XXXX-XXXX" --SourceRsg "SourceRSG" --TargetSubId "XXXX-XXXX-XXXX-XXXX" --TargetRsg "TargetRSG"
```

## Known issues and limitations

- Currently this program only supports subscriptions and resource groups under the same single tenant.
- No know bugs or known issues - if found, please report [here](https://github.com/AaronSaikovski/armv/issues)

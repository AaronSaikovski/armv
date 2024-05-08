<div align="center">

# ARMV - Azure Resource Moveability Validator - v0.0.1-beta

[![Build Status](https://github.com/AaronSaikovski/armv/workflows/build/badge.svg)](https://github.com/AaronSaikovski/armv/actions)
[![Licence](https://img.shields.io/github/license/AaronSaikovski/armv)](LICENSE)

</div>

**Please note that this utility is still in beta and still undergoing extensive testing - please report all bugs and issues [here](https://github.com/AaronSaikovski/armv/issues)**

### Description

This utility performs a validation on whether the specified Azure resources can be moved to the specified target.
The resources to be moved must be in the same source resource group in the source subscription being used.
The target resource group may be in a different subscription.
If validation succeeds, it returns HTTP response code 204 (no content), If validation fails, it returns HTTP response code 409 (Conflict) with an error message.
If the operation fails it returns an \*azcore.ResponseError type.

The expected inputs are a source Azure SubscriptionID and Resource Group, passes these to the [Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources?view=rest-resources-2021-04-01) and validates these source resources against the target SubscriptionID and ResourceGroup and reports accordingly.

The API operation has been abstracted via the Go Azure SDK and also incorporates polling operatons and error handling and types.

**NOTE: This tool will NOT perform the move operation but will generate a report for further analysis.**

This tool is a rewrite of the [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) tool originally written in Python and now greatly enhanced in 100% Go.
This tool utilisies the [Microsoft Go Azure SDK](https://learn.microsoft.com/en-us/azure/developer/go/overview) and is cross platform and designed to be self-contained.

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
./armv --SourceSubscriptionId SOURCESUBSCRIPTIONID --SourceResourceGroup SOURCERESOURCEGROUP --TargetSubscriptionId TARGETSUBSCRIPTIONID --TargetResourceGroup TARGETRESOURCEGROUP
```

## Known issues and limitations

- Currently this utility only supports subscriptions and resource groups under the same single tenant.
- No know bugs or known issues - if found, please report [here](https://github.com/AaronSaikovski/armv/issues)

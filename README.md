<div align="center">

# ARMV - <u>A</u>zure <u>R</u>esource <u>M</u>oveability <u>V</u>alidator

### v0.1.0-alpha.1-release

[![Build Status](https://github.com/AaronSaikovski/armv/workflows/build/badge.svg)](https://github.com/AaronSaikovski/armv/actions)
[![Licence](https://img.shields.io/github/license/AaronSaikovski/armv)](LICENSE)

</div>

**PLEASE NOTE THAT THIS UTILITY IS STILL IN ALPHA AND UNDERGOING EXTENSIVE TESTING\*- PLEASE REPORT ALL BUGS AND ISSUES [HERE](https://github.com/AaronSaikovski/armv/issues)**

**<u>NOTE: THIS TOOL WILL NOT PERFORM THE MOVE OPERATION BUT WILL GENERATE A MOVE REPORT FOR FURTHER ANALYSIS.</u>**

### Description:

This CLI utility performs a validation on whether the specified Azure resources can be moved to the specified target resource group. \
The resources to be moved must be in the same source resource group in the source subscription being used. \
The target resource group may be in a different subscription. \
If validation succeeds, it returns HTTP response code 204 (no content), If validation fails, it returns HTTP response code 409 (Conflict) with an error message. \
If the operation fails it outputs a detailed json error report similar to this example:

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

The expected inputs are a source Azure SubscriptionID and Resource Group, passes these to the [Validate Move Resources API](https://learn.microsoft.com/en-us/rest/api/resources/resources/validate-move-resources?view=rest-resources-2021-04-01) and validates these source resources against the target SubscriptionID and ResourceGroup and reports accordingly.\
The API operation has been abstracted via the Go Azure SDK and also incorporates polling operatons and error handling and types.\
This tool utilises the native [Microsoft Go Azure SDK](https://learn.microsoft.com/en-us/azure/developer/go/overview) and is a full rewrite of the now deprecated [pyazvalidatemoveresources](https://github.com/AaronSaikovski/pyazvalidatemoveresources) tool originally written in Python and now greatly enhanced in 100% Go as a fully standalone and self-contained binary with no redistributable dependencies needed.

### Software Requirements:

- [Go v1.22](https://www.go.dev/dl/) or later needs to be installed to build the code.
- [Azure CLI tools](https://learn.microsoft.com/en-us/cli/azure/) 2.50 or later
- [Taskfile](https://taskfile.dev/) to run the build chain commands listed below.

## Azure Setup:

You must be logged into Azure from the CLI (az login) for this program to work. This program will use the CLIs current logged in identity context. \
Ensure you have run the following:

```bash
az login

# Where "XXXX-XXXX-XXXX-XXXX" is your subscriptionID
az account set --subscription "XXXX-XXXX-XXXX-XXXX"
```

## Installation:

The toolchain is driven by using [Taskfile](https://taskfile.dev/) and all commands are managed via the file `Taskfile.yml`

The list of commands is as follows: \

```bash
* build:             Builds the project in preparation for debug.
* clean:             Removes the old builds and any debug information from the source tree.
* debug:             Runs a debug version of the application with input parameters from the environment file.
* deploy:            Deploy Azure resources using Bicep for testing.
* deps:              Fetches any external dependencies and updates.
* destroy:           Destroy Azure resources for testing.
* docs:              Updates the swagger docs - For APIs.
* generate:          update binary build version using gogenerate.
* goreleaser:        Builds a cross platform release using goreleaser.
* lint:              Lint, format and tidy code.
* release:           Builds the project in preparation for (local)release.
* run:               Builds and runs the program on the target platform.
* seccheck:          Code vulnerability scanner check.
* staticcheck:       Runs static code analyzer staticcheck.
* test:              Executes unit tests.
* version:           Get the Go version.
* vet:               Vet examines Go source code and reports suspicious constructs.
* watch:             Use air server for hot reloading.
```

Execute using the taskfile utility:

```bash
task <command_from_above_list>
```

To get started type,

- `task deps` - to fetch all dependencies and update all dependencies.
- `task build` - to build debug version for your target environment architecture.
- `task release` - Builds a release version for your target environment architecture - outputs to /bin folder.

## Usage

```bash
./armv --SourceSubscriptionId SOURCESUBSCRIPTIONID --SourceResourceGroup SOURCERESOURCEGROUP --TargetSubscriptionId TARGETSUBSCRIPTIONID --TargetResourceGroup TARGETRESOURCEGROUP
```

## Known issues and limitations

- Currently this utility only supports subscriptions and resource groups under the same single tenant.
- No know bugs or known issues - if found, please report [here](https://github.com/AaronSaikovski/armv/issues)

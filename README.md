# AEP Terraform Provider

The AEP Terraform Provider generates a run-time Terraform provider for use with AEP-compliant APIs. This allows API developers who have created AEP-compliant APIs to create or extend a Terraform provider with new resources with near-zero additional development time.

For more information about the AEP project, visit [aep.dev](https://aep.dev).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22 (only needed for building from source or library usage)

## Using the Provider

### Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    aep = {
      source = "aep-dev/aep"
    }
  }
}
```

### Configuration

The provider reads your AEP-compliant OpenAPI spec at startup and dynamically generates Terraform resources for each API resource it finds.

#### Environment Variables

| Variable | Required | Description |
|---|---|---|
| `AEP_OPENAPI` | Yes | URL or file path to your OpenAPI spec (e.g. `http://localhost:8081/openapi.json`) |
| `AEP_PATH_PREFIX` | No | A path prefix prepended to all API methods. Use this if all your OpenAPI paths share a common prefix (e.g. `/cloud/v2`). |

#### Provider Block

```hcl
provider "aep" {
  headers = {
    "Authorization" = "Bearer your-token-here"
    "X-API-Key"     = "your-api-key"
  }
}
```

The `headers` attribute sets custom HTTP headers sent with every API request. This is the primary way to configure authentication.

### Resources and Data Sources

The provider automatically generates its schema from the OpenAPI spec. Run `terraform providers schema -json` to see the available resources and their attributes.

Resources are named `aep_<resource>`, and collection data sources are named `aep_<resource>s`.

#### Example

Given an OpenAPI spec that defines a `publishers` resource with a nested `books` resource:

```hcl
resource "aep_publisher" "example" {
  description = "Example publisher"
}

resource "aep_book" "example" {
  publisher_id = aep_publisher.example.id
  isbn         = "978-0-13-468599-1"
  title        = "The Go Programming Language"
}
```

## Using as a Library

This provider can also be embedded into your own Terraform provider binary. This is useful if you want to customize the provider name, set a fixed OpenAPI path, or bundle it with additional resources.

See `examples/main.go` for a complete example. The key steps are:

```go
package main

import (
    "github.com/aep-dev/terraform-provider-aep/config"
    "github.com/aep-dev/terraform-provider-aep/provider"
)

func main() {
    cfg := config.NewProviderConfigWithOptions(
        "http://localhost:8081/openapi.json",          // openAPIPath
        "",                                            // pathPrefix
        "registry.terraform.io/your-org/your-provider", // registryURL
        "yourprovider",                                // providerPrefix
    )

    c := client.NewClient(http.DefaultClient)
    p, err := provider.NewProvider(&cfg, c, "1.0.0")
    // ... serve with providerserver.Serve()
}
```

When used as a library, resources will be named `yourprovider_<resource>` based on the prefix you choose.

## Developing the Provider

```sh
# Build
go build -v ./...

# Install locally
go install -v ./...

# Run tests
go test -v -cover -timeout=120s -parallel=10 ./...

# Run acceptance tests
TF_ACC=1 go test -v -cover -timeout 120m ./...
```

To point the Terraform CLI at your local build, create a CLI config file:

```hcl
provider_installation {
  dev_overrides {
    "aep-dev/aep" = "/your/gopath/bin"
  }
}
```

Then set `TF_CLI_CONFIG_FILE` to the path of that file.

# AEP Terraform Provider

The AEP Terraform Provider generates a run-time Terraform provider for use with AEP-compliant APIs. This allows API developers who have created AEP-compliant APIs to create or extend a Terraform provider with new resources with near-zero additional development time.

For more information about the AEP project, visit [aep.dev](https://aep.dev)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

This repository is a library that provides both a Terraform provider and Terraform resources. Please see `examples/main.go` to see how to create a new provider.

## Using the provider

This provider automatically generates its schema based off the AEP-compliant OpenAPI spec.

`terraform providers schema -json` will show the current schema.

## Authentication

The provider supports setting custom headers for API requests through the provider configuration. This is useful for setting authentication tokens, API keys, or other required headers.

Example configuration in `provider.tf`:

```hcl
provider "scaffolding" {
  headers = {
    "Authorization" = "Bearer your-token-here"
    "X-API-Key"     = "your-api-key"
  }
}
```

These headers will be sent with every request made by the provider to the API.


## Developing the Provider Locally

`go install` will install the provider to your `$GOPATH/bin` folder.

You'll need a CLI config file with the following code to allow the Terraform CLI to point to your local provider

```
provider_installation {
  dev_overrides {
    "hashicorp/scaffolding" = "/home/go/bin"
  }
}
```

Set `TF_CLI_CONFIG_FILE` to the path of that config file.

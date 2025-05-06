# AEP Terraform Provider

The AEP Terraform Provider generates a run-time Terraform provider for use with AEP-compliant APIs.

For more information about the AEP project, visit [aep.dev](https://aep.dev)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Configuring the provider

The provider can be configured by modifying the following variables in `config/config.go`. When building a version of the provider for distribution, all of these should be hard-coded and not set with environment variables.

- `OpenAPIPath`: The URI where your OpenAPI spec lives. This can also be set via the `AEP_OPENAPI` environment variable.
- `PathPrefix`: A path prefix that will be prepended to all OpenAPI methods. This can also be set via the `AEP_PATH_PREFIX` environment variable.
- `ProviderPrefix`: The name prefix for your provider. All resources will have the format `prefix_resource`.
- `RegistryURL`: The URL for your provider in the Terraform Registry.

Example configuration:
```go
const OpenAPIPath = "http://api.example.com/openapi.json"
const PathPrefix = "/api/v1"
const ProviderPrefix = "mycompany"
const RegistryURL = "registry.terraform.io/mycompany/myprovider"
```

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
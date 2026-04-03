---
page_title: "AEP Provider"
subcategory: ""
description: |-
  The AEP Terraform provider generates a runtime Terraform provider for use with AEP-compliant APIs.
---

# AEP Provider

The AEP Terraform provider generates a runtime Terraform provider for use with [AEP](https://aep.dev)-compliant APIs. This allows API developers who have created AEP-compliant APIs to create or extend a Terraform provider with new resources with near-zero additional development time.

The provider automatically generates its schema based on an AEP-compliant OpenAPI spec.

## Configuration

The provider is configured at build time through the `config/config.go` file:

- **OpenAPIPath** - The URI where your OpenAPI spec lives. Defaults to the `AEP_OPENAPI` environment variable if not set.
- **PathPrefix** - A value prepended to all OpenAPI methods. Useful when all methods share a common prefix. Defaults to the `AEP_PATH_PREFIX` environment variable if not set.
- **ProviderPrefix** - The name of your provider. All resources will have the prefix `prefix_resource`.
- **RegistryURL** - The URL for your provider in the Terraform Registry.

## Schema

### Optional

- `headers` (Map of String) A map of headers that will be sent across the wire.

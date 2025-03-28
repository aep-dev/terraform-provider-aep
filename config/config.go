package config

import "os"

// Only change these values.

// The URI where your OpenAPI spec lives.
// This will default to AEP_OPENAPI if empty.
const OpenAPIPath = "http://localhost:8081/openapi.json"

// A Path prefix.
// This is a value that will prepend all of your OpenAPI methods
// If all of your OpenAPI methods have the same prefix (before resources/{resource} portions), set this.
// This will default to AEP_PATH_PREFIX if empty.
const PathPrefix = ""

// The name of your provider.
// All resources will have the prefix `prefix_resource`.
const ProviderPrefix = "scaffolding"

// The URL for your provider.
// This should be set once you have a listing in the Terraform Registry.
const RegistryURL = "registry.terraform.io/hashicorp/scaffolding"

// Do not change anything below here!

type ProviderConfig struct {
	openAPIPath string
	pathPrefix  string

	ProviderPrefix string
	RegistryURL    string
}

func (c *ProviderConfig) OpenAPIPath() string {
	if c.openAPIPath != "" {
		return c.openAPIPath
	}
	return os.Getenv("AEP_OPENAPI")
}

func (c *ProviderConfig) PathPrefix() string {
	if c.pathPrefix != "" {
		return c.pathPrefix
	}
	return os.Getenv("AEP_PATH_PREFIX")
}

func NewProviderConfig() ProviderConfig {
	return ProviderConfig{
		openAPIPath:    OpenAPIPath,
		pathPrefix:     PathPrefix,
		RegistryURL:    RegistryURL,
		ProviderPrefix: ProviderPrefix,
	}
}

// Only for testing!
func NewProviderConfigForTesting(openAPI string, pathPrefix string, registryUrl string, providerPrefix string) ProviderConfig {
	return ProviderConfig{
		openAPIPath:    openAPI,
		pathPrefix:     pathPrefix,
		RegistryURL:    registryUrl,
		ProviderPrefix: providerPrefix,
	}
}

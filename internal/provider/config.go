package provider

// Only change these values.

// The URI where your OpenAPI spec lives.
const OpenAPIPath = "http://localhost:8081/openapi.json"

// The name of your provider.
// All resources will have the prefix `prefix_resource`
const ProviderPrefix = "scaffolding"

// Do not change anything below here!

type ProviderConfig struct {
	ProviderPrefix string
	OpenAPIPath    string
}

func NewProviderConfig() ProviderConfig {
	return ProviderConfig{
		ProviderPrefix: ProviderPrefix,
		OpenAPIPath:    OpenAPIPath,
	}
}

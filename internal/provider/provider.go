// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/aep-dev/aep-lib-go/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ScaffoldingProvider{}
var _ provider.ProviderWithFunctions = &ScaffoldingProvider{}

// ScaffoldingProvider defines the provider implementation.
type ScaffoldingProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	// generator is set to the information created from the OpenAPI spec.
	generator *GeneratedProviderData

	client *client.Client

	config ProviderConfig
}

// ScaffoldingProviderModel describes the provider data model.
type ScaffoldingProviderModel struct {
	Headers map[string]string `tfsdk:"headers"`
}

func (p *ScaffoldingProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scaffolding"
	resp.Version = p.version
}

func (p *ScaffoldingProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"headers": schema.MapAttribute{
				Description: "A map of headers that will be sent across the wire.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (p *ScaffoldingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ScaffoldingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for k, v := range data.Headers {
		p.client.Headers[k] = v
	}

	// Example client configuration for data sources and resources
	resp.DataSourceData = p.client
	resp.ResourceData = p.client
}

func (p *ScaffoldingProvider) Resources(ctx context.Context) []func() resource.Resource {
	resources := []func() resource.Resource{}
	for name, resource := range p.generator.api.Resources {
		resources = append(resources, NewExampleResourceWithResource(resource, p.generator.api, name, p.generator.openapi, p.generator.resources[name]))
	}
	return resources
}

func (p *ScaffoldingProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *ScaffoldingProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string, g *GeneratedProviderData, client *client.Client, config ProviderConfig) func() provider.Provider {
	return func() provider.Provider {
		return &ScaffoldingProvider{
			version:   version,
			generator: g,
			client:    client,
			config:    config,
		}
	}
}

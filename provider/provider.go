package provider

import (
	"context"

	"github.com/aep-dev/aep-lib-go/pkg/client"
	"github.com/aep-dev/terraform-provider-aep/config"
	internalprovider "github.com/aep-dev/terraform-provider-aep/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type Provider struct {
	config *config.ProviderConfig
	client *client.Client

	provider tfprovider.Provider
}

func NewProvider(config *config.ProviderConfig, client *client.Client, version string) (*Provider, error) {
	gen, err := internalprovider.CreateGeneratedProviderData(context.Background(), config.OpenAPIPath(), config.PathPrefix())
	if err != nil {
		return nil, err
	}

	return &Provider{
		config:   config,
		client:   client,
		provider: internalprovider.New(version, gen, client, *config)(),
	}, nil
}

func (p *Provider) Provider() tfprovider.Provider {
	return p.provider
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return p.provider.(interface {
		Resources(context.Context) []func() resource.Resource
	}).Resources(ctx)
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return p.provider.(interface {
		DataSources(context.Context) []func() datasource.DataSource
	}).DataSources(ctx)
}

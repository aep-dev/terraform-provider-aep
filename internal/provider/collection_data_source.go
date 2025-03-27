// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/client"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/aep-dev/terraform-provider-aep/internal/provider/data"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TODO:
// - Schema:
// - Call the List endpoint
// - Transfer the results properly.
// - Store into state (should act the same as normal?)
// - Write test harness
// - Write tests.

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CollectionDataSource{}

func NewDataSourceWithResource(r *api.Resource, a *api.API, n string, o *openapi.OpenAPI, res *data.ResourceSchema) func() datasource.DataSource {
	return func() datasource.DataSource {
		return &CollectionDataSource{
			resource:       r,
			api:            a,
			name:           n,
			o:              o,
			resourceSchema: res,
		}
	}
}

func NewCollectionDataSource() datasource.DataSource {
	return &CollectionDataSource{}
}

// CollectionDataSource defines the data source implementation.
type CollectionDataSource struct {
	resource *api.Resource
	api      *api.API
	name     string

	// Client will be configured at plan/apply time in the Configure() function.
	client         *client.Client
	o              *openapi.OpenAPI
	resourceSchema *data.ResourceSchema
}

func (d *CollectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.name
}

func (d *CollectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attr := d.resourceSchema.FullCollectionDataSourceSchema(ctx)

	resp.Schema = schema.Schema{
		MarkdownDescription: d.resource.Singular,

		Attributes: attr,
	}
}

func (d *CollectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *CollectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	dataResource := data.NewResource(d.resourceSchema)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &dataResource)...)
	dataResource.Schema = d.resourceSchema

	if resp.Diagnostics.HasError() {
		return
	}

	parameters, err := Parameters(ctx, dataResource, d.resourceSchema)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create parameters, got error: %s", err))
		return
	}

	a, err := d.client.List(ctx, d.resource, d.api.ServerURL, parameters)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	dataState, err := DataSourceState(ctx, a, dataResource, d.resource)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Read: unable to create state, got error: %s", err))
		return

	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, dataState)...)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/client"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-scaffolding/internal/provider/data"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ExampleResource{}
var _ resource.ResourceWithImportState = &ExampleResource{}

func NewExampleResourceWithResource(r *api.Resource, a *api.API, n string, o *openapi.OpenAPI) func() resource.Resource {
	return func() resource.Resource {
		return &ExampleResource{
			resource: r,
			api:      a,
			name:     n,
			o:        o,
		}
	}
}

func NewExampleResource() resource.Resource {
	return &ExampleResource{}
}

// ExampleResource defines the resource implementation.
type ExampleResource struct {
	resource *api.Resource
	api      *api.API
	name     string

	// Client will be configured at plan/apply time in the Configure() function.
	client         *client.Client
	o              *openapi.OpenAPI
	resourceSchema *ResourceSchema
}

func (r *ExampleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.name
}

func (r *ExampleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attr := r.resourceSchema.FullSchema()

	resp.Schema = schema.Schema{
		MarkdownDescription: r.resource.Singular,

		Attributes: attr,
	}
}

func (r *ExampleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *ExampleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	dataPlan := &data.Resource{}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataPlan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	parameters, err := Parameters(ctx, dataPlan, r.resourceSchema)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create parameters, got error: %s", err))
		return
	}

	body, err := Body(ctx, dataPlan, r.resourceSchema)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create body, got error: %s", err))
		return
	}
	tflog.Info(ctx, fmt.Sprintf("headers %q", r.client.Headers))

	a, err := r.client.Create(ctx, r.resource, r.api.ServerURL, body, parameters)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	dataState, err := State(ctx, a, dataPlan, r.resourceSchema)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Create: unable to create state, got error: %s", err))
		return

	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, dataState)...)
}

func (r *ExampleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	dataResource := &data.Resource{}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &dataResource)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jsonDataMap, err := dataResource.ToJSON()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal JSON, got error: %s", err))
		return
	}

	pathInterface, ok := jsonDataMap["path"]
	if !ok {
		resp.Diagnostics.AddError("Client Error", "Unable to find path")
		return
	}
	path, ok := pathInterface.(string)
	if !ok {
		resp.Diagnostics.AddError("Client Error", "Unable to convert path to string")
		return
	}

	a, err := r.client.Get(ctx, r.api.ServerURL, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	dataState, err := State(ctx, a, dataResource, r.resourceSchema)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Read: unable to create state, got error: %s", err))
		return

	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, dataState)...)
}

func (r *ExampleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	dataResource := &data.Resource{}
	dataState := &data.Resource{}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataResource)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &dataState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body, err := Body(ctx, dataResource, r.resourceSchema)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to body, got error: %s", err))
		return
	}

	s, ok := dataState.Values["path"]
	if !ok {
		resp.Diagnostics.AddError("Client Error", "Unable to fetch patch from state")
		return
	}
	if s.String == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to fetch patch from state - pointer empty")
		return

	}

	err = r.client.Update(ctx, r.api.ServerURL, *s.String, body)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
		return
	}

	a, err := r.client.Get(ctx, r.api.ServerURL, *s.String)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Create response: %v", a))

	toBeState, err := State(ctx, a, dataResource, r.resourceSchema)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Update: unable to create state, got error: %s", err))
		return

	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, toBeState)...)
}

func (r *ExampleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	dataResource := &data.Resource{}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &dataResource)...)

	if resp.Diagnostics.HasError() {
		return
	}

	s, ok := dataResource.Values["path"]
	if !ok {
		resp.Diagnostics.AddError("Client Error", "Unable to fetch patch from state")
		return
	}
	if s.String == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to fetch patch from state - pointer empty")
		return

	}

	err := r.client.Delete(ctx, r.api.ServerURL, *s.String)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
		return
	}
}

func (r *ExampleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("path"), req, resp)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-scaffolding/internal/provider/data"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ExampleResource{}
var _ resource.ResourceWithImportState = &ExampleResource{}

func NewExampleResourceWithResource(r *api.Resource, a *api.API, n string) func() resource.Resource {
	return func() resource.Resource {
		return &ExampleResource{
			resource: r,
			api:      a,
			name:     n,
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
	client *http.Client
}

func (r *ExampleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.name
}

func (r *ExampleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		// TODO: Add description.
		MarkdownDescription: r.resource.Singular,

		Attributes: r.schemaAttributes(),
	}
}

func checkIfRequired(requiredProps []string, propName string) bool {
	for _, prop := range requiredProps {
		if prop == propName {
			return true
		}
	}
  return false
}

func (r *ExampleResource) schemaAttributes() map[string]schema.Attribute {
	m := make(map[string]schema.Attribute)
	for name, prop := range r.resource.Schema.Properties {
		required := checkIfRequired(r.resource.Schema.Required, name)
		var a schema.Attribute
		switch prop.Type {
		case "number":
			a = schema.NumberAttribute{
				MarkdownDescription: prop.Description,
				Computed:            prop.ReadOnly,
				Required:            required,
				Optional:            !required,
			}
			m[name] = a
		case "string":
			a = schema.StringAttribute{
				MarkdownDescription: prop.Description,
				Computed:            prop.ReadOnly,
				Optional:            !required,
				Required:            required,
			}
			m[name] = a
		case "boolean":
			a = schema.BoolAttribute{
				MarkdownDescription: prop.Description,
				Computed:            prop.ReadOnly,
				Required:            required,
				Optional:            !required,
			}
			m[name] = a
		case "integer":
			a = schema.Int64Attribute{
				MarkdownDescription: prop.Description,
				Computed:            prop.ReadOnly,
				Required:            required,
				Optional:            !required,
			}
			m[name] = a
		}
	}
	m["id"] = schema.StringAttribute{
		MarkdownDescription: "The id of the resource",
		Required:            true,
	}
	return m
}

func (r *ExampleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

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
	dataResource := &data.Resource{}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataResource)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jsonDataMap, err := dataResource.ToJSON()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal JSON, got error: %s", err))
		return
	}

	a, err := Create(ctx, r.resource, r.client, r.api.ServerURL, jsonDataMap)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("resource state: %v", a))

	err = data.ToResource(a, dataResource)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal example, got error: %s", err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("about to save: %v"))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, dataResource)...)
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

	a, err := Read(ctx, r.resource, r.client, r.api.ServerURL, jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	err = data.ToResource(a, dataResource)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal example, got error: %s", err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("about to save: %v"))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, dataResource)...)
}

func (r *ExampleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	dataResource := &data.Resource{}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataResource)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jsonDataMap, err := dataResource.ToJSON()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal JSON, got error: %s", err))
		return
	}

	err = Update(ctx, r.resource, r.client, r.api.ServerURL, jsonDataMap)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
		return
	}

	a, err := Read(ctx, r.resource, r.client, r.api.ServerURL, jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Create response: %v", a))

	err = data.ToResource(a, dataResource)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal example, got error: %s", err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("about to save: %v"))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, dataResource)...)
}

func (r *ExampleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	err = Delete(ctx, r.resource, r.client, r.api.ServerURL, jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
		return
	}
}

func (r *ExampleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

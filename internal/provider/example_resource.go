// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func (r *ExampleResource) schemaAttributes() map[string]schema.Attribute {
	m := make(map[string]schema.Attribute)
	for name, prop := range r.resource.Schema.Properties {
		var a schema.Attribute
		switch prop.Type {
		case "number":
			a = schema.NumberAttribute{
				MarkdownDescription: "",
				Optional:            true,
			}
			m[name] = a
		case "string":
			a = schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			}
			m[name] = a
		case "boolean":
			a = schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
			}
			m[name] = a
		case "integer":
			a = schema.Int64Attribute{
				MarkdownDescription: "",
				Optional:            true,
			}
			m[name] = a
		}
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
	data := &data.Resource{}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal data to JSON", err.Error())
		return
	}

	var jsonDataMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal JSON to map", err.Error())
		return
	}

	err = Create(r.resource, r.client, r.api.ServerURL, jsonDataMap)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	a, err := Read(r.resource, r.client, r.api.ServerURL, jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &a)...)
}

func (r *ExampleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	data := &data.Resource{}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal data to JSON", err.Error())
		return
	}

	var jsonDataMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal JSON to map", err.Error())
		return
	}

	a, err := Read(r.resource, r.client, r.api.ServerURL, jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &a)...)
}

func (r *ExampleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	data := &data.Resource{}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal data to JSON", err.Error())
		return
	}

	var jsonDataMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal JSON to map", err.Error())
		return
	}

	err = Update(r.resource, r.client, r.api.ServerURL, jsonDataMap)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
		return
	}

	a, err := Read(r.resource, r.client, r.api.ServerURL, jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &a)...)
}

func (r *ExampleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	data := &data.Resource{}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal data to JSON", err.Error())
		return
	}

	var jsonDataMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal JSON to map", err.Error())
		return
	}

	err = Delete(r.resource, r.client, r.api.ServerURL, jsonDataMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
		return
	}
}

func (r *ExampleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

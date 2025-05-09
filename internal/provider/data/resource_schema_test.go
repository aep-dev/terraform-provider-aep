package data

import (
	"context"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaAttributes(t *testing.T) {
	tests := map[string]struct {
		schema   openapi.Schema
		resource *api.Resource
		want     *ResourceSchema
	}{
		"simple": {
			schema: openapi.Schema{
				Properties: map[string]openapi.Schema{
					"foo": {
						Type:        "string",
						Description: "foo description",
					},
					"id": {
						Type:        "string",
						Description: "The id of the resource",
					},
				},
			},
			resource: &api.Resource{
				CreateMethod: &api.CreateMethod{
					SupportsUserSettableCreate: true,
				},
				PatternElems: []string{},
			},
			want: &ResourceSchema{
				Attributes: map[string]*ResourceAttribute{
					"foo": {
						TerraformName: "foo",
						JSONName:      "foo",
						Parameter:     false,
						Type:          STRING,
						Attribute: tfschema.StringAttribute{
							MarkdownDescription: "foo description",
							Optional:            true,
						},
						DatasourceAttribute: dsschema.StringAttribute{
							MarkdownDescription: "foo description",
							Computed:            true,
						},
					},
					"id": {
						TerraformName: "id",
						JSONName:      "id",
						Parameter:     false,
						Type:          STRING,
						Attribute: tfschema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The id of the resource",
						},
						DatasourceAttribute: dsschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The id of the resource",
						},
					},
				},
			},
		},
		"required": {
			schema: openapi.Schema{
				Properties: map[string]openapi.Schema{
					"foo": {
						Type:        "string",
						Description: "foo description",
					},
				},
				Required: []string{"foo"},
			},
			resource: &api.Resource{
				CreateMethod: &api.CreateMethod{
					SupportsUserSettableCreate: false,
				},
				PatternElems: []string{},
			},
			want: &ResourceSchema{
				Attributes: map[string]*ResourceAttribute{
					"foo": {
						TerraformName: "foo",
						JSONName:      "foo",
						Parameter:     false,
						Type:          STRING,
						Attribute: tfschema.StringAttribute{
							MarkdownDescription: "foo description",
							Required:            true,
						},
						DatasourceAttribute: dsschema.StringAttribute{
							MarkdownDescription: "foo description",
							Computed:            true,
						},
					},
					"id": {
						TerraformName: "id",
						JSONName:      "id",
						Parameter:     true,
						Type:          STRING,
						Attribute: tfschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The id of the resource.",
						},
						DatasourceAttribute: dsschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The id of the resource.",
						},
					},
				},
			},
		},
		"nested": {
			schema: openapi.Schema{
				Properties: map[string]openapi.Schema{
					"foo": {
						Type: "object",
						Properties: map[string]openapi.Schema{
							"bar": {
								Type:        "string",
								Description: "bar description",
							},
						},
					},
				},
			},
			resource: &api.Resource{
				CreateMethod: &api.CreateMethod{
					SupportsUserSettableCreate: false,
				},
				PatternElems: []string{},
			},
			want: &ResourceSchema{
				Attributes: map[string]*ResourceAttribute{
					"foo": {
						TerraformName: "foo",
						JSONName:      "foo",
						Parameter:     false,
						Type:          OBJECT,
						Attribute: tfschema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]tfschema.Attribute{
								"bar": tfschema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "bar description",
								},
							},
						},
						DatasourceAttribute: dsschema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]dsschema.Attribute{
								"bar": dsschema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "bar description",
								},
							},
						},
						NestedAttributes: map[string]*ResourceAttribute{
							"bar": {
								TerraformName: "bar",
								JSONName:      "bar",
								Parameter:     false,
								Type:          STRING,
								Attribute: tfschema.StringAttribute{
									MarkdownDescription: "bar description",
									Optional:            true,
								},
								DatasourceAttribute: dsschema.StringAttribute{
									MarkdownDescription: "bar description",
									Computed:            true,
								},
							},
						},
					},
					"id": {
						TerraformName: "id",
						JSONName:      "id",
						Parameter:     true,
						Type:          STRING,
						Attribute: tfschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The id of the resource.",
						},
						DatasourceAttribute: dsschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The id of the resource.",
						},
					},
				},
			},
		},
		"array": {
			schema: openapi.Schema{
				Properties: map[string]openapi.Schema{
					"foo": {
						Type: "array",
						Items: &openapi.Schema{
							Type: "string",
						},
					},
				},
			},
			resource: &api.Resource{
				CreateMethod: &api.CreateMethod{
					SupportsUserSettableCreate: false,
				},
				PatternElems: []string{},
			},
			want: &ResourceSchema{
				Attributes: map[string]*ResourceAttribute{
					"foo": {
						TerraformName: "foo",
						JSONName:      "foo",
						Parameter:     false,
						Type:          ARRAY,
						ListItemType:  STRING,
						Attribute: tfschema.ListAttribute{
							MarkdownDescription: "",
							Optional:            true,
							ElementType:         types.StringType,
						},
						DatasourceAttribute: dsschema.ListAttribute{
							MarkdownDescription: "",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
					"id": {
						TerraformName: "id",
						JSONName:      "id",
						Parameter:     true,
						Type:          STRING,
						Attribute: tfschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The id of the resource.",
						},
						DatasourceAttribute: dsschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The id of the resource.",
						},
					},
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.resource.Schema = &tt.schema
			got, err := NewResourceSchema(context.TODO(), tt.resource, &openapi.OpenAPI{})
			if err != nil {
				t.Errorf("SchemaAttributes() error = %v", err)
				return
			}
			if d := cmp.Diff(got.Attributes, tt.want.Attributes); d != "" {
				t.Errorf("SchemaAttributes() diff: %s", d)
			}
		})
	}
}

func TestListType(t *testing.T) {
	tests := map[string]struct {
		prop    openapi.Schema
		want    attr.Type
		wantErr bool
	}{
		"string": {
			prop: openapi.Schema{
				Items: &openapi.Schema{
					Type: "string",
				},
			},
			want: types.StringType,
		},
		"unknown": {
			prop: openapi.Schema{
				Items: &openapi.Schema{
					Type: "unknown",
				},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := listType(&tt.prop)
			if (err != nil) != tt.wantErr {
				t.Errorf("listType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("listType() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

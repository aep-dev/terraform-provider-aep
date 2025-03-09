package data

import (
	"context"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaAttributes(t *testing.T) {
	tests := map[string]struct {
		schema openapi.Schema
		want   *ResourceSchema
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
			want: &ResourceSchema{
				Attributes: map[string]*ResourceAttribute{
					"foo": {
						TerraformName: "foo",
						JSONName:      "foo",
						Parameter:     false,
						Attribute: tfschema.StringAttribute{
							MarkdownDescription: "foo description",
							Optional:            true,
						},
					},
					"id": {
						TerraformName: "id",
						JSONName:      "id",
						Parameter:     false,
						Attribute: tfschema.StringAttribute{
							Optional:            true,
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
			want: &ResourceSchema{
				Attributes: map[string]*ResourceAttribute{
					"foo": {
						TerraformName: "foo",
						JSONName:      "foo",
						Parameter:     false,
						Attribute: tfschema.StringAttribute{
							MarkdownDescription: "foo description",
							Required:            true,
						},
					},
					"id": {
						TerraformName: "id",
						JSONName:      "id",
						Parameter:     true,
						Attribute: tfschema.StringAttribute{
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
						Required: []string{"bar"},
					},
				},
			},
			want: &ResourceSchema{
				Attributes: map[string]*ResourceAttribute{
					"foo": {
						TerraformName: "foo",
						JSONName:      "foo",
						Parameter:     false,
						Attribute: tfschema.SingleNestedAttribute{
							Attributes: map[string]tfschema.Attribute{
								"bar": tfschema.StringAttribute{
									MarkdownDescription: "bar description",
									Required:            true,
									Optional:            false,
								},
							},
							MarkdownDescription: "",
							Optional:            true,
						},
						NestedAttributes: map[string]*ResourceAttribute{
							"bar": {
								TerraformName: "bar",
								JSONName:      "bar",
								Attribute: tfschema.StringAttribute{
									Required:            true,
									MarkdownDescription: "bar description",
								},
							},
						},
					},
					"id": {
						TerraformName: "id",
						JSONName:      "id",
						Parameter:     true,
						Attribute: tfschema.StringAttribute{
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
			want: &ResourceSchema{
				Attributes: map[string]*ResourceAttribute{
					"foo": {
						TerraformName: "foo",
						JSONName:      "foo",
						Parameter:     false,
						Attribute: tfschema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "",
							Optional:            true,
						},
					},
					"id": {
						TerraformName: "id",
						JSONName:      "id",
						Parameter:     true,
						Attribute: tfschema.StringAttribute{
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
			got, err := NewResourceSchema(context.TODO(), &api.Resource{Schema: &tt.schema, PatternElems: make([]string, 0)}, &openapi.OpenAPI{})
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

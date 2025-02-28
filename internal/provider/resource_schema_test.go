package provider

import (
	"context"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaAttributes(t *testing.T) {
	tests := map[string]struct {
		schema openapi.Schema
		want   map[string]tfschema.Attribute
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
			want: map[string]tfschema.Attribute{
				"foo": tfschema.StringAttribute{
					MarkdownDescription: "foo description",
					Optional:            true,
				},
				"id": tfschema.StringAttribute{
					Computed: true,
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
			want: map[string]tfschema.Attribute{
				"foo": tfschema.StringAttribute{
					MarkdownDescription: "foo description",
					Required:            true,
				},
				"id": tfschema.StringAttribute{
					Computed: true,
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
			want: map[string]tfschema.Attribute{
				"foo": tfschema.SingleNestedAttribute{
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
				"id": tfschema.StringAttribute{
					Computed: true,
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
			want: map[string]tfschema.Attribute{
				"foo": tfschema.ListAttribute{
					ElementType:         types.StringType,
					MarkdownDescription: "",
					Optional:            true,
				},
				"id": tfschema.StringAttribute{
					Computed: true,
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := SchemaAttributes(context.TODO(), tt.schema, &openapi.OpenAPI{})
			if err != nil {
				t.Errorf("SchemaAttributes() error = %v", err)
				return
			}
			if d := cmp.Diff(got, tt.want); d != "" {
				t.Errorf("SchemaAttributes() diff: %s", d)
			}
		})
	}
}

func TestSchemaAttribute(t *testing.T) {
	tests := map[string]struct {
		prop     openapi.Schema
		name     string
		required []string
		want     tfschema.Attribute
		wantErr  bool
	}{
		"string": {
			prop: openapi.Schema{
				Type: "string",
			},
			name: "foo",
			want: tfschema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
		},
		"required": {
			prop: openapi.Schema{
				Type: "string",
			},
			name:     "foo",
			required: []string{"foo"},
			want: tfschema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
		},
		"unknown": {
			prop: openapi.Schema{
				Type: "unknown",
			},
			name:    "foo",
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := schemaAttribute(context.TODO(), tt.prop, tt.name, tt.required, &openapi.OpenAPI{})
			if (err != nil) != tt.wantErr {
				t.Errorf("schemaAttribute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("schemaAttribute() mismatch (-got +want):\n%s", diff)
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
			got, err := listType(tt.prop)
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

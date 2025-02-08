package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SchemaAttributes(schema openapi.Schema) (map[string]tfschema.Attribute, error) {
	m := make(map[string]tfschema.Attribute)
	for name, prop := range schema.Properties {
		a, err := schemaAttribute(prop, name, schema.Required)
		if err != nil {
			return nil, err
		}
		m[ToSnakeCase(name)] = a
	}
	return m, nil
}

// Attributes coming from parents.
func ParameterAttributes(r *api.Resource, o *openapi.OpenAPI) (map[string]tfschema.Attribute, error) {
	if len(r.PatternElems) < 1 {
		return nil, fmt.Errorf("must have at least one parent pattern")
	}
	// Fetch pattern + remove the last part.
	p := fmt.Sprintf("/%s", r.Schema.XAEPResource.Patterns[0])
	parts := strings.Split(p, "/")
	parts = parts[:len(parts)-1]
	p = strings.Join(parts, "/")

	// Fetch create method.
	para, ok := o.Paths[p]
	if !ok {
		return nil, fmt.Errorf("could not find %s in paths", p)
	}
	m := make(map[string]tfschema.Attribute)
	if para != nil && para.Post != nil {
		for _, prop := range para.Post.Parameters {
			m[prop.Name] = tfschema.StringAttribute{
				MarkdownDescription: prop.Description,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			}
		}
	}
	return m, nil
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func schemaAttribute(prop openapi.Schema, name string, requiredProps []string) (tfschema.Attribute, error) {
	required := checkIfRequired(requiredProps, name)
	switch prop.Type {
	case "number":
		return tfschema.NumberAttribute{
			MarkdownDescription: prop.Description,
			Computed:            prop.ReadOnly,
			Required:            required,
			Optional:            !required,
		}, nil
	case "string":
		return tfschema.StringAttribute{
			MarkdownDescription: prop.Description,
			Computed:            prop.ReadOnly,
			Optional:            !required,
			Required:            required,
		}, nil
	case "boolean":
		return tfschema.BoolAttribute{
			MarkdownDescription: prop.Description,
			Computed:            prop.ReadOnly,
			Required:            required,
			Optional:            !required,
		}, nil
	case "integer":
		return tfschema.Int64Attribute{
			MarkdownDescription: prop.Description,
			Computed:            prop.ReadOnly,
			Required:            required,
			Optional:            !required,
		}, nil
	case "object":
		no, err := SchemaAttributes(prop)
		if err != nil {
			return nil, err
		}
		return tfschema.SingleNestedAttribute{
			Attributes:          no,
			MarkdownDescription: prop.Description,
			Computed:            prop.ReadOnly,
			Required:            required,
			Optional:            !required,
		}, nil
	case "array":
		if prop.Items.Type == "object" {
			no, err := SchemaAttributes(*prop.Items)
			if err != nil {
				return nil, err
			}
			return tfschema.ListNestedAttribute{
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: no,
				},
				MarkdownDescription: prop.Description,
				Computed:            prop.ReadOnly,
				Required:            required,
				Optional:            !required,
			}, nil
		} else {
			lt, err := listType(prop)
			if err != nil {
				return nil, err
			}

			return tfschema.ListAttribute{
				ElementType:         lt,
				MarkdownDescription: prop.Description,
				Computed:            prop.ReadOnly,
				Required:            required,
				Optional:            !required,
			}, nil
		}
	default:
		return nil, fmt.Errorf("cannot find type for %s", prop.Type)
	}
}

func listType(prop openapi.Schema) (attr.Type, error) {
	switch prop.Items.Type {
	case "number":
		return types.NumberType, nil
	case "string":
		return types.StringType, nil
	case "boolean":
		return types.BoolType, nil
	case "integer":
		return types.Int64Type, nil
	default:
		return nil, fmt.Errorf("cannot find type for %s", prop.Items.Type)
	}

}

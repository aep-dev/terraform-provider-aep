package data

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceSchema struct {
	Resource *api.Resource

	// Maps Terraform Name -> ResourceAttribute
	Attributes map[string]*ResourceAttribute
}

func FindAttributeByJSONName(name string, attributes map[string]*ResourceAttribute) *ResourceAttribute {
	for _, attr := range attributes {
		if attr.JSONName == name {
			return attr
		}
	}
	return nil
}

type TypeEnum string

const (
	STRING  TypeEnum = "string"
	NUMBER  TypeEnum = "number"
	BOOLEAN TypeEnum = "boolean"
	INTEGER TypeEnum = "integer"
	OBJECT  TypeEnum = "object"
	ARRAY   TypeEnum = "array"
)

type ResourceAttribute struct {
	// The name of the attribute from Terraform's perspective.
	TerraformName string
	// The name of the attribute that should be sent across the wire.
	JSONName string
	// If true, this attribute is not sent across the wire.
	Parameter bool
	// The type of this resource attribute.
	Type TypeEnum
	// Only set for ARRAY types.
	ListItemType TypeEnum
	// The attribute information for the resource.
	Attribute tfschema.Attribute
	// The attribute information for the data source.
	DatasourceAttribute dsschema.Attribute
	// The nested attributes if the type is object.
	// This is most important to gather ResourceAttribute information for other types.
	NestedAttributes map[string]*ResourceAttribute
}

func (r *ResourceSchema) FullSchema() map[string]tfschema.Attribute {
	schema := make(map[string]tfschema.Attribute)
	for _, attr := range r.Attributes {
		schema[attr.TerraformName] = attr.Attribute
	}
	return schema
}

func (r *ResourceSchema) FullCollectionDataSourceSchema(ctx context.Context) map[string]dsschema.Attribute {
	schema := make(map[string]dsschema.Attribute)
	for _, attr := range r.Attributes {
		if attr.Parameter {
			schema[attr.TerraformName] = attr.DatasourceAttribute
		}
	}

	fullResource := make(map[string]dsschema.Attribute)
	for _, attr := range r.Attributes {
		fullResource[attr.TerraformName] = attr.DatasourceAttribute
	}

	// Results are stored here.
	schema["results"] = dsschema.ListNestedAttribute{
		MarkdownDescription: fmt.Sprintf("Contains all of the %s resources", r.Resource.Singular),
		Computed:            true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: fullResource,
		},
	}
	return schema
}

func (r *ResourceSchema) Parameters() map[string]tfschema.Attribute {
	parameters := make(map[string]tfschema.Attribute)
	for _, attr := range r.Attributes {
		if attr.Parameter {
			parameters[attr.TerraformName] = attr.Attribute
		}
	}
	return parameters
}

func (r *ResourceSchema) SchemaAttributes() map[string]tfschema.Attribute {
	schemaAttributes := make(map[string]tfschema.Attribute)
	for _, attr := range r.Attributes {
		if !attr.Parameter {
			schemaAttributes[attr.TerraformName] = attr.Attribute
		}
	}
	return schemaAttributes
}

func NewResourceSchema(ctx context.Context, r *api.Resource, o *openapi.OpenAPI) (*ResourceSchema, error) {
	schema := &ResourceSchema{
		Resource:   r,
		Attributes: make(map[string]*ResourceAttribute),
	}

	// Add all normal schema attributes.
	a := schemaAttributes(ctx, r.Schema, o)
	for _, attr := range a {
		schema.Attributes[attr.TerraformName] = attr
	}

	// Add all parameters.
	if len(r.PatternElems) > 0 {
		for _, elem := range r.PatternElems[:len(r.PatternElems)-1] {
			if strings.HasPrefix(elem, "{") && strings.HasSuffix(elem, "}") {
				paramName := strings.Replace(elem[1:len(elem)-1], "-", "_", -1)
				schema.Attributes[paramName] = &ResourceAttribute{
					TerraformName: paramName,
					JSONName:      paramName,
					Parameter:     true,
					Type:          STRING,
					Attribute: tfschema.StringAttribute{
						MarkdownDescription: paramName,
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					DatasourceAttribute: dsschema.StringAttribute{
						MarkdownDescription: paramName,
						Required:            true,
					},
				}
			}
		}
	}

	if _, ok := schema.Attributes["id"]; !ok {
		if r.CreateMethod != nil && r.CreateMethod.SupportsUserSettableCreate {
			schema.Attributes["id"] = &ResourceAttribute{
				TerraformName: "id",
				JSONName:      "id",
				Parameter:     false,
				Type:          STRING,
				Attribute: tfschema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "The id of the resource.",
				},
				DatasourceAttribute: dsschema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "The id of the resource.",
				},
			}
		} else {
			schema.Attributes["id"] = &ResourceAttribute{
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
			}
		}
	}
	return schema, nil
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func schemaAttributes(ctx context.Context, s *openapi.Schema, o *openapi.OpenAPI) map[string]*ResourceAttribute {
	m := make(map[string]*ResourceAttribute)
	// Add all normal properties.
	for name, prop := range s.Properties {
		a, err := schemaAttribute(ctx, &prop, name, s.Required, o)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("could not create type for %s %v", name, prop))
		} else if a != nil {
			m[name] = a
		}
	}
	return m
}

func schemaAttribute(ctx context.Context, prop *openapi.Schema, name string, requiredProps []string, o *openapi.OpenAPI) (*ResourceAttribute, error) {
	m := &ResourceAttribute{
		TerraformName: strings.Replace(ToSnakeCase(name), "@", "", -1),
		JSONName:      name,
		Parameter:     false,
	}
	required := checkIfRequired(requiredProps, name)

	if name == "etag" {
		return nil, nil
	}

	// GoogleProtobufValue is a type based on its name.
	// It's just a string that stands in for arbitrary JSON.
	if prop.Ref == "#/components/schemas/GoogleProtobufValue" {
		m.Type = STRING
		m.Attribute = tfschema.StringAttribute{
			MarkdownDescription: prop.Description,
			Computed:            prop.ReadOnly,
			Required:            required,
			Optional:            !required,
		}
		m.DatasourceAttribute = dsschema.StringAttribute{
			MarkdownDescription: prop.Description,
			Computed:            true,
		}
		return m, nil
	}

	if prop.Ref != "" {
		s, err := o.DereferenceSchema(*prop)
		if err != nil {
			return nil, err
		}
		if s == nil {
			return nil, fmt.Errorf("ref not found for %s", prop.Ref)
		}
		return schemaAttribute(ctx, s, name, requiredProps, o)
	}

	computed := prop.ReadOnly

	// The path field should always be treated as computed.
	// If the ID is settable, the ID field will be used.
	// If ID is not settable, path should be computed regardless.
	if name == "path" {
		computed = true
	}

	switch prop.Type {
	case "number":
		m.Type = NUMBER
		m.Attribute = tfschema.NumberAttribute{
			MarkdownDescription: prop.Description,
			Computed:            computed,
			Required:            required,
			Optional:            !required,
		}
		m.DatasourceAttribute = dsschema.NumberAttribute{
			MarkdownDescription: prop.Description,
			Computed:            true,
		}
	case "string":
		m.Type = STRING
		m.Attribute = tfschema.StringAttribute{
			MarkdownDescription: prop.Description,
			Computed:            computed,
			Optional:            !required,
			Required:            required,
		}
		m.DatasourceAttribute = dsschema.StringAttribute{
			MarkdownDescription: prop.Description,
			Computed:            true,
		}
	case "boolean":
		m.Type = BOOLEAN
		m.Attribute = tfschema.BoolAttribute{
			MarkdownDescription: prop.Description,
			Computed:            computed,
			Required:            required,
			Optional:            !required,
		}
		m.DatasourceAttribute = dsschema.BoolAttribute{
			MarkdownDescription: prop.Description,
			Computed:            true,
		}
	case "integer":
		m.Type = INTEGER
		m.Attribute = tfschema.Int64Attribute{
			MarkdownDescription: prop.Description,
			Computed:            computed,
			Required:            required,
			Optional:            !required,
		}
		m.DatasourceAttribute = dsschema.Int64Attribute{
			MarkdownDescription: prop.Description,
			Computed:            true,
		}
	case "object":
		m.Type = OBJECT
		no := schemaAttributes(ctx, prop, o)
		m.Attribute = tfschema.SingleNestedAttribute{
			Attributes:          convertToMap(no),
			MarkdownDescription: prop.Description,
			Computed:            computed,
			Required:            required,
			Optional:            !required,
		}
		m.DatasourceAttribute = dsschema.SingleNestedAttribute{
			Attributes:          convertToMapForDatasource(no),
			MarkdownDescription: prop.Description,
			Computed:            true,
		}
		m.NestedAttributes = no
	case "array":
		m.Type = ARRAY
		if prop.Items.Type == "object" {
			m.ListItemType = OBJECT
			no := schemaAttributes(ctx, prop.Items, o)
			m.NestedAttributes = no
			m.Attribute = tfschema.ListNestedAttribute{
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: convertToMap(no),
				},
				MarkdownDescription: prop.Description,
				Computed:            computed,
				Required:            required,
				Optional:            !required,
			}
			m.DatasourceAttribute = dsschema.ListNestedAttribute{
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: convertToMapForDatasource(no),
				},
				MarkdownDescription: prop.Description,
				Computed:            true,
			}
		} else {
			t, err := listType(prop)
			if err != nil {
				return nil, err
			}
			t2, err := listEnumType(prop)
			if err != nil {
				return nil, err
			}
			m.ListItemType = t2
			m.Attribute = tfschema.ListAttribute{
				ElementType: t,
				Computed:    computed,
				Required:    required,
				Optional:    !required,
			}
			m.DatasourceAttribute = dsschema.ListAttribute{
				ElementType: t,
				Computed:    true,
			}
		}
	default:
		return nil, fmt.Errorf("cannot find type for %v", prop)
	}

	return m, nil
}

func listType(prop *openapi.Schema) (attr.Type, error) {
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

func listEnumType(prop *openapi.Schema) (TypeEnum, error) {
	switch prop.Items.Type {
	case "string":
		return STRING, nil
	case "number":
		return NUMBER, nil
	case "boolean":
		return BOOLEAN, nil
	case "integer":
		return INTEGER, nil
	default:
		return "", fmt.Errorf("cannot find type for %s", prop.Items.Type)
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

func convertToMap(l map[string]*ResourceAttribute) map[string]tfschema.Attribute {
	attributeMap := make(map[string]tfschema.Attribute)
	for _, attribute := range l {
		attributeMap[attribute.TerraformName] = attribute.Attribute
	}
	return attributeMap
}

func convertToMapForDatasource(l map[string]*ResourceAttribute) map[string]dsschema.Attribute {
	attributeMap := make(map[string]dsschema.Attribute)
	for _, attribute := range l {
		attributeMap[attribute.TerraformName] = attribute.DatasourceAttribute
	}
	return attributeMap
}

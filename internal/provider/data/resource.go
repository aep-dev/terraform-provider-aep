// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Inspiration: https://github.com/hashicorp/terraform-plugin-framework/issues/1035#issuecomment-2396927170

package data

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tftypes.ValueConverter = &Resource{}
var _ tftypes.ValueCreator = &Resource{}

// Resource acts as an intermediary data structure between Terraform state and JSON requests / responses.
// It has methods to convert to/from Terraform state and to/from JSON (map[string]interface{})
// These conversions happen based on the state in schema.
type Resource struct {
	Values map[string]Value `json:"values"`
	Schema *ResourceSchema

	ObjectType tftypes.Object
}

func NewResource(schema *ResourceSchema) *Resource {
	return &Resource{
		Schema: schema,
	}
}

// GetId returns the ID of the resource.
//
// It assumes the ID value exists and is a string type.
func (r Resource) GetId() string {
	return *r.Values["path"].String
}

// WithType adds type information into a Resource as this is not stored as part
// of our external API.
//
// You must call this function to set the type information before using
// ToTerraform5Value(). The type information can usually be retrieved from the
// Terraform SDK, so this information should be readily available it just needs
// to be added after the Resource has been created.
func (r *Resource) WithType(objectType tftypes.Object) *Resource {
	r.ObjectType = objectType
	return r
}

// ToTerraform5Value ensures that Resource implements the tftypes.ValueCreator
// interface, and so can be converted into Terraform types easily.
func (r Resource) ToTerraform5Value() (interface{}, error) {
	return objectToTerraform5Value(&r.Values, r.ObjectType)
}

// FromTerraform5Value ensures that Resource implements the
// tftypes.ValueConverter interface, and so can be converted from Terraform
// types easily.
func (r *Resource) FromTerraform5Value(value tftypes.Value) error {
	// It has to be an object we are converting from.
	if !value.Type().Is(tftypes.Object{}) {
		return errors.New("can only convert between object types")
	}

	values, err := FromTerraform5Value(value)
	if err != nil {
		return err
	}

	// We know these kinds of conversions are safe now, as we checked the type
	// at the beginning.
	r.Values = *values.Object
	v, ok := value.Type().(tftypes.Object)
	if !ok {
		return fmt.Errorf("value %v is not a object", value.Type())
	}
	r.ObjectType = v
	return nil
}

// The structure of the data is:
//
// {"description": {"string": "my-description"}, "path": {"string": "my-path"}}
//
// This function removes the object type keys.
func (r *Resource) ToJSON() (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	if r.Schema == nil {
		return nil, fmt.Errorf("must set schema on resource")
	}
	// r.Values contains Terraform keys.
	for k, v := range r.Values {
		schemaValue, ok := r.Schema.Attributes[k]
		if !ok {
			return nil, fmt.Errorf("could not find %s in Schema", k)
		}
		convertedValue, err := ConvertValue(v, schemaValue)
		if err != nil {
			return nil, err
		}
		jsonMap[schemaValue.JSONName] = convertedValue
	}
	return jsonMap, nil
}

func ConvertValue(v Value, a *ResourceAttribute) (interface{}, error) {
	if v.Boolean != nil {
		return *v.Boolean, nil
	}
	if v.Number != nil {
		return v.Number.String(), nil
	}
	if v.String != nil {
		return *v.String, nil
	}
	if v.List != nil {
		list := make([]interface{}, len(*v.List))
		for i, item := range *v.List {
			v2, err := ConvertValue(item, a)
			if err != nil {
				return nil, err
			}
			list[i] = v2
		}
		return list, nil
	}
	if v.Map != nil {
		mapJSON := make(map[string]interface{})
		for key, value := range *v.Map {
			convertedValue, err := ConvertValue(value, a)
			if err != nil {
				return nil, err
			}
			mapJSON[key] = convertedValue
		}
		return mapJSON, nil
	}
	if v.Object != nil {
		objectJSON := make(map[string]interface{})
		for key, value := range *v.Object {
			schemaObj, ok := a.NestedAttributes[key]
			if !ok {
				return nil, fmt.Errorf("nested object name %s not found", key)
			}
			convertedValue, err := ConvertValue(value, schemaObj)
			if err != nil {
				return nil, err
			}
			objectJSON[key] = convertedValue
		}
		return objectJSON, nil
	}
	if v.Set != nil {
		set := make([]interface{}, len(*v.Set))
		for i, item := range *v.Set {
			convertedValue, err := ConvertValue(item, a)
			if err != nil {
				return nil, err
			}
			set[i] = convertedValue
		}
		return set, nil
	}
	return nil, fmt.Errorf("unknown type in ToJSON %v", v)
}

// Convert the JSON response to a Resource.
// Only attributes in the plan are used for this conversion.
// This function will not attempt to suppress differences between the response + plan, but those differences will result in a provider error.
func FromJSON(m map[string]interface{}, r *Resource, plan *Resource) error {
	for k, val := range plan.Values {
		// Find matching value in the plan.
		attr, ok := r.Schema.Attributes[k]
		if !ok {
			return fmt.Errorf("no matching resource attribute found for key %s", k)
		}
		v, ok := m[attr.JSONName]
		if !ok {
			return fmt.Errorf("response does not contain key %s", attr.JSONName)
		}
		convertedValue, err := ConvertTypeToValue(v, attr, val)
		if err != nil {
			return err
		}
		r.Values[k] = convertedValue
	}
	return nil
}

func ConvertTypeToValue(v interface{}, r *ResourceAttribute, planValue Value) (Value, error) {
	switch r.Type {
	case STRING:
		str, ok := v.(string)
		if !ok {
			return Value{}, fmt.Errorf("expected string, got %T", v)
		}
		return Value{String: &str}, nil
	case BOOLEAN:
		b, ok := v.(bool)
		if !ok {
			return Value{}, fmt.Errorf("expected boolean, got %T", v)
		}
		return Value{Boolean: &b}, nil
	case NUMBER:
		num, ok := v.(float64)
		if !ok {
			return Value{}, fmt.Errorf("expected number, got %T", v)
		}
		return Value{Number: big.NewFloat(num)}, nil
	case INTEGER:
		intNum, ok := v.(int)
		if !ok {
			return Value{}, fmt.Errorf("expected integer, got %T", v)
		}
		return Value{Number: big.NewFloat(float64(intNum))}, nil
	case OBJECT:
		objectJSON := make(map[string]Value)
		mapValue, ok := v.(map[string]interface{})
		if !ok {
			return Value{}, fmt.Errorf("expected map, got %T", v)
		}

		for key, value := range mapValue {
			schemaObj := FindAttributeByJSONName(key, r.NestedAttributes)
			if schemaObj == nil {
				return Value{}, fmt.Errorf("nested object name %s not found", key)
			}
			val := Value{}
			if planValue.Object != nil {
				if _, ok := (*planValue.Object)[schemaObj.TerraformName]; ok {
					val = (*planValue.Object)[schemaObj.TerraformName]
				}
			}
			convertedValue, err := ConvertTypeToValue(value, schemaObj, val)
			if err != nil {
				return Value{}, err
			}
			objectJSON[schemaObj.TerraformName] = convertedValue
		}
		return Value{Object: &objectJSON}, nil
	case ARRAY:
		arrayValue, ok := v.([]interface{})
		if !ok {
			return Value{}, fmt.Errorf("expected array, got %T", v)
		}
		list := make([]Value, len(arrayValue))
		for i, item := range arrayValue {
			// TODO: Make the plan fetcher work with arrays.
			convertedValue, err := ConvertTypeToValue(item, &ResourceAttribute{NestedAttributes: r.NestedAttributes, Type: r.ListItemType}, Value{})
			if err != nil {
				return Value{}, err
			}
			list[i] = convertedValue
		}
		return Value{List: &list}, nil
	default:
		return Value{}, fmt.Errorf("cannot find type for %v", r)
	}
}

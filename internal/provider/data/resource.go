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

// Resource is the data structure that is actually written into our data stores.
//
// It currently only publicly contains the Values mapping of attribute names to
// actual values. It is designed as a bridge between the Terraform SDK
// representation of a value and a generic JSON representation that can be
// read/written externally. In theory, any terraform object can be represented
// as a Resource. In practice, there will probably be edge cases and types that
// have been missed.
//
// If we could write tftypes.Value into a human friendly format, and read back
// any changes from that then we wouldn't need this bridge. But, we can't do
// that using the current SDK so we handle it ourselves here.
//
// You must call the WithType function manually to attach the object type before
// attempting to convert a Resource into a Terraform SDK value.
//
// The types are attached automatically when converting from a Terraform SDK
// object.
type Resource struct {
	Values map[string]Value `json:"values"`

	objectType tftypes.Object
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
	r.objectType = objectType
	return r
}

// ToTerraform5Value ensures that Resource implements the tftypes.ValueCreator
// interface, and so can be converted into Terraform types easily.
func (r Resource) ToTerraform5Value() (interface{}, error) {
	return objectToTerraform5Value(&r.Values, r.objectType)
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
	r.objectType = v
	return nil
}

// The structure of the data is:
//
// {"description": {"string": "my-description"}, "path": {"string": "my-path"}}
//
// This function removes the object type keys.
func (r *Resource) ToJSON() (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	for k, v := range r.Values {
		convertedValue, err := ConvertValue(v)
		if err != nil {
			return nil, err
		}
		jsonMap[k] = convertedValue
	}
	return jsonMap, nil
}

func ConvertValue(v Value) (interface{}, error) {
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
			v2, err := ConvertValue(item)
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
			convertedValue, err := ConvertValue(value)
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
			convertedValue, err := ConvertValue(value)
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
			convertedValue, err := ConvertValue(item)
			if err != nil {
				return nil, err
			}
			set[i] = convertedValue
		}
		return set, nil
	}
	return nil, fmt.Errorf("unknown type %v", v)
}

func ToResource(m map[string]interface{}, r *Resource) error {
	for k, v := range m {
		r.Values[k] = ConvertTypeToValue(v)
	}
	return nil
}

func ConvertTypeToValue(v interface{}) Value {
	switch v := v.(type) {
	case string:
		return Value{String: &v}
	case bool:
		return Value{Boolean: &v}
	case int:
		return Value{Number: big.NewFloat(float64(v))}
	case float64:
		return Value{Number: big.NewFloat(v)}
	case []interface{}:
		list := make([]Value, len(v))
		for i, item := range v {
			list[i] = ConvertTypeToValue(item)
		}
		return Value{List: &list}
	case map[string]interface{}:
		mapJSON := make(map[string]Value)
		for key, value := range v {
			mapJSON[key] = ConvertTypeToValue(value)
		}
		return Value{Object: &mapJSON}
	default:
		return Value{}
	}
}

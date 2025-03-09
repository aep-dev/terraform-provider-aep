// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package data

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestResource_symmetry(t *testing.T) {
	testCases := []struct {
		TestCase string
		Resource Resource
	}{
		{
			TestCase: "basic",
			Resource: Resource{
				objectType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number": tftypes.Number,
					},
				},
				Values: map[string]Value{
					"number": {Number: big.NewFloat(0)},
				},
			},
		},
		{
			TestCase: "missing_object",
			Resource: Resource{
				objectType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"number": tftypes.Number,
							},
						},
					},
				},
				Values: map[string]Value{},
			},
		},
		{
			TestCase: "missing_object_attribute",
			Resource: Resource{
				objectType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"number": tftypes.Number,
							},
						},
					},
				},
				Values: map[string]Value{
					"object": {
						Object: &map[string]Value{},
					},
				},
			},
		},
		{
			TestCase: "missing_list",
			Resource: Resource{
				objectType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list": tftypes.List{
							ElementType: tftypes.Number,
						},
					},
				},
				Values: map[string]Value{},
			},
		},
		{
			TestCase: "empty_list",
			Resource: Resource{
				objectType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list": tftypes.List{
							ElementType: tftypes.Number,
						},
					},
				},
				Values: map[string]Value{
					"list": {
						List: &[]Value{},
					},
				},
			},
		},
		{
			TestCase: "missing_map",
			Resource: Resource{
				objectType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map": tftypes.Map{
							ElementType: tftypes.Number,
						},
					},
				},
				Values: map[string]Value{},
			},
		},
		{
			TestCase: "missing_set",
			Resource: Resource{
				objectType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set": tftypes.Set{
							ElementType: tftypes.Number,
						},
					},
				},
				Values: map[string]Value{},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.TestCase, func(t *testing.T) {
			checkSymmetry(t, testCase.Resource)
		})
	}
}

func toJson(t *testing.T, obj Resource) string {
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("found unexpected error when marshalling json: %v", err)
	}
	return string(data)
}

func checkResourceEqual(t *testing.T, expected, actual Resource) {
	expectedString := toJson(t, expected)
	actualString := toJson(t, actual)
	if expectedString != actualString {
		t.Fatalf("expected did not match actual\nexpected:\n%s\nactual:\n%s", expectedString, actualString)
	}
}

func checkSymmetry(t *testing.T, resource Resource) {
	raw, err := resource.ToTerraform5Value()
	if err != nil {
		t.Fatalf("found unexpected error in ToTerraform5Value(): %v", err)
	}

	value := tftypes.NewValue(resource.objectType, raw)
	actual := Resource{}
	err = actual.FromTerraform5Value(value)
	if err != nil {
		t.Fatalf("found unexpected error in FromTerraform5Value(): %v", err)
	}

	checkResourceEqual(t, resource, actual)
}

func TestToJSON(t *testing.T) {
	testCases := []struct {
		name     string
		resource Resource
		expected map[string]interface{}
	}{
		{
			name: "simple",
			resource: Resource{
				Values: map[string]Value{
					"foo": {String: String("bar")},
				},
				Schema: &ResourceSchema{
					Attributes: map[string]*ResourceAttribute{
						"foo": {
							TerraformName: "foo",
							JSONName:      "foo@",
							Type:          "string",
						},
					},
				},
			},
			expected: map[string]interface{}{
				"foo@": "bar",
			},
		},
		{
			name: "complex",
			resource: Resource{
				Values: map[string]Value{
					"foo": {String: String("bar")},
					"baz": {Number: BigFloat(big.NewFloat(123))},
				},
				Schema: &ResourceSchema{
					Attributes: map[string]*ResourceAttribute{
						"foo": {
							TerraformName: "foo",
							JSONName:      "foo@",
							Type:          "string",
						},
						"baz": {
							TerraformName: "baz",
							JSONName:      "qux",
							Type:          "number",
						},
					},
				},
			},
			expected: map[string]interface{}{
				"foo@": "bar",
				"qux":  "123",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := testCase.resource.ToJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(testCase.expected, actual) {
				t.Fatalf("expected did not match actual\nexpected:\n%v\nactual:\n%v", testCase.expected, actual)
			}
		})
	}
}

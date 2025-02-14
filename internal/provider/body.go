package provider

import (
	"fmt"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/hashicorp/terraform-provider-scaffolding/internal/provider/data"
)

// Returns the proper formatted body for Create / Update requests.
func Body(d *data.Resource, r *api.Resource) (map[string]interface{}, error) {
	jsonDataMap, err := d.ToJSON()
	if err != nil {
		return nil, err
	}

	attributes, err := SchemaAttributes(*r.Schema)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for key, value := range jsonDataMap {
		if _, ok := attributes[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

// Returns a map that can be used to substitute parent values into a URI.
func Parameters(d *data.Resource, r *api.Resource, o *openapi.OpenAPI) (map[string]string, error) {
	jsonData, err := d.ToJSON()
	if err != nil {
		return nil, err
	}

	parameterAttributes, err := ParameterAttributes(r, o)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for key, value := range jsonData {
		if _, ok := parameterAttributes[key]; ok {
			strValue, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("value %v for key %s is not a string", value, key)
			}
			result[key] = strValue
		}
	}
	return result, nil
}

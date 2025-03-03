package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-scaffolding/internal/provider/data"
)

// Returns the proper formatted body for Create / Update requests.
func Body(ctx context.Context, d *data.Resource, r *ResourceSchema) (map[string]interface{}, error) {
	jsonDataMap, err := d.ToJSON()
	if err != nil {
		return nil, err
	}

	attributes := r.SchemaAttributes()

	result := make(map[string]interface{})
	for key, value := range jsonDataMap {
		if _, ok := attributes[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

// Returns a map that can be used to substitute parent values into a URI.
func Parameters(ctx context.Context, d *data.Resource, r *ResourceSchema) (map[string]string, error) {
	jsonData, err := d.ToJSON()
	if err != nil {
		return nil, err
	}

	parameterAttributes := r.Parameters()

	tflog.Info(ctx, fmt.Sprintf("plan data json: %q", jsonData))
	result := make(map[string]string)
	for key, value := range jsonData {
		tflog.Info(ctx, fmt.Sprintf("key %s", key))
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

// Create state from the API response and plan.
func State(ctx context.Context, resp map[string]interface{}, plan *data.Resource, r *ResourceSchema) (*data.Resource, error) {
	s := r.SchemaAttributes()

	p := r.Parameters()

	result := make(map[string]interface{})
	for k := range s {
		v, ok := resp[k]
		if ok {
			result[k] = v
		}
	}

	// Add parameters back into state.
	// These aren't returned by the API.
	for k := range p {
		v, ok := plan.Values[k]
		if ok && v.String != nil {
			result[k] = *v.String
		}
	}

	_, ok := result["path"]
	if !ok {
		return nil, fmt.Errorf("expected path in response %v", resp)
	}
	result["id"] = result["path"]

	err := data.ToResource(result, plan)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

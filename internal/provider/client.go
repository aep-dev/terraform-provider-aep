package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func Create(ctx context.Context, r *api.Resource, c *http.Client, serverUrl string, body map[string]interface{}) (map[string]interface{}, error) {
	suffix := ""
	if r.CreateMethod.SupportsUserSettableCreate {
		id, ok := body["id"]
		if !ok {
			return nil, fmt.Errorf("id field not found in %v", body)
		}
		idString, ok := id.(string)
		if !ok {
			return nil, fmt.Errorf("id field is not string %v", id)
		}

		suffix = fmt.Sprintf("?id=%s", idString)
	}
	url, err := createBase(ctx, r, body, serverUrl, suffix)
	if err != nil {
		return nil, err
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	tflog.Info(ctx, fmt.Sprintf("Response body: %q", string(respBody)))
	var data map[string]interface{}
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Read(ctx context.Context, r *api.Resource, c *http.Client, serverUrl string, parameters map[string]interface{}) (map[string]interface{}, error) {
	id, ok := parameters["path"]
	if !ok {
		return nil, fmt.Errorf("path field not found in %v", parameters)
	}
	idString, ok := id.(string)
	if !ok {
		return nil, fmt.Errorf("path field is not string %v", id)
	}

	url, err := createBase(ctx, r, parameters, serverUrl, fmt.Sprintf("/%s", idString))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return data, nil
}

func Delete(ctx context.Context, r *api.Resource, c *http.Client, serverUrl string, parameters map[string]interface{}) error {
	id, ok := parameters["path"]
	if !ok {
		return fmt.Errorf("path field not found in %v", parameters)
	}
	idString, ok := id.(string)
	if !ok {
		return fmt.Errorf("path field is not string %v", id)
	}

	url, err := createBase(ctx, r, parameters, serverUrl, fmt.Sprintf("/%s", idString))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %v", err)
	}

	_, err = c.Do(req)
	return err
}

func Update(ctx context.Context, r *api.Resource, c *http.Client, serverUrl string, parameters map[string]interface{}) error {
	id, ok := parameters["path"]
	if !ok {
		return fmt.Errorf("path field not found in %v", parameters)
	}
	idString, ok := id.(string)
	if !ok {
		return fmt.Errorf("path field is not string %v", id)
	}

	url, err := createBase(ctx, r, parameters, serverUrl, fmt.Sprintf("/%s", idString))
	if err != nil {
		return err
	}

	reqBody, err := json.Marshal(parameters)
	if err != nil {
		return fmt.Errorf("error marshalling JSON for request body: %v", err)
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("error creating delete request: %v", err)
	}

	_, err = c.Do(req)
	return err
}

func createBase(ctx context.Context, r *api.Resource, body map[string]interface{}, serverUrl string, suffix string) (string, error) {
	return serverUrl + "/" + r.Plural + suffix, nil
}

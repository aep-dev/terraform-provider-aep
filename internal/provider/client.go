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

func Create(ctx context.Context, r *api.Resource, c *http.Client, serverUrl string, body map[string]interface{}, parameters map[string]string) (map[string]interface{}, error) {
	suffix := ""
	if r.CreateMethod != nil && r.CreateMethod.SupportsUserSettableCreate {
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
	url := createBase(ctx, r, serverUrl, parameters, suffix)
	tflog.Info(ctx, fmt.Sprintf("create url %s", url))

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

func Read(ctx context.Context, c *http.Client, serverUrl string, path string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s", serverUrl, strings.TrimPrefix(path, "/"))

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

func Delete(ctx context.Context, c *http.Client, serverUrl string, path string) error {
	url := fmt.Sprintf("%s/%s", serverUrl, strings.TrimPrefix(path, "/"))

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %v", err)
	}

	_, err = c.Do(req)
	return err
}

func Update(ctx context.Context, c *http.Client, serverUrl string, path string, body map[string]interface{}) error {
	url := fmt.Sprintf("%s/%s", serverUrl, strings.TrimPrefix(path, "/"))

	reqBody, err := json.Marshal(body)
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

func createBase(ctx context.Context, r *api.Resource, serverUrl string, parameters map[string]string, suffix string) string {
	urlElems := []string{serverUrl}
	for i, elem := range r.PatternElems {
		tflog.Info(ctx, fmt.Sprintf("pattern elem %s", elem))
		if i == len(r.PatternElems)-1 {
			tflog.Info(ctx, "skipping")
			continue
		}

		if i%2 == 0 {
			tflog.Info(ctx, fmt.Sprintf("adding elem %s", elem))
			urlElems = append(urlElems, elem)
		} else {
			paramName := elem[1 : len(elem)-1]
			tflog.Info(ctx, fmt.Sprintf("pattern name %s", paramName))
			tflog.Info(ctx, fmt.Sprintf("parameters %q", parameters))
			if value, ok := parameters[paramName]; ok {
				if strings.Contains(value, "/") {
					value = strings.Split(value, "/")[len(strings.Split(value, "/"))-1]
				}
				urlElems = append(urlElems, value)
			}
		}
	}
	if suffix != "" {
		urlElems = append(urlElems, suffix)
	}
	tflog.Info(ctx, fmt.Sprintf("url elems %q", urlElems))
	return strings.Join(urlElems, "/")
}

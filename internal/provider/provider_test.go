// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/jarcoal/httpmock"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"scaffolding": func() (tfprotov6.ProviderServer, error) {
		gen, err := CreateGeneratedProviderData("testdata/oas.yaml")
		if err != nil {
			return nil, fmt.Errorf("unable to create generated data %v", err)
		}
		oas, err := openapi.FetchOpenAPI("testdata/oas.yaml")
		if err != nil {
			return nil, fmt.Errorf("unable to fetch oas spec %v", err)
		}
		mockClient := &http.Client{}
		return providerserver.NewProtocol6WithError(New("test", gen, oas, mockClient)())()
	},
}

func testAccPreCheck(t *testing.T) {
	httpmock.Activate()

	allResources := make(map[string]interface{})
	var bookCounter = 1
	httpmock.RegisterResponder("POST", "http://localhost:8081/publishers",
		func(req *http.Request) (*http.Response, error) {
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, ""), nil
			}
			allResources[fmt.Sprintf("%d", bookCounter)] = requestBody
			requestBody["path"] = fmt.Sprintf("/publishers/%d", bookCounter)
			bookCounter += 1
			jsonRequestBody, err := json.Marshal(requestBody)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return httpmock.NewStringResponse(201, string(jsonRequestBody)), nil
		},
	)

	httpmock.RegisterResponder("GET", "=~^http://localhost:8081/publishers/\\d+",
		func(req *http.Request) (*http.Response, error) {
			publisherID := req.URL.Path[len("/publishers/"):]
			resource, ok := allResources[publisherID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}
			resourceMap, ok := resource.(map[string]interface{})
			if !ok {
				return httpmock.NewStringResponse(500, ""), nil
			}
			resourceMap["path"] = fmt.Sprintf("/publishers/%s", publisherID)
			jsonResource, err := json.Marshal(resource)
			bookCounter += 1
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return httpmock.NewStringResponse(200, string(jsonResource)), nil
		},
	)

	httpmock.RegisterResponder("PATCH", "=~^http://localhost:8081/publishers/\\d+",
		func(req *http.Request) (*http.Response, error) {
			publisherID := req.URL.Path[len("/publishers/"):]
			_, ok := allResources[publisherID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, ""), nil
			}
			allResources[publisherID] = requestBody
			requestBody["path"] = fmt.Sprintf("/publishers/%s", publisherID)
			jsonResource, err := json.Marshal(requestBody)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return httpmock.NewStringResponse(200, string(jsonResource)), nil
		},
	)

	httpmock.RegisterResponder("DELETE", "=~^http://localhost:8081/publishers/\\d+",
		func(req *http.Request) (*http.Response, error) {
			publisherID := req.URL.Path[len("/publishers/"):]
			_, ok := allResources[publisherID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			delete(allResources, publisherID)

			return httpmock.NewStringResponse(200, ""), nil
		},
	)

}

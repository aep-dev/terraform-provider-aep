// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/client"
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
		gen, err := CreateGeneratedProviderData("testdata/oas.yaml", "")
		if err != nil {
			return nil, fmt.Errorf("unable to create generated data %v", err)
		}
		mockClient := client.NewClient(&http.Client{})
		providerConfig := ProviderConfig{
			OpenAPIPath:    "http://localhost:8081/openapi.json",
			ProviderPrefix: "scaffolding",
		}
		return providerserver.NewProtocol6WithError(New("test", gen, mockClient, providerConfig)())()
	},
}

func testAccPreCheck(_ *testing.T) {
	httpmock.Activate()

	allPublishers := make(map[string]interface{})
	var publisherCounter = 1

	allBooks := make(map[string]interface{})
	var bookCounter = 1

	// Books Mock Server.
	httpmock.RegisterResponder("POST", "=~^http://localhost:8081/publishers/\\d+/books",
		func(req *http.Request) (*http.Response, error) {
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, ""), err
			}

			// Ensure publisher has been created.
			publisherNumber := req.URL.Path[len("/publishers/"):]
			publisherNumber = strings.Split(publisherNumber, "/")[0]
			_, ok := allPublishers[publisherNumber]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			// Create book.
			allBooks[fmt.Sprintf("%d", bookCounter)] = requestBody
			requestBody["path"] = fmt.Sprintf("/publishers/%s/books/%d", publisherNumber, bookCounter)
			bookCounter += 1
			jsonRequestBody, err := json.Marshal(requestBody)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), err
			}
			return httpmock.NewStringResponse(201, string(jsonRequestBody)), nil
		},
	)

	httpmock.RegisterResponder("GET", "=~^http://localhost:8081/publishers/\\d+/books/\\d+",
		func(req *http.Request) (*http.Response, error) {
			// Ensure publisher has been created.
			publisherNumber := req.URL.Path[len("/publishers/"):]
			publisherNumber = strings.Split(publisherNumber, "/")[0]
			_, ok := allPublishers[publisherNumber]
			if !ok {
				return httpmock.NewStringResponse(404, fmt.Sprintf("could not find publisher %s", publisherNumber)), nil
			}

			bookID := strings.Split(req.URL.Path, "/")[len(strings.Split(req.URL.Path, "/"))-1]
			resource, ok := allBooks[bookID]
			if !ok {
				return httpmock.NewStringResponse(404, fmt.Sprintf("could not find book %s", bookID)), nil
			}
			resourceMap, ok := resource.(map[string]interface{})
			if !ok {
				return httpmock.NewStringResponse(500, ""), nil
			}
			resourceMap["path"] = fmt.Sprintf("/publishers/%s/books/%s", publisherNumber, bookID)
			jsonResource, err := json.Marshal(resource)
			publisherCounter += 1
			if err != nil {
				return httpmock.NewStringResponse(500, ""), err
			}
			return httpmock.NewStringResponse(200, string(jsonResource)), nil
		},
	)

	httpmock.RegisterResponder("PATCH", "=~^http://localhost:8081/publishers/\\d+/books/\\d+",
		func(req *http.Request) (*http.Response, error) {
			publisherID := strings.Split(req.URL.Path[len("/publishers/"):], "/")[0]
			_, ok := allPublishers[publisherID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			bookID := strings.Split(req.URL.Path, "/")[len(strings.Split(req.URL.Path, "/"))-1]
			_, ok = allBooks[bookID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, ""), err
			}
			requestBody["path"] = fmt.Sprintf("/publishers/%s/books/%s", publisherID, bookID)
			allBooks[bookID] = requestBody
			jsonResource, err := json.Marshal(requestBody)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), err
			}
			return httpmock.NewStringResponse(200, string(jsonResource)), nil
		},
	)

	httpmock.RegisterResponder("DELETE", "=~^http://localhost:8081/publishers/\\d+/books/\\d+",
		func(req *http.Request) (*http.Response, error) {
			publisherID := req.URL.Path[len("/publishers/"):]
			_, ok := allPublishers[publisherID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			bookID := strings.Split(req.URL.Path, "/")[len(strings.Split(req.URL.Path, "/"))-1]
			_, ok = allBooks[bookID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			delete(allBooks, bookID)

			return httpmock.NewStringResponse(200, ""), nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8081/publishers",
		func(req *http.Request) (*http.Response, error) {
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, ""), err
			}
			allPublishers[fmt.Sprintf("%d", publisherCounter)] = requestBody
			requestBody["path"] = fmt.Sprintf("/publishers/%d", publisherCounter)
			publisherCounter += 1
			jsonRequestBody, err := json.Marshal(requestBody)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), err
			}
			return httpmock.NewStringResponse(201, string(jsonRequestBody)), nil
		},
	)

	httpmock.RegisterResponder("GET", "=~^http://localhost:8081/publishers/\\d+",
		func(req *http.Request) (*http.Response, error) {
			publisherID := req.URL.Path[len("/publishers/"):]
			resource, ok := allPublishers[publisherID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}
			resourceMap, ok := resource.(map[string]interface{})
			if !ok {
				return httpmock.NewStringResponse(500, ""), nil
			}
			resourceMap["path"] = fmt.Sprintf("/publishers/%s", publisherID)
			jsonResource, err := json.Marshal(resource)
			publisherCounter += 1
			if err != nil {
				return httpmock.NewStringResponse(500, ""), err
			}
			return httpmock.NewStringResponse(200, string(jsonResource)), nil
		},
	)

	httpmock.RegisterResponder("PATCH", "=~^http://localhost:8081/publishers/\\d+",
		func(req *http.Request) (*http.Response, error) {
			publisherID := req.URL.Path[len("/publishers/"):]
			_, ok := allPublishers[publisherID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, ""), err
			}
			allPublishers[publisherID] = requestBody
			requestBody["path"] = fmt.Sprintf("/publishers/%s", publisherID)
			jsonResource, err := json.Marshal(requestBody)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), err
			}
			return httpmock.NewStringResponse(200, string(jsonResource)), nil
		},
	)

	httpmock.RegisterResponder("DELETE", "=~^http://localhost:8081/publishers/\\d+",
		func(req *http.Request) (*http.Response, error) {
			publisherID := req.URL.Path[len("/publishers/"):]
			_, ok := allPublishers[publisherID]
			if !ok {
				return httpmock.NewStringResponse(404, ""), nil
			}

			delete(allPublishers, publisherID)

			return httpmock.NewStringResponse(200, ""), nil
		},
	)

}

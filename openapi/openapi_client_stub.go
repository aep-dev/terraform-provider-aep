package openapi

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// clientOpenAPIStub is a stubbed client used for testing purposes that implements the ClientOpenAPI interface
type clientOpenAPIStub struct {
	responsePayload     map[string]interface{}
	responseListPayload []map[string]interface{}
	error               error
	returnHTTPCode      int
	idReceived          string
	parentIDsReceived   []string
	telemetryHandler    TelemetryHandler

	funcPatch func() (*http.Response, error)
}

func (c *clientOpenAPIStub) Post(resource SpecResource, requestPayload interface{}, responsePayload interface{}, parentIDs ...string) (*http.Response, error) {
	if c.error != nil {
		return nil, c.error
	}
	c.parentIDsReceived = parentIDs
	switch p := responsePayload.(type) {
	case *map[string]interface{}:
		*p = c.responsePayload
	default:
		panic("unexpected type")
	}
	return c.generateStubResponse(http.StatusCreated), nil
}

func (c *clientOpenAPIStub) Patch(resource SpecResource, id string, requestPayload interface{}, responsePayload interface{}, parentIDs ...string) (*http.Response, error) {
	if c.funcPatch != nil {
		return c.funcPatch()
	}
	if c.error != nil {
		return nil, c.error
	}
	c.idReceived = id
	c.parentIDsReceived = parentIDs
	switch p := responsePayload.(type) {
	case *map[string]interface{}:
		*p = c.responsePayload
	default:
		panic("unexpected type")
	}
	return c.generateStubResponse(http.StatusOK), nil
}

func (c *clientOpenAPIStub) Get(resource SpecResource, id string, responsePayload interface{}, parentIDs ...string) (*http.Response, error) {
	if c.error != nil {
		return nil, c.error
	}
	c.idReceived = id
	c.parentIDsReceived = parentIDs
	switch p := responsePayload.(type) {
	case *map[string]interface{}:
		*p = c.responsePayload
	default:
		panic("unexpected type")
	}

	return c.generateStubResponse(http.StatusOK), nil
}

func (c *clientOpenAPIStub) List(resource SpecResource, responsePayload interface{}, parentIDs ...string) (*http.Response, error) {
	if c.error != nil {
		return nil, c.error
	}
	c.parentIDsReceived = parentIDs
	switch p := responsePayload.(type) {
	case *[]map[string]interface{}:
		*p = c.responseListPayload
	default:
		panic("unexpected type")
	}

	return c.generateStubResponse(http.StatusOK), nil
}

func (c *clientOpenAPIStub) Delete(resource SpecResource, id string, parentIDs ...string) (*http.Response, error) {
	if c.error != nil {
		return nil, c.error
	}
	c.idReceived = id
	c.parentIDsReceived = parentIDs
	delete(c.responsePayload, id)
	return c.generateStubResponse(http.StatusNoContent), nil
}

func (c *clientOpenAPIStub) GetTelemetryHandler() TelemetryHandler {
	return c.telemetryHandler
}

func (c *clientOpenAPIStub) generateStubResponse(defaultHTTPCode int) *http.Response {
	return &http.Response{
		StatusCode: c.returnCode(defaultHTTPCode),
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}
}

func (c *clientOpenAPIStub) returnCode(defaultHTTPCode int) int {
	if c.returnHTTPCode != 0 {
		return c.returnHTTPCode
	}
	return defaultHTTPCode
}

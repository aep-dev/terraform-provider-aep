package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/aep-dev/terraform-provider-aep/openapi/version"
)

type httpMethodSupported string

const (
	httpGet    httpMethodSupported = "GET"
	httpPost   httpMethodSupported = "POST"
	httpPatch  httpMethodSupported = "PATCH"
	httpDelete httpMethodSupported = "DELETE"
)

// ClientOpenAPI defines the behaviour expected to be implemented for the OpenAPI Client used in the Terraform OpenAPI Provider
type ClientOpenAPI interface {
	Post(resource SpecResource, requestPayload interface{}, responsePayload interface{}, parentIDs ...string) (*http.Response, error)
	Patch(resource SpecResource, id string, requestPayload interface{}, responsePayload interface{}, parentIDs ...string) (*http.Response, error)
	Get(resource SpecResource, id string, responsePayload interface{}, parentIDs ...string) (*http.Response, error)
	Delete(resource SpecResource, id string, parentIDs ...string) (*http.Response, error)
	List(resource SpecResource, responsePayload interface{}, parentIDs ...string) (*http.Response, error)
	GetTelemetryHandler() TelemetryHandler
}

// ProviderClient defines a client that is configured based on the OpenAPI server side documentation
// The CRUD operations accept an OpenAPI operation which defines among other things the security scheme applicable to
// the API when making the HTTP requests
type ProviderClient struct {
	openAPIBackendConfiguration SpecBackendConfiguration
	httpClient                  *http.Client
	providerConfiguration       providerConfiguration
	apiAuthenticator            specAuthenticator
	telemetryHandler            TelemetryHandler
}

// Post performs a POST request to the server API based on the resource configuration and the payload passed in
func (o *ProviderClient) Post(resource SpecResource, requestPayload interface{}, responsePayload interface{}, parentIDs ...string) (*http.Response, error) {
	resourceURL, err := o.getResourceURL(resource, parentIDs)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().Post
	return o.performRequest(httpPost, resourceURL, operation, requestPayload, responsePayload)
}

// Patch performs a PATCH request to the server API based on the resource configuration and the payload passed in
func (o *ProviderClient) Patch(resource SpecResource, id string, requestPayload interface{}, responsePayload interface{}, parentIDs ...string) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(resource, parentIDs, id)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().Patch
	return o.performRequest(httpPatch, resourceURL, operation, requestPayload, responsePayload)
}

// Get performs a GET request to the server API based on the resource configuration and the resource instance id passed in
func (o *ProviderClient) Get(resource SpecResource, id string, responsePayload interface{}, parentIDs ...string) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(resource, parentIDs, id)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().Get
	return o.performRequest(httpGet, resourceURL, operation, nil, responsePayload)
}

// List performs a GET request to the root level endpoint of the resource (e,g: GET /v1/groups)
func (o *ProviderClient) List(resource SpecResource, responsePayload interface{}, parentIDs ...string) (*http.Response, error) {
	resourceURL, err := o.getResourceURL(resource, parentIDs)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().List
	return o.performRequest(httpGet, resourceURL, operation, nil, responsePayload)
}

// Delete performs a DELETE request to the server API based on the resource configuration and the resource instance id passed in
func (o *ProviderClient) Delete(resource SpecResource, id string, parentIDs ...string) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(resource, parentIDs, id)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().Delete
	return o.performRequest(httpDelete, resourceURL, operation, nil, nil)
}

// GetTelemetryHandler returns the configured telemetry handler
func (o *ProviderClient) GetTelemetryHandler() TelemetryHandler {
	return o.telemetryHandler
}

func (o *ProviderClient) performRequest(method httpMethodSupported, resourceURL string, operation *specResourceOperation, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {
	reqContext, err := o.apiAuthenticator.prepareAuth(resourceURL, operation.SecuritySchemes, o.providerConfiguration)
	if err != nil {
		return nil, fmt.Errorf("failed to configure the API request for %s %s: %s", method, resourceURL, err)
	}

	err = o.appendOperationHeaders(operation.HeaderParameters, reqContext.headers)
	if err != nil {
		return nil, fmt.Errorf("failed to configure the API request for %s %s: %s", method, resourceURL, err)
	}
	log.Printf("[DEBUG] Performing %s %s", method, reqContext.url)

	userAgentHeader := version.BuildUserAgent(runtime.GOOS, runtime.GOARCH)
	o.appendUserAgentHeader(reqContext.headers, userAgentHeader)

	o.logHeadersSafely(reqContext.headers)

	var body []byte
	if requestPayload != nil {
		body, err = json.Marshal(requestPayload)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(fmt.Sprintf("%s", method), reqContext.url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	for key, value := range reqContext.headers {
		req.Header.Set(key, value)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if responsePayload != nil {
		var body []byte
		if body, err = io.ReadAll(resp.Body); err != nil {
			return nil, err
		}
		resp.Body.Close()                               // close stream so connection is closed gracefully
		resp.Body = io.NopCloser(bytes.NewReader(body)) // create a new reader from bytes read in the response and set the response body (allowing the client to still be able to do res.Body afterwards)
		if len(body) > 0 {
			if err = json.Unmarshal(body, &responsePayload); err != nil {
				return resp, fmt.Errorf("unable to unmarshal response body ['%s'] for request = '%s %s %s'. Response = '%s'", err.Error(), req.Method, req.URL, req.Proto, resp.Status)
			}
		} else {
			return resp, fmt.Errorf("expected a response body but response body received was empty for request = '%s %s %s'. Response = '%s'", req.Method, req.URL, req.Proto, resp.Status)
		}
	}

	return resp, nil
}

func (o *ProviderClient) appendUserAgentHeader(headers map[string]string, value string) {
	headers[userAgentHeader] = value
}

// logHeadersSafely logs the header names sent to the APIs but the values are redacted for security reasons in case
// values contain secrets. However, the logging will display whether the values contained data or not so it's easier
// to debug whether the headers sent had data.
func (o *ProviderClient) logHeadersSafely(headers map[string]string) {
	for headerName, headerValue := range headers {
		if headerValue == "" {
			log.Printf("[DEBUG] Request Header '%s' sent with empty value :(", headerName)
		}
		log.Printf("[DEBUG] Request Header '%s' sent", headerName)
	}
}

// appendOperationHeaders returns a maps containing the headers passed in and adds whatever headers the operation requires. The values
// are retrieved from the provider configuration.
func (o ProviderClient) appendOperationHeaders(operationHeaders []SpecHeaderParam, headers map[string]string) error {
	if operationHeaders != nil && len(operationHeaders) > 0 {
		for _, headerParam := range operationHeaders {
			headerValue := o.providerConfiguration.getHeaderValueFor(headerParam)
			if headerParam.IsRequired && headerValue == "" {
				return fmt.Errorf("required header '%s' is missing the value. Please make sure the property '%s' is configured with a value in the provider's terraform configuration", headerParam.Name, headerParam.GetHeaderTerraformConfigurationName())
			}
			// Setting the actual name of the header with the expectedValue coming from the provider configuration
			headers[headerParam.Name] = o.providerConfiguration.getHeaderValueFor(headerParam)
		}
	}
	return nil
}

func (o ProviderClient) getResourceURL(resource SpecResource, parentIDs []string) (string, error) {
	var host string
	var err error

	isMultiRegion, _, regions, err := o.openAPIBackendConfiguration.IsMultiRegion()
	if err != nil {
		return "", err
	}
	if isMultiRegion {
		// get region value provided by user in the terraform configuration file
		region := o.providerConfiguration.getRegion()
		// otherwise, if not provided falling back to the default value specified in the service provider swagger file
		if region == "" {
			region, err = o.openAPIBackendConfiguration.GetDefaultRegion(regions)
			if err != nil {
				return "", err
			}
		}
		host, err = o.openAPIBackendConfiguration.getHostByRegion(region)
		if err != nil {
			return "", err
		}
	} else {
		host, err = o.openAPIBackendConfiguration.getHost()
		if err != nil {
			return "", err
		}
	}

	basePath := o.openAPIBackendConfiguration.getBasePath()
	resourceRelativePath, err := resource.getResourcePath(parentIDs)
	if err != nil {
		return "", err
	}

	// Fall back to override the host if value is not empty; otherwise global host will be used as usual
	hostOverride, err := resource.getHost()
	if err != nil {
		return "", err
	}
	if hostOverride != "" {
		log.Printf("[INFO] resource '%s' is configured with host override, API calls will be made against '%s' instead of '%s'", resourceRelativePath, hostOverride, host)
		host = hostOverride
	}

	if endPointHost := o.providerConfiguration.getEndPoint(resource.GetResourceName()); endPointHost != "" {
		log.Printf("[INFO] resource '%s' is configured with endpoint override, API calls will be made against '%s' instead of '%s'", resourceRelativePath, endPointHost, host)
		host = endPointHost
	}

	if host == "" || resourceRelativePath == "" {
		return "", fmt.Errorf("host and path are mandatory attributes to get the resource URL - host['%s'], path['%s']", host, resourceRelativePath)
	}

	// TODO: use resource operation schemes if specified
	defaultScheme, err := o.openAPIBackendConfiguration.getHTTPScheme()
	if err != nil {
		return "", err
	}

	path := resourceRelativePath
	if strings.Index(resourceRelativePath, "/") != 0 {
		path = fmt.Sprintf("/%s", resourceRelativePath)
	}

	if basePath != "" && basePath != "/" {
		if strings.Index(basePath, "/") == 0 {
			return fmt.Sprintf("%s://%s%s%s", defaultScheme, host, basePath, path), nil
		}
		return fmt.Sprintf("%s://%s/%s%s", defaultScheme, host, basePath, path), nil
	}
	return fmt.Sprintf("%s://%s%s", defaultScheme, host, path), nil
}

func (o ProviderClient) getResourceIDURL(resource SpecResource, parentIDs []string, id string) (string, error) {
	if strings.Contains(id, "/") {
		return "", fmt.Errorf("instance ID (%s) contains not supported characters (forward slashes)", id)
	}
	url, err := o.getResourceURL(resource, parentIDs)
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("could not build the resourceIDURL: required instance id value is missing")
	}
	if strings.HasSuffix(url, "/") {
		return fmt.Sprintf("%s%s", url, id), nil
	}
	return fmt.Sprintf("%s/%s", url, id), nil
}

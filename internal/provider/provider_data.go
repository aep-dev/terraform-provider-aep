package provider

import (
	"net/http"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

type GeneratedProviderData struct {
	client  *http.Client
	api     *api.API
	openapi *openapi.OpenAPI
}

func CreateGeneratedProviderData(path string, pathPrefix string) (*GeneratedProviderData, error) {
	oas, err := openapi.FetchOpenAPI(path)
	if err != nil {
		return nil, err
	}

	a, err := api.GetAPI(oas, "", pathPrefix)
	if err != nil {
		return nil, err
	}

	return &GeneratedProviderData{
		client:  http.DefaultClient,
		api:     a,
		openapi: oas,
	}, nil
}

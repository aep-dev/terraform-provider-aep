package provider

import (
	"context"
	"net/http"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/aep-dev/terraform-provider-aep/internal/provider/data"
)

type GeneratedProviderData struct {
	client  *http.Client
	api     *api.API
	openapi *openapi.OpenAPI

	resources map[string]*data.ResourceSchema
}

func (p *GeneratedProviderData) Resource(name string) {

}

func CreateGeneratedProviderData(ctx context.Context, path string, pathPrefix string) (*GeneratedProviderData, error) {
	oas, err := openapi.FetchOpenAPI(path)
	if err != nil {
		return nil, err
	}

	a, err := api.GetAPI(oas, "", pathPrefix)
	if err != nil {
		return nil, err
	}

	resources := make(map[string]*data.ResourceSchema)
	for name, resource := range a.Resources {
		resSchema, err := data.NewResourceSchema(context.Background(), resource, oas)
		if err != nil {
			return nil, err
		}
		resources[name] = resSchema
	}

	return &GeneratedProviderData{
		client:    http.DefaultClient,
		api:       a,
		openapi:   oas,
		resources: resources,
	}, nil
}

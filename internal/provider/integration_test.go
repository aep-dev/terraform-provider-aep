package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/client"
	"github.com/aep-dev/terraform-provider-aep/config"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func aepbaseProviderFactory() map[string]func() (tfprotov6.ProviderServer, error) {
	serverURL := os.Getenv("AEPBASE_URL")
	return map[string]func() (tfprotov6.ProviderServer, error){
		"scaffolding": func() (tfprotov6.ProviderServer, error) {
			openAPIURL := serverURL + "/openapi.json"
			gen, err := CreateGeneratedProviderData(context.TODO(), openAPIURL, "")
			if err != nil {
				return nil, fmt.Errorf("unable to create generated data from %s: %v", openAPIURL, err)
			}
			httpClient := client.NewClient(&http.Client{})
			providerConfig := config.NewProviderConfigForTesting(openAPIURL, "", "", "scaffolding")
			return providerserver.NewProtocol6WithError(New("test", gen, httpClient, providerConfig)())()
		},
	}
}

func skipIfNoAepbase(t *testing.T) {
	t.Helper()
	if os.Getenv("AEPBASE_URL") == "" {
		t.Skip("AEPBASE_URL not set, skipping integration test")
	}
}

func resourceDefinitionConfig() string {
	return `
resource "scaffolding_aep-resource-definition" "publisher" {
  singular = "publisher"
  plural   = "publishers"
  schema = jsonencode({
    properties = {
      description = { type = "string" }
    }
  })
}

resource "scaffolding_aep-resource-definition" "book" {
  singular   = "book"
  plural     = "books"
  parents    = ["publisher"]
  depends_on = [scaffolding_aep-resource-definition.publisher]
  schema = jsonencode({
    properties = {
      price     = { type = "string" }
      published = { type = "boolean" }
      isbn = {
        type  = "array"
        items = { type = "string" }
      }
    }
    required = ["price", "published"]
  })
}
`
}

func integrationBookConfig(book string, price string, publisher string) string {
	return fmt.Sprintf(`
resource "scaffolding_publisher" %[3]q {
  description = "pub-description"
}

resource "scaffolding_book" %[1]q {
  price = %[2]q
  publisher_id = scaffolding_publisher.%[3]s.path
  published = true
}
`, book, price, publisher)
}

func TestIntegrationResources(t *testing.T) {
	skipIfNoAepbase(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: aepbaseProviderFactory(),
		Steps: []resource.TestStep{
			// Create resource definitions
			{
				Config: resourceDefinitionConfig(),
			},
			// Create publisher
			{
				Config: resourceDefinitionConfig() + testExamplePublisherConfig("my-pub", "pub-description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "description", "pub-description"),
				),
			},
			// Update publisher
			{
				Config: resourceDefinitionConfig() + testExamplePublisherConfig("my-pub", "pub-description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "description", "pub-description2"),
				),
			},
			// Create book (with publisher)
			{
				Config: resourceDefinitionConfig() + integrationBookConfig("my-book", "1", "my-pub"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "description", "pub-description"),
					resource.TestCheckResourceAttr("scaffolding_book.my-book", "price", "1"),
				),
			},
			// Update book
			{
				Config: resourceDefinitionConfig() + integrationBookConfig("my-book", "3", "my-pub"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "description", "pub-description"),
					resource.TestCheckResourceAttr("scaffolding_book.my-book", "price", "3"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

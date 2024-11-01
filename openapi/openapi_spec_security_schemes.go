package openapi

import "github.com/aep-dev/terraform-provider-aep/openapi/terraformutils"

// SpecSecuritySchemes groups a list of SpecSecurityScheme
type SpecSecuritySchemes []SpecSecurityScheme

func createSecuritySchemes(securitySchemes []map[string][]string) SpecSecuritySchemes {
	schemes := SpecSecuritySchemes{}
	for _, securityScheme := range securitySchemes {
		for securitySchemeName := range securityScheme {
			schemes = append(schemes, SpecSecurityScheme{Name: securitySchemeName})
		}
		// Choosing the first set of security schemes as defined by the service provider. The order defines the priority
		// by which security schemes are selected, in this case the first set. Hence, disregarding the rest of security
		// schemes (if defined)
		break
	}
	return schemes
}

func (s SpecSecuritySchemes) securitySchemeExists(secDef SpecSecurityDefinition) bool {
	for _, securityScheme := range s {
		if securityScheme.GetTerraformConfigurationName() == secDef.GetTerraformConfigurationName() {
			return true
		}
	}
	return false
}

// SpecSecurityScheme defines a security scheme. This struct serves as a translation between the OpenAPI document
// and the scheme that will be used by the OpenAPI Terraform provider when making API calls to the backend
type SpecSecurityScheme struct {
	Name string
}

// GetTerraformConfigurationName returns the scheme name converted to a terraform compliant name if needed following the snake_case naming convention
func (o *SpecSecurityScheme) GetTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(o.Name)
}

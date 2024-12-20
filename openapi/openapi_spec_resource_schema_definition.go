package openapi

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SpecSchemaDefinitionProperties defines a collection of schema definition properties
type SpecSchemaDefinitionProperties []*SpecSchemaDefinitionProperty

// SpecSchemaDefinition defines a struct for a schema definition
type SpecSchemaDefinition struct {
	Properties SpecSchemaDefinitionProperties
}

// ConvertToDataSourceSpecSchemaDefinition transforms the current SpecSchemaDefinition into a data source SpecSchemaDefinition. This
// means that all the properties that form the schema will be made optional, computed and they won't have default values.
func (s *SpecSchemaDefinition) ConvertToDataSourceSpecSchemaDefinition() *SpecSchemaDefinition {
	specSchemaDefinition := &SpecSchemaDefinition{
		Properties: SpecSchemaDefinitionProperties{},
	}
	for _, p := range s.Properties {
		dataSourceSpecSchemaDefinitionProperty := s.convertToDataSourceSpecSchemaDefinitionProperty(*p)
		specSchemaDefinition.Properties = append(specSchemaDefinition.Properties, dataSourceSpecSchemaDefinitionProperty)
	}
	return specSchemaDefinition
}

func (s *SpecSchemaDefinition) convertToDataSourceSpecSchemaDefinitionProperty(specSchemaDefinitionProperty SpecSchemaDefinitionProperty) *SpecSchemaDefinitionProperty {
	if specSchemaDefinitionProperty.IsParentProperty {
		return &specSchemaDefinitionProperty
	}
	specSchemaDefinitionProperty.Required = false
	specSchemaDefinitionProperty.Computed = true
	specSchemaDefinitionProperty.Default = nil
	if specSchemaDefinitionProperty.SpecSchemaDefinition != nil {
		dataSourceObjectSpecSchemaDefinition := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{},
		}
		for _, objectProperty := range specSchemaDefinitionProperty.SpecSchemaDefinition.Properties {
			dataSourceObjectProperty := s.convertToDataSourceSpecSchemaDefinitionProperty(*objectProperty)
			dataSourceObjectSpecSchemaDefinition.Properties = append(dataSourceObjectSpecSchemaDefinition.Properties, dataSourceObjectProperty)
		}
		specSchemaDefinitionProperty.SpecSchemaDefinition = dataSourceObjectSpecSchemaDefinition
	}
	return &specSchemaDefinitionProperty
}

func (s *SpecSchemaDefinition) createResourceSchema() (map[string]*schema.Schema, error) {
	return s.createResourceSchemaIgnoreID(true)
}

func (s *SpecSchemaDefinition) createDataSourceSchema() (map[string]*schema.Schema, error) {
	dataSourceSpecSchemaDefinition := s.ConvertToDataSourceSpecSchemaDefinition()
	terraformSchema, err := dataSourceSpecSchemaDefinition.createResourceSchemaIgnoreID(true)
	if err != nil {
		return nil, err
	}
	return terraformSchema, nil
}

func (s *SpecSchemaDefinition) createResourceSchemaKeepID() (map[string]*schema.Schema, error) {
	return s.createResourceSchemaIgnoreID(false)
}

func (s *SpecSchemaDefinition) createResourceSchemaIgnoreID(ignoreID bool) (map[string]*schema.Schema, error) {
	terraformSchema := map[string]*schema.Schema{}
	for _, property := range s.Properties {
		// Terraform already has a field ID reserved, hence the schema does not need to include an explicit ID property
		if property.isPropertyNamedID() && ignoreID {
			continue
		}
		tfSchema, err := property.terraformSchema()
		if err != nil {
			return nil, err
		}
		terraformSchema[property.GetTerraformCompliantPropertyName()] = tfSchema
	}
	return terraformSchema, nil
}

func (s *SpecSchemaDefinition) getImmutableProperties() []string {
	var immutableProperties []string
	for _, property := range s.Properties {
		if property.isPropertyNamedID() {
			continue
		}
		if property.Immutable {
			immutableProperties = append(immutableProperties, property.Name)
		}
	}
	return immutableProperties
}

// // getResourceIdentifier returns the property name that is supposed to be used as the identifier. The resource id
// // is selected as follows:
// // 1.If the given schema definition contains a property configured with metadata 'x-terraform-id' set to true, that property value
// // will be used to set the state ID of the resource. Additionally, the value will be used when performing GET/PATCH/DELETE requests to
// // identify the resource in question.
// // 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
// // will have a property named 'id'
// // 3. If none of the above requirements is met, an error will be returned
func (s *SpecSchemaDefinition) getResourceIdentifier() (string, error) {
	identifierProperty := ""
	for _, property := range s.Properties {
		if property.isPropertyNamedID() {
			identifierProperty = property.Name
			continue
		}
		if property.IsIdentifier {
			identifierProperty = property.Name
			break
		}
	}
	// if the identifier property is missing, there is not way for the resource to be identified and therefore an error is returned
	if identifierProperty == "" {
		return "", fmt.Errorf("could not find any identifier property in the resource schema definition")
	}
	return identifierProperty, nil
}

// getStatusIdentifier returns the property name that is supposed to be used as the status field. The status field
// is selected as follows:
// 1.If the given schema definition contains a property configured with metadata 'x-terraform-field-status' set to true, that property
// will be used to check the different statues for the asynchronous pooling mechanism.
// 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
// will have a property named 'status'
// 3. If none of the above requirements is met, an error will be returned
func (s *SpecSchemaDefinition) getStatusIdentifier() ([]string, error) {
	return s.getStatusIdentifierFor(s, true, true)
}

func (s *SpecSchemaDefinition) getStatusIdentifierFor(schemaDefinition *SpecSchemaDefinition, shouldIgnoreID, shouldEnforceReadOnly bool) ([]string, error) {
	var statusProperty *SpecSchemaDefinitionProperty
	var statusHierarchy []string
	for _, property := range schemaDefinition.Properties {
		if property.isPropertyNamedID() && shouldIgnoreID {
			continue
		}
		if property.isPropertyNamedStatus() {
			statusProperty = property
			continue
		}
		// field with extTfFieldStatus metadata takes preference over 'status' fields as the service provider is the one acknowledging
		// the fact that this field should be used as identifier of the resource
		if property.IsStatusIdentifier {
			statusProperty = property
			break
		}
	}
	// if the id field is missing and there isn't any properties set with extTfFieldStatus, there is not way for the resource
	// to be identified and therefore an error is returned
	if statusProperty == nil {
		return nil, fmt.Errorf("could not find any status property. Please make sure the resource schema definition has either one property named '%s' or one property is marked with IsStatusIdentifier set to true", statusDefaultPropertyName)
	}
	if !statusProperty.ReadOnly && shouldEnforceReadOnly {
		return nil, fmt.Errorf("schema definition status property '%s' must be readOnly: '%+v' ", statusProperty.Name, statusProperty)
	}

	statusHierarchy = append(statusHierarchy, statusProperty.Name)
	if statusProperty.isObjectProperty() {
		statusIdentifier, err := s.getStatusIdentifierFor(statusProperty.SpecSchemaDefinition, false, false)
		if err != nil {
			return nil, err
		}
		statusHierarchy = append(statusHierarchy, statusIdentifier...)
	}
	return statusHierarchy, nil
}

func (s *SpecSchemaDefinition) getProperty(name string) (*SpecSchemaDefinitionProperty, error) {
	for _, property := range s.Properties {
		if property.Name == name {
			return property, nil
		}
	}
	return nil, fmt.Errorf("property with name '%s' not existing in resource schema definition", name)
}

func (s *SpecSchemaDefinition) getPropertyBasedOnTerraformName(terraformName string) (*SpecSchemaDefinitionProperty, error) {
	for _, property := range s.Properties {
		if property.GetTerraformCompliantPropertyName() == terraformName {
			return property, nil
		}
	}
	return nil, fmt.Errorf("property with terraform name '%s' not existing in resource schema definition", terraformName)
}

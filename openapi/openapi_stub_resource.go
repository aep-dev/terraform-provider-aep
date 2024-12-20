package openapi

// specStubResource is a stub implementation of SpecResource interface which is used for testing purposes
type specStubResource struct {
	name                    string
	host                    string
	path                    string
	shouldIgnore            bool
	schemaDefinition        *SpecSchemaDefinition
	resourceGetOperation    *specResourceOperation
	resourcePostOperation   *specResourceOperation
	resourceListOperation   *specResourceOperation
	resourcePatchOperation  *specResourceOperation
	resourceDeleteOperation *specResourceOperation
	timeouts                *specTimeouts

	parentResourceNames    []string
	fullParentResourceName string

	funcGetResourcePath   func(parentIDs []string) (string, error)
	funcGetResourceSchema func() (*SpecSchemaDefinition, error)
	funcGetTimeouts       func() (*specTimeouts, error)
	error                 error
}

func newSpecStubResource(name, path string, shouldIgnore bool, schemaDefinition *SpecSchemaDefinition) *specStubResource {
	return newSpecStubResourceWithOperations(name, path, shouldIgnore, schemaDefinition, nil, nil, nil, nil)
}

func newSpecStubResourceWithOperations(name, path string, shouldIgnore bool, schemaDefinition *SpecSchemaDefinition, resourcePostOperation, resourcePutOperation, resourceGetOperation, resourceDeleteOperation *specResourceOperation) *specStubResource {
	return &specStubResource{
		name:                    name,
		path:                    path,
		schemaDefinition:        schemaDefinition,
		shouldIgnore:            shouldIgnore,
		resourcePostOperation:   resourcePostOperation,
		resourceGetOperation:    resourceGetOperation,
		resourceDeleteOperation: resourceDeleteOperation,
		resourcePatchOperation:  resourcePutOperation,
		timeouts:                &specTimeouts{},
	}
}

func (s *specStubResource) GetResourceName() string { return s.name }

func (s *specStubResource) getResourcePath(parentIDs []string) (string, error) {
	if s.funcGetResourcePath != nil {
		return s.funcGetResourcePath(parentIDs)
	}
	return s.path, nil
}

func (s *specStubResource) GetResourceSchema() (*SpecSchemaDefinition, error) {
	if s.funcGetResourceSchema != nil {
		return s.funcGetResourceSchema()
	}
	if s.error != nil {
		return nil, s.error
	}
	return s.schemaDefinition, nil
}

func (s *specStubResource) ShouldIgnoreResource() bool { return s.shouldIgnore }

func (s *specStubResource) getResourceOperations() specResourceOperations {
	return specResourceOperations{
		List:   s.resourceListOperation,
		Post:   s.resourcePostOperation,
		Get:    s.resourceGetOperation,
		Patch:  s.resourcePatchOperation,
		Delete: s.resourceDeleteOperation,
	}
}

func (s *specStubResource) getTimeouts() (*specTimeouts, error) {
	if s.funcGetTimeouts != nil {
		return s.funcGetTimeouts()
	}
	return s.timeouts, nil
}

func (s *specStubResource) getHost() (string, error) {
	return s.host, nil
}

func (s *specStubResource) GetParentResourceInfo() *ParentResourceInfo {
	subRes := ParentResourceInfo{}
	if len(s.parentResourceNames) > 0 && s.fullParentResourceName != "" {
		subRes.parentResourceNames = s.parentResourceNames
		subRes.fullParentResourceName = s.fullParentResourceName
		return &subRes
	}
	return nil
}

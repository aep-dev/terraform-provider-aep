package openapi

type specResourceOperations struct {
	List   *specResourceOperation
	Post   *specResourceOperation
	Get    *specResourceOperation
	Patch  *specResourceOperation
	Delete *specResourceOperation
}

// specResourceOperation defines a resource operation
type specResourceOperation struct {
	SecuritySchemes  SpecSecuritySchemes
	HeaderParameters SpecHeaderParameters
	responses        specResponses
}

// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "cellar": Application User Types
//
// Command:
// $ goagen
// --design=github.com/aep-dev/terraform-provider-openapi/examples/goa/api/design
// --out=$(GOPATH)/src/github.com/aep-dev/terraform-provider-openapi/examples/goa/api
// --version=v1.3.1

package app

import (
	"github.com/goadesign/goa"
	"unicode/utf8"
)

// BottlePayload is the type used to create bottles
type bottlePayload struct {
	// Unique bottle ID
	ID *string `form:"id,omitempty" json:"id,omitempty" yaml:"id,omitempty" xml:"id,omitempty"`
	// Name of bottle
	Name *string `form:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
	// Rating of bottle
	Rating *int `form:"rating,omitempty" json:"rating,omitempty" yaml:"rating,omitempty" xml:"rating,omitempty"`
	// Vintage of bottle
	Vintage *int `form:"vintage,omitempty" json:"vintage,omitempty" yaml:"vintage,omitempty" xml:"vintage,omitempty"`
}

// Validate validates the bottlePayload type instance.
func (ut *bottlePayload) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "name"))
	}
	if ut.Vintage == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "vintage"))
	}
	if ut.Rating == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "rating"))
	}
	if ut.Name != nil {
		if utf8.RuneCountInString(*ut.Name) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError(`request.name`, *ut.Name, utf8.RuneCountInString(*ut.Name), 1, true))
		}
	}
	if ut.Rating != nil {
		if *ut.Rating < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError(`request.rating`, *ut.Rating, 1, true))
		}
	}
	if ut.Rating != nil {
		if *ut.Rating > 5 {
			err = goa.MergeErrors(err, goa.InvalidRangeError(`request.rating`, *ut.Rating, 5, false))
		}
	}
	if ut.Vintage != nil {
		if *ut.Vintage < 1900 {
			err = goa.MergeErrors(err, goa.InvalidRangeError(`request.vintage`, *ut.Vintage, 1900, true))
		}
	}
	return
}

// Publicize creates BottlePayload from bottlePayload
func (ut *bottlePayload) Publicize() *BottlePayload {
	var pub BottlePayload
	if ut.ID != nil {
		pub.ID = ut.ID
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	if ut.Rating != nil {
		pub.Rating = *ut.Rating
	}
	if ut.Vintage != nil {
		pub.Vintage = *ut.Vintage
	}
	return &pub
}

// BottlePayload is the type used to create bottles
type BottlePayload struct {
	// Unique bottle ID
	ID *string `form:"id,omitempty" json:"id,omitempty" yaml:"id,omitempty" xml:"id,omitempty"`
	// Name of bottle
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
	// Rating of bottle
	Rating int `form:"rating" json:"rating" yaml:"rating" xml:"rating"`
	// Vintage of bottle
	Vintage int `form:"vintage" json:"vintage" yaml:"vintage" xml:"vintage"`
}

// Validate validates the BottlePayload type instance.
func (ut *BottlePayload) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "name"))
	}

	if utf8.RuneCountInString(ut.Name) < 1 {
		err = goa.MergeErrors(err, goa.InvalidLengthError(`type.name`, ut.Name, utf8.RuneCountInString(ut.Name), 1, true))
	}
	if ut.Rating < 1 {
		err = goa.MergeErrors(err, goa.InvalidRangeError(`type.rating`, ut.Rating, 1, true))
	}
	if ut.Rating > 5 {
		err = goa.MergeErrors(err, goa.InvalidRangeError(`type.rating`, ut.Rating, 5, false))
	}
	if ut.Vintage < 1900 {
		err = goa.MergeErrors(err, goa.InvalidRangeError(`type.vintage`, ut.Vintage, 1900, true))
	}
	return
}

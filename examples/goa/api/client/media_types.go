// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "cellar": Application Media Types
//
// Command:
// $ goagen
// --design=github.com/aep-dev/terraform-provider-openapi/examples/goa/api/design
// --out=$(GOPATH)/src/github.com/aep-dev/terraform-provider-openapi/examples/goa/api
// --version=v1.3.1

package client

import (
	"github.com/goadesign/goa"
	"net/http"
	"unicode/utf8"
)

// DecodeErrorResponse decodes the ErrorResponse instance encoded in resp body.
func (c *Client) DecodeErrorResponse(resp *http.Response) (*goa.ErrorResponse, error) {
	var decoded goa.ErrorResponse
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// bottle media type (default view)
//
// Identifier: application/vnd.gophercon.goa.bottle; view=default
type Bottle struct {
	// Unique bottle ID
	ID string `form:"id" json:"id" yaml:"id" xml:"id"`
	// Name of bottle
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
	// Rating of bottle
	Rating int `form:"rating" json:"rating" yaml:"rating" xml:"rating"`
	// Vintage of bottle
	Vintage int `form:"vintage" json:"vintage" yaml:"vintage" xml:"vintage"`
}

// Validate validates the Bottle media type instance.
func (mt *Bottle) Validate() (err error) {
	if mt.ID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "id"))
	}
	if mt.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}

	if utf8.RuneCountInString(mt.Name) < 1 {
		err = goa.MergeErrors(err, goa.InvalidLengthError(`response.name`, mt.Name, utf8.RuneCountInString(mt.Name), 1, true))
	}
	if mt.Rating < 1 {
		err = goa.MergeErrors(err, goa.InvalidRangeError(`response.rating`, mt.Rating, 1, true))
	}
	if mt.Rating > 5 {
		err = goa.MergeErrors(err, goa.InvalidRangeError(`response.rating`, mt.Rating, 5, false))
	}
	if mt.Vintage < 1900 {
		err = goa.MergeErrors(err, goa.InvalidRangeError(`response.vintage`, mt.Vintage, 1900, true))
	}
	return
}

// DecodeBottle decodes the Bottle instance encoded in resp body.
func (c *Client) DecodeBottle(resp *http.Response) (*Bottle, error) {
	var decoded Bottle
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

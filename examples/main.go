// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/aep-dev/aep-lib-go/pkg/client"
	"github.com/aep-dev/terraform-provider-aep/config"
	"github.com/aep-dev/terraform-provider-aep/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	// Setup config and HTTP Client.
	config := config.NewProviderConfig()
	c := client.NewClient(http.DefaultClient)

	c.RequestLoggingFunction = func(ctx context.Context, req *http.Request, args ...any) {
		tflog.Info(ctx, fmt.Sprintf("Sending %s request to %s", req.Method, req.URL))
	}

	c.ResponseLoggingFunction = func(ctx context.Context, resp *http.Response, args ...any) {}

	// Create the AEP-Terraform provider.
	p, err := provider.NewProvider(&config, c, version)
	if err != nil {
		log.Fatal(err.Error())
	}

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/hashicorp/scaffolding",
		Debug:   debug,
	}

	// Serve the provider.
	err = providerserver.Serve(context.Background(), p.Provider, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}

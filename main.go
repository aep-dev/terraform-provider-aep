// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/aep-dev/aep-lib-go/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-provider-scaffolding/internal/provider"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version    string = "dev"
	path       string = "https://raw.githubusercontent.com/Roblox/creator-docs/refs/heads/main/content/en-us/reference/cloud/cloud.docs.json"
	pathPrefix string = "/cloud/v2"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		// Also update the tfplugindocs generate command to either remove the
		// -provider-name flag or set its value to the updated provider name.
		Address: "registry.terraform.io/hashicorp/scaffolding",
		Debug:   debug,
	}

	gen, err := provider.CreateGeneratedProviderData(path, pathPrefix)
	if err != nil {
		log.Fatal(err.Error())
	}

	c := client.NewClient(http.DefaultClient)

	err = providerserver.Serve(context.Background(), provider.New(version, gen, c, provider.NewProviderConfig()), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}

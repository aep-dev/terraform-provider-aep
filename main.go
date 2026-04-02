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
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	cfg := config.NewProviderConfig()
	c := client.NewClient(http.DefaultClient)

	c.RequestLoggingFunction = func(ctx context.Context, req *http.Request, args ...any) {
		tflog.Info(ctx, fmt.Sprintf("Sending %s request to %s", req.Method, req.URL))
	}

	c.ResponseLoggingFunction = func(ctx context.Context, resp *http.Response, args ...any) {}

	p, err := provider.NewProvider(&cfg, c, version)
	if err != nil {
		log.Fatal(err.Error())
	}

	opts := providerserver.ServeOpts{
		Address: cfg.RegistryURL,
		Debug:   debug,
	}

	err = providerserver.Serve(context.Background(), p.Provider, opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}

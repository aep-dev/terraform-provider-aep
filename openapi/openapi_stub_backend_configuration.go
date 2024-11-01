package openapi

import (
	"fmt"

	"github.com/aep-dev/terraform-provider-aep/openapi/openapiutils"
)

type specStubBackendConfiguration struct {
	host             string
	basePath         string
	httpScheme       string
	regions          []string
	err              error
	hostErr          error
	defaultRegionErr error
	hostByRegionErr  error

	getHTTPSchemeBehavior func() (string, error)
}

func newStubBackendConfiguration(host, basePath string, httpScheme string) *specStubBackendConfiguration {
	return &specStubBackendConfiguration{
		host:       host,
		basePath:   basePath,
		httpScheme: httpScheme,
	}
}

func newStubBackendMultiRegionConfiguration(host string, regions []string) *specStubBackendConfiguration {
	isMultiRegion, _ := openapiutils.IsMultiRegionHost(host)
	if !isMultiRegion {
		return nil
	}
	return &specStubBackendConfiguration{
		host:    host,
		regions: regions,
	}
}

func (s *specStubBackendConfiguration) getHost() (string, error) {
	if s.hostErr != nil {
		return "", s.hostErr
	}
	return s.host, nil
}
func (s *specStubBackendConfiguration) getBasePath() string {
	return s.basePath
}

func (s *specStubBackendConfiguration) getHTTPScheme() (string, error) {
	if s.getHTTPSchemeBehavior != nil {
		return s.getHTTPSchemeBehavior()
	}
	return s.httpScheme, nil
}

func (s *specStubBackendConfiguration) getHostByRegion(region string) (string, error) {
	if s.hostByRegionErr != nil {
		return "", s.hostByRegionErr
	}
	return fmt.Sprintf(s.host, region), nil
}

func (s *specStubBackendConfiguration) GetDefaultRegion(regions []string) (string, error) {
	if s.defaultRegionErr != nil {
		return "", s.defaultRegionErr
	}
	if len(regions) == 0 {
		return "", fmt.Errorf("empty regions provided")
	}
	return s.regions[0], nil
}

func (s *specStubBackendConfiguration) IsMultiRegion() (bool, string, []string, error) {
	if s.err != nil {
		return false, "", nil, s.err
	}
	if len(s.regions) > 0 {
		return true, s.host, s.regions, nil
	}
	return false, "", nil, nil
}

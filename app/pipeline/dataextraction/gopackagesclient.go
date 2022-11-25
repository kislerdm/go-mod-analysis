package main

import (
	"io"
	"net/http"
)

type httpClient interface {
	Get(url string) (*http.Response, error)
}

// GoPackagesClient client to extract data from https://pkg.go.dev.
type GoPackagesClient struct {
	HTTPClient httpClient
}

// ModuleImports contains the modules imported by the given module.
type ModuleImports struct {
	Std    []string
	NonStd []string
}

// GetImports extracts the modules imported by the given module identified by the name.
func (c GoPackagesClient) GetImports(name string) (ModuleImports, error) {
	b, err := c.get(name + "?tag=imports")
	if err != nil {
		return ModuleImports{}, err
	}
	return parseHTMLGoPackageImports(b)
}

func parseHTMLGoPackageImports(b []byte) (ModuleImports, error) {
	panic("todo")
}

// ModuleImportedBy contains the modules which import the given module.
type ModuleImportedBy []string

// GetImportedBy extracts the modules importing the given module identified by the name.
func (c GoPackagesClient) GetImportedBy(name string) (ModuleImportedBy, error) {
	b, err := c.get(name + "?tag=importedby")
	if err != nil {
		return ModuleImportedBy{}, err
	}
	return parseHTMLGoPackageImportedBy(b)
}

func parseHTMLGoPackageImportedBy(b []byte) (ModuleImportedBy, error) {
	panic("todo")
}

func (c GoPackagesClient) get(route string) ([]byte, error) {
	const URL = "https://pkg.go.dev"
	res, err := c.HTTPClient.Get(URL + "/" + route)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 209 {
		panic("to handle status code beyond 209")
	}

	defer func() { _ = res.Body.Close() }()

	return io.ReadAll(res.Body)
}

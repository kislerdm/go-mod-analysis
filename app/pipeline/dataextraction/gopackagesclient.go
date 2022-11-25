package main

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
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
	r, err := c.get(name + "?tag=imports")
	if err != nil {
		return ModuleImports{}, err
	}
	return parseHTMLGoPackageImports(r)
}

func parseHTMLGoPackageImports(r io.Reader) (ModuleImports, error) {
	panic("todo")
}

// ModuleImportedBy contains the modules which import the given module.
type ModuleImportedBy []string

// GetImportedBy extracts the modules importing the given module identified by the name.
func (c GoPackagesClient) GetImportedBy(name string) (ModuleImportedBy, error) {
	r, err := c.get(name + "?tag=importedby")
	if err != nil {
		return ModuleImportedBy{}, err
	}
	return parseHTMLGoPackageImportedBy(r)
}

func parseHTMLGoPackageImportedBy(r io.Reader) (ModuleImportedBy, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var o ModuleImportedBy

	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "u-breakWord" {
					for _, a := range n.Attr {
						if a.Key == "href" {
							o = append(o, strings.TrimPrefix(a.Val, "/"))
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	return o, nil
}

func (c GoPackagesClient) get(route string) (io.Reader, error) {
	const URL = "https://pkg.go.dev"
	res, err := c.HTTPClient.Get(URL + "/" + route)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 209 {
		return nil, errors.New(res.Status + "; status code: " + strconv.Itoa(res.StatusCode))
	}

	return res.Body, nil
}

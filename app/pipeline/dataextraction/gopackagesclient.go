package dataextraction

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type ErrGoPackageClient struct {
	StatusCode int
	Msg        string
}

func (e ErrGoPackageClient) Error() string {
	return "[StatusCode:" + strconv.Itoa(e.StatusCode) + "] " + e.Msg
}

type HttpClient interface {
	Get(url string) (*http.Response, error)
}

// GoPackagesClient client to extract data from https://pkg.go.dev.
type GoPackagesClient struct {
	HTTPClient HttpClient
}

// ModuleImports contains the modules imported by the given module.
type ModuleImports struct {
	Std    []string
	NonStd []string
}

// GetImports extracts the modules imported by the given module identified by the name.
// The name with version concatenated with the @ sign is acceptable: {{name}}@{{version}}
func (c GoPackagesClient) GetImports(name string) (ModuleImports, error) {
	r, err := c.get(name + "?tag=imports")
	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()
	if err != nil {
		return ModuleImports{}, err
	}
	o, err := parseHTMLGoPackageImports(r)
	if err != nil {
		return ModuleImports{}, ErrGoPackageClient{
			StatusCode: 0,
			Msg:        err.Error(),
		}
	}
	return o, nil
}

func parseHTMLGoPackageImports(r io.ReadCloser) (ModuleImports, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return ModuleImports{}, err
	}

	var (
		f           func(*html.Node)
		ulNodesList []*html.Node
	)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "ul" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "Imports-list" {
					ulNodesList = append(ulNodesList, n)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var scanUlNodes func(n *html.Node) []string
	scanUlNodes = func(n *html.Node) []string {
		var (
			o []string
			f func(n *html.Node)
		)

		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "href" {
						o = append(o, strings.TrimPrefix(a.Val, "/"))
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(n)

		return o
	}

	var o ModuleImports
	switch len(ulNodesList) {
	case 1:
		o.Std = scanUlNodes(ulNodesList[0])
	case 2:
		o.NonStd = scanUlNodes(ulNodesList[0])
		o.Std = scanUlNodes(ulNodesList[1])
	default:
		return ModuleImports{}, errors.New("unknown HTML content")
	}

	return o, nil
}

// ModuleImportedBy contains the modules which import the given module.
type ModuleImportedBy []string

// GetImportedBy extracts the modules importing the given module identified by the name.
// The name with version concatenated with the @ sign is acceptable: {{name}}@{{version}}
func (c GoPackagesClient) GetImportedBy(name string) (ModuleImportedBy, error) {
	r, err := c.get(name + "?tag=importedby")
	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()
	if err != nil {
		return ModuleImportedBy{}, err
	}
	o, err := parseHTMLGoPackageImportedBy(r)
	if err != nil {
		return ModuleImportedBy{}, ErrGoPackageClient{
			StatusCode: 0,
			Msg:        err.Error(),
		}
	}
	return o, nil
}

func parseHTMLGoPackageImportedBy(r io.ReadCloser) (ModuleImportedBy, error) {
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

type Meta struct {
	Version                    string
	License                    string
	Repository                 string
	IsModule                   bool
	IsLatestVersion            bool
	IsValidGoMod               bool
	WithRedistributableLicense bool
	IsTaggedVersion            bool
	IsStableVersion            bool
}

// GetMeta extracts the module's metadata:
func (c GoPackagesClient) GetMeta(name string) (Meta, error) {
	r, err := c.get(name)
	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()
	if err != nil {
		return Meta{}, err
	}

	o, err := parseHTMLGoPackageMain(r)
	if err != nil {
		return Meta{}, ErrGoPackageClient{
			StatusCode: 0,
			Msg:        err.Error(),
		}
	}
	return o, nil
}

func parseHTMLGoPackageMain(r io.ReadCloser) (Meta, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return Meta{}, err
	}

	var (
		o              Meta
		f              func(*html.Node)
		cntFlagSummary uint8
	)

	extractFlagFromSummary := func(n *html.Node) bool {
		if n.NextSibling.Data == "img" {
			for _, a := range n.NextSibling.Attr {
				if a.Key == "alt" {
					return a.Val == "checked"
				}
			}
		}
		return false
	}

	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				for _, a := range n.Attr {
					if a.Key == "data-test-id" && a.Val == "UnitHeader-license" {
						o.License = n.LastChild.Data
					}
					if a.Key == "href" && a.Val == "?tab=versions" {
						o.Version = "v" + strings.Split(n.LastChild.Data, "v")[1]
					}
				}
			case "span":
				for _, a := range n.Attr {
					if a.Key == "class" {
						if a.Val == "go-Chip DetailsHeader-span--latest" {
							o.IsLatestVersion = n.LastChild.Data == "Latest"
						}
						if a.Val == "go-Chip go-Chip--inverted" && n.LastChild.Data == "module" {
							o.IsModule = true
						}
					}
				}
			case "summary":
				for _, a := range n.Attr {
					if a.Key == "class" && a.Val == "go-textSubtle" {
						switch cntFlagSummary {
						case 0:
							o.IsValidGoMod = extractFlagFromSummary(n.FirstChild)
						case 1:
							o.WithRedistributableLicense = extractFlagFromSummary(n.FirstChild)
						case 2:
							o.IsTaggedVersion = extractFlagFromSummary(n.FirstChild)
						case 3:
							o.IsStableVersion = extractFlagFromSummary(n.FirstChild)
						}
						cntFlagSummary++
					}
				}
			case "div":
				for _, a := range n.Attr {
					if a.Key == "class" && a.Val == "UnitMeta-repo" {
						for _, aa := range n.FirstChild.NextSibling.Attr {
							if aa.Key == "href" {
								o.Repository = aa.Val
							}
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

func (c GoPackagesClient) get(route string) (io.ReadCloser, error) {
	const URL = "https://pkg.go.dev"
	res, err := c.HTTPClient.Get(URL + "/" + route)
	if err != nil {
		return nil, ErrGoPackageClient{
			StatusCode: -1,
			Msg:        err.Error(),
		}
	}

	if res.StatusCode > 209 {
		return res.Body, ErrGoPackageClient{
			StatusCode: res.StatusCode,
			Msg:        res.Status,
		}
	}

	return res.Body, nil
}

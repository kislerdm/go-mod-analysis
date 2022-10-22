package gomodanalysis

import (
	"golang.org/x/net/html"
	"io"
	"strings"
)

// SearchResultHit a single search result.
// https://github.com/search?type=Code...
type SearchResultHit struct {
	// "/owner/repo"
	Repo     string `json:"repo"`
	FilePath string `json:"file_path"`
}

func readHref(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "href" {
			return a.Val
		}
	}
	return ""
}

// ParseHtmlSearchResults parses html with github search results.
// https://github.com/search?o=desc&q=module+extension%3Amod+language%3AText&s=indexed&type=Code
func ParseHtmlSearchResults(v io.Reader) ([]SearchResultHit, error) {
	var o []SearchResultHit

	doc, err := html.Parse(v)
	if err != nil {
		return o, err
	}

	var f func(n *html.Node, o []SearchResultHit)
	f = func(n *html.Node, o []SearchResultHit) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "title" && strings.HasSuffix(a.Val, ".mod") {
					href := readHref(n)

					o = append(o, SearchResultHit{
						Repo:     strings.Join(strings.Split(href, "/")[:3], "/"),
						FilePath: strings.Replace(href, "blob/", "", -1),
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, o)
		}
	}

	f(doc, o)

	return o, nil
}

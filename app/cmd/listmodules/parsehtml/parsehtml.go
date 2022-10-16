package parsehtml

import (
	"golang.org/x/net/html"
	"io"
	"strings"
)

type HTMLCodeSearchContent struct {
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

func ParseHtml(v io.Reader) ([]HTMLCodeSearchContent, error) {
	o := []HTMLCodeSearchContent{}

	doc, err := html.Parse(v)
	if err != nil {
		return o, err
	}

	var f func(n *html.Node, o []HTMLCodeSearchContent)
	f = func(n *html.Node, o []HTMLCodeSearchContent) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "title" && strings.HasSuffix(a.Val, ".mod") {
					href := readHref(n)

					o = append(o, HTMLCodeSearchContent{
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

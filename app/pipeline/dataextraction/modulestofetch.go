package dataextraction

import "github.com/kislerdm/gomodanalysis/app/pipeline"

type Module struct {
	Name, Version string
}

// ListModulesToFetch lists modules to extract data for.
func ListModulesToFetch(client pipeline.GBQClient) ([]Module, error) {
	var o []Module
	o = []Module{
		{
			Name:    "https://pkg.go.dev/github.com/spf13/cobra",
			Version: "v1.6.1",
		},
	}
	return o, nil
}

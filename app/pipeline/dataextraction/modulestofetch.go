package dataextraction

import "github.com/kislerdm/gomodanalysis/app/pipeline"

type Module struct {
	Name, Version string
}

// ListModulesToFetch lists modules to extract data for.
func ListModulesToFetch(client pipeline.GBQClient) ([]Module, error) {
	var o []Module
	return o, nil
}

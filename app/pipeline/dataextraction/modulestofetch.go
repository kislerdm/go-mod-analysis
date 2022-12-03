package dataextraction

import (
	"context"
	"errors"
	"strconv"

	"github.com/kislerdm/gomodanalysis/app/pipeline"
)

type Module struct {
	Name, Version string
}

// ListModulesToFetch lists modules to extract data for.
func ListModulesToFetch(ctx context.Context, client pipeline.GBQClient) ([]Module, error) {
	const q = "SELECT DISTINCT a.path " +
		"FROM `go-mod-analysis.raw.index` AS a " +
		"LEFT JOIN `go-mod-analysis.raw.pkggodev` AS b USING (path) " +
		"WHERE b.path IS NULL LIMIT 2;"

	r, err := client.Read(ctx, q)
	if err != nil {
		return nil, err
	}

	var o []Module

	for i, row := range r {
		v, ok := row[0].(string)
		if !ok {
			return nil, errors.New("ListModulesToFetch(): cannot parse values of row " + strconv.Itoa(i))
		}
		o = append(
			o, Module{Name: v},
		)
	}

	return o, nil
}

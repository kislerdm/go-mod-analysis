package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/kislerdm/gomodanalysis/app/pipeline/dataextraction/model"
)

type PkgData struct {
	path       string
	meta       Meta
	imports    ModuleImports
	importedBy ModuleImportedBy
}

func (d PkgData) ToGBQ() *model.PkgGoDev {
	return &model.PkgGoDev{
		Path:    d.path,
		Version: d.meta.Version,
		Meta: &model.PkgGoDev_Meta{
			Licence:                    d.meta.License,
			Repository:                 d.meta.Repository,
			IsModule:                   d.meta.IsModule,
			IsLatestVersion:            d.meta.IsLatestVersion,
			IsValidGoMod:               d.meta.IsValidGoMod,
			WithRedistributableLicense: d.meta.WithRedistributableLicense,
			IsTaggedVersion:            d.meta.IsTaggedVersion,
			IsStableVersion:            d.meta.IsStableVersion,
		},
		Imports: &model.PkgGoDev_Imports{
			Std:    d.imports.Std,
			Nonstd: d.imports.NonStd,
		},
		Importedby: d.importedBy,
		Timestamp:  time.Now().UTC().UnixMilli(),
	}
}

type errArr []error

func (e errArr) Error() string {
	o := ""
	for _, err := range e {
		if err != nil {
			o += err.Error() + "\n"
		}
	}
	return o
}

// ExtractGoPkgData extracts module's data from https://pkg.go.dev
func ExtractGoPkgData(name, version string, httpClient HttpClient) (PkgData, error) {
	if version != "" {
		name += "@" + version
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	c := GoPackagesClient{httpClient}

	var wg sync.WaitGroup
	wg.Add(3)
	errs := make(chan error, 3)

	o := PkgData{path: name}

	go func(name string, wg *sync.WaitGroup, e chan error, o *PkgData) {
		defer wg.Done()
		var err error
		o.meta, err = c.GetMeta(name)
		e <- err
	}(name, &wg, errs, &o)

	go func(name string, wg *sync.WaitGroup, e chan error, o *PkgData) {
		defer wg.Done()
		var err error
		o.imports, err = c.GetImports(name)
		e <- err
	}(name, &wg, errs, &o)

	go func(name string, wg *sync.WaitGroup, e chan error, o *PkgData) {
		defer wg.Done()
		var err error
		o.importedBy, err = c.GetImportedBy(name)
		e <- err
	}(name, &wg, errs, &o)

	wg.Wait()

	//var ee errArr
	//for e := range errs {
	//	if e != nil {
	//		ee = append(ee, e)
	//	}
	//}

	return o, nil
}

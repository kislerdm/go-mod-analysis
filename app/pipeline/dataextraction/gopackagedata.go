package dataextraction

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

const (
	errPkgTypeMain       = "pkg.go.dev/main"
	errPkgTypeImports    = "pkg.go.dev/imports"
	errPkgTypeImportedBy = "pkg.go.dev/importedby"
)

type errPkg struct {
	v map[string]ErrGoPackageClient
	m *sync.Mutex
}

func (e errPkg) Add(t string, err error) {
	e.m.Lock()
	er, ok := err.(ErrGoPackageClient)
	if !ok {
		panic("errPkg.Add(): wrong error type")
	}
	e.v[t] = er
	e.m.Unlock()
}

func (e errPkg) Error() string {
	o := ""
	for k, v := range e.v {
		o += "[Type:" + k + "]" + v.Error() + "\n"
	}
	return o
}

func (e errPkg) IsNil() bool {
	return len(e.v) == 0
}

// ExtractGoPkgData extracts module's data from https://pkg.go.dev
func ExtractGoPkgData(name, version string, c *GoPackagesClient) (PkgData, error) {
	if version != "" {
		name += "@" + version
	}

	if c == nil {
		c = &GoPackagesClient{&http.Client{Timeout: 60 * time.Second}}
	}

	var wg sync.WaitGroup
	wg.Add(3)
	errs := errPkg{
		v: map[string]ErrGoPackageClient{},
		m: &sync.Mutex{},
	}

	o := PkgData{path: name}

	go func(name string, wg *sync.WaitGroup, e errPkg, o *PkgData) {
		defer wg.Done()
		var err error
		o.meta, err = c.GetMeta(name)
		if err != nil {
			errs.Add(errPkgTypeMain, err)
		}
	}(name, &wg, errs, &o)

	go func(name string, wg *sync.WaitGroup, e errPkg, o *PkgData) {
		defer wg.Done()
		var err error
		o.imports, err = c.GetImports(name)
		if err != nil {
			errs.Add(errPkgTypeImports, err)
		}
	}(name, &wg, errs, &o)

	go func(name string, wg *sync.WaitGroup, e errPkg, o *PkgData) {
		defer wg.Done()
		var err error
		o.importedBy, err = c.GetImportedBy(name)
		if err != nil {
			errs.Add(errPkgTypeImportedBy, err)
		}
	}(name, &wg, errs, &o)

	wg.Wait()

	if errs.IsNil() {
		return o, nil
	}

	return PkgData{}, errs
}

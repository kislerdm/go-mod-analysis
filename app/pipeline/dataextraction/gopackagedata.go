package dataextraction

import (
	"net/http"
	"sync"
	"time"

	"cloud.google.com/go/bigquery/storage/managedwriter/adapt"
	"github.com/kislerdm/gomodanalysis/app/pipeline/dataextraction/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type PkgData struct {
	path       string
	meta       Meta
	imports    ModuleImports
	importedBy ModuleImportedBy
}

func (d PkgData) Descriptor() *descriptorpb.DescriptorProto {
	m := &model.PkgGoDev{}
	descriptorProto, err := adapt.NormalizeDescriptor(m.ProtoReflect().Descriptor())
	if err != nil {
		panic("PkgData.Descriptor() error: " + err.Error())
	}
	return descriptorProto
}

func (d PkgData) Data() [][]byte {
	b, err := proto.Marshal(
		&model.PkgGoDev{
			Path:    d.path,
			Version: d.meta.Version,
			Meta: &model.PkgGoDev_Meta{
				License:                    d.meta.License,
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
			Timestamp:  time.Now().UTC().UnixMicro(),
		},
	)
	if err != nil {
		panic("PkgData.Data() error: " + err.Error())
	}
	return [][]byte{b}
}

const (
	errPkgTypeMain       = "pkg.go.dev/main"
	errPkgTypeImports    = "pkg.go.dev/imports"
	errPkgTypeImportedBy = "pkg.go.dev/importedby"
)

// ErrExtractGoPkgData error returned by ExtractGoPkgData
type ErrExtractGoPkgData struct {
	v map[string]ErrGoPackageClient
	m *sync.Mutex
}

func (e ErrExtractGoPkgData) Add(t string, err error) {
	e.m.Lock()
	er, ok := err.(ErrGoPackageClient)
	if !ok {
		panic("ErrExtractGoPkgData.Add(): wrong error type")
	}
	e.v[t] = er
	e.m.Unlock()
}

func (e ErrExtractGoPkgData) Error() string {
	o := ""
	for k, v := range e.v {
		o += "[Type:" + k + "]" + v.Error() + "\n"
	}
	return o
}

func (e ErrExtractGoPkgData) IsNil() bool {
	return len(e.v) == 0
}

func (e ErrExtractGoPkgData) IsHTTPStatus(status int) bool {
	for _, err := range e.v {
		if err.StatusCode == status {
			return true
		}
	}
	return false
}

// ExtractGoPkgData extracts module's data from https://pkg.go.dev
func ExtractGoPkgData(name, version string, c *GoPackagesClient) (PkgData, error) {
	o := PkgData{path: name}

	if version != "" {
		name += "@" + version
	}

	if c == nil {
		c = NewGoPackagesClient(&http.Client{Timeout: 60 * time.Second}, 30)
	}

	var wg sync.WaitGroup
	wg.Add(3)
	errs := ErrExtractGoPkgData{
		v: map[string]ErrGoPackageClient{},
		m: &sync.Mutex{},
	}

	go func(name string, wg *sync.WaitGroup, o *PkgData) {
		defer wg.Done()
		var err error
		o.meta, err = c.GetMeta(name)
		if err != nil {
			errs.Add(errPkgTypeMain, err)
		}
	}(name, &wg, &o)

	go func(name string, wg *sync.WaitGroup, o *PkgData) {
		defer wg.Done()
		var err error
		o.imports, err = c.GetImports(name)
		if err != nil {
			errs.Add(errPkgTypeImports, err)
		}
	}(name, &wg, &o)

	go func(name string, wg *sync.WaitGroup, o *PkgData) {
		defer wg.Done()
		var err error
		o.importedBy, err = c.GetImportedBy(name)
		if err != nil {
			errs.Add(errPkgTypeImportedBy, err)
		}
	}(name, &wg, &o)

	wg.Wait()

	if errs.IsNil() {
		return o, nil
	}

	return PkgData{path: name}, errs
}

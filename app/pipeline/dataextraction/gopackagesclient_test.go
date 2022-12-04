package dataextraction

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

//go:embed fixtures
var fixtures embed.FS

type mockHTTP struct{}

func (c mockHTTP) Get(s string) (*http.Response, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	p := "main"
	if len(u.Query()["tab"]) > 0 {
		p = u.Query()["tab"][0]
	}

	if len(u.Query()["mimic_429"]) > 0 {
		return &http.Response{
			Status:     "Mimic Too Many Requests",
			StatusCode: http.StatusTooManyRequests,
		}, nil
	}

	b, err := fixtures.ReadFile("fixtures" + u.Path + "/" + p + ".html")
	switch err.(type) {
	case nil:
		return &http.Response{
			Status:     "OK",
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(b)),
		}, nil
	case *fs.PathError:
		return &http.Response{
			Status:     "Not Found",
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
		}, nil
	default:
		return &http.Response{
			Status:     err.Error(),
			StatusCode: http.StatusInternalServerError,
		}, nil
	}
}

func TestGoPackagesClient_get(t *testing.T) {
	wantImports, err := fixtures.ReadFile("fixtures/go-dockerclient/imports.html")
	if err != nil {
		panic(err)
	}

	wantImportedBy, err := fixtures.ReadFile("fixtures/go-dockerclient/importedby.html")
	if err != nil {
		panic(err)
	}

	type fields struct {
		HTTPClient    HttpClient
		maxBackoffSec int8
	}
	type args struct {
		route string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    io.Reader
		wantErr bool
	}{
		{
			name:   "go-dockerclient: imports",
			fields: fields{HTTPClient: &mockHTTP{}, maxBackoffSec: 1},
			args: args{
				route: "go-dockerclient?tab=imports",
			},
			want:    io.NopCloser(bytes.NewReader(wantImports)),
			wantErr: false,
		},
		{
			name:   "go-dockerclient: importedby",
			fields: fields{HTTPClient: &mockHTTP{}, maxBackoffSec: 1},
			args: args{
				route: "go-dockerclient?tab=importedby",
			},
			want:    io.NopCloser(bytes.NewReader(wantImportedBy)),
			wantErr: false,
		},
		{
			name:   "not-found-package: importedby",
			fields: fields{HTTPClient: &mockHTTP{}, maxBackoffSec: 1},
			args: args{
				route: "not-found-package?tab=importedby",
			},
			want:    io.NopCloser(strings.NewReader("not found")),
			wantErr: true,
		},
		{
			name:   "unhappy path: faulty url",
			fields: fields{HTTPClient: &mockHTTP{}, maxBackoffSec: 1},
			args: args{
				route: "#%%%",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "unhappy path: 429",
			fields: fields{HTTPClient: &mockHTTP{}, maxBackoffSec: 1},
			args: args{
				route: "?mimic_429",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := NewGoPackagesClient(tt.fields.HTTPClient, tt.fields.maxBackoffSec)
				got, err := c.get(tt.args.route)
				if (err != nil) != tt.wantErr {
					t.Errorf("get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("get() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

//go:embed fixtures/bar/importedby.html
var wantImportedBy []byte

func Test_parseHTMLGoPackageImportedBy(t *testing.T) {
	type args struct {
		r io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		want    ModuleImportedBy
		wantErr bool
	}{
		{
			name: "happy path: 8 packages",
			args: args{io.NopCloser(bytes.NewReader(wantImportedBy))},
			want: ModuleImportedBy{
				"bitbucket.org/blackxcloudeng/infra/common/docker",
				"bitbucket.org/blackxcloudeng/infra/prog/weaver",
				"bitbucket.org/blackxcloudeng/infra/prog/weaveutil",
				"bitbucket.org/blackxcloudeng/infra/proxy",
				"bitbucket.org/blackxcloudeng/podman-client/testing",
				"bitbucket.org/blackxcloudeng/scope/app",
				"bitbucket.org/blackxcloudeng/scope/probe/docker",
				"bldy.build/build/namespace/docker",
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := parseHTMLGoPackageImportedBy(tt.args.r)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseHTMLGoPackageImportedBy() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseHTMLGoPackageImportedBy() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestGoPackagesClient_GetImportedBy(t *testing.T) {
	type fields struct {
		HTTPClient    HttpClient
		maxBackoffSec int8
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ModuleImportedBy
		wantErr bool
	}{
		{
			name:   "happy path: 8 packages",
			fields: fields{mockHTTP{}, 1},
			args:   args{"bar"},
			want: ModuleImportedBy{
				"bitbucket.org/blackxcloudeng/infra/common/docker",
				"bitbucket.org/blackxcloudeng/infra/prog/weaver",
				"bitbucket.org/blackxcloudeng/infra/prog/weaveutil",
				"bitbucket.org/blackxcloudeng/infra/proxy",
				"bitbucket.org/blackxcloudeng/podman-client/testing",
				"bitbucket.org/blackxcloudeng/scope/app",
				"bitbucket.org/blackxcloudeng/scope/probe/docker",
				"bldy.build/build/namespace/docker",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := NewGoPackagesClient(tt.fields.HTTPClient, tt.fields.maxBackoffSec)
				got, err := c.GetImportedBy(tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetImportedBy() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetImportedBy() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

//go:embed fixtures/bar/imports.html
var wantImports []byte

func Test_parseHTMLGoPackageImports(t *testing.T) {
	type args struct {
		r io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		want    ModuleImports
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{io.NopCloser(bytes.NewReader(wantImports))},
			want: ModuleImports{
				Std: []string{
					"bufio",
					"bytes",
					"context",
					"crypto/tls",
					"crypto/x509",
					"encoding/base64",
					"encoding/json",
					"errors",
					"fmt",
					"io",
					"math",
					"net",
					"net/http",
					"net/http/httputil",
					"net/url",
					"os",
					"os/exec",
					"path",
					"path/filepath",
					"reflect",
					"runtime",
					"strconv",
					"strings",
					"sync",
					"sync/atomic",
					"time",
				},
				NonStd: []string{
					"github.com/docker/docker/api/types/registry",
					"github.com/docker/docker/api/types/swarm",
					"github.com/docker/docker/pkg/archive",
					"github.com/docker/docker/pkg/fileutils",
					"github.com/docker/docker/pkg/homedir",
					"github.com/docker/docker/pkg/jsonmessage",
					"github.com/docker/docker/pkg/stdcopy",
					"github.com/docker/go-units",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := parseHTMLGoPackageImports(tt.args.r)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseHTMLGoPackageImports() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseHTMLGoPackageImports() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestGoPackagesClient_GetImports(t *testing.T) {
	type fields struct {
		HTTPClient    HttpClient
		maxBackoffSec int8
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ModuleImports
		wantErr bool
	}{
		{
			name:   "happy path: std only",
			fields: fields{mockHTTP{}, 1},
			args:   args{"foo"},
			want: ModuleImports{
				Std: []string{
					"bufio",
					"bytes",
					"context",
					"crypto/tls",
					"crypto/x509",
					"encoding/base64",
					"encoding/json",
					"errors",
					"fmt",
					"io",
					"math",
					"net",
					"net/http",
					"net/http/httputil",
					"net/url",
					"os",
					"os/exec",
					"path",
					"path/filepath",
					"reflect",
					"runtime",
					"strconv",
					"strings",
					"sync",
					"sync/atomic",
					"time",
				},
			},
			wantErr: false,
		},
		{
			name:   "happy path: std and non-std",
			fields: fields{mockHTTP{}, 1},
			args:   args{"bar"},
			want: ModuleImports{
				Std: []string{
					"bufio",
					"bytes",
					"context",
					"crypto/tls",
					"crypto/x509",
					"encoding/base64",
					"encoding/json",
					"errors",
					"fmt",
					"io",
					"math",
					"net",
					"net/http",
					"net/http/httputil",
					"net/url",
					"os",
					"os/exec",
					"path",
					"path/filepath",
					"reflect",
					"runtime",
					"strconv",
					"strings",
					"sync",
					"sync/atomic",
					"time",
				},
				NonStd: []string{
					"github.com/docker/docker/api/types/registry",
					"github.com/docker/docker/api/types/swarm",
					"github.com/docker/docker/pkg/archive",
					"github.com/docker/docker/pkg/fileutils",
					"github.com/docker/docker/pkg/homedir",
					"github.com/docker/docker/pkg/jsonmessage",
					"github.com/docker/docker/pkg/stdcopy",
					"github.com/docker/go-units",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := NewGoPackagesClient(tt.fields.HTTPClient, tt.fields.maxBackoffSec)
				got, err := c.GetImports(tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetImports() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetImports() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

//go:embed fixtures/go-dockerclient/main.html
var wantMain []byte

func Test_parseHTMLGoPackageMain(t *testing.T) {
	type args struct {
		r io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		want    Meta
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{io.NopCloser(bytes.NewReader(wantMain))},
			want: Meta{
				Version:                    "v1.9.0",
				License:                    "BSD-2-Clause",
				Repository:                 "https://github.com/fsouza/go-dockerclient",
				IsModule:                   true,
				IsLatestVersion:            true,
				IsValidGoMod:               true,
				WithRedistributableLicense: true,
				IsTaggedVersion:            true,
				IsStableVersion:            true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := parseHTMLGoPackageMain(tt.args.r)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseHTMLGoPackageMain() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseHTMLGoPackageMain() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestGoPackagesClient_GetMeta(t *testing.T) {
	type fields struct {
		HTTPClient    HttpClient
		maxBackoffSec int8
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Meta
		wantErr bool
	}{
		{
			name:   "happy path",
			fields: fields{mockHTTP{}, 1},
			args:   args{"go-dockerclient"},
			want: Meta{
				Version:                    "v1.9.0",
				License:                    "BSD-2-Clause",
				Repository:                 "https://github.com/fsouza/go-dockerclient",
				IsModule:                   true,
				IsLatestVersion:            true,
				IsValidGoMod:               true,
				WithRedistributableLicense: true,
				IsTaggedVersion:            true,
				IsStableVersion:            true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := NewGoPackagesClient(tt.fields.HTTPClient, tt.fields.maxBackoffSec)
				got, err := c.GetMeta(tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetMeta() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetMeta() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

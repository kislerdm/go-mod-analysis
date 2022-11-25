package main

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"reflect"
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

	b, err := fixtures.ReadFile("fixtures" + u.Path + "/" + u.Query()["tag"][0] + ".html")
	switch err {
	case nil:
		return &http.Response{
			Status:     "OK",
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(b)),
		}, nil
	case fs.ErrNotExist:
		return &http.Response{
			Status:     "Not Found",
			StatusCode: http.StatusNotFound,
		}, err
	default:
		return &http.Response{
			Status:     "Error",
			StatusCode: http.StatusInternalServerError,
		}, err
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
		HTTPClient httpClient
	}
	type args struct {
		route string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:   "go-dockerclient: imports",
			fields: fields{HTTPClient: &mockHTTP{}},
			args: args{
				route: "go-dockerclient?tag=imports",
			},
			want:    wantImports,
			wantErr: false,
		},
		{
			name:   "go-dockerclient: importedby",
			fields: fields{HTTPClient: &mockHTTP{}},
			args: args{
				route: "go-dockerclient?tag=importedby",
			},
			want:    wantImportedBy,
			wantErr: false,
		},
		{
			name:   "not-found-package: importedby",
			fields: fields{HTTPClient: &mockHTTP{}},
			args: args{
				route: "not-found-package?tag=importedby",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := GoPackagesClient{
					HTTPClient: tt.fields.HTTPClient,
				}
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

//go:embed fixtures/foo/importedby.html
var wantImportedBy []byte

func Test_parseHTMLGoPackageImportedBy(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    ModuleImportedBy
		wantErr bool
	}{
		{
			name: "happy path: 8 packages",
			args: args{bytes.NewReader(wantImportedBy)},
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
		HTTPClient httpClient
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
			fields: fields{mockHTTP{}},
			args:   args{"foo"},
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
				c := GoPackagesClient{
					HTTPClient: tt.fields.HTTPClient,
				}
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

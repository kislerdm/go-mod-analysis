package dataextraction

import (
	"reflect"
	"testing"
)

func TestExtractGoPkgData(t *testing.T) {
	type args struct {
		name    string
		version string
		c       *GoPackagesClient
	}
	tests := []struct {
		name    string
		args    args
		want    PkgData
		wantErr bool
	}{
		{
			name: "happy path: bar",
			args: args{
				name:    "bar",
				version: "",
				c:       NewGoPackagesClient(mockHTTP{}, 1),
			},
			want: PkgData{
				path: "bar",
				meta: Meta{
					Version:                    "v0.1.0",
					License:                    "MIT",
					Repository:                 "https://github.com/bar/bar",
					IsModule:                   true,
					IsLatestVersion:            true,
					IsValidGoMod:               true,
					WithRedistributableLicense: true,
					IsTaggedVersion:            true,
					IsStableVersion:            false,
				},
				imports: ModuleImports{
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
				importedBy: ModuleImportedBy{
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
			wantErr: false,
		},
		{
			name: "happy path: package not found",
			args: args{
				name:    "qux",
				version: "",
				c:       NewGoPackagesClient(mockHTTP{}, 1),
			},
			want: PkgData{
				path:       "qux",
				meta:       Meta{},
				imports:    ModuleImports{},
				importedBy: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ExtractGoPkgData(tt.args.name, tt.args.version, tt.args.c)
				if (err != nil) != tt.wantErr {
					t.Errorf("ExtractGoPkgData() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ExtractGoPkgData() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

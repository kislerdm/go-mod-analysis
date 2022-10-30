package moduleversions

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestMinMaxVersions(t *testing.T) {
	type args struct {
		v []string
	}
	tests := []struct {
		name string
		args args
		want Versions
	}{
		{
			name: "single version",
			args: args{
				[]string{"v1.0.0"},
			},
			want: Versions{
				Min: "v1.0.0",
				Max: "v1.0.0",
			},
		},
		{
			name: "array of version",
			args: args{
				[]string{
					"v1.1.22", "v1.1.19",
					"v1.1.6", "v1.1.17",
					"v1.0.4", "v1.1.5",
					"v1.1.21", "v1.1.13",
					"v1.1.0", "v1.1.23-0.20211004211129-b31d40d9a0be", "v1.1.15",
				},
			},
			want: Versions{
				Min: "v1.0.4",
				Max: "v1.1.23-0.20211004211129-b31d40d9a0be",
			},
		},
		{
			name: "empty input",
			args: args{nil},
			want: Versions{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MinMaxVersions(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MinMaxVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type writer struct {
	statusCode int
	body       []byte
}

func (w writer) Header() http.Header {
	return nil
}

func (w writer) Write(bytes []byte) (int, error) {
	w.body = bytes
	return 0, nil
}

func (w writer) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func Test_handler(t *testing.T) {
	reqInput := func(i input) io.Reader {
		o, err := json.Marshal(i)
		if err != nil {
			panic(err)
		}
		return bytes.NewReader(o)
	}

	respBody := func(o output) []byte {
		return nil
	}

	mustReq := func(r *http.Request, err error) *http.Request {
		if err != nil {
			panic(err)
		}
		return r
	}

	type args struct {
		w   http.ResponseWriter
		req *http.Request
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "happy path",
			args: args{
				w: writer{
					statusCode: http.StatusOK,
					body: respBody(output{Replies: []Versions{
						{Min: "v1.0.0", Max: "v1.0.0"},
						{Min: "v1.1.0", Max: "v1.1.23-0.20211004211129-b31d40d9a0be"},
					}}),
				},
				req: mustReq(
					http.NewRequest("POST", "/",
						reqInput(
							input{
								Calls: [][]string{
									{"v1.0.0"},
									{"v1.1.0", "v1.1.23-0.20211004211129-b31d40d9a0be", "v1.1.15"},
								},
							},
						),
					),
				),
			},
		},
		{
			name: "faulty input json",
			args: args{
				w: writer{
					statusCode: http.StatusBadRequest,
					body:       newError("unsupported in: corrupt JSON"),
				},
				req: mustReq(http.NewRequest("POST", "/", strings.NewReader("{"))),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler(tt.args.w, tt.args.req)
		})
	}
}

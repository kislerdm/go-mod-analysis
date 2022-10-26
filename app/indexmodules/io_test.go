package indexmodules

import (
	"gomodanalysis/indexmodules/model"
	"reflect"
	"testing"
	"time"
)

func tsFromStr(s string) int64 {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t.UnixMicro()
}

func Test_extractStringVal(t *testing.T) {
	type args struct {
		vals []byte
		key  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Path",
			args: args{
				vals: []byte(`Path":"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign","Version":"v1.3.31-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.247192Z"}`),
				key:  "Path",
			},
			want: "github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign",
		},
		{
			name: "Version",
			args: args{
				vals: []byte(`Version":"v1.3.31-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.247192Z"}`),
				key:  "Version",
			},
			want: "v1.3.31-0.20221004181815-30a55eb410bc",
		},
		{
			name: "Timestamp",
			args: args{
				vals: []byte(`Timestamp":"2022-10-23T14:22:05.247192Z"}`),
				key:  "Timestamp",
			},
			want: "2022-10-23T14:22:05.247192Z",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractStringVal(tt.args.vals, tt.args.key); got != tt.want {
				t.Errorf("extractStringVal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawData_Decode(t *testing.T) {
	tests := []struct {
		name    string
		v       RawData
		want    []DataRow
		wantErr bool
	}{
		{
			name: "happy path",
			v: RawData(`{"Path":"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign","Version":"v1.3.31-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.247192Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/inspector2","Version":"v1.8.2-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.499347Z"}`),
			want: []DataRow{
				{
					Path:      "github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign",
					Version:   "v1.3.31-0.20221004181815-30a55eb410bc",
					Timestamp: "2022-10-23T14:22:05.247192Z",
				},
				{
					Path:      "github.com/aws/aws-sdk-go-v2/service/inspector2",
					Version:   "v1.8.2-0.20221004181815-30a55eb410bc",
					Timestamp: "2022-10-23T14:22:05.499347Z",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.Decode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToGBQTableFormat(t *testing.T) {
	type args struct {
		v []DataRow
	}
	tests := []struct {
		name    string
		args    args
		want    []*model.Index
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				v: []DataRow{
					{
						Path:      "github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign",
						Version:   "v1.3.31-0.20221004181815-30a55eb410bc",
						Timestamp: "2022-10-23T14:22:05.247192Z",
					},
					{
						Path:      "github.com/aws/aws-sdk-go-v2/service/inspector2",
						Version:   "v1.8.2-0.20221004181815-30a55eb410bc",
						Timestamp: "2022-10-23T14:22:05.499347Z",
					},
				},
			},
			want: []*model.Index{
				{
					Path:      "github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign",
					Version:   "v1.3.31-0.20221004181815-30a55eb410bc",
					Timestamp: tsFromStr("2022-10-23T14:22:05.247192Z"),
				},
				{
					Path:      "github.com/aws/aws-sdk-go-v2/service/inspector2",
					Version:   "v1.8.2-0.20221004181815-30a55eb410bc",
					Timestamp: tsFromStr("2022-10-23T14:22:05.499347Z"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToGBQTableFormat(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToGBQTableFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToGBQTableFormat() got = %v, want %v", got, tt.want)
			}
		})
	}
}

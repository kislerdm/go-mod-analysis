package main

import "testing"

func Test_extractNextLastTime(t *testing.T) {
	type args struct {
		v []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "happy path",
			args: args{[]byte(`{"Path":"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign","Version":"v1.3.31-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.247192Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/inspector2","Version":"v1.8.2-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.499347Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/frauddetector","Version":"v1.20.10-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.968851Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/mgn","Version":"v1.15.13-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.973197Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/personalizeruntime","Version":"v1.12.6-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:06.629783Z"}`)},
			want: "2022-10-23T14:22:06.629783Z",
		},
		{
			name: "happy path: 5 digits after comma",
			args: args{[]byte(`{"Path":"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign","Version":"v1.3.31-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.247192Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/inspector2","Version":"v1.8.2-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.499347Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/frauddetector","Version":"v1.20.10-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.968851Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/mgn","Version":"v1.15.13-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:05.973197Z"}
{"Path":"github.com/aws/aws-sdk-go-v2/service/personalizeruntime","Version":"v1.12.6-0.20221004181815-30a55eb410bc","Timestamp":"2022-10-23T14:22:06.62983Z"}`)},
			want: "2022-10-23T14:22:06.62983Z",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractNextLastTime(tt.args.v); got != tt.want {
				t.Errorf("extractNextLastTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

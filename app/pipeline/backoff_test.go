package pipeline

import (
	"testing"
	"time"
)

func TestBackoff_LinearDelay(t *testing.T) {
	type fields struct {
		MaxDelay time.Duration
		MaxSteps int64
	}
	tests := []struct {
		name    string
		fields  fields
		tries   int
		want    time.Duration
		wantErr bool
	}{
		{
			name: "happy path: 1 step",
			fields: fields{
				MaxDelay: 1 * time.Second,
				MaxSteps: 1,
			},
			tries:   1,
			want:    1 * time.Second,
			wantErr: false,
		},
		{
			name: "unhappy path",
			fields: fields{
				MaxDelay: 1 * time.Second,
				MaxSteps: 1,
			},
			tries:   2,
			want:    0,
			wantErr: true,
		},
		{
			name: "happy path: 2 step",
			fields: fields{
				MaxDelay: 1 * time.Second,
				MaxSteps: 4,
			},
			tries:   2,
			want:    500 * time.Millisecond,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backoff{
				MaxDelay: tt.fields.MaxDelay,
				MaxSteps: tt.fields.MaxSteps,
			}

			for i := 0; i < tt.tries; i++ {
				b.UpCounter()
			}

			got, err := b.LinearDelay()
			if (err != nil) != tt.wantErr {
				t.Errorf("LinearDelay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LinearDelay() got = %v, want %v", got, tt.want)
			}
		})
	}
}

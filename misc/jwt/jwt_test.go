package jwt

import (
	"testing"

	"github.com/xsuners/mo/metadata"
)

func TestNew(t *testing.T) {
	type args struct {
		md *metadata.Metadata
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "新建JWT令牌",
			args: args{
				md: &metadata.Metadata{
					Appid:  10000,
					Id:     100001,
					Device: 1,
				},
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.md)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

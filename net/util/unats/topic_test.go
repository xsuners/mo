package unats

import "testing"

func TestIPSubject(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "example",
			args: args{
				ip: "127.0.0.1",
			},
			want: "127-0-0-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPSubject(tt.args.ip); got != tt.want {
				t.Errorf("IPSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}

package log

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	type args struct {
		c *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "01",
			args: args{
				c: &Config{
					Path:  "test.log",
					Level: LevelDebug,
					Tags: []Tag{
						{
							Key:   "Hello",
							Value: "World",
						},
						{
							Key:   "你好",
							Value: "中国",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			Debugc(nil, "hahahaha", "lalalala")
			Debugfc(nil, ">> %s >> %s", "hahahaha", "lalalala")
			Debugwc(nil, "wc", "键", "值")
			Debugsc(nil, "wc",
				zap.String("键", "值"),
			)

			Debugc(context.Background(), "hahahaha", "lalalala")
			Debugfc(context.Background(), ">> %s >> %s", "hahahaha", "lalalala")
			Debugwc(context.Background(), "wc", "键", "值")
			Debugsc(context.Background(), "wc",
				zap.String("键", "值"),
			)

			if err := Init(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			Debugc(nil, "hahahaha", "lalalala")
			Debugfc(nil, ">> %s >> %s", "hahahaha", "lalalala")
			Debugwc(nil, "wc", "键", "值")
			Debugsc(nil, "wc",
				zap.String("键", "值"),
			)

			Debugc(context.Background(), "hahahaha", "lalalala")
			Debugfc(context.Background(), ">> %s >> %s", "hahahaha", "lalalala")
			Debugwc(context.Background(), "wc", "键", "值")
			Debugsc(context.Background(), "wc",
				zap.String("键", "值"),
			)

		})
	}
}

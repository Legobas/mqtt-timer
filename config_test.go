package main

import (
	"testing"
)

func Test_validate(t *testing.T) {
	type args struct {
		config Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "MQTT URL",
			args: args{
				config: Config{},
			},
			wantErr: true,
		},
		{
			name: "Timer ID",
			args: args{
				config: Config{0, 0, Mqtt{"url", "", "", 0, true}, []Timer{{}}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validate(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

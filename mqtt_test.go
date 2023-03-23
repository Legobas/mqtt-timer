package main

import (
	"testing"
)

func Test_validateMessage(t *testing.T) {
	type args struct {
		msg SetTimer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				msg: SetTimer{},
			},
			wantErr: true,
		},
		{
			name: "empty strings",
			args: args{
				msg: SetTimer{"", "", "", "", "", "", "", nil},
			},
			wantErr: true,
		},
		{
			name: "ok",
			args: args{
				msg: SetTimer{"ok", "", "", "", "", "", "", nil},
			},
			wantErr: false,
		},
		{
			name: "until without interval",
			args: args{
				msg: SetTimer{"id", "", "", "", "until", "", "", nil},
			},
			wantErr: true,
		},
		{
			name: "until with interval",
			args: args{
				msg: SetTimer{"id", "", "", "interval", "until", "", "", nil},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateMessage(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("validateMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

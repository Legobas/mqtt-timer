package main

import (
	"testing"
)

func Test_validateMessage(t *testing.T) {
	enabled:= true
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
			wantErr: true,
		},
		{
			name: "startOnly",
			args: args{
				msg: SetTimer{"id", "", "start", "", "", "", "", nil},
			},
			wantErr: false,
		},
		{
			name: "intervalOnly",
			args: args{
				msg: SetTimer{"id", "", "", "interval", "", "", "", nil},
			},
			wantErr: false,
		},
		{
			name: "until without interval",
			args: args{
				msg: SetTimer{"id", "descr", "start", "", "until", "", "", nil},
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
		{
			name: "enabled with start",
			args: args{
				msg: SetTimer{"id", "", "1 min", "", "", "", "", &enabled},
			},
			wantErr: true,
		},
		{
			name: "enabled with message",
			args: args{
				msg: SetTimer{"id", "", "", "", "", "", "test", &enabled},
			},
			wantErr: true,
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

package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_parseStart(t *testing.T) {
	type args struct {
		startStr string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "empty",
			args: args{""},
			want: time.Now(),
		},
		{
			name: "14:25",
			args: args{"14:25"},
			want: time.Date(0000, 01, 01, 14, 25, 00, 0, time.UTC),
		},
		{
			name: "21:36:24",
			args: args{"21:36:24"},
			want: time.Date(0000, 01, 01, 21, 36, 24, 0, time.UTC),
		},
		{
			name: "300 sec",
			args: args{"300 sec"},
			want: time.Now().Add(time.Duration(300 * time.Second)),
		},
		{
			name: "1 min",
			args: args{"1 min"},
			want: time.Now().Add(time.Duration(time.Minute)),
		},
		{
			name:    "number",
			args:    args{"8258946"},
			want:    time.Now(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStart(tt.args.startStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !(got.Hour() == tt.want.Hour() && got.Minute() == tt.want.Minute()) {
				t.Errorf("parseStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseInterval(t *testing.T) {
	type args struct {
		intervalStr string
		messages    []string
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "empty",
			args: args{"", nil},
			want: time.Duration(30 * time.Second),
		},
		{
			name: "300 sec",
			args: args{"300 sec", nil},
			want: time.Duration(300 * time.Second),
		},
		{
			name: "1 min",
			args: args{"1 min", nil},
			want: time.Duration(time.Minute),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseInterval(tt.args.intervalStr, tt.args.messages); got != tt.want {
				t.Errorf("parseInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseUntil(t *testing.T) {
	type args struct {
		untilStr  string
		startTime time.Time
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 time.Time
	}{
		{
			name:  "empty",
			args:  args{"", time.Now()},
			want:  1,
			want1: time.UnixMicro(0),
		},
		{
			name:  "14:25",
			args:  args{"14:25", time.Date(0000, 01, 01, 00, 10, 00, 0, time.UTC)},
			want:  -1,
			want1: time.Date(0000, 01, 01, 14, 25, 00, 0, time.UTC),
		},
		{
			name:  "21:36:24",
			args:  args{"21:36:24", time.Date(0000, 01, 01, 00, 10, 00, 0, time.UTC)},
			want:  -1,
			want1: time.Date(0000, 01, 01, 21, 36, 24, 0, time.UTC),
		},
		{
			name:  "300 sec",
			args:  args{"300 sec", time.Date(0000, 01, 01, 00, 10, 00, 0, time.UTC)},
			want:  -1,
			want1: time.Date(0000, 01, 01, 00, 15, 00, 0, time.UTC),
		},
		{
			name:  "1 min",
			args:  args{"1 min", time.Date(0000, 01, 01, 00, 10, 00, 0, time.UTC)},
			want:  -1,
			want1: time.Date(0000, 01, 01, 00, 11, 00, 0, time.UTC),
		},
		{
			name:  "100000 times",
			args:  args{"100000 times", time.Now()},
			want:  100000,
			want1: time.UnixMicro(0),
		},
		{
			name:  "10 years",
			args:  args{"10 years", time.Now()},
			want:  1,
			want1: time.UnixMicro(0),
		},
		{
			name:  "number",
			args:  args{"8258946", time.Now()},
			want:  8258946,
			want1: time.UnixMicro(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseUntil(tt.args.untilStr, tt.args.startTime)
			if got != tt.want {
				t.Errorf("parseUntil() got = %v, want %v", got, tt.want)
			}
			if tt.want1 != time.UnixMicro(0) && !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseUntil() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

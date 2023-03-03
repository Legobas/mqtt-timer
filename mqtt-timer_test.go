package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_offsetDuration(t *testing.T) {
	type args struct {
		timer Timer
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "empty",
			args: args{},
			want: time.Duration(0),
		},
		{
			name: "no time",
			args: args{
				timer: Timer{Before: "test"},
			},
			want: time.Duration(0),
		},
		{
			name: "not sec, min",
			args: args{
				timer: Timer{Before: "test tst"},
			},
			want: time.Duration(0),
		},
		{
			name: "not int",
			args: args{
				timer: Timer{Before: "test sec"},
			},
			want: time.Duration(0),
		},
		{
			name: "1 sec",
			args: args{
				timer: Timer{Before: "1 sec"},
			},
			want: time.Duration(0),
		},
		{
			name: "1 second",
			args: args{
				timer: Timer{After: "1 second"},
			},
			want: time.Duration(1000000000),
		},
		{
			name: "10 minutes",
			args: args{
				timer: Timer{After: "10 minutes"},
			},
			want: time.Duration(600000000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := offsetDuration(tt.args.timer); got != tt.want {
				t.Errorf("offsetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_timeBefore(t *testing.T) {
	type args struct {
		timer Timer
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "empty",
			args: args{},
			want: time.Time{},
		},
		{
			name: "no time",
			args: args{
				timer: Timer{Before: "test"},
			},
			want: time.Time{},
		},
		{
			name: "not sec, min",
			args: args{
				timer: Timer{Before: "test tst"},
			},
			want: time.Time{},
		},
		{
			name: "not int",
			args: args{
				timer: Timer{Before: "test sec"},
			},
			want: time.Time{},
		},
		{
			name: "1 min",
			args: args{
				timer: Timer{Time: "00:20", Before: "10 min"},
			},
			want: time.Date(0000, 01, 01, 00, 10, 00, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := timeBefore(tt.args.timer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("timeWithOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_offsetDescr(t *testing.T) {
	type args struct {
		timer Timer
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{},
			want: "at",
		},
		{
			name: "no time",
			args: args{
				timer: Timer{Before: "1 sec"},
			},
			want: "1 sec before",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := offsetDescr(tt.args.timer); got != tt.want {
				t.Errorf("offsetDescr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseSeconds(t *testing.T) {
	type args struct {
		timeExpr string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "empty",
			args: args{},
			want: 0,
		},
		{
			name: "no time",
			args: args{
				timeExpr: "test",
			},
			want: 0,
		},
		{
			name: "not sec, min",
			args: args{
				timeExpr: "test tst",
			},
			want: 0,
		},
		{
			name: "not int",
			args: args{
				timeExpr: "test sec",
			},
			want: 0,
		},
		{
			name: "10 sec",
			args: args{
				timeExpr: "10 seconds",
			},
			want: 10,
		},
		{
			name: "10 min",
			args: args{
				timeExpr: "10 min",
			},
			want: 600,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSeconds(tt.args.timeExpr); got != tt.want {
				t.Errorf("parseSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}

package http

import (
	"reflect"
	"testing"
)

func TestGetHttpPathForFilepath(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Root without dot and slash", args{filepath: ""}, "/"},
		{"Root without dot", args{filepath: "/"}, "/"},
		{"Root", args{filepath: "."}, "/"},
		{"Root trailing slash", args{filepath: "./"}, "/"},
		{"Root leading slash", args{filepath: "/."}, "/"},
		{"File", args{filepath: "./foo.txt"}, "/foo.txt"},
		{"File with trailing slash", args{filepath: "./foo.txt/"}, "/foo.txt"},
		{"File in dir", args{filepath: "./foo/bar.txt"}, "/foo/bar.txt"},
		{"File with dots in name", args{filepath: "./foo.bar.txt"}, "/foo.bar.txt"},
		{"Windows path file", args{filepath: "C:\\foo.bar.txt"}, "/foo.bar.txt"},
		{"Windows path directory", args{filepath: "C:\\foo\\bar\\baz"}, "/foo/bar/baz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHttpPathForFilepath(tt.args.filepath); got != tt.want {
				t.Errorf("GetHttpPathForFilepath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAcceptedQValues(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]float64
		wantErr bool
	}{
		{"Empty string", args{""}, map[string]float64{}, false},
		{"Valid entry", args{"deflate"}, map[string]float64{"deflate": 1.0}, false},
		{"Valid entry with q value", args{"deflate;q=0.3"}, map[string]float64{"deflate": 0.3}, false},
		{"Multiple entries", args{"deflate, gzip, br"}, map[string]float64{"deflate": 1.0, "gzip": 1.0, "br": 1.0}, false},
		{"Multiple entries with q values", args{"deflate;q=1.0, gzip;q=0.3, br;q=0.1"}, map[string]float64{"deflate": 1.0, "gzip": 0.3, "br": 0.1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAcceptedQValues(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAcceptedQValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAcceptedQValues() got = %v, want %v", got, tt.want)
			}
		})
	}
}

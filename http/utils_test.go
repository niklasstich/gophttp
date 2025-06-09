package http

import "testing"

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

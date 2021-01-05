// Copyright (c) 2020 Red Hat, Inc.

package version

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func TestVersion_String(t *testing.T) {
	type fields struct {
		GitCommit string
		BuildDate string
		Release   string
		GoVersion string
		Compiler  string
		Platform  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test",
			fields: fields{
				GitCommit: Commit,
				BuildDate: BuildDate,
				Release:   Release,
				GoVersion: runtime.Version(),
				Compiler:  runtime.Compiler,
				Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			},
			want: fmt.Sprintf("version.test/%s (%s/%s) clusterlifecycle-state-metrics/%s", Release, runtime.GOOS, runtime.GOARCH, Commit),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				GitCommit: tt.fields.GitCommit,
				BuildDate: tt.fields.BuildDate,
				Release:   tt.fields.Release,
				GoVersion: tt.fields.GoVersion,
				Compiler:  tt.fields.Compiler,
				Platform:  tt.fields.Platform,
			}
			if got := v.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	v := Version{
		GitCommit: Commit,
		BuildDate: BuildDate,
		Release:   Release,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	tests := []struct {
		name string
		want Version
	}{
		{
			name: "test",
			want: v,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVersion(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

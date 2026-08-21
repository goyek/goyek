package main

import (
	"flag"
	"io"
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantTasks []string
		wantErr   string
	}{
		{
			name:      "task before flag",
			args:      []string{"ci", "-v"},
			wantTasks: []string{"ci"},
		},
		{
			name:    "task after flag",
			args:    []string{"-v", "ci"},
			wantErr: "unexpected arguments: [ci]",
		},
		{
			name:    "argument after separator",
			args:    []string{"ci", "--", "extra"},
			wantErr: "unexpected arguments: [extra]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := flag.NewFlagSet("build", flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			flags.Bool("v", false, "verbose output")

			gotTasks, err := parseArgs(flags, tt.args)
			if !reflect.DeepEqual(gotTasks, tt.wantTasks) {
				t.Errorf("parseArgs(%v) tasks = %v, want %v", tt.args, gotTasks, tt.wantTasks)
			}
			if gotErr := errorString(err); gotErr != tt.wantErr {
				t.Errorf("parseArgs(%v) error = %q, want %q", tt.args, gotErr, tt.wantErr)
			}
		})
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

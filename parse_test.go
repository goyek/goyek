package goyek_test

import (
	"flag"
	"reflect"
	"testing"

	"github.com/goyek/goyek/v2"
)

//nolint:funlen // Test contains multiple cases.
func TestSplitTasks(t *testing.T) {
	setupFlag := func(fs *flag.FlagSet) {
		fs.Bool("v", false, "verbose")
	}
	checkFlag := func(t *testing.T, fs *flag.FlagSet) {
		v := fs.Lookup("v")
		if v == nil {
			t.Fatal("flag -v not found")
		}
		if v.Value.String() != "true" {
			t.Errorf("flag -v = %v, want true", v.Value)
		}
	}

	tests := []struct {
		name      string
		args      []string
		wantTasks []string
		wantRest  []string
		wantArgs  []string
		setupFlag func(*flag.FlagSet)
		checkFlag func(*testing.T, *flag.FlagSet)
	}{
		{
			name:      "tasks only",
			args:      []string{"task1", "task2"},
			wantTasks: []string{"task1", "task2"},
			wantRest:  nil,
			wantArgs:  []string{},
		},
		{
			name:      "tasks with separator and args",
			args:      []string{"task1", "--", "arg1", "arg2"},
			wantTasks: []string{"task1"},
			wantRest:  []string{"--", "arg1", "arg2"},
			wantArgs:  []string{"arg1", "arg2"},
		},
		{
			name:      "tasks with flags",
			args:      []string{"task1", "task2", "-v"},
			wantTasks: []string{"task1", "task2"},
			wantRest:  []string{"-v"},
			wantArgs:  []string{},
			setupFlag: setupFlag,
			checkFlag: checkFlag,
		},
		{
			name:      "tasks with flags and args",
			args:      []string{"task1", "-v", "--", "arg1", "arg2"},
			wantTasks: []string{"task1"},
			wantRest:  []string{"-v", "--", "arg1", "arg2"},
			wantArgs:  []string{"arg1", "arg2"},
			setupFlag: setupFlag,
			checkFlag: checkFlag,
		},
		{
			name:      "no args",
			args:      []string{},
			wantTasks: nil,
			wantRest:  nil,
			wantArgs:  []string{},
		},
		{
			name:      "only separator",
			args:      []string{"--"},
			wantTasks: nil,
			wantRest:  []string{"--"},
			wantArgs:  []string{},
		},
		{
			name:      "only args after separator",
			args:      []string{"--", "arg1", "arg2"},
			wantTasks: nil,
			wantRest:  []string{"--", "arg1", "arg2"},
			wantArgs:  []string{"arg1", "arg2"},
		},
		{
			name:      "flags only",
			args:      []string{"-v"},
			wantTasks: nil,
			wantRest:  []string{"-v"},
			wantArgs:  []string{},
			setupFlag: setupFlag,
			checkFlag: checkFlag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTasks, gotRest := goyek.SplitTasks(tt.args)

			if !reflect.DeepEqual(gotTasks, tt.wantTasks) {
				t.Errorf("SplitTasks() tasks = %v, want %v", gotTasks, tt.wantTasks)
			}

			if !reflect.DeepEqual(gotRest, tt.wantRest) {
				t.Errorf("SplitTasks() rest = %v, want %v", gotRest, tt.wantRest)
			}

			// Now parse flags if needed
			if tt.setupFlag != nil || tt.checkFlag != nil {
				fs := flag.NewFlagSet("test", flag.ContinueOnError)
				if tt.setupFlag != nil {
					tt.setupFlag(fs)
				}

				if err := fs.Parse(gotRest); err != nil {
					t.Fatalf("flag.Parse() error = %v", err)
				}

				if !reflect.DeepEqual(fs.Args(), tt.wantArgs) {
					t.Errorf("flag.Args() = %v, want %v", fs.Args(), tt.wantArgs)
				}

				if tt.checkFlag != nil {
					tt.checkFlag(t, fs)
				}
			}
		})
	}
}

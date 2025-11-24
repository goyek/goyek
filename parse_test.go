package goyek_test

import (
	"reflect"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestSplitTasks(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantTasks []string
		wantRest  []string
	}{
		{
			name:      "tasks only",
			args:      []string{"task1", "task2"},
			wantTasks: []string{"task1", "task2"},
			wantRest:  nil,
		},
		{
			name:      "tasks with separator and args",
			args:      []string{"task1", "--", "arg1", "arg2"},
			wantTasks: []string{"task1"},
			wantRest:  []string{"--", "arg1", "arg2"},
		},
		{
			name:      "tasks with flags",
			args:      []string{"task1", "task2", "-v"},
			wantTasks: []string{"task1", "task2"},
			wantRest:  []string{"-v"},
		},
		{
			name:      "tasks with flags and args",
			args:      []string{"task1", "-v", "--", "arg1", "arg2"},
			wantTasks: []string{"task1"},
			wantRest:  []string{"-v", "--", "arg1", "arg2"},
		},
		{
			name:      "no args",
			args:      []string{},
			wantTasks: nil,
			wantRest:  nil,
		},
		{
			name:      "only separator",
			args:      []string{"--"},
			wantTasks: nil,
			wantRest:  []string{"--"},
		},
		{
			name:      "only args after separator",
			args:      []string{"--", "arg1", "arg2"},
			wantTasks: nil,
			wantRest:  []string{"--", "arg1", "arg2"},
		},
		{
			name:      "flags only",
			args:      []string{"-v"},
			wantTasks: nil,
			wantRest:  []string{"-v"},
		},
		{
			name:      "multiple flags",
			args:      []string{"-v", "-dry-run", "--no-deps"},
			wantTasks: nil,
			wantRest:  []string{"-v", "-dry-run", "--no-deps"},
		},
		{
			name:      "task followed by multiple flags",
			args:      []string{"build", "-v", "--flag=value"},
			wantTasks: []string{"build"},
			wantRest:  []string{"-v", "--flag=value"},
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
		})
	}
}

package taskflow_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/pellared/taskflow"
)

func runTaskflowWith(flow *taskflow.Taskflow, param taskflow.RegisteredParam, cmd func(*taskflow.TF), args []string) int {
	flow.MustRegister(taskflow.Task{
		Name:        "task",
		Parameters:  []taskflow.RegisteredParam{param},
		Command:     cmd,
		Description: "Sample task for parameter tests",
	})
	completeArgs := []string{"task"}
	completeArgs = append(completeArgs, args...)
	return flow.Run(context.Background(), completeArgs...)
}

func Test_bool_param(t *testing.T) {
	tt := []struct {
		defaultValue bool
		args         []string

		exitCode int
		value    bool
	}{
		{defaultValue: false, args: []string{}, exitCode: taskflow.CodePass, value: false},
		{defaultValue: false, args: []string{"--bool"}, exitCode: taskflow.CodePass, value: true},
		{defaultValue: false, args: []string{"-b"}, exitCode: taskflow.CodePass, value: true},
		{defaultValue: true, args: []string{"-b=false"}, exitCode: taskflow.CodePass, value: false},

		{defaultValue: false, args: []string{"-b", "false"}, exitCode: taskflow.CodeInvalidArgs},
		{defaultValue: false, args: []string{"-b=maybe"}, exitCode: taskflow.CodeInvalidArgs},
	}

	for index, tc := range tt {
		tc := tc
		t.Run("case "+strconv.Itoa(index), func(t *testing.T) {
			flow := taskflow.New()
			param := flow.ConfigureBool(tc.defaultValue, taskflow.ParameterInfo{
				Name:  "bool",
				Short: 'b',
			})
			var got bool
			exitCode := runTaskflowWith(flow, param.RegisteredParam, func(tf *taskflow.TF) { got = param.Get(tf) }, tc.args)

			assertEqual(t, exitCode, tc.exitCode, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_bool_param_help(t *testing.T) {
	flow := taskflow.New()
	param := flow.ConfigureBool(true, taskflow.ParameterInfo{
		Name: "bool",
	})
	exitCode := runTaskflowWith(flow, param.RegisteredParam, func(tf *taskflow.TF) {}, []string{"-h"})

	assertEqual(t, exitCode, 0, "exit code should be OK")
}

func Test_int_param(t *testing.T) {
	tt := []struct {
		defaultValue int
		args         []string

		exitCode int
		value    int
	}{
		{defaultValue: 1, args: []string{}, exitCode: taskflow.CodePass, value: 1},
		{defaultValue: 1, args: []string{"--int"}, exitCode: taskflow.CodePass, value: 1},
		{defaultValue: 1, args: []string{"-i=123"}, exitCode: taskflow.CodePass, value: 123},
		{defaultValue: 1, args: []string{"-i", "456"}, exitCode: taskflow.CodePass, value: 456},

		{defaultValue: 1, args: []string{"-i", "9999999999999999999999"}, exitCode: taskflow.CodeInvalidArgs},
		{defaultValue: 1, args: []string{"-i=abc"}, exitCode: taskflow.CodeInvalidArgs},
	}

	for index, tc := range tt {
		tc := tc
		t.Run("case "+strconv.Itoa(index), func(t *testing.T) {
			flow := taskflow.New()
			param := flow.ConfigureInt(tc.defaultValue, taskflow.ParameterInfo{
				Name:  "int",
				Short: 'i',
			})
			var got int
			exitCode := runTaskflowWith(flow, param.RegisteredParam, func(tf *taskflow.TF) { got = param.Get(tf) }, tc.args)

			assertEqual(t, exitCode, tc.exitCode, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_int_param_help(t *testing.T) {
	flow := taskflow.New()
	param := flow.ConfigureInt(123, taskflow.ParameterInfo{
		Name: "int",
	})
	exitCode := runTaskflowWith(flow, param.RegisteredParam, func(tf *taskflow.TF) {}, []string{"-h"})

	assertEqual(t, exitCode, 0, "exit code should be OK")
}

func Test_string_param(t *testing.T) {
	tt := []struct {
		defaultValue string
		args         []string

		exitCode int
		value    string
	}{
		{defaultValue: "abc", args: []string{}, exitCode: taskflow.CodePass, value: "abc"},
		{defaultValue: "abc", args: []string{"--string", "xyz"}, exitCode: taskflow.CodePass, value: "xyz"},
		{defaultValue: "abc", args: []string{"-s=def"}, exitCode: taskflow.CodePass, value: "def"},
		{defaultValue: "abc", args: []string{"-s", "ghi"}, exitCode: taskflow.CodePass, value: "ghi"},
		{defaultValue: "abc", args: []string{"-s=jkl=mno"}, exitCode: taskflow.CodePass, value: "jkl=mno"},
	}

	for index, tc := range tt {
		tc := tc
		t.Run("case "+strconv.Itoa(index), func(t *testing.T) {
			flow := taskflow.New()
			param := flow.ConfigureString(tc.defaultValue, taskflow.ParameterInfo{
				Name:  "string",
				Short: 's',
			})
			var got string
			exitCode := runTaskflowWith(flow, param.RegisteredParam, func(tf *taskflow.TF) { got = param.Get(tf) }, tc.args)

			assertEqual(t, exitCode, tc.exitCode, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_string_param_help(t *testing.T) {
	flow := taskflow.New()
	param := flow.ConfigureString("abc", taskflow.ParameterInfo{
		Name: "string",
	})
	exitCode := runTaskflowWith(flow, param.RegisteredParam, func(tf *taskflow.TF) {}, []string{"-h"})

	assertEqual(t, exitCode, 0, "exit code should be OK")
}

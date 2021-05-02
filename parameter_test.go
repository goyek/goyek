package taskflow_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/pellared/taskflow"
)

func runTaskflowWith(flow *taskflow.Taskflow, param taskflow.RegisteredParam, cmd func(*taskflow.TF), args []string) int {
	flow.Register(taskflow.Task{
		Name:    "task",
		Params:  taskflow.Params{param},
		Command: cmd,
		Usage:   "Sample task for parameter tests",
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
		{defaultValue: false, args: []string{"-b"}, exitCode: taskflow.CodePass, value: true},
		{defaultValue: true, args: []string{"-b=false"}, exitCode: taskflow.CodePass, value: false},

		{defaultValue: false, args: []string{"-b", "false"}, exitCode: taskflow.CodeInvalidArgs},
		{defaultValue: false, args: []string{"-b=maybe"}, exitCode: taskflow.CodeInvalidArgs},
	}

	for index, tc := range tt {
		tc := tc
		t.Run("case "+strconv.Itoa(index), func(t *testing.T) {
			flow := taskflow.New()
			param := flow.RegisterBoolParam(tc.defaultValue, taskflow.ParamInfo{
				Name: "b",
			})
			var got bool
			exitCode := runTaskflowWith(flow, param, func(tf *taskflow.TF) { got = param.Get(tf) }, tc.args)

			assertEqual(t, exitCode, tc.exitCode, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_bool_param_help(t *testing.T) {
	flow := taskflow.New()
	param := flow.RegisterBoolParam(true, taskflow.ParamInfo{
		Name: "bool",
	})
	exitCode := runTaskflowWith(flow, param, func(tf *taskflow.TF) {}, []string{"-h"})

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
		{defaultValue: 1, args: []string{"-i"}, exitCode: taskflow.CodePass, value: 1},
		{defaultValue: 1, args: []string{"-i=123"}, exitCode: taskflow.CodePass, value: 123},
		{defaultValue: 1, args: []string{"-i", "456"}, exitCode: taskflow.CodePass, value: 456},

		{defaultValue: 1, args: []string{"-i", "9999999999999999999999"}, exitCode: taskflow.CodeInvalidArgs},
		{defaultValue: 1, args: []string{"-i=abc"}, exitCode: taskflow.CodeInvalidArgs},
	}

	for index, tc := range tt {
		tc := tc
		t.Run("case "+strconv.Itoa(index), func(t *testing.T) {
			flow := taskflow.New()
			param := flow.RegisterIntParam(tc.defaultValue, taskflow.ParamInfo{
				Name: "i",
			})
			var got int
			exitCode := runTaskflowWith(flow, param, func(tf *taskflow.TF) { got = param.Get(tf) }, tc.args)

			assertEqual(t, exitCode, tc.exitCode, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_int_param_help(t *testing.T) {
	flow := taskflow.New()
	param := flow.RegisterIntParam(123, taskflow.ParamInfo{
		Name: "int",
	})
	exitCode := runTaskflowWith(flow, param, func(tf *taskflow.TF) {}, []string{"-h"})

	assertEqual(t, exitCode, 0, "exit code should be OK")
}

func Test_string_param(t *testing.T) {
	tt := []struct {
		defaultValue string
		args         []string

		value string
	}{
		{defaultValue: "abc", args: []string{}, value: "abc"},
		{defaultValue: "abc", args: []string{"-s=def"}, value: "def"},
		{defaultValue: "abc", args: []string{"-s", "ghi"}, value: "ghi"},
		{defaultValue: "abc", args: []string{"-s=jkl=mno"}, value: "jkl=mno"},
		{defaultValue: "abc", args: []string{"-s", "param 'that \" may 'need' some \"escaping'"}, value: "param 'that \" may 'need' some \"escaping'"},
	}

	for index, tc := range tt {
		tc := tc
		t.Run("case "+strconv.Itoa(index), func(t *testing.T) {
			flow := taskflow.New()
			param := flow.RegisterStringParam(tc.defaultValue, taskflow.ParamInfo{
				Name: "s",
			})
			var got string
			exitCode := runTaskflowWith(flow, param, func(tf *taskflow.TF) { got = param.Get(tf) }, tc.args)

			assertEqual(t, exitCode, taskflow.CodePass, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_string_param_help(t *testing.T) {
	flow := taskflow.New()
	param := flow.RegisterStringParam("abc", taskflow.ParamInfo{
		Name: "string",
	})
	exitCode := runTaskflowWith(flow, param, func(tf *taskflow.TF) {}, []string{"-h"})

	assertEqual(t, exitCode, 0, "exit code should be OK")
}

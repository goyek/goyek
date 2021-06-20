package goyek_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/goyek/goyek"
)

func runFlowWith(flow *goyek.Flow, param goyek.RegisteredParam, cmd func(*goyek.A), args []string) int {
	flow.Register(goyek.Task{
		Name:   "task",
		Params: goyek.Params{param},
		Action: cmd,
		Usage:  "Sample task for parameter tests",
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
		{defaultValue: false, args: []string{}, exitCode: goyek.CodePass, value: false},
		{defaultValue: false, args: []string{"-b"}, exitCode: goyek.CodePass, value: true},
		{defaultValue: true, args: []string{"-b=false"}, exitCode: goyek.CodePass, value: false},

		{defaultValue: false, args: []string{"-b", "false"}, exitCode: goyek.CodeInvalidArgs},
		{defaultValue: false, args: []string{"-b=maybe"}, exitCode: goyek.CodeInvalidArgs},
	}

	for index, tc := range tt {
		tc := tc
		t.Run("case "+strconv.Itoa(index), func(t *testing.T) {
			flow := &goyek.Flow{}
			param := flow.RegisterBoolParam(goyek.BoolParam{
				Name:    "b",
				Default: tc.defaultValue,
			})
			var got bool
			exitCode := runFlowWith(flow, param, func(a *goyek.A) { got = param.Get(a) }, tc.args)

			assertEqual(t, exitCode, tc.exitCode, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_bool_param_help(t *testing.T) {
	flow := &goyek.Flow{}
	param := flow.RegisterBoolParam(goyek.BoolParam{
		Name:    "bool",
		Default: true,
	})
	exitCode := runFlowWith(flow, param, func(a *goyek.A) {}, []string{"-h"})

	assertEqual(t, exitCode, 0, "exit code should be OK")
}

func Test_int_param(t *testing.T) {
	tt := []struct {
		defaultValue int
		args         []string

		exitCode int
		value    int
	}{
		{defaultValue: 1, args: []string{}, exitCode: goyek.CodePass, value: 1},
		{defaultValue: 1, args: []string{"-i"}, exitCode: goyek.CodePass, value: 1},
		{defaultValue: 1, args: []string{"-i=123"}, exitCode: goyek.CodePass, value: 123},
		{defaultValue: 1, args: []string{"-i", "456"}, exitCode: goyek.CodePass, value: 456},

		{defaultValue: 1, args: []string{"-i", "9999999999999999999999"}, exitCode: goyek.CodeInvalidArgs},
		{defaultValue: 1, args: []string{"-i=abc"}, exitCode: goyek.CodeInvalidArgs},
	}

	for index, tc := range tt {
		tc := tc
		t.Run("case "+strconv.Itoa(index), func(t *testing.T) {
			flow := &goyek.Flow{}
			param := flow.RegisterIntParam(goyek.IntParam{
				Name:    "i",
				Default: tc.defaultValue,
			})
			var got int
			exitCode := runFlowWith(flow, param, func(a *goyek.A) { got = param.Get(a) }, tc.args)

			assertEqual(t, exitCode, tc.exitCode, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_int_param_help(t *testing.T) {
	flow := &goyek.Flow{}
	param := flow.RegisterIntParam(goyek.IntParam{
		Name:    "int",
		Default: 123,
	})
	exitCode := runFlowWith(flow, param, func(a *goyek.A) {}, []string{"-h"})

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
			flow := &goyek.Flow{}
			param := flow.RegisterStringParam(goyek.StringParam{
				Name:    "s",
				Default: tc.defaultValue,
			})
			var got string
			exitCode := runFlowWith(flow, param, func(a *goyek.A) { got = param.Get(a) }, tc.args)

			assertEqual(t, exitCode, goyek.CodePass, "exit code should match")
			assertEqual(t, got, tc.value, "value should match")
		})
	}
}

func Test_string_param_help(t *testing.T) {
	flow := &goyek.Flow{}
	param := flow.RegisterStringParam(goyek.StringParam{
		Name:    "string",
		Default: "abc",
	})
	exitCode := runFlowWith(flow, param, func(a *goyek.A) {}, []string{"-h"})

	assertEqual(t, exitCode, 0, "exit code should be OK")
}

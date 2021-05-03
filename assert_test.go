package goyek_test

import (
	"reflect"
	"strings"
	"testing"
)

func assertTrue(tb testing.TB, got bool, msg string) {
	tb.Helper()
	if got {
		return
	}
	tb.Errorf("%s\ngot: [%v], want: [true]", msg, got)
}

func assertContains(tb testing.TB, got string, want string, msg string) {
	tb.Helper()
	if strings.Contains(got, want) {
		return
	}
	tb.Errorf("%s\ngot: [%s], should contain: [%s]", msg, got, want)
}

func requireEqual(tb testing.TB, got interface{}, want interface{}, msg string) {
	tb.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	tb.Fatalf("%s\ngot: [%v], want: [%v]", msg, got, want)
}

func assertEqual(tb testing.TB, got interface{}, want interface{}, msg string) {
	tb.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	tb.Errorf("%s\ngot: [%v], want: [%v]", msg, got, want)
}

func assertPanics(tb testing.TB, fn func(), msg string) {
	tb.Helper()
	tryPanic := func() bool {
		didPanic := false
		func() {
			defer func() {
				if info := recover(); info != nil {
					didPanic = true
				}
			}()
			fn()
		}()
		return didPanic
	}

	if tryPanic() {
		return
	}
	tb.Errorf("%s\ndid not panic, but expected to do so", msg)
}

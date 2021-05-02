package goyek_test

import (
	"reflect"
	"strings"
	"testing"
)

func assertTrue(t testing.TB, got bool, msg string) {
	if got {
		return
	}
	t.Helper()
	t.Errorf("%s\ngot: [%v], want: [true]", msg, got)
}

func assertContains(t testing.TB, got string, want string, msg string) {
	if strings.Contains(got, want) {
		return
	}
	t.Helper()
	t.Errorf("%s\ngot: [%s], should contain: [%s]", msg, got, want)
}

func requireEqual(t testing.TB, got interface{}, want interface{}, msg string) {
	if reflect.DeepEqual(got, want) {
		return
	}
	t.Helper()
	t.Fatalf("%s\ngot: [%v], want: [%v]", msg, got, want)
}

func assertEqual(t testing.TB, got interface{}, want interface{}, msg string) {
	if reflect.DeepEqual(got, want) {
		return
	}
	t.Helper()
	t.Errorf("%s\ngot: [%v], want: [%v]", msg, got, want)
}

func assertPanics(t testing.TB, fn func(), msg string) {
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
	t.Helper()
	t.Errorf("%s\ndid not panic, but expected to do so", msg)
}

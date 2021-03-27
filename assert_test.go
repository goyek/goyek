package taskflow_test

import (
	"reflect"
	"strings"
	"testing"
)

func assertTrue(t testing.TB, got bool, msg string) {
	if !got {
		t.Helper()
		t.Errorf("%s\ngot: [%v], want: [true]", msg, got)
	}
}

func assertContains(t testing.TB, got string, want string, msg string) {
	if !strings.Contains(got, want) {
		t.Helper()
		t.Errorf("%s\ngot: [%s], should contain: [%s]", msg, got, want)
	}
}

func requireEqual(t testing.TB, want interface{}, got interface{}, msg string) {
	if !reflect.DeepEqual(got, want) {
		t.Helper()
		t.Fatalf("%s\ngot: [%v], want: [%v]", msg, got, want)
	}
}

func assertEqual(t testing.TB, want interface{}, got interface{}, msg string) {
	if !reflect.DeepEqual(got, want) {
		t.Helper()
		t.Errorf("%s\ngot: [%v], want: [%v]", msg, got, want)
	}
}

func requireNoError(t testing.TB, got error, msg string) {
	if got != nil {
		t.Helper()
		t.Fatalf("%s\ngot: [%v], want: [nil]", msg, got)
	}
}

func assertNoError(t testing.TB, got error, msg string) {
	if got != nil {
		t.Helper()
		t.Errorf("%s\ngot: [%v], want: [nil]", msg, got)
	}
}

func assertError(t testing.TB, got error, msg string) {
	if got == nil {
		t.Helper()
		t.Errorf("%s\ngot: [%v], want: [!nil]", msg, got)
	}
}

func assertPanics(t testing.TB, task func(), msg string) {
	tryPanic := func() bool {
		didPanic := false
		func() {
			defer func() {
				if info := recover(); info != nil {
					didPanic = true
				}
			}()
			task()
		}()
		return didPanic
	}

	if !tryPanic() {
		t.Helper()
		t.Errorf("%s\ndid not panic, but expected to do so", msg)
	}
}

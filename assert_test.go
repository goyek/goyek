package taskflow_test

import (
	"reflect"
	"strings"
	"testing"
)

func failedMessage(message ...string) string {
	if len(message) == 0 {
		return "Failed"
	}
	return strings.Join(message, ",")
}

func assertTrue(t testing.TB, got bool, message ...string) {
	if !got {
		t.Helper()
		t.Error(failedMessage(message...))
	}
}

func assertContains(t testing.TB, got string, want string, message ...string) {
	if !strings.Contains(got, want) {
		t.Helper()
		t.Errorf("%s\ngot: [%s], want: [%s]", failedMessage(message...), got, want)
	}
}

func requireEqual(t testing.TB, want interface{}, got interface{}, message ...string) {
	if !reflect.DeepEqual(got, want) {
		t.Helper()
		t.Fatalf("%s\ngot: [%v], want: [%v]", failedMessage(message...), got, want)
	}
}

func assertEqual(t testing.TB, want interface{}, got interface{}, message ...string) {
	if !reflect.DeepEqual(got, want) {
		t.Helper()
		t.Errorf("%s\ngot: [%v], want: [%v]", failedMessage(message...), got, want)
	}
}

func requireNoError(t testing.TB, got error, message ...string) {
	if got != nil {
		t.Helper()
		t.Fatalf("%s\ngot: [%v]", failedMessage(message...), got)
	}
}

func assertNoError(t testing.TB, got error, message ...string) {
	if got != nil {
		t.Helper()
		t.Errorf("%s\ngot: [%v]", failedMessage(message...), got)
	}
}

func assertError(t testing.TB, got error, message ...string) {
	if got == nil {
		t.Helper()
		t.Errorf("%s\ngot: [%v]", failedMessage(message...), got)
	}
}

func assertPanics(t testing.TB, task func(), message ...string) {
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
		t.Error(failedMessage(message...))
	}
}

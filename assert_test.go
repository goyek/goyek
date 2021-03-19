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

func assertTrue(t testing.TB, value bool, message ...string) {
	if !value {
		t.Helper()
		t.Error(failedMessage(message...))
	}
}

func assertContains(t testing.TB, value string, expected string, message ...string) {
	if !strings.Contains(value, expected) {
		t.Helper()
		t.Errorf("%s\nHave [%s], want [%s]", failedMessage(message...), value, expected)
	}
}

func requireEqual(t testing.TB, expected interface{}, value interface{}, message ...string) {
	if !reflect.DeepEqual(value, expected) {
		t.Helper()
		t.Fatalf("%s\nHave [%v], want [%v]", failedMessage(message...), value, expected)
	}
}

func assertEqual(t testing.TB, expected interface{}, value interface{}, message ...string) {
	if !reflect.DeepEqual(value, expected) {
		t.Helper()
		t.Errorf("%s\nHave [%v], want [%v]", failedMessage(message...), value, expected)
	}
}

func requireNoError(t testing.TB, err error, message ...string) {
	if err != nil {
		t.Helper()
		t.Fatalf("%s\nHave [%v]", failedMessage(message...), err)
	}
}

func assertNoError(t testing.TB, err error, message ...string) {
	if err != nil {
		t.Helper()
		t.Errorf("%s\nHave [%v]", failedMessage(message...), err)
	}
}

func assertError(t testing.TB, err error, message ...string) {
	if err == nil {
		t.Helper()
		t.Errorf("%s\nHave [%v]", failedMessage(message...), err)
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

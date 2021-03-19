package taskflow_test

import (
	"reflect"
	"strings"
	"testing"
)

func assertTrue(t testing.TB, value bool, message string) {
	if !value {
		t.Helper()
		t.Error(message)
	}
}

func assertContains(t testing.TB, value string, expected string) {
	if !strings.Contains(value, expected) {
		t.Helper()
		t.Errorf("Have [%s], want [%s]", value, expected)
	}
}

func assertEqual(t testing.TB, value interface{}, expected interface{}, message string) {
	if !reflect.DeepEqual(value, expected) {
		t.Helper()
		t.Errorf("%s\nHave [%v], want [%v]", message, value, expected)
	}
}

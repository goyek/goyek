package taskflow_test

import (
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

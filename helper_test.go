package goyek_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func assertContains(tb testing.TB, got fmt.Stringer, want string, msg string) {
	tb.Helper()
	gotTxt := got.String()
	if strings.Contains(gotTxt, want) {
		return
	}
	tb.Errorf("%s\nGOT:\n%s\nSHOULD CONTAIN:\n%s", msg, gotTxt, want)
}

func requireEqual(tb testing.TB, got interface{}, want interface{}, msg string) {
	tb.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	tb.Fatalf("%s\nGOT: %v\nWANT: %v", msg, got, want)
}

func assertEqual(tb testing.TB, got interface{}, want interface{}, msg string) {
	tb.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	tb.Errorf("%s\nGOT: %v\nWANT: %v", msg, got, want)
}

package goyek_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/goyek/goyek/v2"
)

var noDeps goyek.Deps

func assertTrue(tb testing.TB, got bool, msg string) {
	tb.Helper()
	if got {
		return
	}
	tb.Errorf("%s\nGOT: %v, WANT: true", msg, got)
}

func assertFalse(tb testing.TB, got bool, msg string) {
	tb.Helper()
	if !got {
		return
	}
	tb.Errorf("%s\nGOT: %v, WANT: false", msg, got)
}

func assertContains(tb testing.TB, got fmt.Stringer, want string, msg string) {
	tb.Helper()
	gotTxt := got.String()
	if strings.Contains(gotTxt, want) {
		return
	}
	tb.Errorf("%s\nGOT:\n%s\nSHOULD CONTAIN:\n%s", msg, gotTxt, want)
}

func assertNotContains(tb testing.TB, got fmt.Stringer, want string, msg string) {
	tb.Helper()
	gotTxt := got.String()
	if !strings.Contains(gotTxt, want) {
		return
	}
	tb.Errorf("%s\nGOT:\n%s\nSHOULD NOT CONTAIN:\n%s", msg, gotTxt, want)
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

func assertPass(tb testing.TB, got error, msg string) {
	tb.Helper()
	if got != nil {
		tb.Errorf("%s\nGOT: %v\nWANT: <PASS>", msg, got)
	}
}

func assertFail(tb testing.TB, got error, msg string) {
	tb.Helper()
	var ferr *goyek.FailError
	if !errors.As(got, &ferr) {
		tb.Errorf("%s\nGOT: %v\nWANT: <FAIL>", msg, got)
	}
}

func assertInvalid(tb testing.TB, got error, msg string) {
	tb.Helper()
	var ferr *goyek.FailError
	if errors.As(got, &ferr) || got == nil {
		tb.Errorf("%s\nGOT: %v\nWANT: <INVALID>", msg, got)
	}
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

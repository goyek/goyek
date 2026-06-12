//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package internal_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/goyek/goyek/v3/internal"
)

func TestTerminationSignals(t *testing.T) {
	got := internal.TerminationSignals()
	want := []os.Signal{os.Interrupt}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

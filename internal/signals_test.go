package internal

import (
	"testing"
)

func TestTerminationSignals(t *testing.T) {
	sigs := TerminationSignals()
	if len(sigs) == 0 {
		t.Fatal("TerminationSignals returned no signals")
	}
}

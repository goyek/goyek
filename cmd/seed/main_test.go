package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_greet(t *testing.T) {
	got := greet()

	want := "Hi!"
	assert.Equal(t, want, got, "should properly greet")
}

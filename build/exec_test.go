package main

import (
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestExec_EmptyCommand(t *testing.T) {
	runner := goyek.NewRunner(func(a *goyek.A) {
		if Exec(a, ".", "") {
			a.Error("Exec should return false for empty command")
		}
	})
	result := runner(goyek.Input{TaskName: "test"})
	if result.Status != goyek.StatusFailed {
		t.Errorf("Expected status Failed, got %v", result.Status)
	}
}

func TestExec_WhitespaceCommand(t *testing.T) {
	runner := goyek.NewRunner(func(a *goyek.A) {
		if Exec(a, ".", "   ") {
			a.Error("Exec should return false for whitespace-only command")
		}
	})
	result := runner(goyek.Input{TaskName: "test"})
	if result.Status != goyek.StatusFailed {
		t.Errorf("Expected status Failed, got %v", result.Status)
	}
}

func TestExec_ValidCommand(t *testing.T) {
	runner := goyek.NewRunner(func(a *goyek.A) {
		if !Exec(a, ".", "go version") {
			a.Error("Exec should return true for valid command")
		}
	})
	result := runner(goyek.Input{TaskName: "test"})
	if result.Status != goyek.StatusPassed {
		t.Errorf("Expected status Passed, got %v", result.Status)
	}
}

func TestExecArgs_InvalidCommand(t *testing.T) {
	runner := goyek.NewRunner(func(a *goyek.A) {
		if ExecArgs(a, ".", "non-existing-command") {
			a.Error("ExecArgs should return false for non-existing command")
		}
	})
	result := runner(goyek.Input{TaskName: "test"})
	if result.Status != goyek.StatusFailed {
		t.Errorf("Expected status Failed, got %v", result.Status)
	}
}

func TestExecArgs_ValidCommand(t *testing.T) {
	runner := goyek.NewRunner(func(a *goyek.A) {
		if !ExecArgs(a, ".", "go", "version") {
			a.Error("ExecArgs should return true for valid command")
		}
	})
	result := runner(goyek.Input{TaskName: "test"})
	if result.Status != goyek.StatusPassed {
		t.Errorf("Expected status Passed, got %v", result.Status)
	}
}

package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goyek/goyek/v3"
)

const publicModule = "github.com/goyek/goyek/v3"

var apiDiff = goyek.Define(goyek.Task{
	Name:  "apidiff",
	Usage: "check public API compatibility",
	Action: func(a *goyek.A) {
		if !Exec(a, dirBuild, "go", "install", "golang.org/x/exp/cmd/apidiff") {
			return
		}

		baseline, ok := localAPIBaseline(a)
		if !ok {
			return
		}
		a.Logf("API baseline: %s", baseline)

		compareModuleAPI(a, baseline)
	},
})

func localAPIBaseline(a *goyek.A) (string, bool) {
	a.Helper()
	args := []string{"describe", "--tags", "--abbrev=0", "--match", "v3.*", "HEAD"}
	a.Logf("Run %v in %s", append([]string{"git"}, args...), dirRoot)
	cmd := exec.CommandContext(a.Context(), "git", args...)
	cmd.Dir = dirRoot
	var stdout strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = a.Output()
	if err := cmd.Run(); err != nil {
		a.Errorf("no reachable local v3 release tag found: %v", err)
		return "", false
	}

	tag := strings.TrimSpace(stdout.String())
	if tag == "" {
		a.Error("local API baseline tag is empty")
		return "", false
	}
	return tag, true
}

func compareModuleAPI(a *goyek.A, baseline string) {
	a.Helper()
	tempDir, err := os.MkdirTemp("", "goyek-apidiff-")
	if err != nil {
		a.Error(err)
		return
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			a.Error(err)
		}
	}()

	baselineDir := filepath.Join(tempDir, "baseline")
	if !Exec(a, dirRoot, "git", "worktree", "add", "--detach", baselineDir, baseline) {
		return
	}
	defer Exec(a, dirRoot, "git", "worktree", "remove", "--force", baselineDir)

	oldAPI := filepath.Join(tempDir, "old.api")
	newAPI := filepath.Join(tempDir, "new.api")
	if !Exec(a, baselineDir, "apidiff", "-m", "-w", oldAPI, publicModule) {
		return
	}
	if !Exec(a, dirRoot, "apidiff", "-m", "-w", newAPI, publicModule) {
		return
	}
	checkAPICompatibility(a, oldAPI, newAPI)
}

func checkAPICompatibility(a *goyek.A, oldAPI, newAPI string) {
	a.Helper()
	args := []string{"-m", "-incompatible", oldAPI, newAPI}
	a.Logf("Run %v in %s", append([]string{"apidiff"}, args...), dirRoot)
	cmd := exec.CommandContext(a.Context(), "apidiff", args...) //nolint:gosec // apidiff is a pinned build tool
	cmd.Dir = dirRoot
	var report strings.Builder
	cmd.Stdout = io.MultiWriter(a.Output(), &report)
	cmd.Stderr = a.Output()
	if err := cmd.Run(); err != nil {
		a.Error(err)
		return
	}
	if strings.TrimSpace(report.String()) != "" {
		a.Error("incompatible public API changes found")
	}
}

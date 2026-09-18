package buildgo_test

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	buildgo "github.com/Genekkion/build.go/v1"
	"github.com/Genekkion/build.go/v1/commands/inline"
)

func setupTestCache(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		buildgo.Cleanup()
		_ = os.Chdir(origDir)
	})

	buildgo.Setup()
}

func TestStepRebuildLifecycle(t *testing.T) {
	setupTestCache(t)

	depFile := "dep.txt"
	if err := os.WriteFile(depFile, []byte("initial"), 0o644); err != nil {
		t.Fatalf("failed to write dep file: %v", err)
	}

	var runCount atomic.Int32
	mockCmd, err := inline.NewCmd([]inline.CmdFunc{
		func(ctx context.Context) error {
			runCount.Add(1)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("failed to create command: %v", err)
	}

	absDep, err := filepath.Abs(depFile)
	if err != nil {
		t.Fatalf("failed to get abs path: %v", err)
	}

	step1 := buildgo.NewStep("build", mockCmd).AddFileDeps(absDep)
	if err := step1.Run(context.Background()); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected 1 run on initial build, got %d", runCount.Load())
	}

	step2 := buildgo.NewStep("build", mockCmd).AddFileDeps(absDep)
	if err := step2.Run(context.Background()); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected step to skip, runCount is %d", runCount.Load())
	}

	if err := os.WriteFile(depFile, []byte("updated"), 0o644); err != nil {
		t.Fatalf("failed to update dep file: %v", err)
	}

	step3 := buildgo.NewStep("build", mockCmd).AddFileDeps(absDep)
	if err := step3.Run(context.Background()); err != nil {
		t.Fatalf("third run failed: %v", err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected 2 runs after file update, got %d", runCount.Load())
	}
}

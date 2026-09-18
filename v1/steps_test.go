package buildgo_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/Genekkion/build.go/internal/vfs"
	buildgo "github.com/Genekkion/build.go/v1"
	"github.com/Genekkion/build.go/v1/commands/inline"
)

func setupTestEnv(t *testing.T) *vfs.MemFS {
	t.Helper()
	memFS := vfs.NewMemFS()
	buildgo.Setup(
		buildgo.WithInMemoryDB(),
		buildgo.WithFS(memFS),
	)
	t.Cleanup(func() {
		buildgo.Cleanup()
	})
	return memFS
}

func TestStepRebuildLifecycle(t *testing.T) {
	memFS := setupTestEnv(t)

	depFile := "/virtual/dep.txt"
	if err := memFS.WriteFile(depFile, []byte("initial"), 0o644); err != nil {
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

	step1 := buildgo.NewStep("build", mockCmd).AddFileDeps(depFile)
	if err := step1.Run(context.Background()); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected 1 run on initial build, got %d", runCount.Load())
	}

	step2 := buildgo.NewStep("build", mockCmd).AddFileDeps(depFile)
	if err := step2.Run(context.Background()); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected step to skip, runCount is %d", runCount.Load())
	}

	if err := memFS.WriteFile(depFile, []byte("updated"), 0o644); err != nil {
		t.Fatalf("failed to update dep file: %v", err)
	}

	step3 := buildgo.NewStep("build", mockCmd).AddFileDeps(depFile)
	if err := step3.Run(context.Background()); err != nil {
		t.Fatalf("third run failed: %v", err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected 2 runs after file update, got %d", runCount.Load())
	}
}

func TestStepFailureDoesNotCacheHash(t *testing.T) {
	memFS := setupTestEnv(t)

	depFile := "/virtual/fail_dep.txt"
	if err := memFS.WriteFile(depFile, []byte("failure-content"), 0o644); err != nil {
		t.Fatalf("failed to write dep file: %v", err)
	}

	var shouldFail atomic.Bool
	shouldFail.Store(true)
	var runCount atomic.Int32

	cmd, err := inline.NewCmd([]inline.CmdFunc{
		func(ctx context.Context) error {
			runCount.Add(1)
			if shouldFail.Load() {
				return errors.New("simulated step command failure")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("failed to create command: %v", err)
	}

	// 1. First run fails
	step1 := buildgo.NewStep("failStep", cmd).AddFileDeps(depFile)
	err = step1.Run(context.Background())
	if err == nil {
		t.Fatalf("expected step1 to return error, got nil")
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected 1 invocation, got %d", runCount.Load())
	}

	// 2. Second run without file changes must STILL attempt to run because step1 failed and did not cache
	shouldFail.Store(false)
	step2 := buildgo.NewStep("failStep", cmd).AddFileDeps(depFile)
	err = step2.Run(context.Background())
	if err != nil {
		t.Fatalf("expected step2 to succeed, got %v", err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected 2 invocations (retry succeeded), got %d", runCount.Load())
	}

	// 3. Third run without file changes must now skip since step2 succeeded and cached
	step3 := buildgo.NewStep("failStep", cmd).AddFileDeps(depFile)
	err = step3.Run(context.Background())
	if err != nil {
		t.Fatalf("expected step3 to succeed, got %v", err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected step3 to skip, runCount is %d", runCount.Load())
	}
}

func TestStepMultipleFileDeps(t *testing.T) {
	memFS := setupTestEnv(t)

	f1 := "/virtual/file1.txt"
	f2 := "/virtual/file2.txt"
	if err := memFS.WriteFile(f1, []byte("f1-v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := memFS.WriteFile(f2, []byte("f2-v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	var runCount atomic.Int32
	cmd, err := inline.NewCmd([]inline.CmdFunc{
		func(ctx context.Context) error {
			runCount.Add(1)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// First run
	step1 := buildgo.NewStep("multi", cmd).AddFileDeps(f1, f2)
	if err := step1.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected runCount 1, got %d", runCount.Load())
	}

	// Modify only one file
	if err := memFS.WriteFile(f2, []byte("f2-v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	step2 := buildgo.NewStep("multi", cmd).AddFileDeps(f1, f2)
	if err := step2.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected runCount 2 after modifying one dep, got %d", runCount.Load())
	}
}

package buildgo_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Genekkion/build.go/internal/vfs"
	buildgo "github.com/Genekkion/build.go/v1"
	"github.com/Genekkion/build.go/v1/commands/inline"
)

func setupTestEnv(t *testing.T) (context.Context, *vfs.MemFS) {
	t.Helper()
	memFS := vfs.NewMemFS()
	ctx := buildgo.Setup(
		buildgo.WithInMemoryDB(),
		buildgo.WithFS(memFS),
	)
	t.Cleanup(func() {
		buildgo.Cleanup(ctx)
	})
	return ctx, memFS
}

func TestStepRebuildLifecycle(t *testing.T) {
	t.Parallel()
	ctx, memFS := setupTestEnv(t)

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
	if err := step1.Run(ctx); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected 1 run on initial build, got %d", runCount.Load())
	}

	step2 := buildgo.NewStep("build", mockCmd).AddFileDeps(depFile)
	if err := step2.Run(ctx); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected step to skip, runCount is %d", runCount.Load())
	}

	if err := memFS.WriteFile(depFile, []byte("updated"), 0o644); err != nil {
		t.Fatalf("failed to update dep file: %v", err)
	}

	step3 := buildgo.NewStep("build", mockCmd).AddFileDeps(depFile)
	if err := step3.Run(ctx); err != nil {
		t.Fatalf("third run failed: %v", err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected 2 runs after file update, got %d", runCount.Load())
	}
}

func TestStepFailureDoesNotCacheHash(t *testing.T) {
	t.Parallel()
	ctx, memFS := setupTestEnv(t)

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

	step1 := buildgo.NewStep("failStep", cmd).AddFileDeps(depFile)
	err = step1.Run(ctx)
	if err == nil {
		t.Fatalf("expected step1 to return error, got nil")
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected 1 invocation, got %d", runCount.Load())
	}

	shouldFail.Store(false)
	step2 := buildgo.NewStep("failStep", cmd).AddFileDeps(depFile)
	err = step2.Run(ctx)
	if err != nil {
		t.Fatalf("expected step2 to succeed, got %v", err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected 2 invocations (retry succeeded), got %d", runCount.Load())
	}

	step3 := buildgo.NewStep("failStep", cmd).AddFileDeps(depFile)
	err = step3.Run(ctx)
	if err != nil {
		t.Fatalf("expected step3 to succeed, got %v", err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected step3 to skip, runCount is %d", runCount.Load())
	}
}

func TestStepMultipleFileDeps(t *testing.T) {
	t.Parallel()
	ctx, memFS := setupTestEnv(t)

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

	step1 := buildgo.NewStep("multi", cmd).AddFileDeps(f1, f2)
	if err := step1.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if runCount.Load() != 1 {
		t.Fatalf("expected runCount 1, got %d", runCount.Load())
	}

	if err := memFS.WriteFile(f2, []byte("f2-v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	step2 := buildgo.NewStep("multi", cmd).AddFileDeps(f1, f2)
	if err := step2.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if runCount.Load() != 2 {
		t.Fatalf("expected runCount 2 after modifying one dep, got %d", runCount.Load())
	}
}

func TestStepScopedCacheIsolation(t *testing.T) {
	t.Parallel()
	ctx, memFS := setupTestEnv(t)

	sharedFile := "/virtual/shared.txt"
	if err := memFS.WriteFile(sharedFile, []byte("shared-v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	var runCountA, runCountB atomic.Int32
	cmdA, err := inline.NewCmd([]inline.CmdFunc{
		func(ctx context.Context) error {
			runCountA.Add(1)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	cmdB, err := inline.NewCmd([]inline.CmdFunc{
		func(ctx context.Context) error {
			runCountB.Add(1)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	stepA := buildgo.NewStep("stepA", cmdA).AddFileDeps(sharedFile)
	stepB := buildgo.NewStep("stepB", cmdB).AddFileDeps(sharedFile)

	if err := stepA.Run(ctx); err != nil {
		t.Fatalf("stepA run failed: %v", err)
	}
	if runCountA.Load() != 1 {
		t.Fatalf("expected stepA runCount 1, got %d", runCountA.Load())
	}

	if err := stepB.Run(ctx); err != nil {
		t.Fatalf("stepB run failed: %v", err)
	}
	if runCountB.Load() != 1 {
		t.Fatalf("expected stepB to run (not falsely skip due to stepA), got %d", runCountB.Load())
	}

	stepA2 := buildgo.NewStep("stepA", cmdA).AddFileDeps(sharedFile)
	stepB2 := buildgo.NewStep("stepB", cmdB).AddFileDeps(sharedFile)
	if err := stepA2.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if err := stepB2.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if runCountA.Load() != 1 || runCountB.Load() != 1 {
		t.Fatalf("expected both to skip, got A=%d, B=%d", runCountA.Load(), runCountB.Load())
	}
}

func TestStepCycleDetection(t *testing.T) {
	t.Parallel()
	ctx, _ := setupTestEnv(t)

	noop, _ := inline.NewCmd([]inline.CmdFunc{func(ctx context.Context) error { return nil }})

	t.Run("self-cycle", func(t *testing.T) {
		stepA := buildgo.NewStep("A", noop)
		stepA.DependsOn(stepA)

		err := stepA.Run(ctx)
		if err == nil {
			t.Fatal("expected cycle error, got nil")
		}
		if !strings.Contains(err.Error(), "dependency cycle detected: A -> A") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("two-step cycle", func(t *testing.T) {
		stepA := buildgo.NewStep("A", noop)
		stepB := buildgo.NewStep("B", noop)
		stepA.DependsOn(stepB)
		stepB.DependsOn(stepA)

		err := stepA.Run(ctx)
		if err == nil {
			t.Fatal("expected cycle error, got nil")
		}
		if !strings.Contains(err.Error(), "dependency cycle detected") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("three-step cycle", func(t *testing.T) {
		stepA := buildgo.NewStep("A", noop)
		stepB := buildgo.NewStep("B", noop)
		stepC := buildgo.NewStep("C", noop)
		stepA.DependsOn(stepB)
		stepB.DependsOn(stepC)
		stepC.DependsOn(stepA)

		err := stepA.Run(ctx)
		if err == nil {
			t.Fatal("expected cycle error, got nil")
		}
		if !strings.Contains(err.Error(), "dependency cycle detected: A -> B -> C -> A") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestStepDiamondDependency(t *testing.T) {
	t.Parallel()
	ctx, _ := setupTestEnv(t)

	var runCountA, runCountB, runCountC, runCountD atomic.Int32

	cmdA, _ := inline.NewCmd([]inline.CmdFunc{func(ctx context.Context) error { runCountA.Add(1); return nil }})
	cmdB, _ := inline.NewCmd([]inline.CmdFunc{func(ctx context.Context) error { runCountB.Add(1); return nil }})
	cmdC, _ := inline.NewCmd([]inline.CmdFunc{func(ctx context.Context) error { runCountC.Add(1); return nil }})
	cmdD, _ := inline.NewCmd([]inline.CmdFunc{func(ctx context.Context) error { runCountD.Add(1); return nil }})

	stepA := buildgo.NewStep("A", cmdA)
	stepB := buildgo.NewStep("B", cmdB).DependsOn(stepA)
	stepC := buildgo.NewStep("C", cmdC).DependsOn(stepA)
	stepD := buildgo.NewStep("D", cmdD).DependsOn(stepB, stepC)

	err := stepD.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error in diamond dependency: %v", err)
	}

	if runCountA.Load() != 1 || runCountB.Load() != 1 || runCountC.Load() != 1 || runCountD.Load() != 1 {
		t.Fatalf("expected each step to run once, got A=%d, B=%d, C=%d, D=%d",
			runCountA.Load(), runCountB.Load(), runCountC.Load(), runCountD.Load())
	}
}

func TestUpstreamRebuildPropagation(t *testing.T) {
	t.Parallel()
	ctx, memFS := setupTestEnv(t)

	fileA := "/virtual/a.txt"
	fileB := "/virtual/b.txt"
	if err := memFS.WriteFile(fileA, []byte("a-v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := memFS.WriteFile(fileB, []byte("b-v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	var runCountA, runCountB atomic.Int32
	cmdA, _ := inline.NewCmd([]inline.CmdFunc{func(ctx context.Context) error { runCountA.Add(1); return nil }})
	cmdB, _ := inline.NewCmd([]inline.CmdFunc{func(ctx context.Context) error { runCountB.Add(1); return nil }})

	stepA := buildgo.NewStep("A", cmdA).AddFileDeps(fileA)
	stepB := buildgo.NewStep("B", cmdB).AddFileDeps(fileB).DependsOn(stepA)

	if err := stepB.Run(ctx); err != nil {
		t.Fatalf("initial run failed: %v", err)
	}
	if runCountA.Load() != 1 || runCountB.Load() != 1 {
		t.Fatalf("expected initial run counts A=1, B=1; got A=%d, B=%d", runCountA.Load(), runCountB.Load())
	}

	stepA2 := buildgo.NewStep("A", cmdA).AddFileDeps(fileA)
	stepB2 := buildgo.NewStep("B", cmdB).AddFileDeps(fileB).DependsOn(stepA2)

	if err := stepB2.Run(ctx); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if runCountA.Load() != 1 || runCountB.Load() != 1 {
		t.Fatalf("expected both to skip; got A=%d, B=%d", runCountA.Load(), runCountB.Load())
	}

	if err := memFS.WriteFile(fileA, []byte("a-v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	stepA3 := buildgo.NewStep("A", cmdA).AddFileDeps(fileA)
	stepB3 := buildgo.NewStep("B", cmdB).AddFileDeps(fileB).DependsOn(stepA3)

	if err := stepB3.Run(ctx); err != nil {
		t.Fatalf("third run failed: %v", err)
	}
	if runCountA.Load() != 2 {
		t.Fatalf("expected stepA to rebuild (count 2), got %d", runCountA.Load())
	}
	if runCountB.Load() != 2 {
		t.Fatalf("expected stepB to rebuild due to upstream rebuild (count 2), got %d", runCountB.Load())
	}

	stepA4 := buildgo.NewStep("A", cmdA).AddFileDeps(fileA)
	stepB4 := buildgo.NewStep("B", cmdB).AddFileDeps(fileB).DependsOn(stepA4)

	if err := stepB4.Run(ctx); err != nil {
		t.Fatalf("fourth run failed: %v", err)
	}
	if runCountA.Load() != 2 || runCountB.Load() != 2 {
		t.Fatalf("expected both to skip again; got A=%d, B=%d", runCountA.Load(), runCountB.Load())
	}
}

func TestCleanupSafety(t *testing.T) {
	t.Parallel()
	// Calling Cleanup with a context that has no DB should not panic
	buildgo.Cleanup(context.Background())
	buildgo.Cleanup(context.Background())

	// Calling Cleanup after Setup should cleanly close DB
	memFS := vfs.NewMemFS()
	ctx := buildgo.Setup(buildgo.WithInMemoryDB(), buildgo.WithFS(memFS))
	buildgo.Cleanup(ctx)
}

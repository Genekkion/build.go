package inline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Genekkion/build.go/v1/commands/inline"
)

func TestInlineCmd(t *testing.T) {
	_, err := inline.NewCmd(nil)
	if err == nil {
		t.Fatal("expected error for empty funcs")
	}

	var executions []int
	cmd, err := inline.NewCmd([]inline.CmdFunc{
		func(ctx context.Context) error {
			executions = append(executions, 1)
			return nil
		},
		func(ctx context.Context) error {
			executions = append(executions, 2)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("failed to create inline command: %v", err)
	}

	if err := cmd.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(executions) != 2 || executions[0] != 1 || executions[1] != 2 {
		t.Fatalf("unexpected executions order: %v", executions)
	}

	// Test error propagation
	failErr := errors.New("inline error")
	failCmd, err := inline.NewCmd([]inline.CmdFunc{
		func(ctx context.Context) error {
			return failErr
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := failCmd.Run(context.Background()); !errors.Is(err, failErr) {
		t.Fatalf("expected %v, got %v", failErr, err)
	}
}

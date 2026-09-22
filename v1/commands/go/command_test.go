package cmdgo_test

import (
	"bytes"
	"context"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	cmdgo "github.com/Genekkion/build.go/v1/commands/go"
)

func TestNewCmdTargetNormalization(t *testing.T) {
	tests := []struct {
		name     string
		targets  []string
		expected []string
	}{
		{
			name:     "dot target",
			targets:  []string{"."},
			expected: []string{"."},
		},
		{
			name:     "relative path without dot",
			targets:  []string{"my_pkg"},
			expected: []string{"." + string(filepath.Separator) + "my_pkg"},
		},
		{
			name:     "relative path with dot",
			targets:  []string{"./my_pkg"},
			expected: []string{"./my_pkg"},
		},
		{
			name:     "parent relative path",
			targets:  []string{"../other_pkg"},
			expected: []string{"../other_pkg"},
		},
		{
			name:     "recursive wildcard",
			targets:  []string{"./..."},
			expected: []string{"./..."},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := cmdgo.NewBuildCmd(".", tc.targets, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(cmd.Targets(), tc.expected) {
				t.Fatalf("expected targets %v, got %v", tc.expected, cmd.Targets())
			}
		})
	}
}

func TestNewCmdEmptyTargets(t *testing.T) {
	_, err := cmdgo.NewBuildCmd(".", []string{}, nil)
	if err == nil {
		t.Fatal("expected error for empty targets")
	}
}

func TestGoCmdRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	// Run go test on internal/util/set
	cmd, err := cmdgo.NewTestCmd(repoRoot, []string{"./internal/util/set"}, []string{"-v"},
		cmdgo.WithStdout(&stdout),
		cmdgo.WithStderr(&stderr),
	)
	if err != nil {
		t.Fatalf("failed to create test cmd: %v", err)
	}

	err = cmd.Run(context.Background())
	if err != nil {
		t.Fatalf("command execution failed: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "PASS") {
		t.Fatalf("expected PASS in test output, got: %s", output)
	}
}

package shell_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Genekkion/build.go/v1/commands/shell"
)

func TestShellCmd(t *testing.T) {
	_, err := shell.NewCmd(nil)
	if err == nil {
		t.Fatal("expected error for empty args")
	}

	var stdout bytes.Buffer
	cmd, err := shell.NewCmd([]string{"echo", "hello shell"}, shell.WithStdout(&stdout))
	if err != nil {
		t.Fatalf("failed to create shell command: %v", err)
	}

	if err := cmd.Run(context.Background()); err != nil {
		t.Fatalf("failed to run shell command: %v", err)
	}

	if !strings.Contains(stdout.String(), "hello shell") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

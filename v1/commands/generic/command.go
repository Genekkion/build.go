package generic

import (
	"context"
	"os/exec"
)

// Cmd represents a generic command.
type Cmd struct {
	cfg  Config
	cmd  string
	args []string
}

// NewCmd creates a new generic command.
func NewCmd(args []string, opts ...Option) Cmd {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return Cmd{
		cfg:  cfg,
		cmd:  args[0],
		args: args[1:],
	}
}

// Run runs the command.
func (c Cmd) Run(ctx context.Context) (err error) {
	cmd := exec.CommandContext(ctx, c.cmd, c.args...)
	cmd.Dir = c.cfg.cwd
	cmd.Stdout = c.cfg.stdout
	cmd.Stderr = c.cfg.stderr

	return cmd.Run()
}

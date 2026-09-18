package cmdgo

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"

	buildgo "github.com/Genekkion/build.go/v1"
)

// GoCmd represents a go command.
type GoCmd struct {
	cfg     Config
	cwd     string
	targets []string
	args    []string
}

// NewBuildCmd creates a new go build command.
func NewBuildCmd(cwd string, targets []string, args []string, opts ...Option) (cmd *GoCmd, err error) {
	cmd, err = newCmd(cwd, targets, args, opts...)
	if err != nil {
		return nil, err
	}

	buildgo.Logger.Debug("Go build command created",
		"compilerPath", cmd.cfg.compilerPath,
		"cwd", cmd.cwd,
		"targets", cmd.targets,
		"args", cmd.args,
	)

	args = append([]string{
		cmd.cfg.compilerPath, "build",
	}, cmd.args...)
	args = append(args, cmd.targets...)
	cmd.args = args

	return cmd, nil
}

// NewRunCmd creates a new go run command.
func NewRunCmd(cwd string, targets []string, args []string, opts ...Option) (cmd *GoCmd, err error) {
	cmd, err = newCmd(cwd, targets, args, opts...)
	if err != nil {
		return nil, err
	}

	buildgo.Logger.Debug("Go run command created",
		"compilerPath", cmd.cfg.compilerPath,
		"cwd", cmd.cwd,
		"targets", cmd.targets,
		"args", cmd.args,
	)

	args = append([]string{
		cmd.cfg.compilerPath, "run",
	}, cmd.args...,
	)
	args = append(args, cmd.targets...)
	cmd.args = args

	return cmd, nil
}

// NewTestCmd creates a new go test command.
func NewTestCmd(cwd string, targets []string, args []string, opts ...Option) (cmd *GoCmd, err error) {
	cmd, err = newCmd(cwd, targets, args, opts...)
	if err != nil {
		return nil, err
	}

	buildgo.Logger.Debug("Go test command created",
		"compilerPath", cmd.cfg.compilerPath,
		"cwd", cmd.cwd,
		"targets", cmd.targets,
		"args", cmd.args,
	)

	args = append([]string{
		cmd.cfg.compilerPath, "test",
	}, cmd.args...)
	args = append(args, cmd.targets...)
	cmd.args = args

	return cmd, nil
}

// newCmd creates a new go command.
func newCmd(cwd string, targets []string, args []string, opts ...Option) (cmd *GoCmd, err error) {
	if len(targets) == 0 {
		return nil, errors.New("target is required")
	}

	cmd = &GoCmd{
		cfg:     defaultConfig(),
		cwd:     cwd,
		targets: targets,
		args:    args,
	}
	for _, opt := range opts {
		opt(&cmd.cfg)
	}

	err = cmd.setupTargets()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

// setupTargets normalizes relative target paths for the go toolchain.
func (c *GoCmd) setupTargets() error {
	for i, target := range c.targets {
		if target == "." || target == "..." || strings.HasPrefix(target, "./") || strings.HasPrefix(target, "../") || filepath.IsAbs(target) {
			continue
		}
		c.targets[i] = "." + string(filepath.Separator) + target
	}
	return nil
}

// Targets returns the command targets.
func (c *GoCmd) Targets() []string {
	return c.targets
}

// Args returns the full command arguments.
func (c *GoCmd) Args() []string {
	return c.args
}

// Run runs the go command.
func (c GoCmd) Run(ctx context.Context) error {
	args := c.args
	buildgo.Logger.Debug("Running go command",
		"cwd", c.cwd,
		"args", args,
	)

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = c.cwd
	cmd.Stdout = c.cfg.stdout
	cmd.Stderr = c.cfg.stderr

	return cmd.Run()
}

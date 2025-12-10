package cmdgo

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

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
	}, cmd.targets...)
	args = append(args, cmd.args...)
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
	}, cmd.targets...)
	args = append(args, cmd.args...)
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

// setupTargets recursively collects all files in the target directory. Warning: if the target is a directory, it cannot
// have any other targets apart from the directory itself.
func (c *GoCmd) setupTargets() (err error) {
	for _, target := range c.targets {
		stat, err := os.Stat(target)
		if err != nil {
			return err
		} else if !stat.IsDir() {
			continue
		}

		// target is a directory
		if len(c.targets) != 1 {
			return errors.New("directory target must have exactly one argument")
		}

		files, err := os.ReadDir(target)
		if err != nil {
			return err
		}

		c.targets = make([]string, 0, len(files))
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			c.targets = append(c.targets, filepath.Join(target, file.Name()))
		}
		return nil
	}

	return nil
}

// Run runs the go command.
func (c GoCmd) Run(ctx context.Context) error {
	args := append([]string{
		c.cfg.compilerPath, "run",
	}, c.targets...)
	args = append(args, c.args...)
	buildgo.Logger.Debug("Running go command",
		"cwd", c.cwd,
		"args", args,
	)

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = c.cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

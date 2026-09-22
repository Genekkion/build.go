package buildgo

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/Genekkion/build.go/internal/vfs"
)

// Step represents a single build step.
type Step struct {
	name             string
	commands         []Command
	dependsOn        []*Step
	fileDepsPatterns []string
	done             atomic.Bool
	rebuilt          atomic.Bool
}

// NewStep creates a new step.
func NewStep(name string, commands ...Command) *Step {
	if len(commands) == 0 {
		panic("step needs at least 1 command")
	}

	return &Step{
		name:     name,
		commands: commands,
	}
}

// DependsOn adds a dependency on other steps.
func (s *Step) DependsOn(steps ...*Step) *Step {
	s.dependsOn = append(s.dependsOn, steps...)
	return s
}

// AddFileDeps adds file dependency patterns. Patterns are stored as-is
// and resolved against the context's filesystem at build time.
func (s *Step) AddFileDeps(patterns ...string) *Step {
	s.fileDepsPatterns = append(s.fileDepsPatterns, patterns...)
	return s
}

// SetFileDeps sets the file dependencies.
func (s *Step) SetFileDeps(patterns []string) *Step {
	s.fileDepsPatterns = nil
	return s.AddFileDeps(patterns...)
}

// FileDeps returns the file dependencies.
func (s *Step) FileDeps() []string {
	return s.fileDepsPatterns
}

// Name returns the name of the step.
func (s *Step) Name() string {
	return s.name
}

// Commands returns the commands of the step.
func (s *Step) Commands() []Command {
	return s.commands
}

// Done returns whether the step has been completed.
func (s *Step) Done() bool {
	return s.done.Load()
}

// Rebuilt returns whether the step ran its commands during the current build execution.
func (s *Step) Rebuilt() bool {
	return s.rebuilt.Load()
}

// resolvePatterns resolves raw patterns to absolute paths using the context's filesystem.
func (s *Step) resolvePatterns(ctx context.Context) ([]string, error) {
	fs := vfs.FSFromCtx(ctx)
	resolved := make([]string, len(s.fileDepsPatterns))
	for i, p := range s.fileDepsPatterns {
		abs, err := fs.Abs(p)
		if err != nil {
			Logger.Warn("Unable to resolve file pattern, defaulting to raw path",
				"pattern", p,
				"error", err,
			)
			abs = p
		}
		resolved[i] = abs
	}
	return resolved, nil
}

// needsRebuild returns nil if the step can be skipped, or a map of files which
// hashes are to be updated after the step is run.
func (s *Step) needsRebuild(ctx context.Context) (toSet map[string][]byte, err error) {
	toSet = map[string][]byte{}

	patterns, err := s.resolvePatterns(ctx)
	if err != nil {
		return nil, err
	}
	if len(patterns) == 0 {
		return nil, nil
	}

	fs := vfs.FSFromCtx(ctx)
	for _, fileDep := range patterns {
		files, err := fs.Glob(fileDep)
		if err != nil {
			return nil, err
		}

		if len(files) == 0 {
			Logger.Warn("No files matched dependency pattern",
				"pattern", fileDep,
				"step", s.name,
			)
		} else {
			Logger.Debug("Files matched",
				"pattern", fileDep,
				"files", files,
			)
		}

		for _, fp := range files {
			h, err := needsRebuild(ctx, s.name, fp)
			if err != nil {
				return nil, err
			}
			if h != nil {
				toSet[fp] = h
			}
		}
	}

	return toSet, nil
}

// CheckCycles verifies that there are no circular dependencies reachable from this step.
func (s *Step) CheckCycles() error {
	state := make(map[*Step]int) // 0: unvisited, 1: visiting, 2: visited
	var path []*Step

	var dfs func(curr *Step) error
	dfs = func(curr *Step) error {
		switch state[curr] {
		case 1:
			cycleStart := 0
			for i, node := range path {
				if node == curr {
					cycleStart = i
					break
				}
			}
			var cycleNames []string
			for _, node := range path[cycleStart:] {
				cycleNames = append(cycleNames, node.name)
			}
			cycleNames = append(cycleNames, curr.name)
			return fmt.Errorf("dependency cycle detected: %s", strings.Join(cycleNames, " -> "))
		case 2:
			return nil
		}

		state[curr] = 1
		path = append(path, curr)

		for _, dep := range curr.dependsOn {
			if err := dfs(dep); err != nil {
				return err
			}
		}

		path = path[:len(path)-1]
		state[curr] = 2
		return nil
	}

	return dfs(s)
}

// Run runs the step and all its dependencies after verifying the graph has no cycles.
func (s *Step) Run(ctx context.Context) error {
	if err := s.CheckCycles(); err != nil {
		return err
	}
	_, err := s.run(ctx)
	return err
}

func (s *Step) run(ctx context.Context) (rebuilt bool, err error) {
	if s.Done() {
		return s.rebuilt.Load(), nil
	}

	var depRebuilt bool
	for _, dep := range s.dependsOn {
		if !dep.Done() {
			r, err := dep.run(ctx)
			if err != nil {
				return false, err
			}
			if r {
				depRebuilt = true
			}
		} else if dep.Rebuilt() {
			depRebuilt = true
		}
	}

	var toSet map[string][]byte
	if len(s.fileDepsPatterns) > 0 {
		toSet, err = s.needsRebuild(ctx)
		if err != nil {
			return false, err
		} else if len(toSet) == 0 && !depRebuilt {
			Logger.Info("Skipping step", "step", s.name)
			s.done.Store(true)
			return false, nil
		}
	}

	Logger.Info("Running step", "step", s.name)
	for _, cmd := range s.commands {
		err = cmd.Run(ctx)
		if err != nil {
			Logger.Error("Step failed",
				"step", s.name,
				"error", err,
			)
			return false, err
		}
	}

	Logger.Info("Step completed", "step", s.name)
	s.done.Store(true)
	s.rebuilt.Store(true)

	for fp, h := range toSet {
		err = setHash(ctx, s.name, fp, h)
		if err != nil {
			Logger.Error("Unable to update cache for file",
				"file", fp,
				"error", err,
			)

			return false, err
		}
	}

	return true, nil
}

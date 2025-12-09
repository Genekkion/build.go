package build

import (
	"context"
	"path/filepath"
	"sync/atomic"
)

// Step represents a single build step.
type Step struct {
	name             string
	commands         []Command
	dependsOn        []*Step
	fileDepsPatterns []string
	done             atomic.Bool
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

// AddFileDeps adds file dependencies.
func (s *Step) AddFileDeps(patterns ...string) *Step {
	p := make([]string, len(patterns))
	var err error
	for i := range patterns {
		p[i], err = filepath.Abs(patterns[i])
		if err != nil {
			Logger.Warn("Unable to resolve file pattern, defaulting to relative path",
				"pattern", patterns[i],
				"error", err,
			)
			p[i] = patterns[i]
		}
	}
	Logger.Debug("Adding file dependencies", "patterns", p)
	s.fileDepsPatterns = append(s.fileDepsPatterns, p...)
	return s
}

// SetFileDeps sets the file dependencies.
func (s *Step) SetFileDeps(patterns []string) *Step {
	s.fileDepsPatterns = patterns
	return s
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

// needsRebuild returns nil if the step can be skipped, or a map of files which
// hashes are to be updated after the step is run.
func (s *Step) needsRebuild() (toSet map[string][]byte, err error) {
	toSet = map[string][]byte{}
	if len(s.fileDepsPatterns) == 0 {
		return nil, nil
	}

	for _, fileDep := range s.fileDepsPatterns {
		files, err := filepath.Glob(fileDep)
		if err != nil {
			return nil, err
		}

		Logger.Debug("Files matched",
			"pattern", fileDep,
			"files", files,
		)

		for _, fp := range files {
			h, err := needsRebuild(fp)
			if err != nil {
				return nil, err
			}
			if h != nil {
				toSet[fp] = h
			}
		}
	}

	s.done.Store(len(toSet) == 0)
	return toSet, nil
}

// Run runs the step.
func (s *Step) Run(ctx context.Context) (err error) {
	for _, dep := range s.dependsOn {
		if dep.Done() {
			continue
		}

		err = dep.Run(ctx)
		if err != nil {
			return err
		}
	}

	var toSet map[string][]byte
	if len(s.fileDepsPatterns) > 0 {
		toSet, err = s.needsRebuild()
		if err != nil {
			return err
		} else if len(toSet) == 0 {
			Logger.Info("Skipping step", "step", s.name)
			s.done.Store(true)
			return nil
		}
	}

	Logger.Info("Running step", "step", s.name)
	for _, cmd := range s.commands {
		err = cmd.Run(ctx)
		if err != nil {
			break
		}
	}

	if err != nil {
		Logger.Error("Step failed",
			"step", s.name,
			"error", err,
		)
		return err
	}

	Logger.Info("Step completed", "step", s.name)
	s.done.Store(true)

	for fp, h := range toSet {
		err = SetHash(fp, h)
		if err != nil {
			Logger.Error("Unable to update cache for file",
				"file", fp,
				"error", err,
			)

			return err
		}
	}

	return nil
}

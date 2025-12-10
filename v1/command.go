package buildgo

import "context"

// Command represents a runnable command as part of a build step.
type Command interface {
	// Run executes the command.
	Run(ctx context.Context) error
}

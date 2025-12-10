package inline

import (
	"errors"

	buildgo "github.com/Genekkion/build.go/v1"
)

// CmdFunc represents a command function.
type CmdFunc func() error

// Cmd represents a command.
type Cmd struct {
	funcs []CmdFunc
}

// NewCmd creates a new command.
func NewCmd(funcs []CmdFunc) (cmd *Cmd, err error) {
	if len(funcs) == 0 {
		return nil, errors.New("at least 1 function is required")
	}

	return &Cmd{
		funcs: funcs,
	}, nil
}

// Run runs the command.
func (c Cmd) Run() error {
	for _, f := range c.funcs {
		err := f()
		if err != nil {
			buildgo.Logger.Error("Command failed",
				"error", err,
			)
			return err
		}
	}
	return nil
}

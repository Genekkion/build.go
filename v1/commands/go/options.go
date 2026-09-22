package cmdgo

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	buildgo "github.com/Genekkion/build.go/v1"
)

// Config represents the configuration.
type Config struct {
	compilerPath string
	stdout       io.Writer
	stderr       io.Writer
}

// defaultConfig returns the default configuration.
func defaultConfig() Config {
	compilerPath, err := exec.LookPath("go")
	if err != nil {
		compilerPath = "go"
		buildgo.Logger.Warn(
			fmt.Sprintf(
				"Unable to find default go compiler, resort to using \"%s\"",
				compilerPath,
			),
			"error", err,
		)
	}

	return Config{
		compilerPath: compilerPath,
		stdout:       os.Stdout,
		stderr:       os.Stderr,
	}
}

// Option represents an option.
type Option func(*Config)

// WithCompilerPath sets the compiler path.
func WithCompilerPath(path string) Option {
	return func(cfg *Config) {
		cfg.compilerPath = path
	}
}

// WithStdout sets the stdout writer.
func WithStdout(stdout io.Writer) Option {
	return func(cfg *Config) {
		cfg.stdout = stdout
	}
}

// WithStderr sets the stderr writer.
func WithStderr(stderr io.Writer) Option {
	return func(cfg *Config) {
		cfg.stderr = stderr
	}
}

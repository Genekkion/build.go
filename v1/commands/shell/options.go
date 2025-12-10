package shell

import (
	"io"
	"os"
)

// Config represents the configuration.
type Config struct {
	cwd    string
	stdout io.Writer
	stderr io.Writer
}

// defaultConfig returns the default configuration.
func defaultConfig() Config {
	return Config{
		cwd:    ".",
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Option represents an option.
type Option func(*Config)

// WithCwd sets the working directory.
func WithCwd(cwd string) Option {
	return func(cfg *Config) {
		cfg.cwd = cwd
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

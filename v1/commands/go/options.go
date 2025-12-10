package cmdgo

import (
	"fmt"
	"os/exec"

	buildgo "github.com/Genekkion/build.go/v1"
)

// Config represents the configuration.
type Config struct {
	compilerPath string
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

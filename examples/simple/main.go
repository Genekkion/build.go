package main

import (
	"context"
	"path/filepath"

	buildgo "github.com/Genekkion/build.go/v1"
	cmdgo "github.com/Genekkion/build.go/v1/commands/go"
	"github.com/Genekkion/build.go/v1/commands/shell"
	"github.com/Genekkion/build.go/v1/fpath"
)

func main() {
	buildgo.Setup()
	defer buildgo.Cleanup()

	dir := filepath.Dir(fpath.CurrentFilePath())

	var firstStep *buildgo.Step
	{
		cmd, err := cmdgo.NewRunCmd(dir, []string{"./read_file"}, nil)
		if err != nil {
			panic(err)
		}
		firstStep = buildgo.NewStep("First step", cmd)
		firstStep.AddFileDeps(filepath.Join(dir, "read_file", "file.txt"))
	}

	var second *buildgo.Step
	{
		cmd, err := shell.NewCmd([]string{"echo", "second step"})
		if err != nil {
			panic(err)
		}

		second = buildgo.NewStep("Second step", cmd)
		second.DependsOn(firstStep)
	}

	err := second.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

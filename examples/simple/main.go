package main

import (
	"context"
	"path/filepath"

	buildgo "github.com/Genekkion/build.go/v1"
	"github.com/Genekkion/build.go/v1/commands/generic"
	cmdgo "github.com/Genekkion/build.go/v1/commands/go"
)

func main() {
	buildgo.Setup()
	defer buildgo.Cleanup()

	fp, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}

	var firstStep *buildgo.Step
	{
		cmd, err := cmdgo.NewRunCmd(fp, []string{"read_file"}, nil)
		if err != nil {
			panic(err)
		}
		firstStep = buildgo.NewStep("First step", cmd)
		firstStep.AddFileDeps("read_file/file.txt")
	}

	var second *buildgo.Step
	{
		cmd := generic.NewCmd([]string{"echo", "second step"})
		if err != nil {
			panic(err)
		}

		second = buildgo.NewStep("Second step", cmd)
		second.DependsOn(firstStep)
	}

	err = second.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

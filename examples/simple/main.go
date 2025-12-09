package main

import (
	"context"
	"path/filepath"

	"build/v1"
	"build/v1/commands/generic"
	cmdgo "build/v1/commands/go"
)

func main() {
	build.Setup()
	defer build.Cleanup()

	fp, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}

	var firstStep *build.Step
	{
		cmd, err := cmdgo.NewRunCmd(fp, []string{"read_file"}, nil)
		if err != nil {
			panic(err)
		}
		firstStep = build.NewStep("First step", cmd)
		firstStep.AddFileDeps("read_file/file.txt")
	}

	var second *build.Step
	{
		cmd := generic.NewCmd([]string{"echo", "second step"})
		if err != nil {
			panic(err)
		}

		second = build.NewStep("Second step", cmd)
		second.DependsOn(firstStep)
	}

	err = second.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

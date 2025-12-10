# Build.go

Build.go is meant to be a simple build system, that uses a main go file to specify build steps,
dependencies, and running commands in the right order.

## Usage

Firstly, download the dependency as:
```shell
go get github.com/Genekkion/build.go
```

### Step

A `Step` is a struct which holds details about a build step. You can use it to specify both other `Step`s as dependencies,
or file dependencies. It uses a simple sha256 hash to determine if a file has changed and requires rebuilding of the step.

### Command

A `Command` is a struct which holds details about a command to run. There is both a generic command for running shell
commands, and a special `Go` command for running `go` commands.

```go
package main

import (
    "context"
	buildgo "github.com/Genekkion/build.go/v1"
	"github.com/Genekkion/build.go/v1/commands/generic"
)

func main() {
	// These two functions are for setting up and tearing down the build system.
	// It handles cache-related things like setting up of the cache directory and db.
	// By default, it will use "./.gobuild" as the cache directory.
	buildgo.Setup()
	defer buildgo.Cleanup()

	cmd := generic.NewCmd([]string{
        "echo", "First step!"
    })
    firstStep := buildgo.NewStep("firstStep", cmd)
    
    cmd = generic.NewCmd([]string{
        "echo", "Second step!"
    })
    secondStep := buildgo.NewStep("secondStep", cmd)
    
    secondStep.DependsOn(firstStep)
    
    // By calling the `Run` method on the second step here, it will
    // see that it is dependent on the first step, and will run it first.
    err := secondStep.Run(context.Background())
    if err != nil {
        panic(err)
    }
}
```

## Example

Further examples can be found in the `examples` directory.

package dotnet

import (
	"github.com/jfrog/build-info-go/utils/dotnet"
	"github.com/jfrog/build-info-go/utils/dotnet/solution"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
)

type NugetCommand struct {
	*DotnetCommand
}

func NewNugetCommand() *NugetCommand {
	nugetCmd := NugetCommand{&DotnetCommand{}}
	nugetCmd.SetToolchainType(dotnet.Nuget)
	return &nugetCmd
}

func (nc *NugetCommand) Run() error {
	nc.useNugetAddSource = true
	return nc.Exec()
}

func DependencyTreeCmd() error {
	workspace, err := os.Getwd()
	if err != nil {
		return errorutils.CheckError(err)
	}

	sol, err := solution.Load(workspace, "")
	if err != nil {
		return err
	}

	// Create the tree for each project
	for _, project := range sol.GetProjects() {
		err = project.CreateDependencyTree()
		if err != nil {
			return err
		}
	}
	// Build the tree.
	content, err := sol.Marshal()
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Output(clientutils.IndentJson(content))
	return nil
}

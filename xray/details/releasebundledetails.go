package details

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands"
	"github.com/jfrog/jfrog-client-go/xray/services"
)

type RbDetailsCommand struct {
	serverDetails     *config.ServerDetails
	params            RbDetailsParams
	includeViolations bool
	result            *services.RBScanDetails
}

type RbDetailsParams struct {
	Name    string
	Version string
}

func (rbDetailsCmd *RbDetailsCommand) Result() *services.RBScanDetails {
	return rbDetailsCmd.result
}

func (rbDetailsCmd *RbDetailsCommand) SetParams(params RbDetailsParams) *RbDetailsCommand {
	rbDetailsCmd.params = params
	return rbDetailsCmd
}

func (rbDetailsCmd *RbDetailsCommand) SetServerDetails(server *config.ServerDetails) *RbDetailsCommand {
	rbDetailsCmd.serverDetails = server
	return rbDetailsCmd
}

func (rbDetailsCmd *RbDetailsCommand) SetIncludeViolations(include bool) *RbDetailsCommand {
	rbDetailsCmd.includeViolations = include
	return rbDetailsCmd
}

func (rbDetailsCmd *RbDetailsCommand) ServerDetails() (*config.ServerDetails, error) {
	return rbDetailsCmd.serverDetails, nil
}

func (rbDetailsCmd *RbDetailsCommand) Run() (err error) {
	xrayManager, err := commands.CreateXrayServiceManager(rbDetailsCmd.serverDetails)
	if err != nil {
		return err
	}
	response, err := xrayManager.GetRbDetails(rbDetailsCmd.params.Name, rbDetailsCmd.params.Version, rbDetailsCmd.includeViolations)
	if err != nil {
		return err
	}
	rbDetailsCmd.result = response
	return nil
}

func NewRbDetailsCommand() *RbDetailsCommand {
	return &RbDetailsCommand{}
}

func (rbDetailsCmd *RbDetailsCommand) CommandName() string {
	return "xr_rb_details"
}

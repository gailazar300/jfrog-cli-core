package utils

import (
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
)

// Remove all the none docker CLI flags from args.
func ExtractDockerOptionsFromArgs(args []string) (threads int, serverDetails *config.ServerDetails, detailedSummary, skipLogin bool, cleanArgs []string, buildConfig *utils.BuildConfiguration, err error) {
	cleanArgs = append([]string(nil), args...)
	var serverId string
	cleanArgs, serverId, err = coreutils.ExtractServerIdFromCommand(cleanArgs)
	if err != nil {
		return
	}
	serverDetails, err = config.GetSpecificConfig(serverId, true, true)
	if err != nil {
		return
	}
	cleanArgs, threads, err = coreutils.ExtractThreadsFromArgs(cleanArgs, 3)
	if err != nil {
		return
	}
	cleanArgs, detailedSummary, err = coreutils.ExtractDetailedSummaryFromArgs(cleanArgs)
	if err != nil {
		return
	}
	cleanArgs, skipLogin, err = coreutils.ExtractSkipLoginFromArgs(cleanArgs)
	if err != nil {
		return
	}
	cleanArgs, buildConfig, err = utils.ExtractBuildDetailsFromArgs(cleanArgs)
	return
}

func ExtractDockerScanOptionsFromArgs(args []string) (fail, licenses bool, format, project, watches string, serverDetails *config.ServerDetails, err error) {
	cleanArgs := append([]string(nil), args...)
	var serverId string
	cleanArgs, serverId, err = coreutils.ExtractServerIdFromCommand(cleanArgs)
	if err != nil {
		return
	}
	serverDetails, err = config.GetSpecificConfig(serverId, true, true)
	if err != nil {
		return
	}
	cleanArgs, fail, err = coreutils.ExtractFailFromArgs(cleanArgs)
	if err != nil {
		return
	}
	cleanArgs, format, err = coreutils.ExtractXrayOutputFormatFromArgs(cleanArgs)
	if err != nil {
		return
	}
	cleanArgs, licenses, err = coreutils.ExtractLicensesFromArgs(cleanArgs)
	if err != nil {
		return
	}
	cleanArgs, watches, err = coreutils.ExtractWatchesFromArgs(cleanArgs)
	if err != nil {
		return
	}
	var buildConfig *utils.BuildConfiguration
	_, buildConfig, err = utils.ExtractBuildDetailsFromArgs(cleanArgs)
	project = buildConfig.GetProject()
	return
}
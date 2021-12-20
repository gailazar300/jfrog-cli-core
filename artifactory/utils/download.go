package utils

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/io"
)

func CreateDownloadServiceManager(artDetails *config.ServerDetails, threads, httpRetries, httpRetryWaitTime int, dryRun bool, progressBar io.ProgressMgr) (artifactory.ArtifactoryServicesManager, error) {
	return CreateServiceManagerWithProgressBar(artDetails, threads, httpRetries, httpRetryWaitTime, dryRun, progressBar)
}

type DownloadConfiguration struct {
	Threads         int
	SplitCount      int
	MinSplitSize    int64
	Symlink         bool
	ValidateSymlink bool
}

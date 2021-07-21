package utils

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Check a case that targetRepository is not written in a deployableArtifacts file and needs to be read from the config file.
func TestUnmarshalDeployableArtifacts(t *testing.T) {
	err, cleanUpJfrogHome := CreateTempJfrogConfig()
	defer cleanUpJfrogHome()
	// DeployableArtifact file is being change during runtime so a copy needs to be created.
	tempDeployableArtifacts, err := createTempDeployableArtifactFile()
	defer os.Remove(tempDeployableArtifacts)
	r, err := UnmarshalDeployableArtifacts(path.Join(tempDeployableArtifacts), path.Join(getTestsDataGradlePath(), "config", "gradle.yaml"))
	assert.NoError(t, err)
	for transferDetails := new(clientutils.FileTransferDetails); r.reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
		assert.True(t, strings.HasPrefix(transferDetails.TargetPath, "http://localhost:8080/artifactory/"))
		assert.True(t, strings.Contains(transferDetails.TargetPath, "gradle-local-repo"))
	}
}

func createTempDeployableArtifactFile() (string, error) {
	testsDataGradlePath := getTestsDataGradlePath()
	summary, err := os.Open(path.Join(testsDataGradlePath, "deployableArtifacts", "summary"))
	tmpSummary, err := os.Create(path.Join(testsDataGradlePath, "deployableArtifacts", "tmpSummary"))
	buffer := []byte{}
	summary.Read(buffer)
	tmpSummary.Write(buffer)
	_, err = io.Copy(tmpSummary, summary)
	if err != nil {
		return "", err
	}
	return tmpSummary.Name(), nil
}

func getTestsDataGradlePath() string {
	return path.Join("..", "testdata", "gradle")
}

func CreateTempJfrogConfig() (err error, cleanUp func()) {
	err, cleanUp = tests.SetJfrogHome()
	if err != nil {
		return err, nil
	}
	homedir := os.Getenv(coreutils.HomeDir)

	configuration := `
		{
		  "artifactory": [
			{
			  "url": "http://localhost:8080/artifactory/",
			  "user": "user",
			  "password": "password",
			  "serverId": "name",
			  "isDefault": true
			},
			{
			  "url": "http://localhost:8080/artifactory/",
			  "user": "user",
			  "password": "password",
			  "serverId": "notDefault"
			}
		  ],
		  "version": "2"
		}
	`
	content, err := config.ConvertIfNeeded([]byte(configuration))
	if err != nil {
		return
	}

	versionString := ".v" + strconv.Itoa(coreutils.GetConfigVersion())
	configFile, err := os.Create(filepath.Join(homedir, coreutils.JfrogConfigFile+versionString))

	//configFile, err := os.Create(homedir)
	if err != nil {
		return
	}
	defer configFile.Close()
	_, err = configFile.Write(content)
	if err != nil {
		return
	}
	return
}

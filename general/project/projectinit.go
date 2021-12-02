package project

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
	artifactoryCommandsUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"gopkg.in/yaml.v2"
)

const (
	buildFileName = "build.yaml"
)

type ProjectInitCommand struct {
	projectPath string
	serverId    string
}

func NewProjectInitCommand() *ProjectInitCommand {
	return &ProjectInitCommand{}
}

func (pic *ProjectInitCommand) SetProjectPath(path string) *ProjectInitCommand {
	pic.projectPath = path
	return pic
}

func (pic *ProjectInitCommand) SetServerId(id string) *ProjectInitCommand {
	pic.serverId = id
	return pic
}

func (pic *ProjectInitCommand) Run() (err error) {
	if pic.serverId == "" {
		defaultServer, err := config.GetSpecificConfig("", true, false)
		if err != nil {
			return err
		}
		pic.serverId = defaultServer.ServerId
	}
	technologiesMap, err := coreutils.DetectTechnologies(pic.projectPath, false, false)
	if err != nil {
		return err
	}
	// First create repositories for the detected technologies.
	for tech, detected := range technologiesMap {
		if detected {
			// First create repositories for the detected technology.
			err = createDefaultReposIfNeeded(tech, pic.serverId)
			if err != nil {
				return err
			}
			err = createProjectBuildConfigs(tech, pic.projectPath, pic.serverId)
			if err != nil {
				return err
			}
		}
	}
	// Create build config
	if err = pic.createBuildConfig(); err != nil {
		return
	}

	message :=
		bold("You're all set!") +
			"\n\n" +
			bold("The project config is stored inside the .jfrog directory.") +
			"\n\n" +
			title("Audit your code project for security vulnerabilities by running") +
			"\n" +
			"jf audit\n\n" +
			title("Scan any software package on this machine for security vulnerabilities by running") +
			"\n" +
			"jf scan path/to/dir/or/package\n\n" +
			pic.createBuildMessage(technologiesMap) +
			title("Read more using this link:") +
			"\n" +
			link(coreutils.GettingStartedGuideUrl)
	fmt.Println()
	err = coreutils.PrintTable("", "", message)
	fmt.Println()

	return
}

// Return a string message, which includes all the build and deployment commands, matching the technologiesMap sent.
func (pic *ProjectInitCommand) createBuildMessage(technologiesMap map[coreutils.Technology]bool) string {
	message := ""
	for tech, detected := range technologiesMap {
		if detected {
			switch tech {
			case coreutils.Maven:
				message += "jf mvn install deploy\n"
			case coreutils.Gradle:
				message += "jf gradle artifactoryP\n"
			case coreutils.Npm:
				message += "jf npm install publish\n"
			case coreutils.Go:
				message +=
					"jf go build\n" +
						"jf go-publish v1.0.0\n"
			case coreutils.Pypi:
				message +=
					"jf pip install\n" +
						"jf rt u path/to/package/file default-pypi-local" + comment(" # Publish your pip package") +
						"\n"
			}
		}
	}
	if message != "" {
		message = title("Build the code & deploy the packages by running") +
			"\n" +
			message +
			"\n" +
			title("Publish the build-info to Artifactory") +
			"\n" +
			"jf rt bp\n\n"
	}
	return message
}

// Print the test to the console in green color.
func title(str string) string {
	return colorStr(str, color.Green)
}

// Print the test to the console in cyan color.
func link(str string) string {
	return colorStr(str, color.Cyan)
}

// Print the test to the console with bold style.
func bold(str string) string {
	return colorStr(str, color.Bold)
}

// Print the test to the console in gray color.
func comment(str string) string {
	return colorStr(str, color.Gray)
}

// Print the test to the console with the specified color.
func colorStr(str string, c color.Color) string {
	if coreutils.IsTerminal() {
		return c.Render(str)
	}
	return str
}

type BuildConfigFile struct {
	Version    int    `yaml:"version,omitempty"`
	ConfigType string `yaml:"type,omitempty"`
	BuildName  string `yaml:"name,omitempty"`
}

func (pic *ProjectInitCommand) createBuildConfig() error {
	jfrogProjectDir := filepath.Join(pic.projectPath, ".jfrog", "projects")
	if err := fileutils.CreateDirIfNotExist(jfrogProjectDir); err != nil {
		return errorutils.CheckError(err)
	}
	configFilePath := filepath.Join(jfrogProjectDir, buildFileName)
	projectDirName := filepath.Base(filepath.Dir(pic.projectPath))
	buildConfigFile := &BuildConfigFile{Version: 1, ConfigType: "build", BuildName: projectDirName}
	resBytes, err := yaml.Marshal(&buildConfigFile)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return errorutils.CheckError(ioutil.WriteFile(configFilePath, resBytes, 0644))
}

func createDefaultReposIfNeeded(tech coreutils.Technology, serverId string) error {
	err := CreateDefaultLocalRepo(tech, serverId)
	if err != nil {
		return err
	}
	err = CreateDefaultRemoteRepo(tech, serverId)
	if err != nil {
		return err
	}

	return CreateDefaultVirtualRepo(tech, serverId)
}

func createProjectBuildConfigs(tech coreutils.Technology, projectPath string, serverId string) error {
	jfrogProjectDir := filepath.Join(projectPath, ".jfrog", "projects")
	if err := fileutils.CreateDirIfNotExist(jfrogProjectDir); err != nil {
		return errorutils.CheckError(err)
	}
	techName := strings.ToLower(string(tech))
	// Due to cli-artifactory naming mismatch we have to add this line
	if tech == coreutils.Pypi {
		techName = "pip"
	}
	configFilePath := filepath.Join(jfrogProjectDir, techName+".yaml")
	configFile := artifactoryCommandsUtils.ConfigFile{
		Version:    artifactoryCommandsUtils.BuildConfVersion,
		ConfigType: techName,
	}
	configFile.Resolver = artifactoryUtils.Repository{ServerId: serverId}
	configFile.Deployer = artifactoryUtils.Repository{ServerId: serverId}
	switch tech {
	case coreutils.Maven:
		configFile.Resolver.ReleaseRepo = MavenVirtualDefaultName
		configFile.Resolver.SnapshotRepo = MavenVirtualDefaultName
		configFile.Deployer.ReleaseRepo = MavenVirtualDefaultName
		configFile.Deployer.SnapshotRepo = MavenVirtualDefaultName
	case coreutils.Gradle:
		configFile.Resolver.Repo = GradleVirtualDefaultName
		configFile.Deployer.Repo = GradleVirtualDefaultName
	case coreutils.Npm:
		configFile.Resolver.Repo = NpmVirtualDefaultName
		configFile.Deployer.Repo = NpmVirtualDefaultName
	case coreutils.Go:
		configFile.Resolver.Repo = GoVirtualDefaultName
		configFile.Deployer.Repo = GoVirtualDefaultName
	case coreutils.Pypi:
		configFile.Resolver.Repo = PypiVirtualDefaultName
		configFile.Deployer.Repo = PypiVirtualDefaultName
	}
	resBytes, err := yaml.Marshal(&configFile)
	if err != nil {
		return errorutils.CheckError(err)
	}

	return errorutils.CheckError(ioutil.WriteFile(configFilePath, resBytes, 0644))
}
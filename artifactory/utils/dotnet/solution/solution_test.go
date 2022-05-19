package solution

import (
	"encoding/json"
	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestEmptySolution(t *testing.T) {
	solution, err := Load(".", "")
	if err != nil {
		t.Error(err)
	}

	expected := &buildinfo.BuildInfo{}
	buildInfo, err := solution.BuildInfo("")
	if err != nil {
		t.Error("An error occurred while creating the build info object", err.Error())
	}
	if !reflect.DeepEqual(buildInfo, expected) {
		expectedString, _ := json.Marshal(expected)
		buildInfoString, _ := json.Marshal(buildInfo)
		t.Errorf("Expecting: \n%s \nGot: \n%s", expectedString, buildInfoString)
	}
}

func TestParseSln(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	testdataDir := filepath.Join(pwd, "testdata")

	tests := []struct {
		name     string
		slnPath  string
		expected []string
	}{
		{"oneproject", filepath.Join(testdataDir, "oneproject.sln"), []string{`Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfig", "packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`}},
		{"multiProjects", filepath.Join(testdataDir, "multiprojects.sln"), []string{`Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfigmulti", "packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`, `Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfiganothermulti", "test\packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, err := parseSlnFile(test.slnPath)
			if err != nil {
				t.Error(err)
			}

			replaceCarriageSign(results)

			if !reflect.DeepEqual(test.expected, results) {
				t.Errorf("Expected %s, got %s", test.expected, results)
			}
		})
	}
}

func TestParseProjectLine(t *testing.T) {
	tests := []struct {
		name                 string
		projectLine          string
		expectedProjFilePath string
		expectedProjectName  string
	}{
		{"packagename", `Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagename", "packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`, filepath.Join("jfrog", "path", "test", "packagesconfig.csproj"), "packagename"},
		{"withpath", `Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagename", "packagesconfig/packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`, filepath.Join("jfrog", "path", "test", "packagesconfig", "packagesconfig.csproj"), "packagename"},
		{"sameprojectname", `Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfig", "packagesconfig/packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`, filepath.Join("jfrog", "path", "test", "packagesconfig", "packagesconfig.csproj"), "packagesconfig"},
		{"vbproj", `Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfig", "packagesconfig/packagesconfig.vbproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`, filepath.Join("jfrog", "path", "test", "packagesconfig", "packagesconfig.vbproj"), "packagesconfig"},
	}

	path := filepath.Join("jfrog", "path", "test")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projectName, projFilePath, err := parseProjectLine(test.projectLine, path)
			if err != nil {
				t.Error(err)
			}
			if projFilePath != test.expectedProjFilePath {
				t.Errorf("Expected %s, got %s", test.expectedProjFilePath, projFilePath)
			}
			if projectName != test.expectedProjectName {
				t.Errorf("Expected %s, got %s", test.expectedProjectName, projectName)
			}
		})
	}
}

func TestGetProjectsFromSlns(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	testdataDir := filepath.Join(pwd, "testdata")
	tests := []struct {
		name             string
		solution         solution
		expectedProjects []string
	}{
		{"withoutSlnFile", solution{path: testdataDir, slnFile: "", projects: nil}, []string{`Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfigmulti", "packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`, `Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfiganothermulti", "test\packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`, `Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfig", "packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`},
		},
		{"withSlnFile", solution{path: testdataDir, slnFile: "oneproject.sln", projects: nil}, []string{`Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "packagesconfig", "packagesconfig.csproj", "{D1FFA0DC-0ACC-4108-ADC1-2A71122C09AF}"
EndProject`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, err := test.solution.getProjectsFromSlns()
			if err != nil {
				t.Error(err)
			}
			replaceCarriageSign(results)

			if !reflect.DeepEqual(test.expectedProjects, results) {
				t.Errorf("Expected %s, got %s", test.expectedProjects, results)
			}
		})
	}
}

// If running on Windows, replace \r\n with \n.
func replaceCarriageSign(results []string) {
	if runtime.GOOS == "windows" {
		for i, result := range results {
			results[i] = strings.Replace(result, "\r\n", "\n", -1)
		}
	}
}

//func TestLoad2(t *testing.T) {
//	log.SetDefaultLogger()
//	wd, err := os.Getwd()
//	if err != nil {
//		t.Error(err)
//	}
//	// 'nugetproj' contains 2 'packages.config' files for 2 projects - one file is located in the project's root dir and the other in solutions dir.
//	solutions, err := Load(filepath.Join(wd, "testdata", "nugetproj", "solutions"), "nugetproj.sln")
//	if err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, 2, len(solutions.GetProjects()))
//}

func TestLoad(t *testing.T) {
	log.SetDefaultLogger()
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name             string
		solution         solution
		expectedProjects int
	}{
		{"withoutSlnFile", solution{path: filepath.Join(pwd, "testdata", "nugetproj", "solutions"), slnFile: "nugetproj.sln"}, 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			solutions, err := Load(test.solution.path, test.solution.slnFile)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, test.expectedProjects, len(solutions.GetProjects()))
		})
	}
}

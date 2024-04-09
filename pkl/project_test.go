package pkl_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/apple/pkl-go/pkl"
	"github.com/stretchr/testify/assert"
)

const project1Contents = `
amends "pkl:Project"

evaluatorSettings {
  timeout = 5.min
  rootDir = "."
  noCache = false
  moduleCacheDir = "cache/"
  env {
    ["one"] = "1"
  }
  externalProperties {
    ["two"] = "2"
  }
  modulePath {
    "modulepath1/"
    "modulepath2/"
  }
  allowedModules {
    "foo:"
    "bar:"
  }
  allowedResources {
    "baz:"
    "biz:"
  }
}

package {
  name = "hawk"
  baseUri = "package://example.com/hawk"
  version = "0.5.0"
  description = "Some project about hawks"
  packageZipUrl = "https://example.com/hawk/\(version)/hawk-\(version).zip"
  authors {
    "Birdy Bird <birdy@bird.com>"
  }
  license = "MIT"
  licenseText = """
    # Some License text
    
    This is my license text
    """
  sourceCode = "https://example.com/my/repo"
  sourceCodeUrlScheme = "https://example.com/my/repo/\(version)%{path}"
  documentation = "https://example.com/my/docs"
  website = "https://example.com/my/website"
  apiTests {
    "apiTest1.pkl"
    "apiTest2.pkl"
  }
  exclude { "*.exe" }
  issueTracker = "https://example.com/my/issues"
}

dependencies {
  ["flamingos"] { uri = "package://example.com/flamingos@0.5.0" }
  ["storks"] = import("../storks/PklProject")
}

tests {
  "test1.pkl"
  "test2.pkl"
}
`

const project2Contents = `
amends "pkl:Project"

package {
  name = "storks"
  baseUri = "package://example.com/storks"
  version = "0.5.0"
  packageZipUrl = "https://example.com/stork/\(version)/stork-\(version).zip"
}
`

func writeFile(t *testing.T, filename string, contents string) {
	if err := os.WriteFile(filename, []byte(contents), 0o777); err != nil {
		t.Logf("Failed to write file %s: %s", filename, err)
		t.FailNow()
	}
}

func TestLoadProject(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.Mkdir(tempDir+"/hawks", 0o777)
	_ = os.Mkdir(tempDir+"/storks", 0o777)
	writeFile(t, tempDir+"/hawks/PklProject", project1Contents)
	writeFile(t, tempDir+"/storks/PklProject", project2Contents)
	project, err := pkl.LoadProject(context.Background(), tempDir+"/hawks/PklProject")
	if assert.NoError(t, err) {
		t.Run("projectFileUri", func(t *testing.T) {
			assert.Equal(t, fmt.Sprintf("file://%s/hawks/PklProject", tempDir), project.ProjectFileUri)
		})

		t.Run("evaluatorSettings", func(t *testing.T) {
			fals := false
			expectedSettings := &pkl.ProjectEvaluatorSettings{
				Timeout: pkl.Duration{
					Value: 5,
					Unit:  pkl.Minute,
				},
				NoCache:            &fals,
				RootDir:            ".",
				ModuleCacheDir:     "cache/",
				Env:                map[string]string{"one": "1"},
				ExternalProperties: map[string]string{"two": "2"},
				ModulePath:         []string{"modulepath1/", "modulepath2/"},
				AllowedModules:     []string{"foo:", "bar:"},
				AllowedResources:   []string{"baz:", "biz:"},
			}
			assert.Equal(t, expectedSettings, project.EvaluatorSetings)
		})

		t.Run("package", func(t *testing.T) {
			expectedPackage := &pkl.ProjectPackage{
				Name:                "hawk",
				BaseUri:             "package://example.com/hawk",
				Version:             "0.5.0",
				Description:         "Some project about hawks",
				PackageZipUrl:       "https://example.com/hawk/0.5.0/hawk-0.5.0.zip",
				Authors:             []string{"Birdy Bird <birdy@bird.com>"},
				License:             "MIT",
				LicenseText:         "# Some License text\n\nThis is my license text",
				SourceCode:          "https://example.com/my/repo",
				SourceCodeUrlScheme: "https://example.com/my/repo/0.5.0%{path}",
				Documentation:       "https://example.com/my/docs",
				Website:             "https://example.com/my/website",
				ApiTests:            []string{"apiTest1.pkl", "apiTest2.pkl"},
				Exclude:             []string{"PklProject", "PklProject.deps.json", ".**", "*.exe"},
				IssueTracker:        "https://example.com/my/issues",
				Uri:                 "package://example.com/hawk@0.5.0",
			}
			assert.Equal(t, expectedPackage, project.Package)
		})

		t.Run("dependencies", func(t *testing.T) {
			expectedDependences := &pkl.ProjectDependencies{
				RemoteDependencies: map[string]*pkl.ProjectRemoteDependency{
					"flamingos": {PackageUri: "package://example.com/flamingos@0.5.0"},
				},
				LocalDependencies: map[string]*pkl.ProjectLocalDependency{
					"storks": {
						ProjectFileUri: fmt.Sprintf("file://%s/storks/PklProject", tempDir),
						PackageUri:     "package://example.com/storks@0.5.0",
						Dependencies: &pkl.ProjectDependencies{
							LocalDependencies:  map[string]*pkl.ProjectLocalDependency{},
							RemoteDependencies: map[string]*pkl.ProjectRemoteDependency{},
						},
					},
				},
			}
			assert.Equal(t, expectedDependences, project.Dependencies())
		})

		t.Run("tests", func(t *testing.T) {
			assert.Equal(t, []string{"test1.pkl", "test2.pkl"}, project.Tests)
		})
	}
}

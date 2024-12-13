// ===----------------------------------------------------------------------===//
// Copyright © 2024 Apple Inc. and the Pkl project authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ===----------------------------------------------------------------------===//

package pkl

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const project1Contents = `
@ModuleInfo { minPklVersion = "0.25.0" }
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

const project3Contents = `
amends "pkl:Project"

evaluatorSettings {
  http {
    proxy {
      address = "http://localhost:80"
      noProxy {
        "127.0.0.1"
        "192.168.0.1/24"
        "example.com"
        "localhost:8000"
      }
    }
  }
}

package {
  name = "pigeon"
  baseUri = "package://example.com/pigeon"
  version = "0.26.0"
  description = "Some project about pigeons"
  packageZipUrl = "https://example.com/pigeon/\(version)/pigeon-\(version).zip"
}
`

const project4Contents = `
amends "pkl:Project"

evaluatorSettings {
  externalModuleReaders {
		["scheme1"] {
			executable = "reader1"
		}
		["scheme2"] {
			executable = "reader2"
			arguments { "with"; "args" }
		}
	}
	externalResourceReaders {
		["scheme3"] {
			executable = "reader3"
		}
		["scheme4"] {
			executable = "reader4"
			arguments { "with"; "args" }
		}
	}
}

package {
  name = "pigeon"
  baseUri = "package://example.com/pigeon"
  version = "0.26.0"
  description = "Some project about pigeons"
  packageZipUrl = "https://example.com/pigeon/\(version)/pigeon-\(version).zip"
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
	project, err := LoadProject(context.Background(), tempDir+"/hawks/PklProject")
	if assert.NoError(t, err) {
		t.Run("projectFileUri", func(t *testing.T) {
			assert.Equal(t, fmt.Sprintf("file://%s/hawks/PklProject", tempDir), project.ProjectFileUri)
		})

		t.Run("annotations", func(t *testing.T) {
			manager := NewEvaluatorManager()
			defer manager.Close()
			version, err := manager.(*evaluatorManager).getVersion()
			if err != nil {
				t.Fatal(err)
			}
			if version.isLessThan(pklVersion0_27) {
				t.SkipNow()
			}
			assert.Len(t, project.Annotations, 1)
			assert.Equal(t, project.Annotations[0].Properties["minPklVersion"], "0.25.0")
		})

		t.Run("evaluatorSettings", func(t *testing.T) {
			fals := false
			expectedSettings := &ProjectEvaluatorSettings{
				Timeout: Duration{
					Value: 5,
					Unit:  Minute,
				},
				NoCache:            &fals,
				RootDir:            ".",
				ModuleCacheDir:     "cache/",
				Env:                map[string]string{"one": "1"},
				ExternalProperties: map[string]string{"two": "2"},
				ModulePath:         []string{"modulepath1/", "modulepath2/"},
				AllowedModules:     &[]string{"foo:", "bar:"},
				AllowedResources:   &[]string{"baz:", "biz:"},
			}
			assert.Equal(t, expectedSettings, project.EvaluatorSettings)
		})

		t.Run("package", func(t *testing.T) {
			expectedPackage := &ProjectPackage{
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
			expectedDependences := &ProjectDependencies{
				RemoteDependencies: map[string]*ProjectRemoteDependency{
					"flamingos": {PackageUri: "package://example.com/flamingos@0.5.0"},
				},
				LocalDependencies: map[string]*ProjectLocalDependency{
					"storks": {
						ProjectFileUri: fmt.Sprintf("file://%s/storks/PklProject", tempDir),
						PackageUri:     "package://example.com/storks@0.5.0",
						Dependencies: &ProjectDependencies{
							LocalDependencies:  map[string]*ProjectLocalDependency{},
							RemoteDependencies: map[string]*ProjectRemoteDependency{},
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

func TestLoadProjectWithProxy(t *testing.T) {
	manager := NewEvaluatorManager()
	version, err := manager.(*evaluatorManager).getVersion()
	if err != nil {
		t.Fatal(err)
	}
	if pklVersion0_26.isGreaterThan(version) {
		t.SkipNow()
	}

	tempDir := t.TempDir()
	_ = os.Mkdir(tempDir+"/pigeons", 0o777)
	writeFile(t, tempDir+"/pigeons/PklProject", project3Contents)

	project, err := LoadProject(context.Background(), tempDir+"/pigeons/PklProject")
	if assert.NoError(t, err) {
		t.Run("evaluatorSettings", func(t *testing.T) {
			expectedSettings := &ProjectEvaluatorSettings{
				Http: &ProjectEvaluatorSettingsHttp{
					Proxy: &ProjectEvaluatorSettingsProxy{
						Address: &[]string{"http://localhost:80"}[0],
						NoProxy: &[]string{
							"127.0.0.1",
							"192.168.0.1/24",
							"example.com",
							"localhost:8000",
						},
					},
				},
			}
			assert.Equal(t, expectedSettings, project.EvaluatorSettings)
		})
	}
}

func TestLoadProjectWithExternalReaders(t *testing.T) {
	manager := NewEvaluatorManager()
	version, err := manager.(*evaluatorManager).getVersion()
	if err != nil {
		t.Fatal(err)
	}
	if pklVersion0_27.isGreaterThan(version) {
		t.SkipNow()
	}

	tempDir := t.TempDir()
	_ = os.Mkdir(tempDir+"/pigeons", 0o777)
	writeFile(t, tempDir+"/pigeons/PklProject", project4Contents)

	project, err := LoadProject(context.Background(), tempDir+"/pigeons/PklProject")
	if assert.NoError(t, err) {
		t.Run("evaluatorSettings", func(t *testing.T) {
			expectedSettings := &ProjectEvaluatorSettings{
				ExternalModuleReaders: map[string]ProjectEvaluatorSettingExternalReader{
					"scheme1": {Executable: "reader1"},
					"scheme2": {Executable: "reader2", Arguments: []string{"with", "args"}},
				},
				ExternalResourceReaders: map[string]ProjectEvaluatorSettingExternalReader{
					"scheme3": {Executable: "reader3"},
					"scheme4": {Executable: "reader4", Arguments: []string{"with", "args"}},
				},
			}
			assert.Equal(t, expectedSettings, project.EvaluatorSettings)
		})
	}
}

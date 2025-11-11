//===----------------------------------------------------------------------===//
// Copyright © 2024-2025 Apple Inc. and the Pkl project authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//===----------------------------------------------------------------------===//

package pkl

import (
	"context"

	"github.com/apple/pkl-go/pkl/internal"
)

// needed for mapping Project.RawDependencies, because the value is defined as any.
func init() {
	RegisterStrictMapping("pkl.Project", &Project{})
	RegisterStrictMapping("pkl.Project#RemoteDependency", &ProjectRemoteDependency{})
}

// Project is the go representation of pkl.Project.
type Project struct {
	Package *ProjectPackage `pkl:"package"`

	// internal field; use Project.Dependencies instead.
	// values are either Project or ProjectRemoteDependency
	RawDependencies map[string]any `pkl:"dependencies"`

	dependencies      *ProjectDependencies     `pkl:"-"`
	ProjectFileUri    string                   `pkl:"projectFileUri"`
	Tests             []string                 `pkl:"tests"`
	Annotations       []Object                 `pkl:"annotations"`
	EvaluatorSettings ProjectEvaluatorSettings `pkl:"evaluatorSettings"`
}

// ProjectPackage is the go representation of pkl.Project#Package.
type ProjectPackage struct {
	Name                string   `pkl:"name"`
	BaseUri             string   `pkl:"baseUri"`
	Version             string   `pkl:"version"`
	PackageZipUrl       string   `pkl:"packageZipUrl"`
	Description         string   `pkl:"description"`
	Website             string   `pkl:"website"`
	Documentation       string   `pkl:"documentation"`
	SourceCode          string   `pkl:"sourceCode"`
	SourceCodeUrlScheme string   `pkl:"sourceCodeUrlScheme"`
	License             string   `pkl:"license"`
	LicenseText         string   `pkl:"licenseText"`
	IssueTracker        string   `pkl:"issueTracker"`
	Uri                 string   `pkl:"uri"`
	Authors             []string `pkl:"authors"`
	ApiTests            []string `pkl:"apiTests"`
	Exclude             []string `pkl:"exclude"`
}

// ProjectEvaluatorSettings is the Go representation of pkl.EvaluatorSettings
type ProjectEvaluatorSettings struct {
	ExternalProperties      map[string]string                                `pkl:"externalProperties"`
	Env                     map[string]string                                `pkl:"env"`
	AllowedModules          *[]string                                        `pkl:"allowedModules"`
	AllowedResources        *[]string                                        `pkl:"allowedResources"`
	NoCache                 *bool                                            `pkl:"noCache"`
	Http                    *ProjectEvaluatorSettingsHttp                    `pkl:"http"`
	ExternalModuleReaders   map[string]ProjectEvaluatorSettingExternalReader `pkl:"externalModuleReaders"`
	ExternalResourceReaders map[string]ProjectEvaluatorSettingExternalReader `pkl:"externalResourceReaders"`
	TraceMode               *TraceMode                                       `pkl:"traceMode"`
	ModuleCacheDir          string                                           `pkl:"moduleCacheDir"`
	RootDir                 string                                           `pkl:"rootDir"`
	Color                   string                                           `pkl:"color"`
	ModulePath              []string                                         `pkl:"modulePath"`
	Timeout                 Duration                                         `pkl:"timeout"`
}

// ProjectEvaluatorSettingsHttp is the Go representation of pkl.EvaluatorSettings.Http
type ProjectEvaluatorSettingsHttp struct {
	Proxy    *ProjectEvaluatorSettingsProxy `pkl:"proxy"`
	Rewrites *map[string]string             `pkl:"rewrites"`
}

// ProjectEvaluatorSettingsProxy is the Go representation of pkl.EvaluatorSettings.Proxy
type ProjectEvaluatorSettingsProxy struct {
	Address *string   `pkl:"address"`
	NoProxy *[]string `pkl:"noProxy"`
}

// ProjectEvaluatorSettingExternalReader is the Go representation of pkl.EvaluatorSettings.ExternalReader
type ProjectEvaluatorSettingExternalReader struct {
	Executable string   `pkl:"executable"`
	Arguments  []string `pkl:"arguments"`
}

func (project *Project) Dependencies() *ProjectDependencies {
	if project.dependencies == nil {
		deps := ProjectDependencies{
			LocalDependencies:  make(map[string]*ProjectLocalDependency),
			RemoteDependencies: make(map[string]*ProjectRemoteDependency),
		}
		for name, dep := range project.RawDependencies {
			if proj, ok := dep.(*Project); ok {
				localDep := &ProjectLocalDependency{
					PackageUri:     proj.Package.Uri,
					ProjectFileUri: proj.ProjectFileUri,
					Dependencies:   proj.Dependencies(),
				}
				deps.LocalDependencies[name] = localDep
				continue
			}
			if remoteDep, ok := dep.(*ProjectRemoteDependency); ok {
				deps.RemoteDependencies[name] = remoteDep
				continue
			}
			// If we get here, the most likely explanation is that a Project was manually
			// initialized and RawDependencies was set incorrectly.
			internal.Debug("Invalid dependency type: %+v", dep)
		}
		project.dependencies = &deps
	}
	return project.dependencies
}

// LoadProject loads a project definition from the specified path directory.
func LoadProject(context context.Context, path string) (*Project, error) {
	ev, err := NewEvaluator(context, PreconfiguredOptions)
	if err != nil {
		return nil, err
	}
	return LoadProjectFromEvaluator(context, ev, FileSource(path))
}

// LoadProjectFromEvaluator is like LoadProject, but uses the already provisioned evaluator, and loads the project
// from a ModuleSource rather than a path string.
func LoadProjectFromEvaluator(context context.Context, ev Evaluator, projectSource *ModuleSource) (*Project, error) {
	var proj *Project
	if err := ev.EvaluateOutputValue(context, projectSource, &proj); err != nil {
		return nil, err
	}
	return proj, nil
}

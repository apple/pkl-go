package pkl

import (
	"context"

	"github.com/apple/pkl-go/pkl/internal"
)

// needed for mapping Project.RawDependencies, because the value is defined as any.
func init() {
	RegisterMapping("pkl.Project", Project{})
	RegisterMapping("pkl.Project#RemoteDependency", ProjectRemoteDependency{})
}

// Project is the go representation of pkl.Project.
type Project struct {
	ProjectFileUri    string                    `pkl:"projectFileUri"`
	Package           *ProjectPackage           `pkl:"package"`
	EvaluatorSettings *ProjectEvaluatorSettings `pkl:"evaluatorSettings"`
	Tests             []string                  `pkl:"tests"`

	// internal field; use Project.Dependencies instead.
	// values are either *Project or *ProjectRemoteDependency
	RawDependencies map[string]any `pkl:"dependencies"`

	dependencies *ProjectDependencies `pkl:"-"`
}

// ProjectPackage is the go representation of pkl.Project#Package.
type ProjectPackage struct {
	Name                string   `pkl:"name"`
	BaseUri             string   `pkl:"baseUri"`
	Version             string   `pkl:"version"`
	PackageZipUrl       string   `pkl:"packageZipUrl"`
	Description         string   `pkl:"description"`
	Authors             []string `pkl:"authors"`
	Website             string   `pkl:"website"`
	Documentation       string   `pkl:"documentation"`
	SourceCode          string   `pkl:"sourceCode"`
	SourceCodeUrlScheme string   `pkl:"sourceCodeUrlScheme"`
	License             string   `pkl:"license"`
	LicenseText         string   `pkl:"licenseText"`
	IssueTracker        string   `pkl:"issueTracker"`
	ApiTests            []string `pkl:"apiTests"`
	Exclude             []string `pkl:"exclude"`
	Uri                 string   `pkl:"uri"`
}

// ProjectEvaluatorSettings is the Go representation of pkl.Project#EvaluatorSettings
type ProjectEvaluatorSettings struct {
	ExternalProperties map[string]string `pkl:"externalProperties"`
	Env                map[string]string `pkl:"env"`
	AllowedModules     []string          `pkl:"allowedModules"`
	AllowedResources   []string          `pkl:"allowedResources"`
	NoCache            *bool             `pkl:"noCache"`
	ModulePath         []string          `pkl:"modulePath"`
	Timeout            Duration          `pkl:"timeout"`
	ModuleCacheDir     string            `pkl:"moduleCacheDir"`
	RootDir            string            `pkl:"rootDir"`
}

func (project *Project) Dependencies() *ProjectDependencies {
	if project.dependencies == nil {
		var deps ProjectDependencies
		deps.LocalDependencies = make(map[string]*ProjectLocalDependency)
		deps.RemoteDependencies = make(map[string]*ProjectRemoteDependency)
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
	return LoadProjectFromEvaluator(context, ev, path)
}

func LoadProjectFromEvaluator(context context.Context, ev Evaluator, path string) (*Project, error) {
	var proj Project
	if err := ev.EvaluateOutputValue(context, FileSource(path), &proj); err != nil {
		return nil, err
	}
	return &proj, nil
}

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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/apple/pkl-go/pkl/internal/msgapi"
)

// EvaluatorOptions is the set of options available to control Pkl evaluation.
type EvaluatorOptions struct {
	// Properties is the set of properties available to the `prop:` resource reader.
	Properties map[string]string

	// Env is the set of environment variables available to the `env:` resource reader.
	Env map[string]string

	// ModulePaths is the set of directories, ZIP archives, or JAR archives to search when
	// resolving `modulepath`: resources and modules.
	//
	// This option must be non-emptyMirror if ModuleReaderModulePath or ResourceModulePath are used.
	ModulePaths []string

	// Logger is the logging interface for messages emitted by the Pkl evaluator.
	Logger Logger

	// OutputFormat controls the renderer to be used when rendering the `output.text`
	// property of a module.
	//
	// The supported built-in values are:
	//   - `"json"`
	//   - `"jsonnet"`
	//   - `"pcf"` (default)
	//   - `"plist"`
	//   - `"properties"`
	//   - `"textproto"`
	//   - `"xml"`
	//   - `"yaml"`
	OutputFormat string

	// AllowedModules defines URI patterns that determine which modules are permitted to be loaded and evaluated.
	// Patterns are regular expressions in the dialect understood by [java.util.regex.Pattern].
	//
	// [java.util.regex.Pattern]: https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/util/regex/Pattern.html
	AllowedModules []string

	// AllowedResources defines URI patterns that determine which resources are permitted to be loaded and evaluated.
	// Patterns are regular expressions in the dialect understood by [java.util.regex.Pattern].
	//
	// [java.util.regex.Pattern]: https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/util/regex/Pattern.html
	AllowedResources []string

	// ResourceReaders are the resource readers to be used by the evaluator.
	ResourceReaders []ResourceReader

	// ModuleReaders are the set of custom module readers to be used by the evaluator.
	ModuleReaders []ModuleReader

	// CacheDir is the directory where `package:` modules are cached.
	//
	// If empty, no cacheing is performed.
	CacheDir string

	// RootDir is the root directory for file-based reads within a Pkl program.
	//
	// Attempting to read past the root directory is an error.
	RootDir string

	// ProjectBaseURI sets the project base path for the evaluator.
	//
	// Setting this determines how Pkl resolves dependency notation imports.
	// It causes Pkl to look for the resolved dependencies relative to this base URI,
	// and load resolved dependencies from `PklProject.deps.json` within the base path represented.
	//
	// NOTE:
	// Setting this option is not equivalent to setting the `--project-dir` flag from the CLI.
	// When the `--project-dir` flag is set, the CLI will evaluate the PklProject file,
	// and then applies any evaluator settings and dependencies set in the PklProject file
	// for the main evaluation.
	//
	// In contrast, this option only determines how Pkl considers whether files are part of a
	// project.
	// It is meant to be set by lower level logic in Go that first evaluates the PklProject,
	// which then configures EvaluatorOptions accordingly.
	//
	// To emulate the CLI's `--project-dir` flag, create an evaluator with NewProjectEvaluator,
	// or EvaluatorManager.NewProjectEvaluator.
	ProjectBaseURI string

	// DeclaredProjectDepenedencies is set of dependencies available to modules within ProjectBaseURI.
	//
	// When importing dependencies, a PklProject.deps.json file must exist within ProjectBaseURI
	// that contains the project's resolved dependencies.
	DeclaredProjectDependencies *ProjectDependencies
}

type ProjectRemoteDependency struct {
	PackageUri string     `pkl:"uri"`
	Checksums  *Checksums `pkl:"checksums"`
}

func (dep *ProjectRemoteDependency) toMessage() *msgapi.ProjectOrDependency {
	return &msgapi.ProjectOrDependency{
		PackageUri: dep.PackageUri,
		Checksums:  dep.Checksums.toMessage(),
		Type:       "remote",
	}
}

type Checksums struct {
	Sha256 string `pkl:"sha256"`
}

func (checksums *Checksums) toMessage() *msgapi.Checksums {
	if checksums == nil {
		return nil
	}
	return &msgapi.Checksums{Sha256: checksums.Sha256}
}

type ProjectLocalDependency struct {
	PackageUri string

	ProjectFileUri string

	Dependencies *ProjectDependencies
}

func (dep *ProjectLocalDependency) toMessage() *msgapi.ProjectOrDependency {
	return &msgapi.ProjectOrDependency{
		PackageUri:     dep.PackageUri,
		ProjectFileUri: dep.ProjectFileUri,
		Type:           "local",
		Dependencies:   dep.Dependencies.toMessage(),
	}
}

type ProjectDependencies struct {
	LocalDependencies map[string]*ProjectLocalDependency

	RemoteDependencies map[string]*ProjectRemoteDependency
}

func (p *ProjectDependencies) toMessage() map[string]*msgapi.ProjectOrDependency {
	if p == nil {
		return nil
	}
	ret := make(map[string]*msgapi.ProjectOrDependency, len(p.LocalDependencies)+len(p.RemoteDependencies))
	for name, dep := range p.LocalDependencies {
		ret[name] = dep.toMessage()
	}
	for name, dep := range p.RemoteDependencies {
		ret[name] = dep.toMessage()
	}
	return ret
}

func (e *EvaluatorOptions) toMessage() *msgapi.CreateEvaluator {
	var resourceReaders []*msgapi.ResourceReader
	for _, reader := range e.ResourceReaders {
		resourceReaders = append(resourceReaders, &msgapi.ResourceReader{
			Scheme:              reader.Scheme(),
			IsGlobbable:         reader.IsGlobbable(),
			HasHierarchicalUris: reader.HasHierarchicalUris(),
		})
	}
	var moduleReaders []*msgapi.ModuleReader
	for _, reader := range e.ModuleReaders {
		moduleReaders = append(moduleReaders, &msgapi.ModuleReader{
			Scheme:              reader.Scheme(),
			IsGlobbable:         reader.IsGlobbable(),
			HasHierarchicalUris: reader.HasHierarchicalUris(),
			IsLocal:             reader.IsLocal(),
		})
	}
	return &msgapi.CreateEvaluator{
		ResourceReaders:  resourceReaders,
		ModuleReaders:    moduleReaders,
		Env:              e.Env,
		Properties:       e.Properties,
		ModulePaths:      e.ModulePaths,
		AllowedModules:   e.AllowedModules,
		AllowedResources: e.AllowedResources,
		CacheDir:         e.CacheDir,
		OutputFormat:     e.OutputFormat,
		RootDir:          e.RootDir,
		Project:          e.project(),
	}
}

func (e *EvaluatorOptions) project() *msgapi.ProjectOrDependency {
	if e.ProjectBaseURI == "" {
		return nil
	}
	return &msgapi.ProjectOrDependency{
		ProjectFileUri: e.ProjectBaseURI + "/PklProject",
		Dependencies:   e.DeclaredProjectDependencies.toMessage(),
	}
}

// WithOsEnv enables reading `env` values from the current environment.
var WithOsEnv = func(opts *EvaluatorOptions) {
	if opts.Env == nil {
		opts.Env = make(map[string]string)
	}
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			opts.Env[e[:i]] = e[i+1:]
		}
	}
}

func buildEvaluatorOptions(fns ...func(*EvaluatorOptions)) *EvaluatorOptions {
	o := &EvaluatorOptions{}
	for _, f := range fns {
		f(o)
	}
	// repl:text is the URI of the module used to hold expressions. It should always be allowed.
	o.AllowedModules = append(o.AllowedModules, "repl:text")
	return o
}

// WithDefaultAllowedResources enables reading http, https, file, env, prop, modulepath, and package resources.
var WithDefaultAllowedResources = func(opts *EvaluatorOptions) {
	opts.AllowedResources = append(opts.AllowedResources, "http:", "https:", "file:", "env:", "prop:", "modulepath:", "package:", "projectpackage:")
}

// WithDefaultAllowedModules enables reading stdlib, repl, file, http, https, modulepath, and package modules.
var WithDefaultAllowedModules = func(opts *EvaluatorOptions) {
	opts.AllowedModules = append(opts.AllowedModules, "pkl:", "repl:", "file:", "http:", "https:", "modulepath:", "package:", "projectpackage:")
}

// WithDefaultCacheDir sets the cache directory to Pkl's default location.
// It panics if the home directory cannot be determined.
var WithDefaultCacheDir = func(opts *EvaluatorOptions) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	opts.CacheDir = filepath.Join(dirname, ".pkl/cache")
}

// WithResourceReader sets up the given resource reader, and also adds the reader's scheme to the evaluator's
// allowed resources list.
var WithResourceReader = func(reader ResourceReader) func(opts *EvaluatorOptions) {
	return func(opts *EvaluatorOptions) {
		opts.ResourceReaders = append(opts.ResourceReaders, reader)
		opts.AllowedResources = append(opts.AllowedResources, reader.Scheme()+":")
	}
}

// WithModuleReader sets up the given module reader, and also adds the reader's scheme to the
// evaluator's allowed modules list.
var WithModuleReader = func(reader ModuleReader) func(opts *EvaluatorOptions) {
	return func(opts *EvaluatorOptions) {
		opts.ModuleReaders = append(opts.ModuleReaders, reader)
		opts.AllowedModules = append(opts.AllowedModules, reader.Scheme()+":")
	}
}

// WithFs sets up a ModuleReader and ResourceReader that associates the provided scheme with files
// from fs.
//
// For example, this may come from files within embed.FS.
//
// In Pkl terms, files within this file system are interpreted as based off the root path "/".
// For example, the path "foo.txt" within the provided file system is matched to path "/foo.txt"
// in Pkl code.
//
// If on Pkl 0.22 or lower, triple-dot imports and globbing are not supported.
//
// Modules and resources may be globbed within Pkl via `import*` and `read*`.
// Modules may be imported via triple-dot imports.
//
// Pkl has a built-in file system that reads from the host disk.
// This behavior may be overwritten by setting the scheme as `file`.
//
//goland:noinspection GoUnusedGlobalVariable
var WithFs = func(fs fs.FS, scheme string) func(opts *EvaluatorOptions) {
	return func(opts *EvaluatorOptions) {
		reader := &fsReader{fs: fs, scheme: scheme}
		WithModuleReader(&fsModuleReader{reader})(opts)
		WithResourceReader(&fsResourceReader{reader})(opts)
	}
}

// WithProjectEvaluatorSettings configures the evaluator with settings from the given
// ProjectEvaluatorSettings.
var WithProjectEvaluatorSettings = func(project *Project) func(opts *EvaluatorOptions) {
	return func(opts *EvaluatorOptions) {
		evaluatorSettings := project.EvaluatorSettings
		if evaluatorSettings == nil {
			return
		}
		opts.Properties = evaluatorSettings.ExternalProperties
		opts.Env = evaluatorSettings.Env
		opts.AllowedModules = evaluatorSettings.AllowedModules
		opts.AllowedResources = evaluatorSettings.AllowedResources
		if evaluatorSettings.NoCache != nil && *evaluatorSettings.NoCache {
			opts.CacheDir = ""
		} else {
			opts.CacheDir = evaluatorSettings.ModuleCacheDir
		}
		opts.RootDir = evaluatorSettings.RootDir
	}
}

// WithProjectDependencies configures the evaluator with dependencies from the specified project.
var WithProjectDependencies = func(project *Project) func(opts *EvaluatorOptions) {
	return func(opts *EvaluatorOptions) {
		opts.ProjectBaseURI = strings.TrimSuffix(project.ProjectFileUri, "/PklProject")
		opts.DeclaredProjectDependencies = project.Dependencies()
	}
}

var WithProject = func(project *Project) func(opts *EvaluatorOptions) {
	return func(opts *EvaluatorOptions) {
		WithProjectEvaluatorSettings(project)(opts)
		WithProjectDependencies(project)(opts)
	}
}

// PreconfiguredOptions configures an evaluator with:
//   - allowance for "file", "http", "https", "env", "prop", "package resource schemes
//   - allowance for "repl", "file", "http", "https", "pkl", "package" module schemes
//   - environment variables from the host environment
//   - ~/.pkl/cache as the cache directory
//   - no-op logging
//
// It panics if the home directory cannot be determined.
//
//goland:noinspection GoUnusedGlobalVariable
var PreconfiguredOptions = func(opts *EvaluatorOptions) {
	WithDefaultAllowedResources(opts)
	WithOsEnv(opts)
	WithDefaultAllowedModules(opts)
	WithDefaultCacheDir(opts)
	opts.Logger = NoopLogger
}

// MaybePreconfiguredOptions is like PreconfiguredOptions, except it only applies options
// if they have not already been set.
//
// It panics if the home directory cannot be determined.
var MaybePreconfiguredOptions = func(opts *EvaluatorOptions) {
	if len(opts.AllowedResources) == 0 {
		WithDefaultAllowedResources(opts)
	}
	if len(opts.Env) == 0 {
		WithOsEnv(opts)
	}
	if len(opts.AllowedModules) == 0 {
		WithDefaultAllowedModules(opts)
	}
	if opts.CacheDir == "" {
		WithDefaultCacheDir(opts)
	}
	if opts.Logger == nil {
		opts.Logger = NoopLogger
	}
}

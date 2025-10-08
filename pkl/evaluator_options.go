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
	"fmt"
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

	// Settings for controlling how Pkl talks over HTTP(S).
	//
	// Added in Pkl 0.26.
	// If the underlying Pkl does not support HTTP options, NewEvaluator will return with an error.
	Http *Http

	// ExternalModuleReaders registers external commands that implement module reader schemes.
	//
	// Added in Pkl 0.27.
	// If the underlying Pkl does not support external readers, evaluation will fail when a registered scheme is used.
	ExternalModuleReaders map[string]ExternalReader

	// ExternalResourceReaders registers external commands that implement resource reader schemes.
	//
	// Added in Pkl 0.27.
	// If the underlying Pkl does not support external readers, evaluation will fail when a registered scheme is used.
	ExternalResourceReaders map[string]ExternalReader

	// TraceMode dictates how Pkl will format messages logged by `trace()`.
	//
	// Added in Pkl 0.30.
	// If the underlying Pkl does not support trace modes, this option will be ignored.
	TraceMode TraceMode
}

type TraceMode string

const (
	// TraceCompact causes all structures passed to trace() will be emitted on a single line.
	TraceCompact TraceMode = "compact"
	// TracePretty causes all structures passed to trace() will be indented and emitted across multiple lines.
	TracePretty TraceMode = "pretty"
)

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

type Http struct {
	// PEM format certificates to trust when making HTTP requests.
	//
	// If empty, Pkl will trust its own built-in certificates.
	CaCertificates []byte

	// Configuration of the HTTP proxy to use.
	//
	// If nil, uses the operating system's proxy configuration.
	Proxy *Proxy

	// HTTP URI rewrite rules.
	//
	// Added in Pkl 0.29.
	// If the underlying Pkl does not support HTTP rewrites, NewEvaluator will return an error.
	//
	// Each key-value pair designates a source prefix to a target prefix.
	// Each rewrite rule must start with `http://` or `https://`, and end with `/`.
	//
	// This option is often used for setting up package mirroring.
	//
	// The following example will rewrite a request https://example.com/foo/bar to https://my.other.website/foo/bar:
	//
	//		Rewrites: map[string]string{
	//			"https://example.com/": "https://my.other.website/"
	//		}
	Rewrites map[string]string
}

func (http *Http) toMessage() *msgapi.Http {
	if http == nil {
		return nil
	}
	return &msgapi.Http{
		CaCertificates: http.CaCertificates,
		Proxy:          http.Proxy.toMessage(),
		Rewrites:       http.Rewrites,
	}
}

type Proxy struct {
	// The proxy to use for HTTP(S) connections.
	//
	// Only HTTP proxies are supported.
	// The address must start with "http://", and cannot contain anything other than a host and an optional port.
	//
	// Example:
	//
	//  	Address: "http://my.proxy.example.com:5080"
	Address string

	// Hosts to which all connections should bypass a proxy.
	//
	// Values can be either hostnames, or IP addresses.
	// IP addresses can optionally be provided using [CIDR notation].
	//
	// The only wildcard is `"*"`, which disables all proxying.
	//
	// A hostname matches all subdomains.
	// For example, `example.com` matches `foo.example.com`, but not `fooexample.com`.
	// A hostname that is prefixed with a dot matches the hostname itself,
	// so `.example.com` matches `example.com`.
	//
	// Optionally, a port can be specified.
	// If a port is omitted, all ports are matched.
	//
	// Example:
	//
	// 		NoProxy: []string{
	//			"127.0.0.1",
	//			"169.254.0.0/16",
	//			"example.com",
	//			"localhost:5050",
	//		}
	//
	// [CIDR notation]: https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation
	NoProxy []string
}

func (p *Proxy) toMessage() *msgapi.Proxy {
	if p == nil {
		return nil
	}
	return &msgapi.Proxy{
		Address: p.Address,
		NoProxy: p.NoProxy,
	}
}

type ExternalReader struct {
	Executable string
	Arguments  []string
}

func (r *ExternalReader) toMessage() *msgapi.ExternalReader {
	if r == nil {
		return nil
	}
	return &msgapi.ExternalReader{
		Executable: r.Executable,
		Arguments:  r.Arguments,
	}
}

func (e *EvaluatorOptions) toMessage() *msgapi.CreateEvaluator {
	return &msgapi.CreateEvaluator{
		ResourceReaders:         resourceReadersToMessage(e.ResourceReaders),
		ModuleReaders:           moduleReadersToMessage(e.ModuleReaders),
		Env:                     e.Env,
		Properties:              e.Properties,
		ModulePaths:             e.ModulePaths,
		AllowedModules:          e.AllowedModules,
		AllowedResources:        e.AllowedResources,
		CacheDir:                e.CacheDir,
		OutputFormat:            e.OutputFormat,
		RootDir:                 e.RootDir,
		Project:                 e.project(),
		Http:                    e.Http.toMessage(),
		ExternalModuleReaders:   externalReadersToMessage(e.ExternalModuleReaders),
		ExternalResourceReaders: externalReadersToMessage(e.ExternalResourceReaders),
		TraceMode:               string(e.TraceMode),
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

func buildEvaluatorOptions(version *semver, fns ...func(*EvaluatorOptions)) (*EvaluatorOptions, error) {
	o := &EvaluatorOptions{}
	for _, f := range fns {
		f(o)
	}
	// repl:text is the URI of the module used to hold expressions. It should always be allowed.
	o.AllowedModules = append(o.AllowedModules, "repl:text")
	if o.Http != nil && pklVersion0_26.isGreaterThan(version) {
		return nil, fmt.Errorf("http options are not supported on Pkl versions lower than 0.26")
	}
	if (len(o.ExternalModuleReaders) > 0 || len(o.ExternalResourceReaders) > 0) && pklVersion0_27.isGreaterThan(version) {
		return nil, fmt.Errorf("external reader options are not supported on Pkl versions lower than 0.27")
	}
	return o, nil
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
		opts.Properties = evaluatorSettings.ExternalProperties
		opts.Env = evaluatorSettings.Env
		if evaluatorSettings.AllowedModules != nil {
			opts.AllowedModules = *evaluatorSettings.AllowedModules
		}
		if evaluatorSettings.AllowedResources != nil {
			opts.AllowedResources = *evaluatorSettings.AllowedResources
		}
		if evaluatorSettings.NoCache != nil && *evaluatorSettings.NoCache {
			opts.CacheDir = ""
		} else {
			opts.CacheDir = evaluatorSettings.ModuleCacheDir
		}
		opts.RootDir = evaluatorSettings.RootDir
		if evaluatorSettings.Http != nil {
			opts.Http = &Http{}
			if evaluatorSettings.Http.Proxy != nil {
				opts.Http.Proxy = &Proxy{NoProxy: opts.Http.Proxy.NoProxy}
				if evaluatorSettings.Http.Proxy.Address != nil {
					opts.Http.Proxy.Address = *evaluatorSettings.Http.Proxy.Address
				}
			}
			if evaluatorSettings.Http.Rewrites != nil {
				opts.Http.Rewrites = *evaluatorSettings.Http.Rewrites
			}
		}
		if evaluatorSettings.ExternalModuleReaders != nil {
			opts.ExternalModuleReaders = make(map[string]ExternalReader, len(evaluatorSettings.ExternalModuleReaders))
			for scheme, reader := range evaluatorSettings.ExternalModuleReaders {
				opts.ExternalModuleReaders[scheme] = ExternalReader(reader)
				if evaluatorSettings.AllowedModules == nil { // if no explicit allowed modules are set in the project, allow declared external module readers
					WithDefaultAllowedModules(opts)
					opts.AllowedModules = append(opts.AllowedModules, scheme+":")
				}
			}
		}
		if evaluatorSettings.ExternalResourceReaders != nil {
			opts.ExternalResourceReaders = make(map[string]ExternalReader, len(evaluatorSettings.ExternalResourceReaders))
			for scheme, reader := range evaluatorSettings.ExternalResourceReaders {
				opts.ExternalResourceReaders[scheme] = ExternalReader(reader)
				if evaluatorSettings.AllowedResources == nil { // if no explicit allowed resources are set in the project, allow declared external resource readers
					WithDefaultAllowedResources(opts)
					opts.AllowedResources = append(opts.AllowedResources, scheme+":")
				}
			}
		}
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

var WithExternalModuleReader = func(scheme string, spec ExternalReader) func(opts *EvaluatorOptions) {
	return func(opts *EvaluatorOptions) {
		if opts.ExternalModuleReaders == nil {
			opts.ExternalModuleReaders = map[string]ExternalReader{}
		}
		opts.ExternalModuleReaders[scheme] = spec
		opts.AllowedModules = append(opts.AllowedModules, scheme+":")
	}
}

var WithExternalResourceReader = func(scheme string, spec ExternalReader) func(opts *EvaluatorOptions) {
	return func(opts *EvaluatorOptions) {
		if opts.ExternalResourceReaders == nil {
			opts.ExternalResourceReaders = map[string]ExternalReader{}
		}
		opts.ExternalResourceReaders[scheme] = spec
		opts.AllowedResources = append(opts.AllowedResources, scheme+":")
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

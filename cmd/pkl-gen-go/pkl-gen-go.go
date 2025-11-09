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

package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/apple/pkl-go/cmd/pkl-gen-go/generatorsettings"
	"github.com/apple/pkl-go/cmd/pkl-gen-go/pkg"
	"github.com/apple/pkl-go/pkl"
	"github.com/spf13/cobra"
)

var (
	// Version of pkl-gen-go.
	//
	// This gets replaced by ldflags when built through CI,
	// or by init when installed via go install.
	Version = "development"

	generatorSettingsPath string
	generateScript        string
	mappings              map[string]string
	basePath              string
	allowedModules        []string
	allowedResources      []string
	dryRun                bool
	projectDir            string
	cacheDir              string
	suppressWarnings      bool
	outputPath            string
	printVersion          bool

	settings generatorsettings.GeneratorSettings
	err      error
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" || Version != "development" {
		return
	}
	Version = strings.TrimPrefix(info.Main.Version, "v")
}

var command = cobra.Command{
	Use:   "pkl-gen-go [flags] <module>",
	Short: "Generates Go bindings for a Pkl module",
	Long: `Generates Go bindings for a Pkl module.

PACKAGE MAPPINGS
	To generate Go, all Pkl modules must have a known Go package name. The package name
	may come from one of three sources:

	  1. The @go.Package annotation on a module
	  2. A generator settings Pkl file
	  3. A --mapping argument

GENERATOR SETTINGS FILE
	Code generation may be configured using a settings file. By default, pkl-gen-go will look 
	for file called "generator-settings.pkl" in the current working directory, and the path can
	be configured using the --generator-settings flag.

	The generator settings file should amend module 
	package://pkg.pkl-lang.org/pkl-go/pkl.golang@<VERSION>#/GeneratorSettings.pkl

CONFIGURING OUTPUT PATH
	By default, the full path of each module is written as a relative path to the current working
	directory. This behavior changes by setting a base path either as a CLI flag, or in the
	generator settings file.

	When using a base path, any package that does not belong to the path will be skipped from
	code generation.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if printVersion {
			fmt.Println(Version)
			return nil
		}
		settings, err = loadGeneratorSettings()
		if err != nil {
			return fmt.Errorf("failed to load generator settings: %w", err)
		}
		evaluator, err := newEvaluator()
		if err != nil {
			return fmt.Errorf("failed to create evaluator: %w", err)
		}
		if outputPath == "" {
			outputPath, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		if err = pkg.GenerateGo(evaluator, args[0], settings, suppressWarnings, outputPath); err != nil {
			_, _ = fmt.Fprint(os.Stderr, err.Error())
			os.Exit(1)
		}
		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if printVersion {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
}

func init() {
	flags := command.Flags()
	flags.StringVar(&generatorSettingsPath, "generator-settings", "", "The path to a generator settings file")
	flags.StringVar(&generateScript, "generate-script", "", "The Generate.pkl script to use")
	flags.StringToStringVar(&mappings, "mapping", nil, "The mapping of a Pkl module name to a Go package name")
	flags.StringVar(&basePath, "base-path", "", "The base path used to determine relative output")
	flags.StringVar(&outputPath, "output-path", "", "The output directory to write generated sources into")
	flags.BoolVar(&suppressWarnings, "suppress-format-warning", false, "Suppress warnings around formatting issues")
	flags.StringSliceVar(&allowedModules, "allowed-modules", nil, "URI patterns that determine which modules can be loaded and evaluated")
	flags.StringSliceVar(&allowedResources, "allowed-resources", nil, "URI patterns that determine which resources can be loaded and evaluated")
	flags.StringVar(&projectDir, "project-dir", "", "The project directory to load dependency and evaluator settings from")
	flags.StringVar(&cacheDir, "cache-dir", "", "The cache directory for storing packages")
	flags.BoolVar(&dryRun, "dry-run", false, "Print out the names of the files that will be generated, but don't write any files")
	flags.BoolVar(&printVersion, "version", false, "Print the version and exit")
}

func newEvaluator() (pkl.Evaluator, error) {
	projectDirFlag := ""
	if settings.ProjectDir != nil {
		projectDirFlag = *settings.ProjectDir
	}
	projectDir := findProjectDir(projectDirFlag)
	if projectDir == nil {
		return pkl.NewEvaluator(context.Background(), evaluatorOptions)
	}
	return pkl.NewProjectEvaluator(context.Background(), projectDir, evaluatorOptions)
}

func evaluatorOptions(opts *pkl.EvaluatorOptions) {
	pkl.MaybePreconfiguredOptions(opts)
	opts.Logger = pkl.StderrLogger
	if len(settings.AllowedModules) > 0 {
		opts.AllowedModules = settings.AllowedModules
	}
	if len(settings.AllowedResources) > 0 {
		opts.AllowedResources = settings.AllowedResources
	}
	if settings.CacheDir != nil {
		opts.CacheDir = *settings.CacheDir
	}
	if cacerts, err := cacertsFromHomeDir(); len(cacerts) > 0 && err == nil {
		if opts.Http == nil {
			opts.Http = &pkl.Http{}
		}
		opts.Http.CaCertificates = cacerts
	} else if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "[warn] Failed to load cacerts: "+err.Error())
	}
}

// load certs from ~/.pkl/cacerts if exists.
func cacertsFromHomeDir() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cacertsDir := filepath.Join(home, ".pkl", "cacerts")
	stat, err := os.Stat(cacertsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if stat.IsDir() {
		var ret []byte
		files, err := os.ReadDir(cacertsDir)
		if err != nil {
			return nil, err
		}
		for i := range files {
			bytes, err := os.ReadFile(filepath.Join(cacertsDir, files[i].Name()))
			if err != nil {
				return nil, err
			}
			ret = append(ret, bytes...)
		}
		return ret, nil
	}
	return nil, nil
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	} else if err != nil {
		panic(err)
	}
	return true
}

//goland:noinspection GoBoolExpressions
func generatorSettingsSource() *pkl.ModuleSource {
	if Version == "development" {
		_, filename, _, ok := runtime.Caller(1)
		if !ok {
			panic("Failed to get path to pkl-gen-go.go")
		}
		dirPath := filepath.Dir(filename)
		return pkl.FileSource(dirPath, "../../codegen/src/GeneratorSettings.pkl")
	}
	return pkl.UriSource(fmt.Sprintf("package://pkg.pkl-lang.org/pkl-go/pkl.golang@%s#/GeneratorSettings.pkl", Version))
}

// mimick logic for finding project dir in the pkl CLI.
func doFindProjectDir(dir string) string {
	if fileExists(filepath.Join(dir, "PklProject")) {
		return dir
	}
	parent := filepath.Dir(dir)
	if parent == dir {
		return ""
	}
	return doFindProjectDir(parent)
}

func findProjectDir(projectDirFlag string) *url.URL {
	if projectDirFlag != "" {
		normalized, err := filepath.Abs(projectDirFlag)
		if err != nil {
			panic(err)
		}
		return &url.URL{Scheme: "file", Path: normalized}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	projectDir := doFindProjectDir(cwd)
	if projectDir == "" {
		return nil
	}
	return &url.URL{Scheme: "file", Path: projectDir}
}

// Loads the settings for controlling codegen.
// Uses a Pkl evaluator that is separate from what's used for actually running codegen.
func loadGeneratorSettings() (generatorsettings.GeneratorSettings, error) {
	resolvedProjectDir := findProjectDir(projectDir)
	var evaluator pkl.Evaluator
	var err error
	opts := func(opts *pkl.EvaluatorOptions) {
		if cacheDir != "" {
			opts.CacheDir = cacheDir
		}
	}
	if resolvedProjectDir != nil {
		evaluator, err = pkl.NewProjectEvaluator(context.Background(), resolvedProjectDir, evaluatorOptions, opts)
	} else {
		evaluator, err = pkl.NewEvaluator(context.Background(), evaluatorOptions, opts)
	}
	if err != nil {
		panic(err)
	}
	var source *pkl.ModuleSource
	if generatorSettingsPath != "" {
		source = pkl.FileSource(generatorSettingsPath)
	} else if fileExists("generator-settings.pkl") {
		source = pkl.FileSource("generator-settings.pkl")
	} else {
		source = generatorSettingsSource()
	}
	s, err := generatorsettings.Load(context.Background(), evaluator, source)
	if err != nil {
		return s, err
	}
	settingsFilePath := path.Dir(source.Uri.Path)
	if s.ProjectDir != nil && !path.IsAbs(*s.ProjectDir) {
		normalized := path.Join(settingsFilePath, *s.ProjectDir)
		s.ProjectDir = &normalized
	}
	if s.CacheDir != nil && !path.IsAbs(*s.CacheDir) {
		normalized := path.Join(settingsFilePath, *s.CacheDir)
		s.CacheDir = &normalized
	}

	// load overrides from CLI flags
	if generateScript != "" {
		settings.GeneratorScriptPath = generateScript
	}
	if len(mappings) != 0 {
		settings.PackageMappings = mappings
	}
	if basePath != "" {
		settings.BasePath = basePath
	}
	if len(allowedModules) > 0 {
		settings.AllowedModules = allowedModules
	}
	if len(allowedResources) > 0 {
		settings.AllowedResources = allowedResources
	}
	if projectDir != "" {
		normalized, err := filepath.Abs(projectDir)
		if err != nil {
			return s, err
		}
		settings.ProjectDir = &normalized
	}
	if cacheDir != "" {
		normalized, err := filepath.Abs(cacheDir)
		if err != nil {
			return s, err
		}
		settings.CacheDir = &normalized
	}
	settings.DryRun = dryRun

	return s, nil
}

func main() {
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
